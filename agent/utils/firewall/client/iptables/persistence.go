package iptables

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
)

const (
	BasicBeforeFileName = "1panel_basic_before.rules"
	BasicFileName       = "1panel_basic.rules"
	BasicAfterFileName  = "1panel_basic_after.rules"
	InputFileName       = "1panel_input.rules"
	OutputFileName      = "1panel_out.rules"
	ForwardFileName     = "1panel_forward.rules"
	ForwardFileName1    = "1panel_forward_pre.rules"
	ForwardFileName2    = "1panel_forward_post.rules"
)

func SaveRulesToFile(tab, chain, fileName string) error {
	rulesFile := path.Join(global.Dir.FirewallDir, fileName)

	stdout, err := RunWithStd(tab, fmt.Sprintf("-S %s", chain))
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

	file, err := os.Create(rulesFile)
	if err != nil {
		return fmt.Errorf("failed to create rules file: %w", err)
	}
	defer file.Close()

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

	global.LOG.Infof("persistence rules to %s successful", rulesFile)
	return nil
}

func LoadRulesFromFile(tab, chain, fileName string) error {
	if err := AddChain(tab, chain); err != nil {
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
	if err := ClearChain(tab, chain); err != nil {
		global.LOG.Warnf("clear existing rules from %s failed, err: %v", chain, err)
	}

	for _, rule := range rules {
		if strings.HasPrefix(rule, fmt.Sprintf("-A %s", chain)) {
			if err := Run(tab, rule); err != nil {
				global.LOG.Errorf("apply rule '%s' failed, err: %v", rule, err)
			}
		}
	}

	return nil
}
