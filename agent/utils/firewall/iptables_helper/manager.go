package iptables_helper

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/mattn/go-shellwords"
)

func Cleanup() error {
	if err := disableBase(); err != nil {
		return err
	}
	if err := cleanupBaseChains(false); err != nil {
		return err
	}
	if commands, err := lifecycle.ResolveIptablesCommands(); err == nil && commands.IPv6Available() {
		if err := cleanupBaseChains(true); err != nil && !errors.Is(err, filter.ErrFamilyUnavailable) {
			return err
		}
	}
	for _, file := range []string{BasicBeforeFileName, BasicFileName, BasicAfterFileName,
		IPv6FileName(BasicBeforeFileName), IPv6FileName(BasicFileName), IPv6FileName(BasicAfterFileName)} {
		if err := os.Remove(filepath.Join(global.Dir.FirewallDir, file)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func Operate(operation firewall.BaseOperation, requiredPorts []firewall.PortWhitelist) error {
	switch operation {
	case firewall.BaseOperationInit, firewall.BaseOperationBind:
		if _, err := lifecycle.ResolveIptablesCommands(); err != nil {
			return fmt.Errorf("failed to find iptables")
		}
		return enableBase(true, requiredPorts)
	case firewall.BaseOperationBindWithoutInit:
		return enableBase(false, requiredPorts)
	case firewall.BaseOperationUnbind:
		return disableBase()
	default:
		return fmt.Errorf("unsupported iptables base operation %q", operation)
	}
}

func enableBase(prepare bool, requiredPorts []firewall.PortWhitelist) error {
	if prepare {
		if err := ensureBaseChainsFamily(false); err != nil {
			return err
		}
		if err := applyRequiredFirewallPortWhiteListRules(requiredPorts, false, true, false); err != nil {
			return err
		}
		if err := saveBaseChainsFamily(false); err != nil {
			return err
		}
	}
	if err := setBaseChainBindings(false, true); err != nil {
		return err
	}
	if prepare {
		commands, err := lifecycle.ResolveIptablesCommands()
		if err != nil {
			return err
		}
		if commands.IPv6Available() {
			if err := EnsureIPv6BaseChains(requiredPorts); err != nil {
				return err
			}
		}
		if err := syncRequiredPorts(requiredPorts, true); err != nil {
			return err
		}
	} else if err := BindIPv6BaseChains(); err != nil {
		return err
	}
	return nil
}

func disableBase() error {
	if err := setBaseChainBindings(false, false); err != nil {
		return err
	}
	if err := UnbindIPv6BaseChains(); err != nil && !errors.Is(err, filter.ErrFamilyUnavailable) {
		return err
	}
	return nil
}

func ensureBaseChainsFamily(ipv6 bool) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	executable := commands.Restore4
	var output string
	if ipv6 {
		if !commands.IPv6Available() {
			return nil
		}
		executable = commands.Restore6
		output, err = RunIPv6WithStd(FilterTab, "-S")
	} else {
		output, err = RunWithStd(FilterTab, "-S")
	}
	if err != nil {
		return err
	}
	lines := make([]string, 0, len(BasicChains()))
	for _, chain := range BasicChains() {
		if !containsIptablesRule(output, "-N "+chain) {
			lines = append(lines, "-N "+chain)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	if err := restoreRules(executable, "*filter\n"+strings.Join(lines, "\n")+"\nCOMMIT\n"); err != nil {
		return fmt.Errorf("batch create base chains: %w", err)
	}
	return nil
}

func cleanupBaseChains(ipv6 bool) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	executable := commands.Restore4
	var output string
	if ipv6 {
		if !commands.IPv6Available() {
			return nil
		}
		executable = commands.Restore6
		output, err = RunIPv6WithStd(FilterTab, "-S")
	} else {
		output, err = RunWithStd(FilterTab, "-S")
	}
	if err != nil {
		return err
	}
	lines := make([]string, 0, len(BasicChains())*2)
	for _, chain := range []string{BasicAfterChain, BasicChain, BasicBeforeChain} {
		if containsIptablesRule(output, "-N "+chain) {
			lines = append(lines, "-F "+chain, "-X "+chain)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	if err := restoreRules(executable, "*filter\n"+strings.Join(lines, "\n")+"\nCOMMIT\n"); err != nil {
		return fmt.Errorf("batch delete base chains: %w", err)
	}
	return nil
}

func setBaseChainBindings(ipv6, bind bool) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	executable := commands.Restore4
	var output string
	if ipv6 {
		if !commands.IPv6Available() {
			return nil
		}
		executable = commands.Restore6
		output, err = RunIPv6WithStd(FilterTab, "-S", InputChain)
	} else {
		output, err = RunWithStd(FilterTab, "-S", InputChain)
	}
	if err != nil {
		return err
	}
	script := buildBaseChainBindingsRestoreScript(output, bind)
	if script == "" {
		return nil
	}
	if err := restoreRules(executable, script); err != nil {
		family := "IPv4"
		if ipv6 {
			family = "IPv6"
		}
		return fmt.Errorf("batch update %s base chain bindings: %w", family, err)
	}
	return nil
}

func buildBaseChainBindingsRestoreScript(output string, bind bool) string {
	lines := baseChainBindingCommands(output, bind)
	if len(lines) == 0 {
		return ""
	}
	return "*filter\n" + strings.Join(lines, "\n") + "\nCOMMIT\n"
}

func baseChainBindingCommands(output string, bind bool) []string {
	lines := make([]string, 0, len(BasicChains())*2)
	for _, chain := range BasicChains() {
		binding := "-A " + InputChain + " -j " + chain
		for _, current := range strings.Split(output, "\n") {
			if strings.TrimSpace(current) == binding {
				lines = append(lines, "-D "+InputChain+" -j "+chain)
			}
		}
	}
	if bind {
		for index, chain := range BasicChains() {
			lines = append(lines, fmt.Sprintf("-I %s %d -j %s", InputChain, index+1, chain))
		}
	}
	return lines
}

func saveBaseChainsFamily(ipv6 bool) error {
	read := RunWithStd
	if ipv6 {
		read = RunIPv6WithStd
	}
	output, err := read(FilterTab, "-S")
	if err != nil {
		return err
	}
	for _, item := range []struct{ chain, file string }{
		{BasicBeforeChain, BasicBeforeFileName},
		{BasicChain, BasicFileName},
		{BasicAfterChain, BasicAfterFileName},
	} {
		if !containsIptablesRule(output, "-N "+item.chain) {
			return fmt.Errorf("cannot save missing iptables chain %s", item.chain)
		}
		file := item.file
		if ipv6 {
			file = IPv6FileName(file)
		}
		if err := writeChainRules(output, item.chain, file); err != nil {
			return err
		}
	}
	return nil
}

func RestoreBaseChains(requiredPorts []firewall.PortWhitelist) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	if err := ensureBaseChainsFamily(false); err != nil {
		return err
	}
	input, err := buildBaseChainsRestoreScript(global.Dir.FirewallDir, false, requiredPorts...)
	if err != nil {
		return err
	}
	if err := restoreRules(commands.Restore4, input); err != nil {
		return fmt.Errorf("batch restore IPv4 base chains: %w", err)
	}
	if !commands.IPv6Available() {
		return nil
	}
	if err := ensureBaseChainsFamily(true); err != nil {
		return err
	}
	input, err = buildBaseChainsRestoreScript(global.Dir.FirewallDir, true, requiredPorts...)
	if err != nil {
		return err
	}
	if err := restoreRules(commands.Restore6, input); err != nil {
		return fmt.Errorf("batch restore IPv6 base chains: %w", err)
	}
	return nil
}

func buildBaseChainsRestoreScript(firewallDir string, ipv6 bool, requiredPorts ...firewall.PortWhitelist) (string, error) {
	var script strings.Builder
	script.WriteString("*filter\n")
	for _, chain := range BasicChains() {
		script.WriteString("-F ")
		script.WriteString(chain)
		script.WriteByte('\n')
	}
	for _, item := range []struct{ chain, file string }{
		{BasicBeforeChain, BasicBeforeFileName},
		{BasicChain, BasicFileName},
		{BasicAfterChain, BasicAfterFileName},
	} {
		fileName := item.file
		if ipv6 {
			fileName = IPv6FileName(fileName)
		}
		data, err := os.ReadFile(filepath.Join(firewallDir, fileName))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		prefix := "-A " + item.chain + " "
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, prefix) || strings.ContainsAny(line, "\r\n") {
				continue
			}
			script.WriteString(line)
			script.WriteByte('\n')
		}
	}
	family := constant.FirewallFamilyIPv4
	if ipv6 {
		family = constant.FirewallFamilyIPv6
	}
	defaults, err := baseDefaultRules(requiredPorts, family)
	if err != nil {
		return "", err
	}
	for _, rule := range defaults {
		if !containsIptablesRule(script.String(), rule) {
			if strings.HasPrefix(rule, "-A "+BasicBeforeChain+" ") && strings.Contains(rule, " --dport ") {
				rule = strings.Replace(rule, "-A "+BasicBeforeChain+" ", "-I "+BasicBeforeChain+" 1 ", 1)
			}
			script.WriteString(rule + "\n")
		}
	}
	script.WriteString("COMMIT\n")
	return script.String(), nil
}

func syncRequiredPorts(requiredPorts []firewall.PortWhitelist, withSave bool) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
		ipv6 := family == constant.FirewallFamilyIPv6
		if ipv6 && !commands.IPv6Available() {
			continue
		}
		checkChain := CheckChainExist
		if ipv6 {
			checkChain = CheckIPv6ChainExist
		}
		initialized, err := checkChain(FilterTab, BasicBeforeChain)
		if ipv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		if !initialized {
			continue
		}
		if err := applyRequiredFirewallPortWhiteListRules(requiredPorts, withSave, false, ipv6); err != nil {
			return err
		}
	}
	return nil
}

func applyRequiredFirewallPortWhiteListRules(portWhiteList []firewall.PortWhitelist, withSave, includeDefaults, ipv6 bool) error {
	ports, err := firewall.NormalizeRequiredPorts(portWhiteList)
	if err != nil {
		return err
	}
	rules := firewall.ExpandPortWhitelist(ports)
	run := RunWithStd
	save := SaveRulesToFile
	beforeFile, afterFile := BasicBeforeFileName, BasicAfterFileName
	if ipv6 {
		run = RunIPv6WithStd
		save = SaveIPv6RulesToFile
		beforeFile, afterFile = IPv6FileName(beforeFile), IPv6FileName(afterFile)
	}
	family := constant.FirewallFamilyIPv4
	if ipv6 {
		family = constant.FirewallFamilyIPv6
	}
	beforeRaw, err := run(FilterTab, "-S", BasicBeforeChain)
	if err != nil {
		return err
	}
	afterRaw, err := run(FilterTab, "-S", BasicAfterChain)
	if err != nil {
		return err
	}
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	restore := commands.Restore4
	if ipv6 {
		restore = commands.Restore6
	}
	script := buildRequiredPortsRestoreScript(rules, family, beforeRaw, afterRaw, includeDefaults)
	if script != "" {
		if err := restoreRules(restore, script); err != nil {
			return fmt.Errorf("batch sync required firewall ports with %s: %w", restore, err)
		}
	}
	if !withSave {
		return nil
	}
	if err := save(FilterTab, BasicBeforeChain, beforeFile); err != nil {
		return err
	}
	return save(FilterTab, BasicAfterChain, afterFile)
}

func buildRequiredPortsRestoreScript(desired []firewall.SystemPort, family string, beforeRaw, afterRaw string, includeDefaults bool) string {
	var commands []string
	for _, line := range []string{"-A " + BasicBeforeChain + " " + IoRuleIn, "-A " + BasicBeforeChain + " " + EstablishedRule} {
		if !containsIptablesRule(beforeRaw, line) {
			commands = append(commands, line)
		}
	}
	for _, rule := range desired {
		line := iptablesSystemPortRuleLine(rule)
		if rule.Family == family && !containsIptablesRule(beforeRaw, line) {
			commands = append(commands, strings.Replace(line, "-A "+BasicBeforeChain+" ", "-I "+BasicBeforeChain+" 1 ", 1))
			beforeRaw += "\n" + line
		}
	}
	if includeDefaults {
		for _, rule := range []string{DropAllTcp, DropAllUdp} {
			line := "-A " + BasicAfterChain + " " + rule
			if !containsIptablesRule(afterRaw, line) {
				commands = append(commands, line)
			}
		}
	}
	if len(commands) == 0 {
		return ""
	}
	return "*filter\n" + strings.Join(commands, "\n") + "\nCOMMIT\n"
}

func iptablesSystemPortRuleLine(rule firewall.SystemPort) string {
	parts := []string{"-A", BasicBeforeChain}
	if rule.SourceAddress != "" {
		parts = append(parts, "-s", rule.SourceAddress)
	}
	return strings.Join(append(parts, "-p", rule.Protocol, "-m", rule.Protocol, "--dport", rule.Port, "-j", "ACCEPT"), " ")
}

func containsIptablesRule(output, rule string) bool {
	return countIptablesRule(output, rule) > 0
}

func countIptablesRule(output, rule string) int {
	canonical := func(value string) string {
		fields, err := shellwords.Parse(value)
		if err != nil || len(fields)%2 != 0 {
			return strings.TrimSpace(value)
		}
		var options []string
		for index := 0; index < len(fields); index += 2 {
			key, value := fields[index], fields[index+1]
			if key == "--comment" || key == "-m" && (value == "comment" || value == "tcp" || value == "udp") {
				continue
			}
			if prefix, err := netip.ParsePrefix(value); err == nil {
				prefix = prefix.Masked()
				value = prefix.String()
				if prefix.Bits() == prefix.Addr().BitLen() {
					value = prefix.Addr().String()
				}
			}
			options = append(options, key+" "+value)
		}
		sort.Strings(options)
		return strings.Join(options, " ")
	}
	rule = canonical(rule)
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if canonical(line) == rule {
			count++
		}
	}
	return count
}
