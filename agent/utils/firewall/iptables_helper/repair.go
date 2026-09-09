package iptables_helper

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func (m *Manager) RepairBaseChains() error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
		ipv6 := family == constant.FirewallFamilyIPv6
		run, executable := RunWithStd, commands.Restore4
		if ipv6 {
			if !commands.IPv6Available() {
				continue
			}
			run, executable = RunIPv6WithStd, commands.Restore6
		}
		output, err := run(FilterTab, "-S")
		if ipv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		script, err := buildBaseChainsRepairScript(global.Dir.FirewallDir, output, ipv6, func() ([]string, error) {
			ports, err := m.loadRequiredPorts()
			if err != nil {
				return nil, err
			}
			return baseDefaultRules(m.panelPort(), ports, family)
		})
		if err != nil {
			return fmt.Errorf("prepare %s base chain repair: %w", family, err)
		}
		if script != "" {
			if err := restoreRules(executable, script); err != nil {
				return fmt.Errorf("repair %s base chains: %w", family, err)
			}
		}
	}
	return nil
}

func buildBaseChainsRepairScript(dir, output string, ipv6 bool, loadDefaults func() ([]string, error)) (string, error) {
	var chains, rules, defaults []string
	bound := true
	for _, item := range []struct{ chain, file string }{
		{BasicBeforeChain, BasicBeforeFileName},
		{BasicChain, BasicFileName},
		{BasicAfterChain, BasicAfterFileName},
	} {
		bound = bound && containsIptablesRule(output, "-A "+InputChain+" -j "+item.chain)
		if containsIptablesRule(output, "-N "+item.chain) {
			continue
		}
		chains = append(chains, "-N "+item.chain)
		fileName := item.file
		if ipv6 {
			fileName = IPv6FileName(fileName)
		}
		data, err := os.ReadFile(filepath.Join(dir, fileName))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		lines := strings.Split(string(data), "\n")
		if errors.Is(err, os.ErrNotExist) && item.chain != BasicChain {
			if defaults == nil {
				defaults, err = loadDefaults()
				if err != nil {
					return "", err
				}
			}
			lines = defaults
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "-A "+item.chain+" ") && !strings.ContainsAny(line, "\r\n") {
				rules = append(rules, line)
			}
		}
	}
	lines := append(chains, rules...)
	if !bound {
		lines = append(lines, baseChainBindingCommands(output, true)...)
	}
	if len(lines) == 0 {
		return "", nil
	}
	return "*filter\n" + strings.Join(lines, "\n") + "\nCOMMIT\n", nil
}
