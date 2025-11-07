package client

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
)

const (
	// RulesFileName is the name of the rules persistence file
	RulesFileName = "1panel_basic.rules"
)

// SaveRulesToFile saves all rules from 1PANEL_BASIC chain to a text file
func (ipt *Iptables) SaveRulesToFile() error {
	rulesFile := path.Join(global.Dir.FirewallDir, RulesFileName)

	// Get all rules from 1PANEL_BASIC chain
	stdout, err := ipt.out(FilterTab, fmt.Sprintf("-S %s", Chain1PanelBasic))
	if err != nil {
		return fmt.Errorf("failed to list %s rules: %w", Chain1PanelBasic, err)
	}

	// Parse and save only the rule lines (skip chain creation lines)
	var rules []string
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Only save -A (append) rules, skip -N (new chain) and -P (policy) lines
		if strings.HasPrefix(line, fmt.Sprintf("-A %s", Chain1PanelBasic)) {
			rules = append(rules, line)
		}
	}

	// Write rules to file
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

	global.LOG.Infof("Successfully saved %d rules to %s", len(rules), rulesFile)
	return nil
}

// LoadRulesFromFile loads rules from file and applies them to 1PANEL_BASIC chain
// Uses complete override strategy: clears existing rules before loading
func (ipt *Iptables) LoadRulesFromFile() error {
	rulesFile := path.Join(global.Dir.FirewallDir, RulesFileName)

	// Check if rules file exists
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		global.LOG.Infof("Rules file %s does not exist, skipping restore", rulesFile)
		return nil
	}

	// Read rules from file
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
			continue // Skip empty lines and comments
		}
		rules = append(rules, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	// Clear existing rules in 1PANEL_BASIC chain (complete override strategy)
	if err := ipt.ClearChainRules(Chain1PanelBasic); err != nil {
		global.LOG.Warnf("Failed to clear existing rules from %s: %v", Chain1PanelBasic, err)
	}

	// Apply each rule
	appliedCount := 0
	for _, rule := range rules {
		// Convert -A to iptables command format
		// Example: "-A 1PANEL_BASIC -p tcp --dport 80 -j ACCEPT"
		// becomes: "1PANEL_BASIC -p tcp --dport 80 -j ACCEPT" (remove -A)
		if strings.HasPrefix(rule, fmt.Sprintf("-A %s", Chain1PanelBasic)) {
			// Remove "-A " prefix and execute
			ruleArgs := strings.TrimPrefix(rule, "-A ")
			if err := ipt.run(FilterTab, "-A "+ruleArgs); err != nil {
				global.LOG.Errorf("Failed to apply rule '%s': %v", rule, err)
				continue
			}
			appliedCount++
		}
	}

	global.LOG.Infof("Successfully loaded and applied %d/%d rules from %s", appliedCount, len(rules), rulesFile)
	return nil
}

// ClearChainRules clears all rules in the specified chain
func (ipt *Iptables) ClearChainRules(chainName string) error {
	// Use -F (flush) to clear all rules in the chain
	if err := ipt.run(FilterTab, fmt.Sprintf("-F %s", chainName)); err != nil {
		return fmt.Errorf("failed to flush chain %s: %w", chainName, err)
	}
	global.LOG.Infof("Successfully cleared all rules from chain %s", chainName)
	return nil
}
