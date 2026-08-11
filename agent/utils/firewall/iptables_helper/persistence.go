package iptables_helper

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	BasicBeforeFileName = "1panel_basic_before.rules"
	BasicFileName       = "1panel_basic.rules"
	BasicAfterFileName  = "1panel_basic_after.rules"
)

func SaveRulesToFile(tab, chain, fileName string) error {
	return saveRulesToFile("iptables", tab, chain, fileName)
}

func SaveIPv6RulesToFile(tab, chain, fileName string) error {
	return saveRulesToFile("ip6tables", tab, chain, fileName)
}

func IPv6FileName(fileName string) string {
	return "ipv6_" + fileName
}

func saveRulesToFile(executable, tab, chain, fileName string) error {
	rulesFile := path.Join(global.Dir.FirewallDir, fileName)

	var stdout string
	var err error
	if executable == "ip6tables" {
		stdout, err = RunIPv6WithStd(tab, "-S", chain)
	} else {
		stdout, err = RunWithStd(tab, "-S", chain)
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
	return loadRulesFromFile("iptables", "iptables-restore", tab, chain, fileName)
}

func LoadIPv6RulesFromFile(tab, chain, fileName string) error {
	return loadRulesFromFile("ip6tables", "ip6tables-restore", tab, chain, fileName)
}

func loadRulesFromFile(executable, restoreExecutable, tab, chain, fileName string) error {
	var err error
	if executable == "ip6tables" {
		err = AddIPv6Chain(tab, chain)
	} else {
		err = AddChain(tab, chain)
	}
	if err != nil {
		global.LOG.Errorf("create chain %s failed: %v", chain, err)
		return err
	}
	rulesFile := path.Join(global.Dir.FirewallDir, fileName)
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(rulesFile)
	if err != nil {
		global.LOG.Errorf("read rules from file %s failed, err: %v", rulesFile, err)
		return err
	}
	rules := strings.Split(string(data), "\n")
	if executable == "ip6tables" {
		err = RunIPv6(tab, "-F", chain)
	} else {
		err = ClearChain(tab, chain)
	}
	if err != nil {
		return fmt.Errorf("clear existing rules from %s: %w", chain, err)
	}

	var restoreErrors []error
	for _, rule := range rules {
		if strings.HasPrefix(rule, fmt.Sprintf("-A %s", chain)) {
			if err := restoreRule(restoreExecutable, tab, rule); err != nil {
				restoreErrors = append(restoreErrors, fmt.Errorf("apply rule %q: %w", rule, err))
			}
		}
	}

	return errors.Join(restoreErrors...)
}

func restoreRule(executable, tab, rule string) error {
	restoreInput := fmt.Sprintf("*%s\n%s\nCOMMIT\n", tab, rule)
	commandName, commandArgs := cmd.WrapWithOptionalSudo(executable, "-n")
	_, err := cmd.NewCommandMgr().RunPipe(cmd.PipeCommand{
		Name:  commandName,
		Args:  commandArgs,
		Stdin: bytes.NewReader([]byte(restoreInput)),
	})
	return err
}
