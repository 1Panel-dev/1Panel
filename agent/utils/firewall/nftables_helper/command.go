package nftables_helper

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const (
	TableName        = "nft_1panel_filter"
	InputChain       = "NFT_1PANEL_INPUT"
	BasicBeforeChain = "NFT_1PANEL_BASIC_BEFORE"
	BasicChain       = "NFT_1PANEL_BASIC"
	BasicAfterChain  = "NFT_1PANEL_BASIC_AFTER"
)

func TableFamily(family filter.Family) string {
	if family == filter.FamilyIPv6 {
		return "ip6"
	}
	return "ip"
}

func BasicChains() []string {
	return []string{BasicBeforeChain, BasicChain, BasicAfterChain}
}

func run(args ...string) (string, error) {
	return cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout("nft", args...)
}

func runCommand(args ...string) error {
	return cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudo("nft", args...)
}

func runBatch(commands ...[]string) error {
	script, err := buildBatchScript(commands...)
	if err != nil || script == "" {
		return err
	}
	manager := cmd.NewCommandMgr(
		cmd.WithTimeout(60*time.Second),
		cmd.WithStdin(strings.NewReader(script)),
	)
	return manager.RunWithOptionalSudo("nft", "-f", "-")
}

func buildBatchScript(commands ...[]string) (string, error) {
	var script strings.Builder
	for _, command := range commands {
		if len(command) == 0 {
			continue
		}
		for _, token := range command {
			if strings.ContainsAny(token, "\r\n") {
				return "", fmt.Errorf("invalid newline in nftables batch command")
			}
		}
		script.WriteString(strings.Join(command, " "))
		script.WriteByte('\n')
	}
	return script.String(), nil
}
