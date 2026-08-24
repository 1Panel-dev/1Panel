package providers

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Iptables struct{ executable string }

func NewIptables(executable string) (*Iptables, error) {
	if executable == "" {
		executable = "iptables"
	}
	return &Iptables{executable: executable}, nil
}

func (i *Iptables) Name() string {
	return "iptables"
}

func (i *Iptables) Status() (bool, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout(i.executable, "-L", "-n")
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
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout(i.executable, "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get iptables version: %w", err)
	}
	parts := strings.Fields(stdout)
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "v"), nil
	}
	return strings.TrimSpace(stdout), nil
}
