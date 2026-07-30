package forwarding

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

type legacyIptablesBackend interface {
	Run(table string, args ...string) error
	RunWithStd(table string, args ...string) (string, error)
	AddChainWithAppend(table, parentChain, chain string) error
	SaveRulesToFile(table, chain, fileName string) error
	LoadRulesFromFile(table, chain, fileName string) error
}

type systemIptablesBackend struct{}

func (systemIptablesBackend) Run(table string, args ...string) error {
	return iptables.Run(table, args...)
}

func (systemIptablesBackend) RunWithStd(table string, args ...string) (string, error) {
	return iptables.RunWithStd(table, args...)
}

func (systemIptablesBackend) AddChainWithAppend(table, parentChain, chain string) error {
	return iptables.AddChainWithAppend(table, parentChain, chain)
}

func (systemIptablesBackend) SaveRulesToFile(table, chain, fileName string) error {
	return iptables.SaveRulesToFile(table, chain, fileName)
}

func (systemIptablesBackend) LoadRulesFromFile(table, chain, fileName string) error {
	return iptables.LoadRulesFromFile(table, chain, fileName)
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

type legacyNATAdapter struct {
	provider string
	backend  legacyIptablesBackend
	system   forwardingSystem
}

func newLegacyNATAdapter(provider string) *legacyNATAdapter {
	return &legacyNATAdapter{
		provider: provider,
		backend:  systemIptablesBackend{},
		system:   defaultForwardingSystem{},
	}
}

func (l *legacyNATAdapter) Name() string {
	return l.provider
}

func (l *legacyNATAdapter) List() ([]Rule, error) {
	stdout, err := l.backend.RunWithStd(iptables.NatTab, "-nvL", ChainPreRouting, "--line-numbers")
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT rules: %w", err)
	}
	return parseLegacyRules(stdout), nil
}

func (l *legacyNATAdapter) Operate(rule Rule, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}
	if rule.Protocol == "" || rule.Port == "" || rule.TargetPort == "" {
		return fmt.Errorf("protocol, port, and target port are required")
	}
	var err error
	if operation == "add" {
		err = l.add(rule)
	} else {
		err = l.remove(rule)
	}
	if err != nil {
		return err
	}
	l.persist()
	return nil
}

func (l *legacyNATAdapter) add(rule Rule) error {
	srcPort := strings.ReplaceAll(rule.Port, "-", ":")
	targetPort := strings.ReplaceAll(rule.TargetPort, "-", ":")
	if isRemoteTarget(rule.TargetIP) {
		args := []string{"-A", ChainPreRouting}
		if rule.Interface != "" {
			args = append(args, "-i", rule.Interface)
		}
		args = append(args, "-p", rule.Protocol, "--dport", srcPort, "-j", "DNAT", "--to-destination", rule.TargetIP+":"+rule.TargetPort)
		if err := l.backend.Run(iptables.NatTab, args...); err != nil {
			return err
		}
		if err := l.backend.Run(iptables.NatTab, "-A", ChainPostRouting, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "MASQUERADE"); err != nil {
			return err
		}
		if err := l.backend.Run(iptables.FilterTab, "-A", ChainForward, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "ACCEPT"); err != nil {
			return err
		}
		return l.backend.Run(iptables.FilterTab, "-A", ChainForward, "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", targetPort, "-j", "ACCEPT")
	}

	args := []string{"-A", ChainPreRouting}
	if rule.Interface != "" {
		args = append(args, "-i", rule.Interface)
	}
	args = append(args, "-p", rule.Protocol, "--dport", srcPort, "-j", "REDIRECT", "--to-port", rule.TargetPort)
	return l.backend.Run(iptables.NatTab, args...)
}

func (l *legacyNATAdapter) remove(rule Rule) error {
	targetPort := strings.ReplaceAll(rule.TargetPort, "-", ":")
	if err := l.backend.Run(iptables.NatTab, "-D", ChainPreRouting, rule.Num); err != nil {
		return err
	}
	if !isRemoteTarget(rule.TargetIP) {
		return nil
	}
	if err := l.backend.Run(iptables.NatTab, "-D", ChainPostRouting, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := l.backend.Run(iptables.FilterTab, "-D", ChainForward, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "ACCEPT"); err != nil {
		return err
	}
	return l.backend.Run(iptables.FilterTab, "-D", ChainForward, "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", targetPort, "-j", "ACCEPT")
}

func isRemoteTarget(target string) bool {
	return target != "" && target != "127.0.0.1" && target != "localhost"
}

func (l *legacyNATAdapter) persist() {
	for _, item := range []struct {
		table string
		chain string
		file  string
	}{
		{iptables.FilterTab, ChainForward, ForwardFile},
		{iptables.NatTab, ChainPreRouting, PreRoutingFile},
		{iptables.NatTab, ChainPostRouting, PostRoutingFile},
	} {
		if err := l.backend.SaveRulesToFile(item.table, item.chain, item.file); err != nil {
			global.LOG.Errorf("persistence for %s failed, err: %v", item.chain, err)
		}
	}
}

func (l *legacyNATAdapter) Enable() error {
	if err := l.system.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	if data, err := l.system.ReadFile("/etc/sysctl.conf"); err == nil && !strings.Contains(string(data), "net.ipv4.ip_forward") {
		content := strings.TrimRight(string(data), "\n") + "\nnet.ipv4.ip_forward = 1\n"
		_ = l.system.WriteFile("/etc/sysctl.conf", []byte(content), constant.FilePerm)
	}
	_ = l.system.RunWithOptionalSudo("sysctl", "-p")

	for _, item := range []struct {
		table  string
		parent string
		chain  string
	}{
		{iptables.NatTab, "PREROUTING", ChainPreRouting},
		{iptables.NatTab, "POSTROUTING", ChainPostRouting},
		{iptables.FilterTab, "FORWARD", ChainForward},
	} {
		if err := l.backend.AddChainWithAppend(item.table, item.parent, item.chain); err != nil {
			return err
		}
	}
	return nil
}

func (l *legacyNATAdapter) InitStatus() (bool, bool) {
	data, err := l.system.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil || strings.TrimSpace(string(data)) == "0" {
		return false, false
	}
	natRules, err := l.backend.RunWithStd(iptables.NatTab, "-S")
	if err != nil {
		return false, false
	}
	natInit, natBind := checkInitAndBind(
		[]string{"-N " + ChainPreRouting, "-N " + ChainPostRouting},
		[]string{"-A PREROUTING -j " + ChainPreRouting, "-A POSTROUTING -j " + ChainPostRouting},
		strings.Split(natRules, "\n"),
	)
	if !natInit {
		return false, false
	}
	filterRules, err := l.backend.RunWithStd(iptables.FilterTab, "-S")
	if err != nil {
		return false, false
	}
	filterInit, filterBind := checkInitAndBind(
		[]string{"-N " + ChainForward},
		[]string{"-A FORWARD -j " + ChainForward},
		strings.Split(filterRules, "\n"),
	)
	return natInit && filterInit, natBind && filterBind
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

func (l *legacyNATAdapter) Replay() error {
	for _, item := range []struct {
		table string
		chain string
		file  string
	}{
		{iptables.FilterTab, ChainForward, ForwardFile},
		{iptables.NatTab, ChainPreRouting, PreRoutingFile},
		{iptables.NatTab, ChainPostRouting, PostRoutingFile},
	} {
		if err := l.backend.LoadRulesFromFile(item.table, item.chain, item.file); err != nil {
			return err
		}
	}
	return nil
}

func parseLegacyRules(stdout string) []Rule {
	var rules []Rule
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			continue
		}
		rule := Rule{
			Num:       fields[0],
			Protocol:  loadProtocol(fields[4]),
			Interface: fields[6],
			Port:      loadSourcePort(fields[11]),
		}
		if len(fields) == 15 && fields[13] == "ports" {
			rule.TargetPort = fields[14]
		}
		if len(fields) == 13 && strings.HasPrefix(fields[12], "to:") {
			parts := strings.Split(fields[12], ":")
			if len(parts) > 2 {
				rule.TargetPort = parts[2]
				rule.TargetIP = parts[1]
			}
		}
		if rule.TargetIP == "" {
			rule.TargetIP = "127.0.0.1"
		}
		rule.TargetPort = strings.TrimPrefix(rule.TargetPort, ":")
		rules = append(rules, rule)
	}
	return rules
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
