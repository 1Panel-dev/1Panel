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
	BasicFileName    = "1panel_basic.rules"
	InputFileName    = "1panel_input.rules"
	OutputFileName   = "1panel_out.rules"
	ForwardFileName  = "1panel_forward.rules"
	ForwardFileName1 = "1panel_forward_pre.rules"
	ForwardFileName2 = "1panel_forward_post.rules"
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
	rulesFile := path.Join(global.Dir.FirewallDir, fileName)
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(rulesFile)
	if err != nil {
		return fmt.Errorf("failed to open rules file: %w", err)
	}
	defer file.Close()

	var rules []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rules = append(rules, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	if err := ClearChain(tab, chain); err != nil {
		global.LOG.Warnf("Failed to clear existing rules from %s: %v", chain, err)
	}

	appliedCount := 0
	for _, rule := range rules {
		if strings.HasPrefix(rule, fmt.Sprintf("-A %s", chain)) {
			ruleArgs := strings.TrimPrefix(rule, "-A ")
			if err := Run(tab, "-A "+ruleArgs); err != nil {
				global.LOG.Errorf("Failed to apply rule '%s': %v", rule, err)
				continue
			}
			appliedCount++
		}
	}

	return nil
}
