package providers

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Nftables struct{}

func NewNftables() (*Nftables, error) { return &Nftables{}, nil }

func (n *Nftables) Name() string { return "nftables" }

func (n *Nftables) Status() (bool, error) {
	_, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithOptionalSudoAndStdout("nft", "list", "tables")
	return err == nil, err
}

func (n *Nftables) Start() error   { return nil }
func (n *Nftables) Stop() error    { return nil }
func (n *Nftables) Restart() error { return nil }

func (n *Nftables) Version() (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("nft", "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get nftables version: %w", err)
	}
	fields := strings.Fields(strings.TrimSpace(stdout))
	if len(fields) >= 2 {
		return strings.TrimPrefix(fields[1], "v"), nil
	}
	return strings.TrimSpace(stdout), nil
}
