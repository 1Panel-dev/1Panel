package systemctl

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ServiceConfig struct {
	ID          string
	DisplayName string
	ServiceName map[string]string
	UseSocket   bool
	Description string
}

type ServiceHandle struct {
	config      *ServiceConfig
	manager     ServiceManager
	timeout     time.Duration
	ctx         context.Context
	statusCache *ServiceStatusCache
}

var (
	logger          = logrus.New()
	managerPriority = []string{"systemd", "openrc", "sysvinit", "snap"}
)

type ServiceStatusCache struct {
	sync.RWMutex
	items map[string]statusEntry
}

type statusEntry struct {
	status    string
	timestamp time.Time
}

const cacheTTL = 30 * time.Second

func NewServiceStatusCache() *ServiceStatusCache {
	return &ServiceStatusCache{
		items: make(map[string]statusEntry),
	}
}

func (c *ServiceStatusCache) Get(key string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	entry, exists := c.items[key]
	if !exists || time.Since(entry.timestamp) > cacheTTL {
		return "", false
	}
	return entry.status, true
}

func (c *ServiceStatusCache) Set(key, status string) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = statusEntry{
		status:    status,
		timestamp: time.Now(),
	}
}

func NewServiceHandle(config *ServiceConfig) (*ServiceHandle, error) {
	for _, name := range managerPriority {
		mgr, ok := managers[name]
		if !ok || !mgr.IsAvailable() {
			continue
		}

		exists, err := mgr.ServiceExists(config)
		if err != nil {
			logger.Warnf("Service check failed for %s: %v", name, err)
			continue
		}

		if exists {
			return &ServiceHandle{
				config:      config,
				manager:     mgr,
				timeout:     30 * time.Second,
				ctx:         context.Background(),
				statusCache: NewServiceStatusCache(),
			}, nil
		}
	}
	return nil, fmt.Errorf("no available manager for %q", config.ID)
}

func (h *ServiceHandle) WithTimeout(timeout time.Duration) *ServiceHandle {
	h.timeout = timeout
	return h
}

func (h *ServiceHandle) Execute(action string) (string, error) {
	ctx, cancel := context.WithTimeout(h.ctx, h.timeout)
	defer cancel()

	cmdArgs, err := h.manager.BuildCommand(action, h.config)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()
	result := string(output)

	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return result, ServiceError{
			Action:   action,
			Service:  h.config.ServiceName[h.manager.Name()],
			Output:   result,
			Wrapped:  err,
			ExitCode: exitCode,
		}
	}
	return result, nil
}
func (h *ServiceHandle) IsActive() (bool, error) {
	output, err := h.Execute("status")
	if err != nil {
		var se ServiceError
		if errors.As(err, &se) && isInactiveCode(se.ExitCode) {
			return false, nil
		}
		return false, err
	}
	return h.manager.ParseActiveStatus(output, h.config)
}

func (h *ServiceHandle) ManagerName() string {
	return h.manager.Name()
}

func (h *ServiceHandle) Config() ServiceConfig {
	if h == nil {
		return ServiceConfig{}
	}
	return ServiceConfig{
		ID:          h.config.ID,
		ServiceName: cloneStringMap(h.config.ServiceName),
		Description: h.config.Description,
		DisplayName: h.config.DisplayName,
		UseSocket:   h.config.UseSocket,
	}
}
func (h *ServiceHandle) GetServicePath() (string, error) {
	if h == nil || h.manager == nil {
		return "", errors.New("invalid service handle")
	}

	managerName := h.ManagerName()
	serviceName := h.config.ServiceName[managerName]

	if serviceName == "" {
		return "", ServicePathError{
			Manager: managerName,
			Service: h.config.ID,
		}
	}

	switch managerName {
	case "systemd":
		return fmt.Sprintf("/etc/systemd/system/%s", serviceName), nil
	case "openrc", "sysvinit":
		return fmt.Sprintf("/etc/init.d/%s", serviceName), nil
	default:
		return "", fmt.Errorf("unsupported init system: %s", managerName)
	}
}

func cloneStringMap(m map[string]string) map[string]string {
	clone := make(map[string]string, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

func (h *ServiceHandle) GetServiceFileName() (string, error) {
	managerName := h.ManagerName()
	serviceName := h.config.ServiceName[managerName]
	if serviceName == "" {
		return "", fmt.Errorf("service name not configured for %s", managerName)
	}
	return serviceName, nil

}
func (h *ServiceHandle) GetBinaryPath() (string, error) {
	return "/usr/local/bin", nil // 标准路径可配置化
}
