package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
)

type Openrc struct {
	toolCmd string
	Client  *ssh.SSHClient
}

func NewOpenrc() *Openrc {
	return &Openrc{toolCmd: "rc-service"}
}

func (s *Openrc) Name() string {
	return "openrc"
}
func (s *Openrc) IsActive(serviceName string) (bool, error) {
	out, err := cmd.RunDefaultWithStdoutBashCf("if service %s status >/dev/null 2>&1; then echo 'active'; else echo 'inactive'; fi", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}
func (s *Openrc) IsEnable(serviceName string) (bool, error) {
	out, err := cmd.RunDefaultWithStdoutBashCf("if ls /etc/rc*.d/S*%s >/dev/null 2>&1; then echo 'enabled'; else echo 'disabled'; fi", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}
func (s *Openrc) IsExist(serviceName string) (bool, error) {
	if _, err := os.Stat(filepath.Join("/etc/init.d", serviceName)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat /etc/init.d/%s failed: %w", serviceName, err)
	}
	return true, nil
}
func (s *Openrc) Status(serviceName string) (string, error) {
	return run(s.Client, s.toolCmd, serviceName, "status")
}

func (s *Openrc) Operate(operate, serviceName string) error {
	switch operate {
	case "enable":
		return handlerErr(run(s.Client, "rc-update", "add", serviceName, "default"))
	case "disable":
		return handlerErr(run(s.Client, "rc-update", "del", serviceName, "default"))
	default:
		return handlerErr(run(s.Client, s.toolCmd, serviceName, operate))
	}
}

func (s *Openrc) Reload() error {
	return nil
}
