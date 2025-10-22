package manager

import (
	"strings"
)

type Systemd struct{ toolCmd string }

func NewSystemd() *Systemd {
	return &Systemd{toolCmd: "systemctl"}
}

func (s *Systemd) Name() string {
	return "systemd"
}
func (s *Systemd) IsActive(serviceName string) (bool, error) {
	out, err := run(s.toolCmd, "is-active", serviceName)
	if err != nil && out != "inactive\n" {
		if NewSnap().IsActive(serviceName) {
			return true, nil
		}
		return false, err
	}
	return out == "active\n", nil
}

func (s *Systemd) IsEnable(serviceName string) (bool, error) {
	out, err := run(s.toolCmd, "is-enabled", serviceName)
	if err != nil && out != "disabled\n" {
		if serviceName == "sshd" && out == "alias\n" {
			return s.IsEnable("ssh")
		}
		if NewSnap().IsEnable(serviceName) {
			return true, nil
		}
		return false, err
	}
	return out == "enabled\n", nil
}

func (s *Systemd) IsExist(serviceName string) (bool, error) {
	out, err := run(s.toolCmd, "is-enabled", serviceName)
	if err != nil && out != "enabled\n" {
		if strings.Contains(out, "disabled") {
			return true, err
		}
		if NewSnap().IsExist(serviceName) {
			return true, nil
		}
		return false, err
	}
	return true, err
}

func (s *Systemd) Status(serviceName string) (string, error) {
	return run(s.toolCmd, "status", serviceName)
}
func (s *Systemd) Operate(operate, serviceName string) error {
	out, err := run(s.toolCmd, operate, serviceName)
	if err != nil {
		if serviceName == "sshd" && strings.Contains(out, "alias name or linked unit file") {
			return s.Operate(operate, "ssh")
		}
		if err := NewSnap().Operate(operate, serviceName); err == nil {
			return nil
		}
		return handlerErr(run(s.toolCmd, operate, serviceName))
	}
	return nil
}

func (s *Systemd) Reload() error {
	out, err := run(s.toolCmd, "daemon-reload")
	return handlerErr(out, err)
}
