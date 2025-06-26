package systemctl

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
)

type ServiceConfig struct {
	ServiceName map[string]string
}

type ServiceHandler struct {
	config  *ServiceConfig
	manager ServiceManager
}

func NewServiceHandler(serviceNames map[string]string) *ServiceHandler {
	mgr := GetGlobalManager()
	if mgr == nil {
		global.LOG.Error("failed to get global service manager when creating ServiceHandler")
		return nil
	}
	return &ServiceHandler{
		config: &ServiceConfig{
			ServiceName: serviceNames,
		},
		manager: mgr,
	}
}

type ServiceStatus struct {
	IsActive  bool   `json:"isActive"`
	IsEnabled bool   `json:"isEnabled"`
	IsExists  bool   `json:"isExists"`
	Output    string `json:"output"`
}
type ServiceIsActive struct {
	IsActive bool   `json:"isActive"`
	Output   string `json:"output"`
}

type ServiceIsEnabled struct {
	IsEnabled bool   `json:"isEnabled"`
	Output    string `json:"output"`
}

type ServiceResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

var (
	BinaryPath         = "/usr/local/bin" // 1panl service default path
	ErrServiceNotExist = errors.New("service does not exist")
)

func defaultServiceConfig(serviceName string) map[string]string {
	mgr := getManagerName()
	if mgr == "" {
		global.LOG.Error("failed to get manager name for default service config")
		return nil
	}
	return map[string]string{
		mgr: serviceName,
	}
}

func (h *ServiceHandler) ManagerName() string { return h.manager.Name() }
func getManagerName() string {
	if mgr := GetGlobalManager(); mgr != nil {
		return mgr.Name()
	}
	global.LOG.Error("failed to get global service manager")
	return ""
}

func (h *ServiceHandler) GetServiceName() string {
	manager := h.ManagerName()
	if manager == "" {
		global.LOG.Error("manager name is empty when getting service name")
		return ""
	}
	return h.config.ServiceName[manager]
}

func (h *ServiceHandler) GetServicePath() (string, error) {
	manager := h.ManagerName()
	serviceName := h.config.ServiceName[manager]

	if serviceName == "" {
		err := fmt.Errorf("service name not found for %s", manager)
		global.LOG.Errorf("GetServicePath error: %v", err)
		return "", err
	}

	cleanPath := filepath.Clean(serviceName)
	if strings.Contains(cleanPath, "..") {
		err := fmt.Errorf("invalid path: %q", cleanPath)
		global.LOG.Errorf("GetServicePath security check failed: %v", err)
		return "", err
	}
	switch manager {
	case "systemd":
		return findSystemdPath(cleanPath)
	case "openrc", "sysvinit":
		return checkInitDPath(cleanPath)
	default:
		err := fmt.Errorf("unsupported init system: %s", manager)
		global.LOG.Errorf("GetServicePath error: %v", err)
		return "", err
	}
}

func findSystemdPath(name string) (string, error) {
	paths := []string{"/etc/systemd/system", "/usr/lib/systemd/system",
		"/usr/share/systemd/system", "/usr/local/lib/systemd/system"}

	for _, p := range paths {
		if path := filepath.Join(p, name); FileExist(path) {
			return path, nil
		}
	}
	err := fmt.Errorf("service path not found for %s", name)
	global.LOG.Errorf("findSystemdPath error: %v", err)
	return "", err
}

func checkInitDPath(name string) (string, error) {
	path := filepath.Join("/etc/init.d", name)
	if !FileExist(path) {
		err := fmt.Errorf("service path not found for %s", name)
		global.LOG.Errorf("checkInitDPath error: %v", err)
		return "", err
	}
	return path, nil
}

func (h *ServiceHandler) ExecuteAction(action string) (ServiceResult, error) {
	successMsg := fmt.Sprintf("%s : %s completed", action, h.GetServiceName())
	return h.executeAction(action, successMsg)
}

func (h *ServiceHandler) CheckStatus() (ServiceStatus, error) {
	manager := GetGlobalManager()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type result struct {
		isActive  bool
		isEnabled bool
		output    string
		err       error
	}
	var status ServiceStatus
	var errs []error

	results := make(chan result, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		res := result{}
		cmd, err := manager.BuildCommand("status", h.config)
		if err != nil {
			res.err = fmt.Errorf("build status command failed: %w", err)
			results <- res
			return
		}

		output, err := executeCommand(ctx, cmd[0], cmd[1:]...)
		if err != nil {
			res.err = fmt.Errorf("status check failed: %w", err)
			results <- res
			return
		}

		isActive, err := manager.ParseStatus(string(output), h.config, "active")
		if err != nil {
			res.err = fmt.Errorf("parse status failed: %w", err)
			results <- res
			return
		}
		res.isActive = isActive
		res.output = string(output)
		results <- res
	}()

	go func() {
		defer wg.Done()
		res := result{}
		cmd, err := manager.BuildCommand("is-enabled", h.config)
		if err != nil {
			res.err = fmt.Errorf("build is-enabled command failed: %w", err)
			results <- res
			return
		}

		output, err := executeCommand(ctx, cmd[0], cmd[1:]...)
		if err != nil {
			res.err = fmt.Errorf("enabled check failed: %w", err)
			results <- res
			return
		}

		isEnabled, err := manager.ParseStatus(string(output), h.config, "enabled")
		if err != nil {
			res.err = fmt.Errorf("parse enabled status failed: %w", err)
			results <- res
			return
		}
		res.isEnabled = isEnabled
		results <- res
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.err != nil {
			errs = append(errs, res.err)
			continue
		}
		status.IsActive = res.isActive
		status.IsEnabled = res.isEnabled
		if res.output != "" {
			status.Output = res.output
		}
	}

	if len(errs) > 0 {
		return status, errors.Join(errs...)
	}
	return status, nil
}

func (h *ServiceHandler) IsExists() (ServiceStatus, error) {
	manager := GetGlobalManager()
	isExist, _ := manager.ServiceExists(h.config)
	return ServiceStatus{
		IsExists: isExist,
	}, nil
}
func (h *ServiceHandler) IsActive() (ServiceStatus, error) {
	manager := GetGlobalManager()
	if manager == nil {
		global.LOG.Error("service manager not initialized during active check")
		return ServiceStatus{}, fmt.Errorf("service manager not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	activeCmd, err := manager.BuildCommand("status", h.config)
	if err != nil {
		global.LOG.Errorf("Build status command failed: %v", err)
		return ServiceStatus{}, fmt.Errorf("build status command failed: %w", err)
	}

	output, err := executeCommand(ctx, activeCmd[0], activeCmd[1:]...)
	if err != nil {
		if strings.Contains(err.Error(), "inactive") {
			return ServiceStatus{
				IsExists: false,
			}, nil
		}
		global.LOG.Errorf("Active check execution failed: %v", err)
		return ServiceStatus{}, fmt.Errorf("status check failed: %w", err)
	}

	isActive, err := manager.ParseStatus(string(output), h.config, "active")
	if err != nil {
		global.LOG.Warnf("Status parse error: %v", err)
	}
	return ServiceStatus{
		IsActive: isActive,
		Output:   string(output),
	}, nil
}

func (h *ServiceHandler) IsEnabled() (ServiceStatus, error) {
	manager := GetGlobalManager()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	enabledCmd, err := manager.BuildCommand("is-enabled", h.config)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ServiceStatus{
				IsEnabled: false,
			}, nil
		}
		global.LOG.Errorf("Build is-enabled command failed: %v", err)
		return ServiceStatus{}, fmt.Errorf("build enabled check command failed: %w", err)
	}

	output, err := executeCommand(ctx, enabledCmd[0], enabledCmd[1:]...)
	if err != nil {
		if strings.Contains(err.Error(), "disabled") {
			return ServiceStatus{
				IsEnabled: false,
			}, nil
		}
		// 	// isEnabled, err := h.ParseStatus(string(output), h.config, "enabled")
		// global.LOG.Errorf("Enabled check execution failed: %v", err)
		// return ServiceStatus{}, fmt.Errorf("enabled check failed: %w", err)
	}

	isEnabled, err := manager.ParseStatus(string(output), h.config, "enabled")
	if err != nil {
		global.LOG.Warnf("Enabled status parse error: %v", err)
	}
	return ServiceStatus{
		IsEnabled: isEnabled,
		Output:    string(output),
	}, nil
}

func (h *ServiceHandler) StartService() (ServiceResult, error) {
	return h.ExecuteAction("start")
}

func (h *ServiceHandler) StopService() (ServiceResult, error) {
	return h.ExecuteAction("stop")
}

func (h *ServiceHandler) RestartService() (ServiceResult, error) {
	return h.ExecuteAction("restart")
}

func (h *ServiceHandler) EnableService() (ServiceResult, error) {
	return h.ExecuteAction("enable")
}

func (h *ServiceHandler) DisableService() (ServiceResult, error) {
	return h.ExecuteAction("disable")
}

func (h *ServiceHandler) executeAction(action, successMsg string) (ServiceResult, error) {
	manager := GetGlobalManager()
	if manager == nil {
		global.LOG.Error("service manager not initialized during action execution")
		return ServiceResult{}, fmt.Errorf("service manager not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultCommandTimeout)
	defer cancel()

	cmdArgs, err := manager.BuildCommand(action, h.config)
	if err != nil {
		global.LOG.Errorf("Build command failed for action %s: %v", action, err)
		return ServiceResult{}, fmt.Errorf("build command failed: %w", err)
	}

	output, err := executeCommand(ctx, cmdArgs[0], cmdArgs[1:]...)
	if err != nil {
		global.LOG.Errorf("%s operation failed: %v", action, err)
		return ServiceResult{
			Success: false,
			Message: fmt.Sprintf("%s failed", action),
			Output:  string(output),
		}, fmt.Errorf("%s operation failed: %w", action, err)
	}

	global.LOG.Infof("[%s]: %s", manager.Name(), successMsg)
	return ServiceResult{
		Success: true,
		Message: successMsg,
		Output:  string(output),
	}, nil
}

func (h *ServiceHandler) ReloadManager() error {
	if err := ReinitializeManager(); err != nil {
		global.LOG.Errorf("Failed to reload service manager: %v", err)
		return fmt.Errorf("failed to reload service manager: %w", err)
	}
	global.LOG.Info("Service manager reloaded successfully")
	return nil
}

var (
	ExecuteCommand = executeCommand
)
