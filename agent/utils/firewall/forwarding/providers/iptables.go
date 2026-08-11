package providers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
)

type iptablesBackend interface {
	Run(table string, args ...string) error
	RunWithStd(table string, args ...string) (string, error)
	AddChainWithAppend(table, parentChain, chain string) error
	SaveRulesToFile(table, chain, fileName string) error
	LoadRulesFromFile(table, chain, fileName string) error
}

type systemIptablesBackend struct{}

func (systemIptablesBackend) Run(table string, args ...string) error {
	return iptables_helper.Run(table, args...)
}

func (systemIptablesBackend) RunWithStd(table string, args ...string) (string, error) {
	return iptables_helper.RunWithStd(table, args...)
}

func (systemIptablesBackend) AddChainWithAppend(table, parentChain, chain string) error {
	return iptables_helper.AddChainWithAppend(table, parentChain, chain)
}

func (systemIptablesBackend) SaveRulesToFile(table, chain, fileName string) error {
	return iptables_helper.SaveRulesToFile(table, chain, fileName)
}

func (systemIptablesBackend) LoadRulesFromFile(table, chain, fileName string) error {
	return iptables_helper.LoadRulesFromFile(table, chain, fileName)
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
	return parseIptablesRules(stdout), nil
}

func (l *iptablesNATAdapter) Operate(rule forwarding.Rule, operation forwarding.OperationType) error {
	if operation != forwarding.OperationAdd && operation != forwarding.OperationRemove {
		return buserr.New("ErrCmdIllegal")
	}
	if rule.Protocol == "" || rule.Port == "" || rule.TargetPort == "" {
		return fmt.Errorf("protocol, port, and target port are required")
	}
	var err error
	if operation == forwarding.OperationAdd {
		err = l.add(rule)
	} else {
		err = l.remove(rule)
	}
	if err != nil {
		return err
	}
	return l.persist()
}

func (l *iptablesNATAdapter) add(rule forwarding.Rule) error {
	srcPort := strings.ReplaceAll(rule.Port, "-", ":")
	targetPort := strings.ReplaceAll(rule.TargetPort, "-", ":")
	if isRemoteTarget(rule.TargetIP) {
		args := []string{"-A", forwarding.ChainPreRouting}
		if rule.Interface != "" {
			args = append(args, "-i", rule.Interface)
		}
		args = append(args, "-p", rule.Protocol, "--dport", srcPort, "-j", "DNAT", "--to-destination", rule.TargetIP+":"+rule.TargetPort)
		if err := l.backend.Run(iptables_helper.NatTab, args...); err != nil {
			return err
		}
		if err := l.backend.Run(iptables_helper.NatTab, "-A", forwarding.ChainPostRouting, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "MASQUERADE"); err != nil {
			return err
		}
		if err := l.backend.Run(iptables_helper.FilterTab, "-A", forwarding.ChainForward, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "ACCEPT"); err != nil {
			return err
		}
		return l.backend.Run(iptables_helper.FilterTab, "-A", forwarding.ChainForward, "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", targetPort, "-j", "ACCEPT")
	}

	args := []string{"-A", forwarding.ChainPreRouting}
	if rule.Interface != "" {
		args = append(args, "-i", rule.Interface)
	}
	args = append(args, "-p", rule.Protocol, "--dport", srcPort, "-j", "REDIRECT", "--to-port", rule.TargetPort)
	return l.backend.Run(iptables_helper.NatTab, args...)
}

func (l *iptablesNATAdapter) remove(rule forwarding.Rule) error {
	targetPort := strings.ReplaceAll(rule.TargetPort, "-", ":")
	if err := l.backend.Run(iptables_helper.NatTab, "-D", forwarding.ChainPreRouting, rule.Num); err != nil {
		return err
	}
	if !isRemoteTarget(rule.TargetIP) {
		return nil
	}
	if err := l.backend.Run(iptables_helper.NatTab, "-D", forwarding.ChainPostRouting, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := l.backend.Run(iptables_helper.FilterTab, "-D", forwarding.ChainForward, "-d", rule.TargetIP, "-p", rule.Protocol, "--dport", targetPort, "-j", "ACCEPT"); err != nil {
		return err
	}
	return l.backend.Run(iptables_helper.FilterTab, "-D", forwarding.ChainForward, "-s", rule.TargetIP, "-p", rule.Protocol, "--sport", targetPort, "-j", "ACCEPT")
}

func isRemoteTarget(target string) bool {
	return target != "" && target != "127.0.0.1" && target != "localhost"
}

func (l *iptablesNATAdapter) persist() error {
	var persistErrors []error
	for _, item := range []struct {
		table string
		chain string
		file  string
	}{
		{iptables_helper.FilterTab, forwarding.ChainForward, forwarding.ForwardFile},
		{iptables_helper.NatTab, forwarding.ChainPreRouting, forwarding.PreRoutingFile},
		{iptables_helper.NatTab, forwarding.ChainPostRouting, forwarding.PostRoutingFile},
	} {
		if err := l.backend.SaveRulesToFile(item.table, item.chain, item.file); err != nil {
			persistErrors = append(persistErrors, fmt.Errorf("persist %s: %w", item.chain, err))
		}
	}
	return errors.Join(persistErrors...)
}

func (l *iptablesNATAdapter) Enable() error {
	if err := l.system.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	data, err := l.system.ReadFile("/etc/sysctl.conf")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read /etc/sysctl.conf: %w", err)
	}
	content := enableIPv4Forwarding(string(data))
	if err := l.system.WriteFile("/etc/sysctl.conf", []byte(content), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to persist IP forwarding: %w", err)
	}
	if err := l.system.RunWithOptionalSudo("sysctl", "-p"); err != nil {
		return fmt.Errorf("failed to apply IP forwarding: %w", err)
	}

	for _, item := range []struct {
		table  string
		parent string
		chain  string
	}{
		{iptables_helper.NatTab, "PREROUTING", forwarding.ChainPreRouting},
		{iptables_helper.NatTab, "POSTROUTING", forwarding.ChainPostRouting},
		{iptables_helper.FilterTab, "FORWARD", forwarding.ChainForward},
	} {
		if err := l.backend.AddChainWithAppend(item.table, item.parent, item.chain); err != nil {
			return err
		}
	}
	return nil
}

func enableIPv4Forwarding(content string) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	found := false
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "net.ipv4.ip_forward" {
			lines[index] = "net.ipv4.ip_forward = 1"
			found = true
		}
	}
	if !found {
		lines = append(lines, "net.ipv4.ip_forward = 1")
	}
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n") + "\n"
}

func (l *iptablesNATAdapter) InitStatus() (bool, bool, error) {
	data, err := l.system.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return false, false, fmt.Errorf("read IPv4 forwarding status: %w", err)
	}
	if strings.TrimSpace(string(data)) == "0" {
		return false, false, nil
	}
	natRules, err := l.backend.RunWithStd(iptables_helper.NatTab, "-S")
	if err != nil {
		return false, false, fmt.Errorf("list NAT initialization rules: %w", err)
	}
	natInit, natBind := checkInitAndBind(
		[]string{"-N " + forwarding.ChainPreRouting, "-N " + forwarding.ChainPostRouting},
		[]string{"-A PREROUTING -j " + forwarding.ChainPreRouting, "-A POSTROUTING -j " + forwarding.ChainPostRouting},
		strings.Split(natRules, "\n"),
	)
	if !natInit {
		return false, false, nil
	}
	filterRules, err := l.backend.RunWithStd(iptables_helper.FilterTab, "-S")
	if err != nil {
		return false, false, fmt.Errorf("list filter initialization rules: %w", err)
	}
	filterInit, filterBind := checkInitAndBind(
		[]string{"-N " + forwarding.ChainForward},
		[]string{"-A FORWARD -j " + forwarding.ChainForward},
		strings.Split(filterRules, "\n"),
	)
	return natInit && filterInit, natBind && filterBind, nil
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
	}
	return nil
}

func parseIptablesRules(stdout string) []forwarding.Rule {
	var rules []forwarding.Rule
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			continue
		}
		rule := forwarding.Rule{
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
