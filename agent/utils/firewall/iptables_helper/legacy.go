package iptables_helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/mattn/go-shellwords"
)

const (
	LegacyInputChain     = "1PANEL_INPUT"
	LegacyOutputChain    = "1PANEL_OUTPUT"
	legacyInputFileName  = "1panel_input.rules"
	legacyOutputFileName = "1panel_out.rules"
)

func CleanupLegacyAdvancedChains(ctx context.Context) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	output, err := RunWithStdContext(ctx, FilterTab, "-S")
	if err != nil {
		return fmt.Errorf("inspect legacy iptables advanced chains: %w", err)
	}
	if script := buildLegacyAdvancedChainCleanupScript(output); script != "" {
		if err := restoreRules(commands.Restore4, script); err != nil {
			return fmt.Errorf("remove legacy iptables advanced chains: %w", err)
		}
	}
	for _, name := range []string{legacyInputFileName, legacyOutputFileName} {
		file := filepath.Join(global.Dir.FirewallDir, name)
		if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove legacy iptables rules file %s: %w", file, err)
		}
	}
	return nil
}

func buildLegacyAdvancedChainCleanupScript(output string) string {
	legacyChains := map[string]struct{}{LegacyInputChain: {}, LegacyOutputChain: {}}
	existing := make(map[string]bool, len(legacyChains))
	deletions := make([]string, 0)
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		fields, err := shellwords.Parse(line)
		if err != nil {
			continue
		}
		if len(fields) == 2 && fields[0] == "-N" {
			if _, legacy := legacyChains[fields[1]]; legacy {
				existing[fields[1]] = true
			}
			continue
		}
		if len(fields) < 4 || fields[0] != "-A" {
			continue
		}
		for index := 2; index+1 < len(fields); index++ {
			if fields[index] != "-j" && fields[index] != "-g" {
				continue
			}
			if _, legacy := legacyChains[fields[index+1]]; legacy {
				deletions = append(deletions, strings.Replace(line, "-A ", "-D ", 1))
			}
			break
		}
	}
	for _, chain := range []string{LegacyInputChain, LegacyOutputChain} {
		if existing[chain] {
			deletions = append(deletions, "-F "+chain, "-X "+chain)
		}
	}
	if len(deletions) == 0 {
		return ""
	}
	return "*filter\n" + strings.Join(deletions, "\n") + "\nCOMMIT\n"
}
