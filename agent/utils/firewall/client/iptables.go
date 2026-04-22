package client

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

type Iptables struct{}

func NewIptables() (*Iptables, error) {
	return &Iptables{}, nil
}

func (i *Iptables) Name() string {
	return "iptables"
}

func (i *Iptables) Status() (bool, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashC("iptables -L -n | head -1")
	if err != nil {
		return false, err
	}
	return strings.Contains(stdout, "Chain"), nil
}

func (i *Iptables) Start() error {
	return nil
}

func (i *Iptables) Stop() error {
	return nil
}

func (i *Iptables) Restart() error {
	return nil
}

func (i *Iptables) Reload() error {
	return nil
}

func (i *Iptables) Version() (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashC("iptables --version")
	if err != nil {
		return "", fmt.Errorf("failed to get iptables version: %w", err)
	}
	parts := strings.Fields(stdout)
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "v"), nil
	}
	return strings.TrimSpace(stdout), nil
}

func (i *Iptables) ListPort() ([]FireInfo, error) {
	var datas []FireInfo
	basicRules, err := iptables.ReadFilterRulesByChain(iptables.Chain1PanelBasic)
	if err != nil {
		return nil, err
	}
	beforeRules, _ := iptables.ReadFilterRulesByChain(iptables.Chain1PanelBasicBefore)
	basicRules = append(basicRules, beforeRules...)
	for _, item := range basicRules {
		if len(item.DstPort) == 0 {
			continue
		}
		if item.Strategy == "drop" || item.Strategy == "reject" {
			item.Strategy = "drop"
		}

		datas = append(datas, FireInfo{
			Chain:    item.Chain,
			Address:  item.SrcIP,
			Protocol: item.Protocol,
			Port:     item.DstPort,
			Strategy: item.Strategy,
			Family:   "ipv4",
		})
	}

	return datas, nil
}

func (i *Iptables) ListAddress() ([]FireInfo, error) {
	var datas []FireInfo
	basicRules, err := iptables.ReadFilterRulesByChain(iptables.Chain1PanelBasic)
	if err != nil {
		return nil, err
	}
	for _, item := range basicRules {
		if len(item.DstPort) != 0 || len(item.SrcPort) != 0 {
			continue
		}
		if item.Strategy == "drop" || item.Strategy == "reject" {
			item.Strategy = "drop"
		}
		datas = append(datas, FireInfo{
			Address:  item.SrcIP,
			Strategy: item.Strategy,
			Family:   "ipv4",
		})
	}
	return datas, nil
}

func (i *Iptables) Port(port FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}
	if len(port.Chain) == 0 {
		port.Chain = iptables.Chain1PanelBasic
	}

	portSpec, err := normalizePortSpec(port.Port)
	if err != nil {
		return err
	}

	protocol := port.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	action := "ACCEPT"
	if port.Strategy == "drop" {
		action = "DROP"
	}

	ruleArgs := []string{fmt.Sprintf("-p %s", protocol)}
	ruleArgs = append(ruleArgs, fmt.Sprintf("--dport %s", portSpec), fmt.Sprintf("-j %s", action))
	ruleSpec := strings.Join(ruleArgs, " ")
	if operation == "add" {
		if err := iptables.AddRule(iptables.FilterTab, port.Chain, ruleSpec); err != nil {
			return err
		}
	} else {
		if err := iptables.DeleteRule(iptables.FilterTab, port.Chain, ruleSpec); err != nil {
			return err
		}
	}

	name := iptables.BasicFileName
	if port.Chain == iptables.Chain1PanelBasicBefore {
		name = iptables.BasicBeforeFileName
	}
	if err := iptables.SaveRulesToFile(iptables.FilterTab, port.Chain, name); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelBasic, err)
	}
	return nil
}

func (i *Iptables) RichRules(rule FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}
	if len(rule.Chain) == 0 {
		rule.Chain = iptables.Chain1PanelBasic
	}

	address := strings.TrimSpace(rule.Address)
	if strings.EqualFold(address, "Anywhere") {
		address = ""
	}

	action := "ACCEPT"
	if rule.Strategy == "drop" {
		action = "DROP"
	}

	var ruleArgs []string
	if address != "" {
		ruleArgs = append(ruleArgs, fmt.Sprintf("-s %s", address))
	}

	protocol := strings.TrimSpace(rule.Protocol)
	if rule.Port != "" && protocol == "" {
		protocol = "tcp"
	}

	if protocol != "" {
		ruleArgs = append(ruleArgs, fmt.Sprintf("-p %s", protocol))
	}

	if rule.Port != "" {
		portSegment, err := normalizePortSpec(rule.Port)
		if err != nil {
			return err
		}
		if protocol == "" {
			return fmt.Errorf("protocol is required when specifying a port")
		}
		ruleArgs = append(ruleArgs, fmt.Sprintf("--dport %s", portSegment))
	}

	ruleArgs = append(ruleArgs, fmt.Sprintf("-j %s", action))
	ruleSpec := strings.Join(ruleArgs, " ")
	if operation == "add" {
		if err := iptables.AddRule(iptables.FilterTab, rule.Chain, ruleSpec); err != nil {
			return err
		}
	} else {
		if err := iptables.DeleteRule(iptables.FilterTab, rule.Chain, ruleSpec); err != nil {
			return err
		}
	}

	name := iptables.BasicFileName
	if rule.Chain == iptables.Chain1PanelBasicBefore {
		name = iptables.BasicBeforeFileName
	}
	if err := iptables.SaveRulesToFile(iptables.FilterTab, rule.Chain, name); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelBasic, err)
	}
	return nil
}

func (i *Iptables) PortForward(info Forward, operation string) error {
	return iptablesPortForward(info, operation)
}

func (i *Iptables) EnableForward() error {
	return EnableIptablesForward()
}

func (i *Iptables) ListForward() ([]FireInfo, error) {
	return iptablesListForward()
}

func EnableIptablesForward() error {
	if err := cmd.RunDefaultBashC("echo 1 > /proc/sys/net/ipv4/ip_forward"); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	_ = cmd.RunDefaultBashC("grep -q '^net.ipv4.ip_forward' /etc/sysctl.conf || echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf")
	_ = cmd.RunDefaultBashC("sysctl -p")

	if err := iptables.AddChainWithAppend(iptables.NatTab, "PREROUTING", iptables.Chain1PanelPreRouting); err != nil {
		return err
	}
	if err := iptables.AddChainWithAppend(iptables.NatTab, "POSTROUTING", iptables.Chain1PanelPostRouting); err != nil {
		return err
	}
	if err := iptables.AddChainWithAppend(iptables.FilterTab, "FORWARD", iptables.Chain1PanelForward); err != nil {
		return err
	}

	return nil
}

func iptablesPortForward(info Forward, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}
	if info.Protocol == "" || info.Port == "" || info.TargetPort == "" {
		return fmt.Errorf("protocol, port, and target port are required")
	}
	if operation == "add" {
		if err := iptables.AddForward(info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface, true); err != nil {
			return err
		}
	} else {
		if err := iptables.DeleteForward(info.Num, info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface); err != nil {
			return err
		}
	}
	forwardPersistence()
	return nil
}

func forwardPersistence() {
	if err := iptables.SaveRulesToFile(iptables.FilterTab, iptables.Chain1PanelForward, iptables.ForwardFileName); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelForward, err)
	}
	if err := iptables.SaveRulesToFile(iptables.NatTab, iptables.Chain1PanelPreRouting, iptables.ForwardFileName1); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelPreRouting, err)
	}
	if err := iptables.SaveRulesToFile(iptables.NatTab, iptables.Chain1PanelPostRouting, iptables.ForwardFileName2); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelPostRouting, err)
	}
}

func iptablesListForward() ([]FireInfo, error) {
	natList, err := iptables.ListForward(iptables.Chain1PanelPreRouting)
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT rules: %w", err)
	}

	var datas []FireInfo
	for _, nat := range natList {
		datas = append(datas, FireInfo{
			Num:        nat.Num,
			Protocol:   nat.Protocol,
			Port:       strings.TrimPrefix(nat.SrcPort, ":"),
			TargetIP:   nat.Destination,
			TargetPort: strings.TrimPrefix(nat.DestPort, ":"),
			Interface:  nat.InIface,
		})
	}

	return datas, nil
}

func parsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port out of range: %d", port)
	}
	return port, nil
}

func normalizePortSpec(port string) (string, error) {
	value := strings.TrimSpace(port)
	if value == "" {
		return "", fmt.Errorf("port is required")
	}

	separator := ""
	if strings.Contains(value, "-") {
		separator = "-"
	} else if strings.Contains(value, ":") {
		separator = ":"
	}

	if separator != "" {
		parts := strings.Split(value, separator)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid port range: %s", port)
		}
		start, err := parsePort(strings.TrimSpace(parts[0]))
		if err != nil {
			return "", err
		}
		end, err := parsePort(strings.TrimSpace(parts[1]))
		if err != nil {
			return "", err
		}
		if start > end {
			return "", fmt.Errorf("invalid port range: %d-%d", start, end)
		}
		return fmt.Sprintf("%d:%d", start, end), nil
	}

	single, err := parsePort(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", single), nil
}
