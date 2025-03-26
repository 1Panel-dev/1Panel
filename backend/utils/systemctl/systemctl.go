package systemctl

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/pkg/errors"
)

var (
	DockerService = &ServiceConfig{
		ID:          "docker",
		DisplayName: "Docker Engine",
		ServiceName: map[string]string{
			"systemd":  "docker.service",
			"openrc":   "docker",
			"sysvinit": "dockerd",
			"snap":     "docker",
		},
		UseSocket:   true,
		Description: "Container runtime service",
	}

	PanelService = &ServiceConfig{
		ID:          "1panel",
		DisplayName: "1Panel Service",
		ServiceName: map[string]string{
			"systemd":  "1panel.service",
			"openrc":   "1paneld",
			"sysvinit": "1paneld",
		},
		UseSocket:   false,
		Description: "1Panel service management",
	}
)

func RestartDocker() error {
	h, err := NewServiceHandle(DockerService)
	if err != nil {
		global.LOG.Errorf("Create docker service handle failed: %v", err)
		return errors.WithMessage(err, "create docker service handle failed")
	}

	_, err = h.WithTimeout(30 * time.Second).Execute("restart")
	if err != nil {
		if serr, ok := err.(ServiceError); ok {
			global.LOG.Errorf("Restart docker failed [%s], Output: %s", serr.Wrapped, serr.Output)
		}
		return errors.WithMessage(err, "restart docker failed")
	}
	return nil
}

func SystemRestart() error {
	h, err := NewServiceHandle(PanelService)
	if err != nil {
		global.LOG.Errorf("Create 1panel service handle failed: %v", err)
		return errors.WithMessage(err, "create 1panel service handle failed")
	}
	_, err = h.WithTimeout(30 * time.Second).Execute("restart")
	if serr, ok := err.(ServiceError); ok {
		global.LOG.Errorf("Restart  1panel failed [%s], Output: %s", serr.Wrapped, serr.Output)
	}
	return errors.WithMessage(err, "restart 1panel failed")
}

func IsActive(config *ServiceConfig) (bool, error) {
	h, err := NewServiceHandle(config)
	if err != nil {
		return false, err
	}

	output, err := h.Execute("status")
	if err != nil {
		var se ServiceError
		if errors.As(err, &se) && isInactiveCode(se.ExitCode) {
			return false, nil
		}
		return false, err
	}

	return h.manager.ParseActiveStatus(output, config)
}

func IsEnabled(config *ServiceConfig) (bool, error) {
	h, err := NewServiceHandle(config)
	if err != nil {
		return false, err
	}

	output, err := h.Execute("is-enabled")
	if err != nil {
		var se ServiceError
		if errors.As(err, &se) && isDisabledCode(se.ExitCode) {
			return false, nil
		}
		return false, err
	}

	return h.manager.ParseEnabledStatus(output, config)
}

func IsExist(config *ServiceConfig) (bool, error) {
	h, err := NewServiceHandle(config)
	if err != nil {
		return false, err
	}

	_, err = h.Execute("status")
	if err != nil {
		var se ServiceError
		if errors.As(err, &se) && se.ExitCode == 3 { // 3 表示单元未找到
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func Operate(operate string, config *ServiceConfig) error {
	h, err := NewServiceHandle(config)
	if err != nil {
		return err
	}
	_, err = h.Execute(operate)
	return err
}

func isInactiveCode(code int) bool {
	return code == 3 || code > 0
}

func isDisabledCode(code int) bool {
	return code == 1 || code > 0
}

// 保留原systemctl函数，保持兼容
func RunSystemCtl(args ...string) (string, error) {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %w", err)
	}
	return string(output), nil
}

func SctlIsActive(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-active", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}

func IsEnable(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}

func SctlIsExist(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		if strings.Contains(out, "disabled") {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}
func handlerErr(out string, err error) error {
	if err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func SctlRestart(serviceName string) error {
	out, err := RunSystemCtl("restart", serviceName)
	return handlerErr(out, err)
}
