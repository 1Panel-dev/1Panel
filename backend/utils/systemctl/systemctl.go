package systemctl

import (
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
			"sysvinit": "docker",
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
		global.LOG.Errorf("Create service handle failed: %v", err)
		return errors.WithMessage(err, "create service handle failed")
	}

	_, err = h.WithTimeout(30 * time.Second).Execute("restart")
	if err != nil {
		if serr, ok := err.(ServiceError); ok {
			global.LOG.Errorf("Restart docker failed [%s], Output: %s", serr.Wrapped, serr.Output)
		}
		return errors.WithMessage(err, "restart service failed")
	}
	return nil
}

func SystemRestart() error {
	h, err := NewServiceHandle(PanelService)
	if err != nil {
		return err
	}
	_, err = h.WithTimeout(30 * time.Second).Execute("restart")
	return err
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
