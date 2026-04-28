package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Sysvinit struct{ toolCmd string }

func NewSysvinit() *Sysvinit {
	return &Sysvinit{toolCmd: "service"}
}

func (s *Sysvinit) Name() string {
	return "sysvinit"
}
func (s *Sysvinit) IsActive(serviceName string) (bool, error) {
	_, err := cmd.NewCommandMgr().RunWithStdout("service", serviceName, "status")
	return err == nil, nil
}
func (s *Sysvinit) IsEnable(serviceName string) (bool, error) {
	return isSysvServiceEnabled(serviceName)
}
func (s *Sysvinit) IsExist(serviceName string) (bool, error) {
	_, err := os.Stat(filepath.Join("/etc/init.d", serviceName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat /etc/init.d/%s failed: %w", serviceName, err)
	}
	return true, nil
}
func (s *Sysvinit) Status(serviceName string) (string, error) {
	return run(s.toolCmd, serviceName, "status")
}

func (s *Sysvinit) Operate(operate, serviceName string) error {
	return handlerErr(run(s.toolCmd, serviceName, operate))
}

func (s *Sysvinit) Reload() error {
	return nil
}
