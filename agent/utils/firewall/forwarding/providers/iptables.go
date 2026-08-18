package providers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

type iptablesBackend interface {
	IPv6Available() bool
	Run(table string, args ...string) error
	RunWithStd(table string, args ...string) (string, error)
	RunIPv6(table string, args ...string) error
	RunIPv6WithStd(table string, args ...string) (string, error)
	Restore(family, input string) error
	LoadRulesFromFile(table, chain, fileName string) error
	LoadIPv6RulesFromFile(table, chain, fileName string) error
}

type systemIptablesBackend struct{}

func (systemIptablesBackend) IPv6Available() bool {
	commands, err := lifecycle.ResolveIptablesCommands()
	return err == nil && commands.IPv6Available()
}

func (systemIptablesBackend) Run(table string, args ...string) error {
	return iptables_helper.Run(table, args...)
}

func (systemIptablesBackend) RunWithStd(table string, args ...string) (string, error) {
	return iptables_helper.RunWithStd(table, args...)
}

func (systemIptablesBackend) RunIPv6(table string, args ...string) error {
	return iptables_helper.RunIPv6(table, args...)
}

func (systemIptablesBackend) RunIPv6WithStd(table string, args ...string) (string, error) {
	return iptables_helper.RunIPv6WithStd(table, args...)
}

func (systemIptablesBackend) Restore(family, input string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	executable := commands.Restore4
	if family == forwarding.FamilyIPv6 {
		executable = commands.Restore6
		if executable == "" {
			return fmt.Errorf("ip6tables-restore command family is unavailable")
		}
	}
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(input)))
	return manager.RunWithOptionalSudo(executable, "--noflush", "--wait")
}

func (systemIptablesBackend) LoadRulesFromFile(table, chain, fileName string) error {
	return iptables_helper.LoadRulesFromFile(table, chain, fileName)
}

func (systemIptablesBackend) LoadIPv6RulesFromFile(table, chain, fileName string) error {
	return iptables_helper.LoadIPv6RulesFromFile(table, chain, fileName)
}

type forwardingSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	RunWithOptionalSudo(name string, args ...string) error
}

type defaultForwardingSystem struct{}

func (defaultForwardingSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (defaultForwardingSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return cmd.WriteFileWithOptionalSudo(name, data, perm)
}

func (defaultForwardingSystem) RunWithOptionalSudo(name string, args ...string) error {
	return cmd.NewCommandMgr().RunWithOptionalSudo(name, args...)
}

type iptablesNATAdapter struct {
	provider string
	backend  iptablesBackend
	system   forwardingSystem
}

func newIptablesNATAdapter(provider string) *iptablesNATAdapter {
	return &iptablesNATAdapter{
		provider: provider,
		backend:  systemIptablesBackend{},
		system:   defaultForwardingSystem{},
	}
}

func (l *iptablesNATAdapter) Name() string {
	return l.provider
}

func (l *iptablesNATAdapter) List() ([]forwarding.Rule, error) {
	stdout, err := l.backend.RunWithStd(iptables_helper.NatTab, "-nvL", forwarding.ChainPreRouting, "--line-numbers")
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT rules: %w", err)
	}
	rules := parseIptablesRules(stdout, forwarding.FamilyIPv4)
	if !l.backend.IPv6Available() {
		return rules, nil
	}
	stdout, err = l.backend.RunIPv6WithStd(iptables_helper.NatTab, "-nvL", forwarding.ChainPreRouting, "--line-numbers")
	if err != nil {
		return nil, fmt.Errorf("failed to list IPv6 NAT rules: %w", err)
	}
	return append(rules, parseIptablesRules(stdout, forwarding.FamilyIPv6)...), nil
}

func (l *iptablesNATAdapter) Reconcile(rules []forwarding.Rule) error {
	byFamily := map[string][]forwarding.Rule{
		forwarding.FamilyIPv4: nil,
		forwarding.FamilyIPv6: nil,
	}
	for _, rule := range rules {
		normalized, err := NormalizeRule(rule)
		if err != nil {
			return err
		}
		if normalized.Family == forwarding.FamilyIPv6 && !l.backend.IPv6Available() {
			return fmt.Errorf("ip6tables command family is unavailable")
		}
		byFamily[normalized.Family] = append(byFamily[normalized.Family], normalized)
	}
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		if family == forwarding.FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		script, err := buildIptablesForwardRestoreScript(byFamily[family])
		if err != nil {
			return err
		}
		if err := l.backend.Restore(family, script); err != nil {
			return fmt.Errorf("restore %s forwarding rules: %w", family, err)
		}
	}
	return nil
}

func buildIptablesForwardRestoreScript(rules []forwarding.Rule) (string, error) {
	natRules := [][]string{{"-F", forwarding.ChainPreRouting}, {"-F", forwarding.ChainPostRouting}}
	filterRules := [][]string{{"-F", forwarding.ChainForward}}
	for _, rule := range rules {
		sourcePort := strings.ReplaceAll(rule.Port, "-", ":")
		targetPort := strings.ReplaceAll(rule.TargetPort, "-", ":")
		preRouting := []string{"-A", forwarding.ChainPreRouting}
		if rule.Interface != "" {
			preRouting = append(preRouting, "-i", rule.Interface)
		}
		preRouting = append(preRouting, "-p", rule.Protocol, "--dport", sourcePort)
		if !isRemoteTarget(rule.Family, rule.TargetIP) {
			natRules = append(natRules, append(preRouting, "-j", "REDIRECT", "--to-port", rule.TargetPort))
			continue
		}
		natRules = append(natRules,
			append(preRouting, "-j", "DNAT", "--to-destination", forwardingTarget(rule)),
			[]string{"-A", forwarding.ChainPostRouting, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "MASQUERADE"},
		)
		filterRules = append(filterRules,
			[]string{"-A", forwarding.ChainForward, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "ACCEPT"},
			[]string{"-A", forwarding.ChainForward, "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", targetPort, "-j", "ACCEPT"},
		)
	}
	var script strings.Builder
	for _, table := range []struct {
		name  string
		rules [][]string
	}{{iptables_helper.NatTab, natRules}, {iptables_helper.FilterTab, filterRules}} {
		script.WriteByte('*')
		script.WriteString(table.name)
		script.WriteByte('\n')
		for _, rule := range table.rules {
			for index, token := range rule {
				if token == "" || strings.ContainsAny(token, " \t\r\n\"'") {
					return "", fmt.Errorf("invalid iptables-restore token %q", token)
				}
				if index > 0 {
					script.WriteByte(' ')
				}
				script.WriteString(token)
			}
			script.WriteByte('\n')
		}
		script.WriteString("COMMIT\n")
	}
	return script.String(), nil
}

func forwardingTarget(rule forwarding.Rule) string {
	if rule.Family == forwarding.FamilyIPv6 {
		return "[" + rule.TargetIP + "]:" + rule.TargetPort
	}
	return rule.TargetIP + ":" + rule.TargetPort
}

func isRemoteTarget(family, target string) bool {
	if family == forwarding.FamilyIPv6 {
		return target != "" && target != "::1" && target != "localhost"
	}
	return target != "" && target != "127.0.0.1" && target != "localhost"
}

func (l *iptablesNATAdapter) Enable() error {
	if err := l.system.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	if l.backend.IPv6Available() {
		if err := l.system.WriteFile("/proc/sys/net/ipv6/conf/all/forwarding", []byte("1"), constant.FilePerm); err != nil {
			return fmt.Errorf("failed to enable IPv6 forwarding: %w", err)
		}
	}
	data, err := l.system.ReadFile("/etc/sysctl.conf")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read /etc/sysctl.conf: %w", err)
	}
	content := enableForwardingSysctls(string(data), l.backend.IPv6Available())
	if err := l.system.WriteFile("/etc/sysctl.conf", []byte(content), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to persist IP forwarding: %w", err)
	}
	if err := l.system.RunWithOptionalSudo("sysctl", "-p"); err != nil {
		return fmt.Errorf("failed to apply IP forwarding: %w", err)
	}

	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		if family == forwarding.FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		if err := l.batchEnsureChains(family); err != nil {
			return err
		}
	}
	return nil
}

func (l *iptablesNATAdapter) batchEnsureChains(family string) error {
	list := l.backend.RunWithStd
	if family == forwarding.FamilyIPv6 {
		list = l.backend.RunIPv6WithStd
	}
	outputs := make(map[string]string, 2)
	for _, table := range []string{iptables_helper.NatTab, iptables_helper.FilterTab} {
		output, err := list(table, "-S")
		if err != nil {
			return err
		}
		outputs[table] = output
	}
	script := buildIptablesForwardLifecycleScript(outputs, true)
	if script == "" {
		return nil
	}
	if err := l.backend.Restore(family, script); err != nil {
		return fmt.Errorf("batch initialize %s forwarding chains: %w", family, err)
	}
	return nil
}

func (l *iptablesNATAdapter) Cleanup() error {
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		if family == forwarding.FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		list := l.backend.RunWithStd
		if family == forwarding.FamilyIPv6 {
			list = l.backend.RunIPv6WithStd
		}
		outputs := make(map[string]string, 2)
		for _, table := range []string{iptables_helper.NatTab, iptables_helper.FilterTab} {
			rules, err := list(table, "-S")
			if err != nil {
				return err
			}
			outputs[table] = rules
		}
		script := buildIptablesForwardLifecycleScript(outputs, false)
		if script != "" {
			if err := l.backend.Restore(family, script); err != nil {
				return fmt.Errorf("batch delete %s forwarding chains: %w", family, err)
			}
		}
	}
	for _, file := range []string{forwarding.ForwardFile, forwarding.PreRoutingFile, forwarding.PostRoutingFile,
		iptables_helper.IPv6FileName(forwarding.ForwardFile), iptables_helper.IPv6FileName(forwarding.PreRoutingFile), iptables_helper.IPv6FileName(forwarding.PostRoutingFile)} {
		if err := os.Remove(filepath.Join(global.Dir.FirewallDir, file)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func buildIptablesForwardLifecycleScript(outputs map[string]string, create bool) string {
	items := []struct{ table, parent, chain string }{
		{iptables_helper.NatTab, "PREROUTING", forwarding.ChainPreRouting},
		{iptables_helper.NatTab, "POSTROUTING", forwarding.ChainPostRouting},
		{iptables_helper.FilterTab, "FORWARD", forwarding.ChainForward},
	}
	byTable := make(map[string][]string, 2)
	for _, item := range items {
		output := outputs[item.table]
		chainExists := containsExactLine(output, "-N "+item.chain)
		bindingExists := containsExactLine(output, "-A "+item.parent+" -j "+item.chain)
		if create {
			if !chainExists {
				byTable[item.table] = append(byTable[item.table], "-N "+item.chain)
			}
			if !bindingExists {
				byTable[item.table] = append(byTable[item.table], "-A "+item.parent+" -j "+item.chain)
			}
			continue
		}
		if bindingExists {
			byTable[item.table] = append(byTable[item.table], "-D "+item.parent+" -j "+item.chain)
		}
		if chainExists {
			byTable[item.table] = append(byTable[item.table], "-F "+item.chain, "-X "+item.chain)
		}
	}
	var script strings.Builder
	for _, table := range []string{iptables_helper.NatTab, iptables_helper.FilterTab} {
		lines := byTable[table]
		if len(lines) == 0 {
			continue
		}
		script.WriteByte('*')
		script.WriteString(table)
		script.WriteByte('\n')
		script.WriteString(strings.Join(lines, "\n"))
		script.WriteString("\nCOMMIT\n")
	}
	return script.String()
}

func containsExactLine(output, want string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func enableIPv4Forwarding(content string) string {
	return enableForwardingSysctls(content, false)
}

func enableForwardingSysctls(content string, withIPv6 bool) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	wanted := map[string]string{"net.ipv4.ip_forward": "net.ipv4.ip_forward = 1"}
	if withIPv6 {
		wanted["net.ipv6.conf.all.forwarding"] = "net.ipv6.conf.all.forwarding = 1"
	}
	found := make(map[string]bool, len(wanted))
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if replacement, ok := wanted[key]; ok {
			lines[index] = replacement
			found[key] = true
		}
	}
	for _, key := range []string{"net.ipv4.ip_forward", "net.ipv6.conf.all.forwarding"} {
		if replacement, ok := wanted[key]; ok && !found[key] {
			lines = append(lines, replacement)
		}
	}
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n") + "\n"
}

func (l *iptablesNATAdapter) InitStatus() (bool, bool, error) {
	ipv4Init, ipv4Bind, err := l.familyInitStatus(forwarding.FamilyIPv4)
	if err != nil {
		return false, false, err
	}
	if !l.backend.IPv6Available() {
		return ipv4Init, ipv4Bind, nil
	}
	ipv6Init, ipv6Bind, err := l.familyInitStatus(forwarding.FamilyIPv6)
	if err != nil {
		return false, false, err
	}
	return ipv4Init && ipv6Init, ipv4Bind && ipv6Bind, nil
}

func (l *iptablesNATAdapter) familyInitStatus(family string) (bool, bool, error) {
	sysctlPath := "/proc/sys/net/ipv4/ip_forward"
	label := "IPv4"
	list := l.backend.RunWithStd
	if family == forwarding.FamilyIPv6 {
		sysctlPath = "/proc/sys/net/ipv6/conf/all/forwarding"
		label = "IPv6"
		list = l.backend.RunIPv6WithStd
	}
	data, err := l.system.ReadFile(sysctlPath)
	if err != nil {
		return false, false, fmt.Errorf("read %s forwarding status: %w", label, err)
	}
	forwardingEnabled := strings.TrimSpace(string(data)) != "0"
	natRules, err := list(iptables_helper.NatTab, "-S")
	if err != nil {
		return false, false, fmt.Errorf("list %s NAT initialization rules: %w", label, err)
	}
	natInit, natBind := checkInitAndBind(
		[]string{"-N " + forwarding.ChainPreRouting, "-N " + forwarding.ChainPostRouting},
		[]string{"-A PREROUTING -j " + forwarding.ChainPreRouting, "-A POSTROUTING -j " + forwarding.ChainPostRouting},
		strings.Split(natRules, "\n"),
	)
	if !natInit {
		return false, false, nil
	}
	filterRules, err := list(iptables_helper.FilterTab, "-S")
	if err != nil {
		return false, false, fmt.Errorf("list %s filter initialization rules: %w", label, err)
	}
	filterInit, filterBind := checkInitAndBind(
		[]string{"-N " + forwarding.ChainForward},
		[]string{"-A FORWARD -j " + forwarding.ChainForward},
		strings.Split(filterRules, "\n"),
	)
	return natInit && filterInit, forwardingEnabled && natBind && filterBind, nil
}

func (l *iptablesNATAdapter) FamilyStatus(family string) (bool, bool, error) {
	if family == forwarding.FamilyIPv6 && !l.backend.IPv6Available() {
		return false, false, nil
	}
	return l.familyInitStatus(family)
}

func checkInitAndBind(initRules, bindRules, lines []string) (bool, bool) {
	for _, rule := range initRules {
		if !containsExactRule(lines, rule) {
			return false, false
		}
	}
	for _, rule := range bindRules {
		if !containsExactRule(lines, rule) {
			return true, false
		}
	}
	return true, true
}

func containsExactRule(lines []string, rule string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) == strings.TrimSpace(rule) {
			return true
		}
	}
	return false
}

func (l *iptablesNATAdapter) Replay() error {
	for _, item := range []struct {
		table string
		chain string
		file  string
	}{
		{iptables_helper.FilterTab, forwarding.ChainForward, forwarding.ForwardFile},
		{iptables_helper.NatTab, forwarding.ChainPreRouting, forwarding.PreRoutingFile},
		{iptables_helper.NatTab, forwarding.ChainPostRouting, forwarding.PostRoutingFile},
	} {
		if err := l.backend.LoadRulesFromFile(item.table, item.chain, item.file); err != nil {
			return err
		}
		if l.backend.IPv6Available() {
			if err := l.backend.LoadIPv6RulesFromFile(item.table, item.chain, iptables_helper.IPv6FileName(item.file)); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseIptablesRules(stdout, family string) []forwarding.Rule {
	var rules []forwarding.Rule
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			continue
		}
		rule := forwarding.Rule{
			Num:       fields[0],
			Family:    family,
			Protocol:  loadProtocol(fields[4]),
			Interface: fields[6],
			Port:      loadSourcePort(fields[11]),
		}
		if len(fields) == 15 && fields[13] == "ports" {
			rule.TargetPort = fields[14]
		}
		if len(fields) == 13 && strings.HasPrefix(fields[12], "to:") {
			rule.TargetIP, rule.TargetPort = parseIptablesTarget(strings.TrimPrefix(fields[12], "to:"))
		}
		if rule.TargetIP == "" {
			if family == forwarding.FamilyIPv6 {
				rule.TargetIP = "::1"
			} else {
				rule.TargetIP = "127.0.0.1"
			}
		}
		rule.TargetPort = strings.TrimPrefix(rule.TargetPort, ":")
		rules = append(rules, rule)
	}
	return rules
}

func parseIptablesTarget(value string) (string, string) {
	if strings.HasPrefix(value, "[") {
		separator := strings.LastIndex(value, "]:")
		if separator > 0 {
			return value[1:separator], value[separator+2:]
		}
		return "", ""
	}
	separator := strings.LastIndex(value, ":")
	if separator <= 0 {
		return "", ""
	}
	return value[:separator], value[separator+1:]
}

func loadProtocol(protocol string) string {
	switch protocol {
	case "0":
		return "all"
	case "1":
		return "icmp"
	case "6":
		return "tcp"
	case "17":
		return "udp"
	default:
		return protocol
	}
}

func loadSourcePort(value string) string {
	port := ""
	if strings.Contains(value, "dpt:") {
		port = strings.ReplaceAll(value, "dpt:", "")
	}
	if strings.Contains(value, "dpts:") {
		port = strings.ReplaceAll(value, "dpts:", "")
	}
	return strings.ReplaceAll(port, ":", "-")
}
