package iptables_helper

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

const (
	BasicBeforeFileName = "1panel_basic_before.rules"
	BasicFileName       = "1panel_basic.rules"
	BasicAfterFileName  = "1panel_basic_after.rules"
)

func SaveRulesToFile(tab, chain, fileName string) error {
	return SaveRulesToFileContext(context.Background(), tab, chain, fileName)
}

func SaveRulesToFileContext(ctx context.Context, tab, chain, fileName string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	return saveRulesToFile(ctx, commands.IPv4, tab, chain, fileName)
}

func SaveIPv6RulesToFile(tab, chain, fileName string) error {
	return SaveIPv6RulesToFileContext(context.Background(), tab, chain, fileName)
}

func SaveIPv6RulesToFileContext(ctx context.Context, tab, chain, fileName string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	if !commands.IPv6Available() {
		return fmt.Errorf("ip6tables command family is unavailable")
	}
	return saveRulesToFile(ctx, commands.IPv6, tab, chain, fileName)
}

func IPv6FileName(fileName string) string {
	return "ipv6_" + fileName
}

func saveRulesToFile(ctx context.Context, executable, tab, chain, fileName string) error {
	rulesFile := path.Join(global.Dir.FirewallDir, fileName)

	var stdout string
	var err error
	if strings.HasPrefix(path.Base(executable), "ip6tables") {
		stdout, err = RunIPv6WithStdContext(ctx, tab, "-S", chain)
	} else {
		stdout, err = RunWithStdContext(ctx, tab, "-S", chain)
	}
	if err != nil {
		return fmt.Errorf("failed to list %s rules: %w", chain, err)
	}
	var rules []string
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, fmt.Sprintf("-A %s", chain)) {
			rules = append(rules, line)
		}
	}

	mode := os.FileMode(0644)
	if info, statErr := os.Stat(rulesFile); statErr == nil {
		mode = info.Mode().Perm()
	}
	file, err := os.CreateTemp(global.Dir.FirewallDir, "."+fileName+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary rules file: %w", err)
	}
	temporaryFile := file.Name()
	committed := false
	defer func() {
		_ = file.Close()
		if !committed {
			_ = os.Remove(temporaryFile)
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return fmt.Errorf("failed to set rules file permissions: %w", err)
	}

	writer := bufio.NewWriter(file)
	for _, rule := range rules {
		_, err := writer.WriteString(rule + "\n")
		if err != nil {
			return fmt.Errorf("failed to write rule to file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush rules to file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync rules file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close rules file: %w", err)
	}
	if err := os.Rename(temporaryFile, rulesFile); err != nil {
		return fmt.Errorf("failed to replace rules file: %w", err)
	}
	committed = true

	global.LOG.Infof("persistence rules to %s successful", rulesFile)
	return nil
}

func LoadRulesFromFile(tab, chain, fileName string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	return loadRulesFromFile(commands.IPv4, commands.Restore4, tab, chain, fileName)
}

func LoadIPv6RulesFromFile(tab, chain, fileName string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	if !commands.IPv6Available() {
		return fmt.Errorf("ip6tables command family is unavailable")
	}
	return loadRulesFromFile(commands.IPv6, commands.Restore6, tab, chain, fileName)
}

func loadRulesFromFile(executable, restoreExecutable, tab, chain, fileName string) error {
	var exists bool
	var err error
	if strings.HasPrefix(path.Base(executable), "ip6tables") {
		exists, err = CheckIPv6ChainExist(tab, chain)
	} else {
		exists, err = CheckChainExist(tab, chain)
	}
	if err != nil {
		global.LOG.Errorf("inspect chain %s failed: %v", chain, err)
		return err
	}
	rulesFile := path.Join(global.Dir.FirewallDir, fileName)
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		if exists {
			return nil
		}
		return restoreRules(restoreExecutable, fmt.Sprintf("*%s\n-N %s\nCOMMIT\n", tab, chain))
	}
	data, err := os.ReadFile(rulesFile)
	if err != nil {
		global.LOG.Errorf("read rules from file %s failed, err: %v", rulesFile, err)
		return err
	}
	rules := strings.Split(string(data), "\n")
	var restoreInput strings.Builder
	restoreInput.WriteByte('*')
	restoreInput.WriteString(tab)
	restoreInput.WriteByte('\n')
	if !exists {
		restoreInput.WriteString("-N ")
		restoreInput.WriteString(chain)
		restoreInput.WriteByte('\n')
	}
	restoreInput.WriteString("-F ")
	restoreInput.WriteString(chain)
	restoreInput.WriteByte('\n')
	for _, rule := range rules {
		if strings.HasPrefix(rule, fmt.Sprintf("-A %s", chain)) {
			restoreInput.WriteString(rule)
			restoreInput.WriteByte('\n')
		}
	}
	restoreInput.WriteString("COMMIT\n")
	if err := restoreRules(restoreExecutable, restoreInput.String()); err != nil {
		return fmt.Errorf("batch restore rules for %s: %w", chain, err)
	}
	return nil
}

func restoreRules(executable, input string) error {
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(input)))
	return manager.RunWithOptionalSudo(executable, "--noflush", "--wait")
}
