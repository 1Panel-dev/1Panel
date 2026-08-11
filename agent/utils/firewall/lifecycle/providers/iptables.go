package providers

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Iptables struct{}

func NewIptables() (*Iptables, error) {
	return &Iptables{}, nil
}

func (i *Iptables) Name() string {
	return "iptables"
}

func (i *Iptables) Status() (bool, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("iptables", "-L", "-n")
	if err != nil {
		return false, err
	}
	firstLine := strings.Split(strings.TrimSpace(stdout), "\n")[0]
	return strings.Contains(firstLine, "Chain"), nil
}

func (i *Iptables) Start() error {
	return nil
}

func (i *Iptables) Stop() error {
	return nil
}

func (i *Iptables) Restart() error {
	return nil
}

func (i *Iptables) Version() (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("iptables", "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get iptables version: %w", err)
	}
	parts := strings.Fields(stdout)
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "v"), nil
	}
	return strings.TrimSpace(stdout), nil
}
