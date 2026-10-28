package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/task"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/mcpruntime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *MCPRuntimeService) SubmitInstall(ctx context.Context, req dto.MCPRuntimeInstall) (dto.MCPRuntimeInstallTask, error) {
	if !mcpInstalling.CompareAndSwap(false, true) {
		return dto.MCPRuntimeInstallTask{}, contract.Fail("MCP_CONFLICT", 409, "An MCP installation is already running")
	}
	if !mcpRuntimeMutation.TryLock() {
		mcpInstalling.Store(false)
		return dto.MCPRuntimeInstallTask{}, contract.Fail("MCP_CONFLICT", 409, "An MCP service operation is already running")
	}
	accepted := false
	defer func() {
		if !accepted {
			mcpInstalling.Store(false)
			mcpRuntimeMutation.Unlock()
		}
	}()
	if err := ctx.Err(); err != nil {
		return dto.MCPRuntimeInstallTask{}, err
	}
	if _, _, err := s.installSettings(ctx, req); err != nil {
		return dto.MCPRuntimeInstallTask{}, err
	}
	req.AllowedIPs = append([]string{}, req.AllowedIPs...)
	if req.Website != nil {
		website := *req.Website
		domain, err := normalizeMCPWebsiteDomain(website.Domain)
		if err != nil {
			return dto.MCPRuntimeInstallTask{}, contract.Invalid()
		}
		parsed, _ := url.Parse("http://" + domain)
		if net.ParseIP(parsed.Hostname()) != nil || website.EnableSSL != (website.SSLID != 0) {
			return dto.MCPRuntimeInstallTask{}, contract.Invalid()
		}
		row, err := s.websiteSetting(ctx)
		if err != nil {
			return dto.MCPRuntimeInstallTask{}, err
		}
		if row.WebsiteID != 0 || row.WebsiteBindingID != "" || row.WebsiteOperation != "" {
			return dto.MCPRuntimeInstallTask{}, contract.Fail("MCP_CONFLICT", 409, "Review the existing MCP website configuration before installation")
		}
		website.Domain = domain
		req.Website = &website
	}
	installation, err := task.NewTaskWithOps("1Panel MCP", task.TaskInstall, task.TaskScopePanelMCP, "", 0)
	if err != nil {
		return dto.MCPRuntimeInstallTask{}, err
	}
	if err := repo.NewITaskRepo().Save(ctx, installation.Task); err != nil {
		installation.Close()
		installation.DeleteLogFile()
		return dto.MCPRuntimeInstallTask{}, err
	}
	installation.AddSubTaskWithContext(i18n.GetMsgByKey("MCPInstallPrepare"), func(ctx context.Context, t *task.Task) (resultErr error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				resultErr = errors.New(i18n.GetMsgByKey("MCPInstallFailed"))
			}
		}()
		return s.executeInstallTask(ctx, req, t)
	}, nil, 0, 8*time.Minute)
	accepted = true
	go func() {
		defer mcpRuntimeMutation.Unlock()
		defer mcpInstalling.Store(false)
		defer installation.Close()
		if err := installation.Execute(); err != nil {
			if installation.Task.Status == constant.StatusExecuting {
				installation.Task.Status = constant.StatusFailed
				installation.Task.ErrorMsg = i18n.GetMsgByKey("MCPInstallFailed")
				installation.Task.EndAt = time.Now()
			}
			if global.LOG != nil {
				global.LOG.Errorf("MCP installation task %s failed: %v", installation.TaskID, err)
			}
		}
		if err := repo.NewITaskRepo().Update(context.Background(), installation.Task); err != nil && global.LOG != nil {
			global.LOG.Errorf("MCP installation task %s status could not be saved: %v", installation.TaskID, err)
		}
	}()
	return dto.MCPRuntimeInstallTask{TaskID: installation.TaskID}, nil
}

func (s *MCPRuntimeService) executeInstallTask(ctx context.Context, req dto.MCPRuntimeInstall, t *task.Task) error {
	report := mcpInstallProgressLogger(t)
	ctx = mcpruntime.WithProgress(ctx, report)
	installCtx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	_, err := s.install(installCtx, req)
	cancel()
	if err != nil {
		return safeMCPInstallTaskError(err)
	}
	if req.Website == nil {
		return nil
	}
	t.Log(i18n.GetMsgByKey("MCPInstallBindWebsite"))
	websiteCtx, websiteCancel := context.WithTimeout(ctx, 90*time.Second)
	defer websiteCancel()
	settings, err := s.settings(websiteCtx)
	if err == nil {
		_, err = s.mutateWebsiteLocked(websiteCtx, dto.MCPWebsitePendingRequest{
			Operation: "bind", RequestID: uuid.NewString(), Domain: req.Website.Domain,
			EnableSSL: req.Website.EnableSSL, SSLID: req.Website.SSLID,
		}, settings.Revision)
	}
	if err != nil {
		t.Log(i18n.GetMsgByKey("MCPInstallWebsiteFailed"))
		return errors.New(i18n.GetMsgByKey("MCPInstallWebsiteFailed"))
	}
	return nil
}

func safeMCPInstallTaskError(err error) error {
	var recovery *mcpRuntimeRecoveryError
	if errors.As(err, &recovery) {
		return errors.New(i18n.GetMsgByKey("MCPInstallRecoveryFailed"))
	}
	var operation *mcpruntime.OperationError
	if errors.As(err, &operation) {
		return operation
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return errors.New(i18n.GetMsgByKey("MCPInstallFailed"))
}

func mcpInstallProgressLogger(t *task.Task) func(mcpruntime.Progress) {
	var mu sync.Mutex
	last := make(map[string]time.Time)
	return func(progress mcpruntime.Progress) {
		mu.Lock()
		defer mu.Unlock()
		key := progress.Status + ":" + progress.ID
		now := time.Now()
		if (progress.Status == "downloading" || progress.Status == "extracting") &&
			(progress.Total == 0 || progress.Current != progress.Total) && now.Sub(last[key]) < time.Second {
			return
		}
		last[key] = now
		stage := map[string]string{
			"snapshot": "MCPInstallPrepare", "checking_image": "MCPInstallPullImage",
			"local_image": "MCPInstallLocalImage", "verifying_image": "MCPInstallVerifyImage", "verified_image": "MCPInstallVerifiedImage",
			"catalog": "MCPInstallCatalog", "starting": "MCPInstallStarting",
			"saving": "MCPInstallSaving", "installed": "MCPInstallCompleted", "rollback": "MCPInstallRollback",
			"recovery_failed": "MCPInstallRecoveryFailed",
		}[progress.Status]
		if stage != "" {
			t.Log(i18n.GetMsgByKey(stage))
			return
		}
		message := map[string]string{
			"pulling": "Pulling fs layer", "waiting": "Waiting", "downloading": "Downloading",
			"extracting": "Extracting", "verifying_checksum": "Verifying Checksum",
			"download_complete": "Download complete", "pull_complete": "Pull complete", "already_exists": "Already exists",
		}[progress.Status]
		if message == "" {
			return
		}
		if progress.ID != "" {
			message = progress.ID + ": " + message
		}
		if progress.Total > 0 {
			message += fmt.Sprintf(" %d/%d bytes (%.0f%%)", progress.Current, progress.Total, 100*float64(progress.Current)/float64(progress.Total))
		} else if progress.Current > 0 {
			message += fmt.Sprintf(" %d bytes", progress.Current)
		}
		t.Log(message)
	}
}

func (s *MCPRuntimeService) InstallationTask(ctx context.Context) (string, bool, error) {
	var latest model.Task
	err := global.TaskDB.WithContext(ctx).Where("type = ?", task.TaskScopePanelMCP).Order("created_at DESC").First(&latest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	return latest.ID, latest.Status == constant.StatusExecuting, err
}

func (s *MCPRuntimeService) RecoverInstallationTasks(processStartedAt time.Time) error {
	return global.TaskDB.Model(&model.Task{}).Where("type = ? AND status = ? AND created_at < ?", task.TaskScopePanelMCP, constant.StatusExecuting, processStartedAt).
		Updates(map[string]any{"status": constant.StatusFailed, "error_msg": i18n.GetMsgByKey("MCPInstallInterrupted"), "end_at": time.Now()}).Error
}
