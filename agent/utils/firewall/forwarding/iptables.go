package forwarding

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	firewallutil "github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/mattn/go-shellwords"
)

type iptablesBackend interface {
	IPv6Available() bool
	Run(table string, args ...string) error
	RunWithStd(table string, args ...string) (string, error)
	RunIPv6(table string, args ...string) error
	RunIPv6WithStd(table string, args ...string) (string, error)
	Restore(ctx context.Context, family, input string) error
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
	if len(args) == 1 && args[0] == "-S" {
		return iptables_helper.ReadTable(context.Background(), table, false)
	}
	return iptables_helper.RunWithStd(table, args...)
}

func (systemIptablesBackend) RunIPv6(table string, args ...string) error {
	return iptables_helper.RunIPv6(table, args...)
}

func (systemIptablesBackend) RunIPv6WithStd(table string, args ...string) (string, error) {
	if len(args) == 1 && args[0] == "-S" {
		return iptables_helper.ReadTable(context.Background(), table, true)
	}
	return iptables_helper.RunIPv6WithStd(table, args...)
}

func (systemIptablesBackend) Restore(ctx context.Context, family, input string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return err
	}
	executable := commands.Restore4
	if family == FamilyIPv6 {
		executable = commands.Restore6
		if executable == "" {
			return fmt.Errorf("ip6tables-restore command family is unavailable")
		}
	}
	var stderr strings.Builder
	manager := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(input)), cmd.WithStderr(&stderr))
	err = manager.RunWithOptionalSudo(executable, "--noflush", "--wait")
	if err == nil && strings.TrimSpace(stderr.String()) != "" {
		err = fmt.Errorf("firewall command warning: %s", strings.TrimSpace(stderr.String()))
	}
	return firewallutil.WrapBatchCommandError(executable+" --noflush --wait", input, err)
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

type Iptables struct {
	provider string
	backend  iptablesBackend
	system   forwardingSystem
}

func NewIptables(provider string) *Iptables {
	return &Iptables{
		provider: provider,
		backend:  systemIptablesBackend{},
		system:   defaultForwardingSystem{},
	}
}

func (l *Iptables) Name() string {
	return l.provider
}

func (l *Iptables) List() ([]Rule, error) {
	stdout, err := l.backend.RunWithStd(iptables_helper.NatTab, "-S")
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT rules: %w", err)
	}
	rules := parseIptablesRules(stdout, FamilyIPv4)
	if !l.backend.IPv6Available() {
		return rules, nil
	}
	stdout, err = l.backend.RunIPv6WithStd(iptables_helper.NatTab, "-S")
	if err != nil {
		return nil, fmt.Errorf("failed to list IPv6 NAT rules: %w", err)
	}
	return append(rules, parseIptablesRules(stdout, FamilyIPv6)...), nil
}

func (l *Iptables) ReplaceRules(rules []Rule) error {
	byFamily := map[string][]Rule{
		FamilyIPv4: nil,
		FamilyIPv6: nil,
	}
	for _, rule := range rules {
		normalized, err := NormalizeRule(rule)
		if err != nil {
			return err
		}
		if normalized.Family == FamilyIPv6 && !l.backend.IPv6Available() {
			return fmt.Errorf("ip6tables command family is unavailable")
		}
		byFamily[normalized.Family] = append(byFamily[normalized.Family], normalized)
	}
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		if family == FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		if err := l.batchEnsureChains(family); err != nil {
			return err
		}
		script, err := buildIptablesForwardScript(byFamily[family], OperationAdd, true)
		if err != nil {
			return err
		}
		if err := l.backend.Restore(context.Background(), family, script); err != nil {
			return fmt.Errorf("restore %s forwarding rules: %w", family, err)
		}
	}
	return nil
}

func (l *Iptables) CreateRules(ctx context.Context, rules []Rule) error {
	if len(rules) == 0 {
		return nil
	}
	script, err := buildIptablesForwardScript(rules, OperationAdd, false)
	if err != nil {
		return err
	}
	family := rules[0].Family
	if family == "" {
		family = FamilyIPv4
	}
	return l.backend.Restore(ctx, family, script)
}

func (l *Iptables) DeleteRules(ctx context.Context, rules []Rule) error {
	if len(rules) == 0 {
		return nil
	}
	script, err := buildIptablesForwardScript(rules, OperationRemove, false)
	if err != nil {
		return err
	}
	family := rules[0].Family
	if family == "" {
		family = FamilyIPv4
	}
	return l.backend.Restore(ctx, family, script)
}

func buildIptablesForwardScript(rules []Rule, operation OperationType, replace bool) (string, error) {
	var natRules, filterRules [][]string
	if replace {
		natRules = [][]string{{"-F", ChainPreRouting}, {"-F", ChainPostRouting}}
		filterRules = [][]string{{"-F", ChainForward}}
	}
	verb := "-A"
	if operation == OperationRemove {
		verb = "-D"
	} else if operation != OperationAdd {
		return "", fmt.Errorf("unsupported forwarding operation %q", operation)
	}
	family := ""
	for _, rule := range rules {
		normalized, err := NormalizeRule(rule)
		if err != nil {
			return "", err
		}
		rule = normalized
		if family != "" && family != rule.Family {
			return "", fmt.Errorf("iptables forwarding batch must use one address family")
		}
		family = rule.Family
		sourcePort := strings.ReplaceAll(rule.Port, "-", ":")
		targetPort := strings.ReplaceAll(rule.TargetPort, "-", ":")
		preRouting := []string{verb, ChainPreRouting}
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
			[]string{verb, ChainPostRouting, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "MASQUERADE"},
		)
		filterRules = append(filterRules,
			[]string{verb, ChainForward, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "ACCEPT"},
			[]string{verb, ChainForward, "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", targetPort, "-j", "ACCEPT"},
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

func forwardingTarget(rule Rule) string {
	if rule.Family == FamilyIPv6 {
		return "[" + rule.TargetIP + "]:" + rule.TargetPort
	}
	return rule.TargetIP + ":" + rule.TargetPort
}

func isRemoteTarget(family, target string) bool {
	if family == FamilyIPv6 {
		return target != "" && target != "::1" && target != "localhost"
	}
	return target != "" && target != "127.0.0.1" && target != "localhost"
}

func (l *Iptables) Enable() error {
	if err := ensureForwardingSysctls(l.system, l.backend.IPv6Available()); err != nil {
		return err
	}

	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		if family == FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		if err := l.batchEnsureChains(family); err != nil {
			return err
		}
	}
	return nil
}

func (l *Iptables) batchEnsureChains(family string) error {
	list := l.backend.RunWithStd
	if family == FamilyIPv6 {
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
	if err := l.backend.Restore(context.Background(), family, script); err != nil {
		return fmt.Errorf("batch initialize %s forwarding chains: %w", family, err)
	}
	return nil
}

func (l *Iptables) Cleanup() error {
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		if family == FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		list := l.backend.RunWithStd
		if family == FamilyIPv6 {
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
			if err := l.backend.Restore(context.Background(), family, script); err != nil {
				return fmt.Errorf("batch delete %s forwarding chains: %w", family, err)
			}
		}
	}
	for _, file := range []string{ForwardFile, PreRoutingFile, PostRoutingFile,
		iptables_helper.IPv6FileName(ForwardFile), iptables_helper.IPv6FileName(PreRoutingFile), iptables_helper.IPv6FileName(PostRoutingFile)} {
		if err := os.Remove(filepath.Join(global.Dir.FirewallDir, file)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func buildIptablesForwardLifecycleScript(outputs map[string]string, create bool) string {
	items := []struct{ table, parent, chain string }{
		{iptables_helper.NatTab, "PREROUTING", ChainPreRouting},
		{iptables_helper.NatTab, "POSTROUTING", ChainPostRouting},
	}
	byTable := make(map[string][]string, 2)
	for _, item := range items {
		output := outputs[item.table]
		chainExists := containsExactLine(output, "-N "+item.chain)
		binding := "-A " + item.parent + " -j " + item.chain
		bindingCount := countExactLines(output, binding)
		if create {
			if !chainExists {
				byTable[item.table] = append(byTable[item.table], "-N "+item.chain)
			}
			if bindingCount == 0 {
				byTable[item.table] = append(byTable[item.table], "-A "+item.parent+" -j "+item.chain)
			}
			continue
		}
		for range bindingCount {
			byTable[item.table] = append(byTable[item.table], "-D "+item.parent+" -j "+item.chain)
		}
		if chainExists {
			byTable[item.table] = append(byTable[item.table], "-F "+item.chain, "-X "+item.chain)
		}
	}

	filterOutput := outputs[iptables_helper.FilterTab]
	filterChainExists := containsExactLine(filterOutput, "-N "+ChainForward)
	filterBinding := "-A FORWARD -j " + ChainForward
	filterBindingCount := countExactLines(filterOutput, filterBinding)
	if create {
		if !filterChainExists {
			byTable[iptables_helper.FilterTab] = append(byTable[iptables_helper.FilterTab], "-N "+ChainForward)
		}
		if !forwardBindingEffective(filterOutput) {
			for range filterBindingCount {
				byTable[iptables_helper.FilterTab] = append(byTable[iptables_helper.FilterTab], "-D FORWARD -j "+ChainForward)
			}
			byTable[iptables_helper.FilterTab] = append(byTable[iptables_helper.FilterTab], canonicalForwardBindingRule(filterOutput))
		}
	} else {
		for range filterBindingCount {
			byTable[iptables_helper.FilterTab] = append(byTable[iptables_helper.FilterTab], "-D FORWARD -j "+ChainForward)
		}
		if filterChainExists {
			byTable[iptables_helper.FilterTab] = append(byTable[iptables_helper.FilterTab], "-F "+ChainForward, "-X "+ChainForward)
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

func countExactLines(output, want string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == want {
			count++
		}
	}
	return count
}

func forwardBindingEffective(output string) bool {
	binding := "-A FORWARD -j " + ChainForward
	bindingPosition := 0
	terminalPosition := 0
	position := 0
	bindings := 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "-A FORWARD ") {
			continue
		}
		position++
		if line == binding {
			bindings++
			bindingPosition = position
		}
		if terminalPosition == 0 && isUnconditionalForwardTerminal(line) {
			terminalPosition = position
		}
	}
	return bindings == 1 && (terminalPosition == 0 || bindingPosition < terminalPosition)
}

func canonicalForwardBindingRule(output string) string {
	binding := "-A FORWARD -j " + ChainForward
	position := 1
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "-A FORWARD ") || line == binding {
			continue
		}
		if isUnconditionalForwardTerminal(line) {
			return fmt.Sprintf("-I FORWARD %d -j %s", position, ChainForward)
		}
		position++
	}
	return "-A FORWARD -j " + ChainForward
}

func isUnconditionalForwardTerminal(line string) bool {
	fields := strings.Fields(line)
	if len(fields) < 4 || fields[0] != "-A" || fields[1] != "FORWARD" || fields[2] != "-j" {
		return false
	}
	switch fields[3] {
	case "ACCEPT", "DROP", "REJECT", "RETURN":
		return true
	default:
		return false
	}
}

func containsExactLine(output, want string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func (l *Iptables) InitStatus() (bool, bool, error) {
	ipv4Init, ipv4Bind, err := l.familyInitStatus(FamilyIPv4)
	if err != nil {
		return false, false, err
	}
	if !l.backend.IPv6Available() {
		return ipv4Init, ipv4Bind, nil
	}
	ipv6Init, ipv6Bind, err := l.familyInitStatus(FamilyIPv6)
	if err != nil {
		return false, false, err
	}
	return ipv4Init && ipv6Init, ipv4Bind && ipv6Bind, nil
}

func (l *Iptables) familyInitStatus(family string) (bool, bool, error) {
	sysctlPath := "/proc/sys/net/ipv4/ip_forward"
	label := "IPv4"
	list := l.backend.RunWithStd
	if family == FamilyIPv6 {
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
		[]string{"-N " + ChainPreRouting, "-N " + ChainPostRouting},
		[]string{"-A PREROUTING -j " + ChainPreRouting, "-A POSTROUTING -j " + ChainPostRouting},
		strings.Split(natRules, "\n"),
	)
	if !natInit {
		return false, false, nil
	}
	filterRules, err := list(iptables_helper.FilterTab, "-S")
	if err != nil {
		return false, false, fmt.Errorf("list %s filter initialization rules: %w", label, err)
	}
	filterInit, _ := checkInitAndBind(
		[]string{"-N " + ChainForward},
		nil,
		strings.Split(filterRules, "\n"),
	)
	filterBind := forwardBindingEffective(filterRules)
	return natInit && filterInit, forwardingEnabled && natBind && filterInit && filterBind, nil
}

func (l *Iptables) FamilyStatus(family string) (bool, bool, error) {
	if family == FamilyIPv6 && !l.backend.IPv6Available() {
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

func (l *Iptables) Replay() error {
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		if family == FamilyIPv6 && !l.backend.IPv6Available() {
			continue
		}
		if err := l.batchEnsureChains(family); err != nil {
			return err
		}
	}
	for _, item := range []struct {
		table string
		chain string
		file  string
	}{
		{iptables_helper.FilterTab, ChainForward, ForwardFile},
		{iptables_helper.NatTab, ChainPreRouting, PreRoutingFile},
		{iptables_helper.NatTab, ChainPostRouting, PostRoutingFile},
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

func parseIptablesRules(stdout, family string) []Rule {
	var rules []Rule
	num := 0
lines:
	for _, line := range strings.Split(stdout, "\n") {
		fields, err := shellwords.Parse(line)
		if err != nil || len(fields) < 2 || fields[0] != "-A" || fields[1] != ChainPreRouting {
			continue
		}
		num++
		rule := Rule{Num: strconv.Itoa(num), Family: family}
		target := ""
		for index := 2; index < len(fields); index++ {
			var value *string
			switch fields[index] {
			case "!":
				continue lines
			case "-p", "--protocol":
				value = &rule.Protocol
			case "-i", "--in-interface":
				value = &rule.Interface
			case "--dport", "--destination-port":
				value = &rule.Port
			case "-j", "--jump":
				value = &target
			case "--to-destination":
				if index+1 >= len(fields) {
					continue lines
				}
				index++
				rule.TargetIP, rule.TargetPort = parseIptablesTarget(fields[index])
			case "--to-ports", "--to-port":
				value = &rule.TargetPort
			case "-m", "--match", "--comment":
				if index+1 >= len(fields) {
					continue lines
				}
				index++
			}
			if value != nil {
				if index+1 >= len(fields) {
					continue lines
				}
				index++
				*value = fields[index]
			}
		}
		if rule.Protocol == "" || rule.Port == "" || rule.TargetPort == "" {
			continue
		}
		switch target {
		case "REDIRECT":
			rule.TargetIP = "127.0.0.1"
			if family == FamilyIPv6 {
				rule.TargetIP = "::1"
			}
		case "DNAT":
			if rule.TargetIP == "" {
				continue
			}
		default:
			continue
		}
		rule.Protocol = loadProtocol(rule.Protocol)
		rule.Port = strings.ReplaceAll(rule.Port, ":", "-")
		rule.TargetPort = strings.ReplaceAll(rule.TargetPort, ":", "-")
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
