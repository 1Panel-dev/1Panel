package iptables_helper

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

type Manager struct {
	UpdateSetting     func(key, value string) error
	PanelPort         func() string
	LoadRequiredPorts func() ([]firewall.PortWhitelist, error)
}

func (m *Manager) Cleanup() error {
	if err := m.disableBase(); err != nil {
		return err
	}
	if err := cleanupBaseChains(false); err != nil {
		return err
	}
	if commands, err := lifecycle.ResolveIptablesCommands(); err == nil && commands.IPv6Available() {
		if err := cleanupBaseChains(true); err != nil {
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

func (m *Manager) Operate(operation firewall.BaseOperation) error {
	switch operation {
	case firewall.BaseOperationInit, firewall.BaseOperationBind:
		if _, err := lifecycle.ResolveIptablesCommands(); err != nil {
			return fmt.Errorf("failed to find iptables")
		}
		return m.enableBase(true)
	case firewall.BaseOperationBindWithoutInit:
		return m.enableBase(false)
	case firewall.BaseOperationUnbind:
		return m.disableBase()
	default:
		return fmt.Errorf("unsupported iptables base operation %q", operation)
	}
}

func (m *Manager) enableBase(prepare bool) error {
	if prepare {
		if err := ensureBaseChains(); err != nil {
			return err
		}
		if err := m.initPreRules(); err != nil {
			return err
		}
		if err := saveBaseChains(); err != nil {
			return err
		}
	}
	if err := bindBaseChains(); err != nil {
		return err
	}
	if prepare {
		if err := m.ensureIPv6BaseChains(); err != nil {
			return err
		}
	} else if err := BindIPv6BaseChains(); err != nil {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusEnable)
}

func (m *Manager) disableBase() error {
	if err := setBaseChainBindings(false, false); err != nil {
		return err
	}
	if err := UnbindIPv6BaseChains(); err != nil {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusDisable)
}

func ensureBaseChains() error {
	return ensureBaseChainsFamily(false)
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

func bindBaseChains() error {
	return setBaseChainBindings(false, true)
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
	if len(lines) == 0 {
		return ""
	}
	return "*filter\n" + strings.Join(lines, "\n") + "\nCOMMIT\n"
}

func saveBaseChains() error {
	for _, item := range []struct{ chain, file string }{
		{BasicBeforeChain, BasicBeforeFileName},
		{BasicChain, BasicFileName},
		{BasicAfterChain, BasicAfterFileName},
	} {
		if err := SaveRulesToFile(FilterTab, item.chain, item.file); err != nil {
			return err
		}
	}
	return nil
}

func RestoreBaseChains(panelPort string) error {
	port, err := strconv.Atoi(panelPort)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid panel port %q", panelPort)
	}
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	if err := ensureBaseChains(); err != nil {
		return err
	}
	input, err := buildBaseChainsRestoreScript(global.Dir.FirewallDir, panelPort, false)
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
	input, err = buildBaseChainsRestoreScript(global.Dir.FirewallDir, panelPort, true)
	if err != nil {
		return err
	}
	if err := restoreRules(commands.Restore6, input); err != nil {
		return fmt.Errorf("batch restore IPv6 base chains: %w", err)
	}
	return nil
}

func buildBaseChainsRestoreScript(firewallDir, panelPort string, ipv6 bool) (string, error) {
	var script strings.Builder
	script.WriteString("*filter\n")
	for _, chain := range BasicChains() {
		script.WriteString("-F ")
		script.WriteString(chain)
		script.WriteByte('\n')
	}
	panelRule := "-A " + BasicBeforeChain + " -p tcp -m tcp --dport " + panelPort + " -j ACCEPT"
	panelRuleFound := false
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
			if line == panelRule {
				panelRuleFound = true
			}
			script.WriteString(line)
			script.WriteByte('\n')
		}
	}
	if !panelRuleFound {
		script.WriteString(panelRule)
		script.WriteByte('\n')
	}
	script.WriteString("COMMIT\n")
	return script.String(), nil
}

func (m *Manager) initPreRules() error {
	requiredPorts, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	return applyRequiredFirewallPortWhiteListRules(requiredPorts, false, true)
}

func (m *Manager) ensureIPv6BaseChains() error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return nil
	}
	return EnsureIPv6BaseChains(m.panelPort())
}

func (m *Manager) SyncRequiredPorts(withSave bool) error {
	requiredPorts, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	return applyRequiredFirewallPortWhiteListRules(requiredPorts, withSave, false)
}

func applyRequiredFirewallPortWhiteListRules(portWhiteList []firewall.PortWhitelist, withSave, includeDefaults bool) error {
	portWhiteList, err := validateRequiredFirewallPorts(portWhiteList)
	if err != nil {
		return err
	}
	beforeRules, err := ReadFilterRulesByChain(BasicBeforeChain)
	if err != nil {
		return err
	}
	afterRules, err := ReadFilterRulesByChain(BasicAfterChain)
	if err != nil {
		return err
	}
	beforeRaw, err := RunWithStd(FilterTab, "-S", BasicBeforeChain)
	if err != nil {
		return err
	}
	afterRaw, err := RunWithStd(FilterTab, "-S", BasicAfterChain)
	if err != nil {
		return err
	}
	script := buildRequiredPortsRestoreScript(portWhiteList, beforeRules, afterRules, beforeRaw, afterRaw, includeDefaults)
	if script != "" {
		commands, resolveErr := lifecycle.ResolveIptablesCommands()
		if resolveErr != nil {
			return resolveErr
		}
		if err := restoreRules(commands.Restore4, script); err != nil {
			return fmt.Errorf("batch sync required IPv4 firewall ports: %w", err)
		}
	}
	if !withSave {
		return nil
	}
	if err := SaveRulesToFile(FilterTab, BasicBeforeChain, BasicBeforeFileName); err != nil {
		return err
	}
	return SaveRulesToFile(FilterTab, BasicAfterChain, BasicAfterFileName)
}

func validateRequiredFirewallPorts(ports []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
	result := make([]firewall.PortWhitelist, 0, len(ports))
	for _, port := range ports {
		port.Protocol = strings.ToLower(strings.TrimSpace(port.Protocol))
		if port.Protocol != "tcp" && port.Protocol != "udp" {
			return nil, fmt.Errorf("unsupported required firewall port protocol %q", port.Protocol)
		}
		portNumber, err := strconv.Atoi(strings.TrimSpace(port.Port))
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("invalid required firewall port %q", port.Port)
		}
		port.Port = strconv.Itoa(portNumber)
		result = append(result, port)
	}
	return firewall.NormalizePortWhitelist(result), nil
}

func buildRequiredPortsRestoreScript(
	desired []firewall.PortWhitelist,
	beforeRules, afterRules []FilterRules,
	beforeRaw, afterRaw string,
	includeDefaults bool,
) string {
	desiredKeys := firewall.PortWhitelistMap(desired)
	kept := make(map[string]struct{}, len(desired))
	commands := make([]string, 0)
	for _, rule := range beforeRules {
		if !simpleAcceptedPortRule(rule) {
			continue
		}
		key := firewall.PortWhitelistKey(firewall.PortWhitelist{Protocol: rule.Protocol, Port: rule.DstPort})
		if _, wanted := desiredKeys[key]; wanted {
			if _, alreadyKept := kept[key]; !alreadyKept {
				kept[key] = struct{}{}
				continue
			}
		}
		commands = append(commands, iptablesPortRuleLine("-D", BasicBeforeChain, rule.Protocol, rule.DstPort))
	}
	for _, rule := range afterRules {
		if simpleAcceptedPortRule(rule) && rule.Protocol == "udp" {
			commands = append(commands, iptablesPortRuleLine("-D", BasicAfterChain, rule.Protocol, rule.DstPort))
		}
	}

	if includeDefaults {
		for _, rule := range []string{
			"-A " + BasicBeforeChain + " " + IoRuleIn,
			"-A " + BasicBeforeChain + " " + EstablishedRule,
		} {
			count := countIptablesRule(beforeRaw, rule)
			for duplicate := 1; duplicate < count; duplicate++ {
				commands = append(commands, strings.Replace(rule, "-A ", "-D ", 1))
			}
			if count == 0 {
				commands = append(commands, rule)
			}
		}
	}
	for _, port := range desired {
		if _, exists := kept[firewall.PortWhitelistKey(port)]; exists {
			continue
		}
		commands = append(commands, iptablesPortRuleLine("-A", BasicBeforeChain, port.Protocol, port.Port))
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

func simpleAcceptedPortRule(rule FilterRules) bool {
	return rule.Strategy == "accept" && (rule.Protocol == "tcp" || rule.Protocol == "udp") && rule.DstPort != "" &&
		rule.SrcIP == "" && rule.DstIP == "" && rule.SrcPort == ""
}

func iptablesPortRuleLine(operation, chain, protocol, port string) string {
	return strings.Join([]string{operation, chain, "-p", protocol, "-m", protocol, "--dport", port, "-j", "ACCEPT"}, " ")
}

func containsIptablesRule(output, rule string) bool {
	return countIptablesRule(output, rule) > 0
}

func countIptablesRule(output, rule string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == rule {
			count++
		}
	}
	return count
}

func (m *Manager) updateSetting(key, value string) error {
	if m != nil && m.UpdateSetting != nil {
		return m.UpdateSetting(key, value)
	}
	return nil
}

func (m *Manager) panelPort() string {
	if m != nil && m.PanelPort != nil {
		return m.PanelPort()
	}
	return ""
}

func (m *Manager) loadRequiredPorts() ([]firewall.PortWhitelist, error) {
	if m != nil && m.LoadRequiredPorts != nil {
		return m.LoadRequiredPorts()
	}
	return nil, fmt.Errorf("load required firewall ports is not configured")
}
