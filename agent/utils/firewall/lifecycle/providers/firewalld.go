package providers

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

type Firewalld struct{}

func NewFirewalld() (*Firewalld, error) {
	return &Firewalld{}, nil
}

func (f *Firewalld) Name() string {
	return "firewalld"
}

func (f *Firewalld) Status() (bool, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--state")
	if err != nil {
		if firewalldStopped(stdout, err) {
			return false, nil
		}
		return false, fmt.Errorf("load firewall status failed: %w", err)
	}
	return strings.TrimSpace(stdout) == "running", nil
}

func firewalldStopped(stdout string, err error) bool {
	message := strings.ToLower(strings.TrimSpace(stdout))
	if err != nil {
		message += " " + strings.ToLower(err.Error())
	}
	return strings.Contains(message, "not running")
}

func (f *Firewalld) Version() (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--version")
	if err != nil {
		return "", fmt.Errorf("load the firewall version failed, %v", err)
	}
	return strings.ReplaceAll(stdout, "\n ", ""), nil
}

func (f *Firewalld) Start() error {
	if err := controller.HandleStart("firewalld"); err != nil {
		return fmt.Errorf("enable the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewalld) Stop() error {
	if err := controller.HandleStop("firewalld"); err != nil {
		return fmt.Errorf("stop the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewalld) Restart() error {
	if err := controller.HandleRestart("firewalld"); err != nil {
		return fmt.Errorf("restart the firewall failed, err: %v", err)
	}
	return nil
}
