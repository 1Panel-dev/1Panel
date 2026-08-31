package nftables_helper

import (
	"context"
	"fmt"
	"os"
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
	return RunScript(script)
}

func RunScript(script string) error {
	return RunScriptContext(context.Background(), script)
}

func RunScriptContext(ctx context.Context, script string) error {
	return runScriptFile(script, func(file string) error {
		return cmd.NewCommandMgr(
			cmd.WithContext(ctx),
			cmd.WithTimeout(60*time.Second),
		).RunWithOptionalSudo("nft", "-f", file)
	})
}

func runScriptFile(script string, run func(string) error) error {
	file, err := os.CreateTemp("", "1panel-nft-*.nft")
	if err != nil {
		return err
	}
	name := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(name)
	}()
	if _, err := file.WriteString(script); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return run(name)
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

func LoadInitStatus(tab string) (bool, bool, error) {
	if tab != "base" {
		return false, false, nil
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		initialized, bound, err := loadFamilyInitStatus(family)
		if err != nil || !initialized {
			return false, false, err
		}
		if !bound {
			return true, false, nil
		}
	}
	return true, true, nil
}

func LoadFamilyInitStatus(family filter.Family, tab string) (bool, bool, error) {
	if tab != "base" {
		return false, false, nil
	}
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return false, false, nil
	}
	return loadFamilyInitStatus(family)
}

func LoadFamilyBindStatus(family filter.Family) (bool, error) {
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return false, nil
	}
	stdout, err := run("list", "chain", TableFamily(family), TableName, InputChain)
	if err != nil {
		return false, err
	}
	return hasBaseChainBinding(stdout), nil
}

func hasBaseChainBinding(output string) bool {
	for _, chain := range BasicChains() {
		if strings.Contains(output, "jump "+chain) {
			return true
		}
	}
	return false
}

func loadFamilyInitStatus(family filter.Family) (bool, bool, error) {
	for _, chain := range BasicChains() {
		if _, err := run("list", "chain", TableFamily(family), TableName, chain); err != nil {
			return false, false, nil
		}
	}
	stdout, err := run("list", "chain", TableFamily(family), TableName, InputChain)
	if err != nil {
		return false, false, nil
	}
	for _, chain := range BasicChains() {
		if !strings.Contains(stdout, "jump "+chain) {
			return true, false, nil
		}
	}
	return true, true, nil
}
