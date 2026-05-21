package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Openrc struct{ toolCmd string }

func NewOpenrc() *Openrc {
	return &Openrc{toolCmd: "rc-service"}
}

func (s *Openrc) Name() string {
	return "openrc"
}
func (s *Openrc) IsActive(serviceName string) (bool, error) {
	_, err := cmd.NewCommandMgr().RunWithStdout("service", serviceName, "status")
	return err == nil, nil
}
func (s *Openrc) IsEnable(serviceName string) (bool, error) {
	return isSysvServiceEnabled(serviceName)
}
func (s *Openrc) IsExist(serviceName string) (bool, error) {
	_, err := os.Stat(filepath.Join("/etc/init.d", serviceName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat /etc/init.d/%s failed: %w", serviceName, err)
	}
	return true, nil
}
func (s *Openrc) Status(serviceName string) (string, error) {
	return run(s.toolCmd, serviceName, "status")
}

func (s *Openrc) Operate(operate, serviceName string) error {
	switch operate {
	case "enable":
		return handlerErr(run("rc-update", "add", serviceName, "default"))
	case "disable":
		return handlerErr(run("rc-update", "del", serviceName, "default"))
	default:
		return handlerErr(run(s.toolCmd, serviceName, operate))
	}
}

func (s *Openrc) Reload() error {
	return nil
}

func isSysvServiceEnabled(serviceName string) (bool, error) {
	entries, err := os.ReadDir("/etc")
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !strings.HasPrefix(name, "rc") || !strings.HasSuffix(name, ".d") {
			continue
		}
		items, err := os.ReadDir(filepath.Join("/etc", name))
		if err != nil {
			continue
		}
		for _, item := range items {
			itemName := item.Name()
			if strings.HasPrefix(itemName, "S") && strings.HasSuffix(itemName, serviceName) {
				return true, nil
			}
		}
	}
	return false, nil
}
