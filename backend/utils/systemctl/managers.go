package systemctl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var managers = make(map[string]ServiceManager)

type ServiceManager interface {
	Name() string
	IsAvailable() bool
	ServiceExists(*ServiceConfig) (bool, error)
	BuildCommand(string, *ServiceConfig) ([]string, error)
	ParseActiveStatus(string, *ServiceConfig) (bool, error)
	ParseEnabledStatus(string, *ServiceConfig) (bool, error)
}

func init() {
	RegisterManager(&systemdManager{})
	RegisterManager(&openrcManager{})
	RegisterManager(&sysvinitManager{})
}

func RegisterManager(m ServiceManager) {
	managers[m.Name()] = m
}

type systemdManager struct{}

func (m *systemdManager) Name() string { return "systemd" }
func (m *systemdManager) IsAvailable() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

func (m *systemdManager) ServiceExists(config *ServiceConfig) (bool, error) {
	name := config.ServiceName[m.Name()]
	if name == "" {
		return false, nil
	}

	cmd := exec.Command("systemctl", "list-units", "--all", "--no-legend", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("systemctl list-units failed: %w", err)
	}

	if strings.Contains(string(out), name) {
		return true, nil
	}
	return m.checkUnitFiles(name)
}

func (m *systemdManager) checkUnitFiles(name string) (bool, error) {
	cmd := exec.Command("systemctl", "list-unit-files", "--no-legend", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("systemctl list-unit-files failed: %w", err)
	}
	return strings.Contains(string(out), name), nil
}

func (m *systemdManager) BuildCommand(action string, config *ServiceConfig) ([]string, error) {
	service := config.ServiceName[m.Name()]
	if service == "" {
		return nil, fmt.Errorf("missing systemd service name")
	}

	commands := []string{"systemctl", action, service}
	if config.UseSocket && (action == "restart" || action == "start") {
		commands = append(commands, service+".socket")
	}
	return commands, nil
}

func (m *systemdManager) ParseActiveStatus(output string, _ *ServiceConfig) (bool, error) {
	return strings.TrimSpace(output) == "active", nil
}

func (m *systemdManager) ParseEnabledStatus(output string, _ *ServiceConfig) (bool, error) {
	return strings.TrimSpace(output) == "enabled", nil
}

type sysvinitManager struct{}

func (m *sysvinitManager) Name() string { return "sysvinit" }
func (m *sysvinitManager) IsAvailable() bool {
	if _, err := os.Stat("/etc/init.d"); err != nil {
		return false
	}
	_, err := exec.LookPath("service")
	return err == nil
}

func (m *sysvinitManager) ServiceExists(config *ServiceConfig) (bool, error) {
	name := config.ServiceName[m.Name()]
	if name == "" {
		return false, nil
	}
	_, err := os.Stat(fmt.Sprintf("/etc/init.d/%s", name))
	return !os.IsNotExist(err), nil
}

func (m *sysvinitManager) BuildCommand(action string, config *ServiceConfig) ([]string, error) {
	service := config.ServiceName[m.Name()]
	if service == "" {
		return nil, fmt.Errorf("missing sysvinit service name")
	}
	return []string{"service", service, action}, nil
}

func (m *sysvinitManager) ParseActiveStatus(output string, config *ServiceConfig) (bool, error) {
	return strings.Contains(strings.ToLower(output), "running"), nil
}

func (m *sysvinitManager) ParseEnabledStatus(output string, config *ServiceConfig) (bool, error) {
	name := config.ServiceName[m.Name()]
	_, err := os.Stat(fmt.Sprintf("/etc/rc3.d/S??%s", name))
	return !os.IsNotExist(err), nil
}

type openrcManager struct{}

func (m *openrcManager) Name() string { return "openrc" }
func (m *openrcManager) IsAvailable() bool {
	_, err := exec.LookPath("rc-service")
	return err == nil
}

func (m *openrcManager) ServiceExists(config *ServiceConfig) (bool, error) {
	name := config.ServiceName[m.Name()]
	if name == "" {
		return false, nil
	}

	cmd := exec.Command("rc-service", "-l")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("rc-service list failed: %w", err)
	}
	return strings.Contains(string(out), name), nil
}

func (m *openrcManager) BuildCommand(action string, config *ServiceConfig) ([]string, error) {
	service := config.ServiceName[m.Name()]
	if service == "" {
		return nil, fmt.Errorf("missing openrc service name")
	}
	return []string{"rc-service", service, action}, nil
}

func (m *openrcManager) ParseActiveStatus(output string, _ *ServiceConfig) (bool, error) {
	return strings.Contains(output, "started"), nil
}

func (m *openrcManager) ParseEnabledStatus(output string, _ *ServiceConfig) (bool, error) {
	return strings.Contains(output, "enabled"), nil
}
