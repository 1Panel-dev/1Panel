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
	if out == "alias\n" && serviceName == "sshd.service" {
		return s.IsEnable("ssh")
	}
	if err != nil && out != "disabled\n" {
		if NewSnap().IsEnable(serviceName) {
			return true, nil
		}
		return false, err
	}
	return out == "enabled\n", nil
}

func (s *Systemd) IsExist(serviceName string) (bool, error) {
	out, err := run(s.toolCmd, "is-enabled", serviceName)
	switch strings.TrimSpace(out) {
	case "enabled", "enabled-runtime", "disabled", "static", "indirect", "generated", "transient", "alias", "linked", "linked-runtime", "masked", "masked-runtime":
		return true, err
	}
	if err != nil && out != "enabled\n" {
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
		if fallbackName := systemdAliasFallbackName(serviceName); fallbackName != "" && strings.Contains(out, "alias name or linked unit file") {
			return s.Operate(operate, fallbackName)
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

func systemdAliasFallbackName(serviceName string) string {
	switch serviceName {
	case "sshd", "sshd.service":
		return "ssh"
	case "sshd.socket":
		return "ssh.socket"
	default:
		return ""
	}
}
