package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/global"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/mcpruntime"
	"gorm.io/gorm"
)

var mcpRuntimeMutation sync.Mutex
var mcpInstalling atomic.Bool
var ErrMCPSettingsRevision = errors.New("MCP settings changed; reload the page before saving")

func lockMCPRuntimeMutation() error {
	conflict := func() error { return contract.Fail("MCP_CONFLICT", 409, "An MCP service operation is already running") }
	if mcpInstalling.Load() {
		return conflict()
	}
	mcpRuntimeMutation.Lock()
	if mcpInstalling.Load() {
		mcpRuntimeMutation.Unlock()
		return conflict()
	}
	return nil
}

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

type mcpRuntimeRecoveryError struct{ cause error }

func (e *mcpRuntimeRecoveryError) Error() string {
	return "MCP runtime recovery is incomplete; check the service state before retrying"
}

func (e *mcpRuntimeRecoveryError) Unwrap() error { return e.cause }

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
		if state.AdapterVersion == "" || !mcpruntime.ValidDockerImageID(state.ImageID) {
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
	websiteTransport  mcpWebsiteTransport
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
	saved, err := (repo.MCPRepo{DB: global.DB.WithContext(ctx)}).Setting()
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
	if err := lockMCPRuntimeMutation(); err != nil {
		return dto.MCPRuntimeSettings{}, err
	}
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
	if req.ExternalURL != old.ExternalURL {
		if row, loadErr := s.websiteSetting(ctx); loadErr != nil {
			return old, loadErr
		} else if row.WebsiteID != 0 || row.WebsiteBindingID != "" {
			return old, contract.Fail("MCP_CONFLICT", 409, "The website controls the MCP endpoint")
		}
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
		store := repo.MCPRepo{DB: tx}
		seed := model.MCPSetting{ID: 1, RuntimeType: "docker", DockerImage: mcpruntime.DefaultDockerImage, ListenPort: mcpruntime.DefaultPort, Revision: 1}
		if err := store.CreateSettingIfAbsent(&seed); err != nil {
			return err
		}
		rowsAffected, err := store.UpdateRuntimeSetting(model.MCPSetting{
			RuntimeType: "docker", DockerImage: req.DockerImage,
			Enabled: req.Enabled, ListenPort: req.ListenPort, ExternalURL: req.ExternalURL, Revision: revision + 1,
			AccessPolicy: req.AccessPolicy, AllowedIPs: encodeMCPIPs(req.AllowedIPs),
		}, revision)
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
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

	result := dto.MCPRuntimeStatus{
		RuntimeType: "docker", DockerImage: settings.DockerImage, ContainerName: mcpruntime.DockerContainerName, ImageID: state.ImageID,
		Enabled: settings.Enabled, Installed: state.Installed, Status: state.Status, Ready: state.Ready, Reason: state.Reason,
		AdapterVersion: state.AdapterVersion, ContractVersion: mcpruntime.ContractVersion, Protocol: mcpruntime.Protocol, Endpoint: settings.ExternalURL,
		InternalEndpoint: fmt.Sprintf("http://127.0.0.1:%d/mcp", settings.ListenPort),
	}
	website, websiteErr := s.WebsiteStatus(ctx)
	if websiteErr != nil {
		return result, websiteErr
	}
	if website.State != "unbound" {
		result.Website = &website
		result.Endpoint = website.Endpoint
	}
	return result, nil
}

func (s *MCPRuntimeService) Operate(ctx context.Context, operation string) (dto.MCPRuntimeStatus, error) {
	if err := lockMCPRuntimeMutation(); err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
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
	if err := lockMCPRuntimeMutation(); err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	defer mcpRuntimeMutation.Unlock()
	return s.install(ctx, req)
}

func (s *MCPRuntimeService) installSettings(ctx context.Context, req dto.MCPRuntimeInstall) (old, next dto.MCPRuntimeSettings, err error) {
	old, err = s.settings(ctx)
	if err != nil {
		return old, next, err
	}
	if req.Revision != old.Revision {
		return old, next, ErrMCPSettingsRevision
	}
	next = old
	next.RuntimeType, next.DockerImage, next.ListenPort, next.ExternalURL = "docker", req.DockerImage, req.ListenPort, req.ExternalURL
	next.AccessPolicy, next.AllowedIPs = req.AccessPolicy, append([]string{}, req.AllowedIPs...)
	if next.ExternalURL != old.ExternalURL {
		if row, loadErr := s.websiteSetting(ctx); loadErr != nil {
			return old, next, loadErr
		} else if row.WebsiteID != 0 || row.WebsiteBindingID != "" {
			return old, next, contract.Fail("MCP_CONFLICT", 409, "The website controls the MCP endpoint")
		}
	}
	if err := validateMCPRuntimeSettings(next); err != nil {
		return old, next, contract.Fail("MCP_INVALID_ARGUMENTS", 400, err.Error())
	}
	if err := normalizeMCPAccessSettings(&next); err != nil {
		return old, next, contract.Fail("MCP_INVALID_ARGUMENTS", 400, err.Error())
	}
	return old, next, nil
}

func (s *MCPRuntimeService) install(ctx context.Context, req dto.MCPRuntimeInstall) (dto.MCPRuntimeStatus, error) {
	old, next, err := s.installSettings(ctx, req)
	if err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "snapshot"})
	snapshot, err := s.snapshot(ctx, old)
	if err != nil {
		return dto.MCPRuntimeStatus{}, err
	}
	if !snapshot.installed {
		next.Enabled = true
	}
	prepared := old
	policyPrepared := !snapshot.installed || next.AccessPolicy != old.AccessPolicy || !reflect.DeepEqual(next.AllowedIPs, old.AllowedIPs)
	if policyPrepared {
		prepared.AccessPolicy, prepared.AllowedIPs = next.AccessPolicy, next.AllowedIPs
		if err := s.persist(prepared, old.Revision); err != nil {
			return dto.MCPRuntimeStatus{}, err
		}
		prepared.Revision++
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
		mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "rollback"})
		current, loadErr := s.Settings()
		if loadErr != nil {
			mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "recovery_failed"})
			return &mcpRuntimeRecoveryError{cause: errors.Join(operationErr, loadErr)}
		}
		if policyPrepared && current.Revision == prepared.Revision {
			if recoveryErr := rollback(old); recoveryErr != nil {
				mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "recovery_failed"})
				return &mcpRuntimeRecoveryError{cause: errors.Join(operationErr, recoveryErr)}
			}
			if recoveryErr := s.persist(old, prepared.Revision); recoveryErr != nil {
				mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "recovery_failed"})
				return &mcpRuntimeRecoveryError{cause: errors.Join(operationErr, recoveryErr)}
			}
			return operationErr
		}
		if recoveryErr := rollback(current); recoveryErr != nil {
			mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "recovery_failed"})
			return &mcpRuntimeRecoveryError{cause: errors.Join(operationErr, recoveryErr)}
		}
		return operationErr
	}
	if err := installMCPRuntimeImage(ctx, manager, next.Enabled); err != nil {
		return dto.MCPRuntimeStatus{}, recoverFailure(err)
	}
	mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "saving"})
	if err := s.persist(next, prepared.Revision); err != nil {
		return dto.MCPRuntimeStatus{}, recoverFailure(err)
	}
	mcpruntime.ReportProgress(ctx, mcpruntime.Progress{Status: "installed"})
	return s.Status(ctx)
}

func installMCPRuntimeImage(ctx context.Context, manager mcpRuntimeManager, enabled bool) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("MCP Docker image installation failed")
		}
	}()
	return manager.InstallImage(ctx, enabled)
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
