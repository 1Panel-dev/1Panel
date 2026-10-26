package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/mcpruntime"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var mcpRuntimeMutation sync.Mutex
var ErrMCPSettingsRevision = errors.New("MCP settings changed; reload the page before saving")

type mcpRuntimeManager interface {
	Prepare(context.Context) error
	Catalog(context.Context) ([]contract.Tool, error)
	Status(context.Context, bool) mcpruntime.State
	Apply(context.Context, bool, bool) error
	InstallImage(context.Context, bool) error
	Restore(context.Context, string, bool) error
	RestoreAbsent(context.Context) error
	Logs(context.Context) (string, error)
}

type mcpRuntimeSnapshot struct {
	settings    dto.MCPRuntimeSettings
	imageID     string
	installed   bool
	unavailable bool
}

func (s *MCPRuntimeService) snapshot(ctx context.Context, settings dto.MCPRuntimeSettings) (mcpRuntimeSnapshot, error) {
	snapshot := mcpRuntimeSnapshot{settings: settings}
	state := s.manager(settings).Status(ctx, settings.Enabled)
	switch state.Status {
	case "not_installed", "stopped", "running", "unhealthy":
	default:
		snapshot.unavailable = true
		return snapshot, errors.New("cannot determine the existing MCP Docker instance; retry before changing it")
	}
	snapshot.installed = state.Installed
	if state.Installed {
		if state.AdapterVersion == "" || mcpruntime.ValidateDockerImage(mcpruntime.DockerImageRepository+"@"+state.ImageID) != nil {
			snapshot.unavailable = true
			return snapshot, errors.New("the existing MCP Docker instance has invalid metadata or image identity")
		}
		snapshot.imageID = state.ImageID
	} else if state.Status != "not_installed" {
		snapshot.unavailable = true
		return snapshot, errors.New("cannot determine the existing MCP Docker installation")
	}
	return snapshot, nil
}

type MCPRuntimeService struct {
	newRuntimeManager func(dto.MCPRuntimeSettings) mcpRuntimeManager
}

func (s *MCPRuntimeService) manager(settings dto.MCPRuntimeSettings) mcpRuntimeManager {
	if s.newRuntimeManager != nil {
		return s.newRuntimeManager(settings)
	}
	return mcpruntime.NewDocker(settings.ListenPort, settings.DockerImage)
}

func NewMCPRuntimeService() *MCPRuntimeService { return &MCPRuntimeService{} }

func (s *MCPRuntimeService) PrepareOnStartup(ctx context.Context) error {
	mcpRuntimeMutation.Lock()
	defer mcpRuntimeMutation.Unlock()
	settings, err := s.Settings()
	if err != nil {
		return err
	}
	return s.manager(settings).Prepare(ctx)
}

func (s *MCPRuntimeService) Settings() (dto.MCPRuntimeSettings, error) {
	return s.settings(context.Background())
}

func (s *MCPRuntimeService) settings(ctx context.Context) (dto.MCPRuntimeSettings, error) {
	settings := dto.MCPRuntimeSettings{
		RuntimeType: "docker", DockerImage: mcpruntime.DefaultDockerImage,
		ListenPort: mcpruntime.DefaultPort, Revision: 1,
		AccessPolicy: "inherit", AllowedIPs: []string{},
	}
	var saved model.MCPSetting
	err := global.DB.WithContext(ctx).First(&saved, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return settings, nil
	}
	if err != nil {
		return settings, err
	}
	entries, err := contract.DecodeAllowedIPs(saved.AllowedIPs)
	if err != nil {
		return settings, err
	}
	settings = dto.MCPRuntimeSettings{RuntimeType: saved.RuntimeType, DockerImage: saved.DockerImage, Enabled: saved.Enabled, ListenPort: saved.ListenPort, ExternalURL: saved.ExternalURL, Revision: saved.Revision, AccessPolicy: saved.AccessPolicy, AllowedIPs: entries}
	return settings, validateMCPRuntimeSettings(settings)
}

func (s *MCPRuntimeService) Catalog(ctx context.Context) ([]contract.Tool, error) {
	settings, err := s.settings(ctx)
	if err != nil {
		return nil, err
	}
	return s.manager(settings).Catalog(ctx)
}

func (s *MCPRuntimeService) Update(req dto.MCPRuntimeSettingsUpdate) (dto.MCPRuntimeSettings, error) {
	mcpRuntimeMutation.Lock()
	defer mcpRuntimeMutation.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	return s.update(ctx, req, false)
}

func (s *MCPRuntimeService) update(ctx context.Context, req dto.MCPRuntimeSettingsUpdate, restart bool) (dto.MCPRuntimeSettings, error) {
	old, err := s.Settings()
	if err != nil {
		return old, err
	}
	if req.Revision != old.Revision {
		return old, ErrMCPSettingsRevision
	}
	if err := validateMCPRuntimeSettings(req); err != nil {
		return old, contract.Fail("MCP_INVALID_ARGUMENTS", 400, err.Error())
	}
	if err := normalizeMCPAccessSettings(&req); err != nil {
		return old, contract.Fail("MCP_INVALID_ARGUMENTS", 400, err.Error())
	}
	snapshot, snapshotErr := s.snapshot(ctx, old)
	if snapshotErr != nil && req.Enabled {
		return old, snapshotErr
	}
	manager := s.manager(req)
	recoverFailure := func(operationErr error) (dto.MCPRuntimeSettings, error) {
		current, loadErr := s.Settings()
		if loadErr != nil {
			return old, errors.Join(operationErr, loadErr, errors.New("MCP settings are unavailable; runtime recovery requires retrying after the database is available"))
		}
		return current, errors.Join(operationErr, s.restore(current, snapshot))
	}
	if err := manager.Apply(ctx, req.Enabled, restart || (!old.Enabled && req.Enabled) || (old.Enabled && old.ListenPort != req.ListenPort)); err != nil {
		return recoverFailure(err)
	}
	if err := s.persist(req, old.Revision); err != nil {
		return recoverFailure(err)
	}
	req.Revision++
	return req, nil
}

func (s *MCPRuntimeService) persist(req dto.MCPRuntimeSettings, revision uint64) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		seed := model.MCPSetting{ID: 1, RuntimeType: "docker", DockerImage: mcpruntime.DefaultDockerImage, ListenPort: mcpruntime.DefaultPort, Revision: 1}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&seed).Error; err != nil {
			return err
		}
		result := tx.Model(&model.MCPSetting{}).Where("id = ? AND revision = ?", 1, revision).Updates(map[string]interface{}{
			"runtime_type": "docker", "docker_image": req.DockerImage,
			"enabled": req.Enabled, "listen_port": req.ListenPort, "external_url": req.ExternalURL, "revision": revision + 1,
			"access_policy": req.AccessPolicy, "allowed_ips": encodeMCPIPs(req.AllowedIPs),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrMCPSettingsRevision
		}
		return nil
	})
}

func (s *MCPRuntimeService) restore(settings dto.MCPRuntimeSettings, snapshot mcpRuntimeSnapshot) error {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	manager := s.manager(settings)
	old := snapshot.settings
	if sameMCPRuntimeConfiguration(settings, old) {
		if snapshot.unavailable {
			return errors.New("the previous MCP Docker identity was unavailable; the runtime could not be restored automatically")
		}
		if snapshot.imageID != "" {
			return manager.Restore(ctx, snapshot.imageID, settings.Enabled)
		}
	}
	return manager.Apply(ctx, settings.Enabled, true)
}

func sameMCPRuntimeConfiguration(a, b dto.MCPRuntimeSettings) bool {
	return a.DockerImage == b.DockerImage && a.ListenPort == b.ListenPort && a.Enabled == b.Enabled
}

func (s *MCPRuntimeService) Status(ctx context.Context) (dto.MCPRuntimeStatus, error) {
	settings, err := s.Settings()
	if err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	state := s.manager(settings).Status(ctx, settings.Enabled)

	return dto.MCPRuntimeStatus{
		RuntimeType: "docker", DockerImage: settings.DockerImage, ContainerName: mcpruntime.DockerContainerName, ImageID: state.ImageID,
		Enabled: settings.Enabled, Installed: state.Installed, Status: state.Status, Ready: state.Ready, Reason: state.Reason,
		AdapterVersion: state.AdapterVersion, ContractVersion: mcpruntime.ContractVersion, Protocol: mcpruntime.Protocol, Endpoint: settings.ExternalURL,
		InternalEndpoint: fmt.Sprintf("http://127.0.0.1:%d/mcp", settings.ListenPort),
	}, nil
}

func (s *MCPRuntimeService) Operate(ctx context.Context, operation string) (dto.MCPRuntimeStatus, error) {
	mcpRuntimeMutation.Lock()
	defer mcpRuntimeMutation.Unlock()
	settings, err := s.Settings()
	if err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	switch operation {
	case "start", "restart":
		settings.Enabled = true
	case "stop":
		settings.Enabled = false
	default:
		return dto.MCPRuntimeStatus{}, contract.Invalid()
	}
	if _, err := s.update(ctx, settings, operation == "restart"); err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	return s.Status(ctx)
}

func (s *MCPRuntimeService) Install(ctx context.Context, req dto.MCPRuntimeInstall) (dto.MCPRuntimeStatus, error) {
	mcpRuntimeMutation.Lock()
	defer mcpRuntimeMutation.Unlock()
	old, err := s.settings(ctx)
	if err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	if req.Revision != old.Revision {
		return dto.MCPRuntimeStatus{}, ErrMCPSettingsRevision
	}
	next := old
	next.RuntimeType, next.DockerImage, next.ListenPort, next.ExternalURL = "docker", req.DockerImage, req.ListenPort, req.ExternalURL
	if err := validateMCPRuntimeSettings(next); err != nil {
		return dto.MCPRuntimeStatus{}, contract.Fail("MCP_INVALID_ARGUMENTS", 400, err.Error())
	}
	snapshot, err := s.snapshot(ctx, old)
	if err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	if !snapshot.installed {
		next.Enabled = true
	}
	manager := s.manager(next)
	rollback := func(authoritative dto.MCPRuntimeSettings) error {
		if !snapshot.installed && reflect.DeepEqual(authoritative, old) {
			restoreCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			return manager.RestoreAbsent(restoreCtx)
		}
		return s.restore(authoritative, snapshot)
	}
	recoverFailure := func(operationErr error) error {
		current, loadErr := s.Settings()
		if loadErr != nil {
			return errors.Join(operationErr, loadErr, errors.New("MCP settings are unavailable; runtime recovery requires retrying after the database is available"))
		}
		return errors.Join(operationErr, rollback(current))
	}
	if err := manager.InstallImage(ctx, next.Enabled); err != nil {
		return dto.MCPRuntimeStatus{}, recoverFailure(err)
	}
	if err := s.persist(next, old.Revision); err != nil {
		return dto.MCPRuntimeStatus{}, recoverFailure(err)
	}
	return s.Status(ctx)
}

func (s *MCPRuntimeService) Logs(ctx context.Context) (dto.MCPRuntimeLogs, error) {
	settings, err := s.Settings()
	if err != nil {
		return dto.MCPRuntimeLogs{}, err
	}
	content, err := s.manager(settings).Logs(ctx)
	return dto.MCPRuntimeLogs{Content: content}, err
}

func validateMCPRuntimeSettings(req dto.MCPRuntimeSettings) error {
	if req.AccessPolicy == "" {
		return errors.New("MCP source policy is required")
	}
	if err := normalizeMCPAccessSettings(&req); err != nil {
		return err
	}
	if req.RuntimeType != "docker" {
		return errors.New("1Panel MCP supports only the Docker runtime")
	}
	if err := mcpruntime.ValidateDockerImage(req.DockerImage); err != nil {
		return err
	}
	if req.ListenPort < 1024 || req.ListenPort > 65535 {
		return errors.New("MCP listen port must be between 1024 and 65535")
	}
	if req.ExternalURL == "" {
		return nil
	}
	if len(req.ExternalURL) > 2048 || strings.ContainsAny(req.ExternalURL, "\r\n") {
		return errors.New("invalid MCP external URL")
	}
	u, err := url.Parse(req.ExternalURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "/mcp" {
		return errors.New("MCP external URL must be an HTTP or HTTPS URL ending in /mcp without credentials or query parameters")
	}
	return nil
}
