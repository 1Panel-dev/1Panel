package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
)

type Sysvinit struct {
	toolCmd string
	Client  *ssh.SSHClient
}

func NewSysvinit() *Sysvinit {
	return &Sysvinit{toolCmd: "service"}
}

func (s *Sysvinit) Name() string {
	return "sysvinit"
}
func (s *Sysvinit) IsActive(serviceName string) (bool, error) {
	out, err := cmd.RunDefaultWithStdoutBashCf("if service %s status >/dev/null 2>&1; then echo 'active'; else echo 'inactive'; fi", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}
func (s *Sysvinit) IsEnable(serviceName string) (bool, error) {
	out, err := cmd.RunDefaultWithStdoutBashCf("if ls /etc/rc*.d/S*%s >/dev/null 2>&1; then echo 'enabled'; else echo 'disabled'; fi", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}
func (s *Sysvinit) IsExist(serviceName string) (bool, error) {
	if _, err := os.Stat(filepath.Join("/etc/init.d", serviceName)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat /etc/init.d/%s failed: %w", serviceName, err)
	}
	return true, nil
}
func (s *Sysvinit) Status(serviceName string) (string, error) {
	return run(s.Client, s.toolCmd, serviceName, "status")
}

func (s *Sysvinit) Operate(operate, serviceName string) error {
	return handlerErr(run(s.Client, s.toolCmd, serviceName, operate))
}

func (s *Sysvinit) Reload() error {
	return nil
}
