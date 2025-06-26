package systemctl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
)

var (
	managers        = make(map[string]ServiceManager)
	mu              sync.RWMutex
	globalManager   ServiceManager
	managerPriority = []string{"systemd", "openrc", "sysvinit"}
)

const (
	defaultCommandTimeout = 60 * time.Second
	serviceCheckTimeout   = 5 * time.Second
)

type ServiceManager interface {
	Name() string
	IsAvailable() bool
	ServiceExists(*ServiceConfig) (bool, error)
	BuildCommand(string, *ServiceConfig) ([]string, error)
	ParseStatus(string, *ServiceConfig, string) (bool, error)
	FindServices(string) ([]string, error)
}

type baseManager struct {
	name         string
	cmdTool      string
	activeRegex  *regexp.Regexp
	enabledRegex *regexp.Regexp
}

func (b *baseManager) Name() string { return b.name }

func isRootUser() bool {
	return os.Geteuid() == 0
}
func (b *baseManager) buildBaseCommand() []string {
	var cmdArgs []string
	if !isRootUser() {
		cmdArgs = append(cmdArgs, "sudo")
	}
	cmdArgs = append(cmdArgs, b.cmdTool)
	return cmdArgs
}
func (b *baseManager) commonServiceExists(config *ServiceConfig, checkFn func(string) (bool, error)) (bool, error) {
	if name := config.ServiceName[b.name]; name != "" {
		exists, checkErr := checkFn(name)
		if checkErr != nil {
			return false, nil
		}
		return exists, nil
	}
	return false, nil
}
func (b *baseManager) ParseStatus(output string, _ *ServiceConfig, statusType string) (bool, error) {
	if output == "" {
		return false, nil
	}
	switch statusType {
	case "active":
		if b.activeRegex == nil {
			return false, nil
		}
		return b.activeRegex.MatchString(output), nil
	case "enabled":
		if b.enabledRegex == nil {
			return false, nil
		}
		return b.enabledRegex.MatchString(output), nil
	default:
		return false, nil
	}
}

func registerManager(m ServiceManager) {
	mu.Lock()
	defer mu.Unlock()
	managers[m.Name()] = m
}
func init() {
	for _, mgr := range []ServiceManager{
		newSystemdManager(),
		newOpenrcManager(),
		newSysvinitManager(),
	} {
		registerManager(mgr)
	}
}
func InitializeGlobalManager() (err error) {

	for _, name := range managerPriority {
		if mgr, ok := managers[name]; ok && mgr.IsAvailable() {
			if testManager(mgr) {
				globalManager = mgr
				global.LOG.Infof("Initialized service manager: %s", name)
				return
			}
		}
	}
	err = fmt.Errorf("no available service manager found (tried: %v)", managerPriority)

	return
}

func testManager(mgr ServiceManager) bool {
	_, err := mgr.BuildCommand("status", &ServiceConfig{
		ServiceName: map[string]string{mgr.Name(): "test-service"},
	})
	return err == nil
}

func GetGlobalManager() ServiceManager {
	if globalManager == nil {
		mu.Lock()
		defer mu.Unlock()
		if globalManager == nil {
			return initializeWithRetry()
		}
	}
	return globalManager
}
func initializeWithRetry() ServiceManager {
	const (
		maxRetries  = 5
		initialWait = 1 * time.Second
	)
	backoff := initialWait
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := InitializeGlobalManager(); err == nil {
			return globalManager
		}

		logMessage := fmt.Sprintf("Manager init attempt %d/%d failed", attempt, maxRetries)
		if global.LOG != nil {
			global.LOG.Warn(logMessage)
		} else {
			fmt.Printf("[WARN] %s\n", logMessage)
		}

		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	if global.LOG != nil {
		global.LOG.Error("All manager initialization attempts failed")
	} else {
		fmt.Println("[FATAL] All manager initialization attempts failed")
	}
	panic("unable to initialize service manager")
}
func executeCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultCommandTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, command, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return nil, &CommandError{
			Cmd:    cmd.String(),
			Output: buf.String(),
			Err:    err,
		}
	}
	return buf.Bytes(), nil
}

type systemdManager struct{ baseManager }

func newSystemdManager() ServiceManager {
	return &systemdManager{baseManager{
		name:         "systemd",
		cmdTool:      "systemctl",
		activeRegex:  regexp.MustCompile(`(?i)Active:\s+active\b`),
		enabledRegex: regexp.MustCompile(`(?i)^\s*enabled\s*$`),
	}}
}

func (m *systemdManager) IsAvailable() bool {
	_, err := exec.LookPath(m.cmdTool)
	return err == nil
}

func (m *systemdManager) ServiceExists(config *ServiceConfig) (bool, error) {
	return m.commonServiceExists(config, func(name string) (bool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), serviceCheckTimeout)
		defer cancel()
		out, cmdErr := executeCommand(ctx, m.cmdTool, "is-enabled", name)
		if cmdErr == nil || strings.Contains(string(out), "disabled") {
			return true, nil
		}
		return false, nil
	})
}

func (m *systemdManager) BuildCommand(action string, config *ServiceConfig) ([]string, error) {
	cmdArgs := m.buildBaseCommand()
	service := config.ServiceName[m.name]
	switch action {
	case "is-enabled":
		cmdArgs = append(cmdArgs, "is-enabled", service)
	default:
		cmdArgs = append(cmdArgs, action, service)
	}
	return cmdArgs, nil
}

func (m *systemdManager) ParseStatus(output string, config *ServiceConfig, statusType string) (bool, error) {
	if strings.Contains(output, "could not be found") {
		return false, nil
	}
	result, err := m.baseManager.ParseStatus(output, config, statusType)
	if err != nil {
		return false, nil
	}
	return result, nil
}
func (m *systemdManager) FindServices(keyword string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := executeCommand(ctx, m.cmdTool, "list-unit-files", "--type=service", "--no-legend")
	if err != nil {
		return nil, fmt.Errorf("failed to list systemd services: %w", err)
	}

	var services []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		serviceName := fields[0]
		if strings.Contains(serviceName, keyword) {
			services = append(services, serviceName)
		}
	}
	return services, nil
}

type sysvinitManager struct{ baseManager }

func newSysvinitManager() ServiceManager {
	return &sysvinitManager{baseManager{
		name:         "sysvinit",
		cmdTool:      "service",
		activeRegex:  regexp.MustCompile(`(?i)(?:^|\s)\b(running|active)\b(?:$|\s)`),
		enabledRegex: regexp.MustCompile(`(?i)(?:^|\s)\b(enabled)\b(?:$|\s)`),
	}}
}

func (m *sysvinitManager) IsAvailable() bool {
	_, err := exec.LookPath(m.cmdTool)
	return err == nil
}

func (m *sysvinitManager) ServiceExists(config *ServiceConfig) (bool, error) {
	return m.commonServiceExists(config, func(name string) (bool, error) {
		_, err := os.Stat(filepath.Join("/etc/init.d", name))
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("stat /etc/init.d/%s failed: %w", name, err)
		}
		return true, nil
	})
}

func (m *sysvinitManager) BuildCommand(action string, config *ServiceConfig) ([]string, error) {
	service := config.ServiceName[m.name]
	switch action {
	case "is-enabled":
		return []string{
			"sh",
			"-c",
			fmt.Sprintf("if ls /etc/rc*.d/S*%s >/dev/null 2>&1; then echo 'enabled'; else echo 'disabled'; fi", service)}, nil
	case "is-active", "status":
		return []string{
			"sh",
			"-c",
			fmt.Sprintf("if service %s status >/dev/null 2>&1; then echo 'active'; else echo 'inactive'; fi", service),
		}, nil
	default:
		cmdArgs := m.buildBaseCommand()
		cmdArgs = append(cmdArgs, service, action)
		return cmdArgs, nil
	}
}

func (m *sysvinitManager) ParseStatus(output string, config *ServiceConfig, statusType string) (bool, error) {
	serviceName := config.ServiceName[m.name]
	switch statusType {
	case "enabled":
		if strings.Contains(output, "no such file or directory") {
			return false, nil
		}
		return strings.TrimSpace(output) != "", nil
	case "active":
		if strings.Contains(output, "not found") {
			return false, nil
		}
		if strings.Contains(output, "running") || strings.Contains(output, "active") {
			return true, nil
		}
	default:
		result, err := m.baseManager.ParseStatus(output, config, statusType)
		if err != nil {
			global.LOG.Debugf("[sysvinit] Status parse failed. [ServiceName:%s] Type: %s, Output: %q", serviceName, statusType, output)
		}
		return result, err
	}
	return false, fmt.Errorf("unsupported status type: %s", statusType)
}

func (m *sysvinitManager) FindServices(keyword string) ([]string, error) {
	files, err := os.ReadDir("/etc/init.d/")
	if err != nil {
		return nil, fmt.Errorf("failed to read init.d directory: %w", err)
	}

	var services []string
	for _, file := range files {
		if strings.Contains(file.Name(), keyword) {
			services = append(services, file.Name())
		}
	}
	return services, nil
}

type openrcManager struct{ baseManager }

func newOpenrcManager() ServiceManager {
	return &openrcManager{baseManager{
		name:    "openrc",
		cmdTool: "rc-service",
		// activeRegex:  regexp.MustCompile(`(?i)^\s*status:\s+(started|running|active)\s*$`),
		// enabledRegex: regexp.MustCompile(`(?i)^[^\|]+\|\s*(default|enabled)\b.*$`),
		activeRegex:  regexp.MustCompile(`(?i)(?:^|\s)\b(running|active)\b(?:$|\s)`),
		enabledRegex: regexp.MustCompile(`(?i)(?:^|\s)\b(enabled)\b(?:$|\s)`),
	}}
}
func (m *openrcManager) IsAvailable() bool {
	_, err := exec.LookPath(m.cmdTool)
	return err == nil
}

func (m *openrcManager) ServiceExists(config *ServiceConfig) (bool, error) {
	return m.commonServiceExists(config, func(name string) (bool, error) {
		_, err := os.Stat(filepath.Join("/etc/init.d", name))
		if os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("stat /etc/init.d/%s failed: %w", name, err)
		}
		return true, nil
	})
}

func (m *openrcManager) BuildCommand(action string, config *ServiceConfig) ([]string, error) {
	cmdArgs := m.buildBaseCommand()
	service := config.ServiceName[m.name]
	switch action {
	case "is-enabled":
		return []string{
			"sh",
			"-c",
			fmt.Sprintf("if ls /etc/runlevels/default/%s >/dev/null 2>&1; then echo 'enabled'; else echo 'disabled'; fi", service)}, nil
	case "is-active", "status":
		return []string{
			"sh",
			"-c",
			fmt.Sprintf("if service %s status >/dev/null 2>&1; then echo 'active'; else echo 'inactive'; fi", service),
		}, nil
	case "enable":
		return []string{"rc-update", "add", service, "default"}, nil
	case "disable":
		return []string{"rc-update", "del", service, "default"}, nil
	default:
		cmdArgs = append(cmdArgs, service, action)
		return cmdArgs, nil
	}
}

func (m *openrcManager) ParseStatus(output string, config *ServiceConfig, statusType string) (bool, error) {
	if strings.Contains(output, "does not exist") {
		return false, nil
	}
	result, err := m.baseManager.ParseStatus(output, config, statusType)
	if err != nil {
		return false, nil
	}
	return result, nil
}
func (m *openrcManager) FindServices(keyword string) ([]string, error) {
	files, err := os.ReadDir("/etc/init.d/")
	if err != nil {
		return nil, fmt.Errorf("failed to read init.d directory: %w", err)
	}
	keyword = strings.ToLower(keyword)
	var services []string
	for _, file := range files {
		if strings.Contains(file.Name(), keyword) {
			services = append(services, file.Name())
		}
	}
	return services, nil
}

type CommandError struct {
	Cmd    string
	Err    error
	Output string
}

func (e CommandError) Error() string {
	return fmt.Sprintf("command %q failed: %v \nOutput: %s",
		e.Cmd, e.Err, e.Output)
}

func (e CommandError) Unwrap() error { return e.Err }

func ReinitializeManager() error {
	mu.Lock()
	defer mu.Unlock()
	globalManager = nil
	return InitializeGlobalManager()
}

func SetManagerPriority(order []string) {
	mu.Lock()
	defer mu.Unlock()
	managerPriority = order
}
