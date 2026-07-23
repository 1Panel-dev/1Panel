package client

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("iptables", "-L", "-n")
	if err != nil {
		return false, err
	}
	firstLine := strings.Split(strings.TrimSpace(stdout), "\n")[0]
	return strings.Contains(firstLine, "Chain"), nil
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
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("iptables", "--version")
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

	ruleArgs, err := buildIptablesPortRuleArgs(port)
	if err != nil {
		return err
	}
	if operation == "add" {
		if err := iptables.AddRule(iptables.FilterTab, port.Chain, ruleArgs...); err != nil {
			return err
		}
	} else {
		if err := iptables.DeleteRule(iptables.FilterTab, port.Chain, ruleArgs...); err != nil {
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

	ruleArgs, err := buildIptablesRichRuleArgs(rule)
	if err != nil {
		return err
	}
	if operation == "add" {
		if err := iptables.AddRule(iptables.FilterTab, rule.Chain, ruleArgs...); err != nil {
			return err
		}
	} else {
		if err := iptables.DeleteRule(iptables.FilterTab, rule.Chain, ruleArgs...); err != nil {
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

func (i *Iptables) ExpandPortRule(rule FireInfo) []PortUnit {
	chain := rule.Chain
	if len(chain) == 0 {
		chain = iptables.Chain1PanelBasic
	}
	return expandPortRule(rule, chain)
}

func (i *Iptables) ApplyPortUnit(unit PortUnit, operation string) error {
	apply := unit.Apply
	apply.Chain = unit.Chain
	if needsRichRule(apply) {
		return i.RichRules(apply, operation)
	}
	return i.Port(apply, operation)
}

func (i *Iptables) ExpandAddressRule(rule FireInfo) []AddressUnit {
	return expandAddressRule(rule, iptables.Chain1PanelBasic)
}

func (i *Iptables) ApplyAddressUnit(unit AddressUnit, operation string) error {
	apply := unit.Apply
	apply.Chain = unit.Chain
	return i.RichRules(apply, operation)
}

func (i *Iptables) AddPortWhiteList(list PortWhiteList) error {
	list.Previous = nil
	return i.SyncPortWhiteList(list)
}

func (i *Iptables) SyncPortWhiteList(list PortWhiteList) error {
	if isInit, _ := iptables.LoadInitStatus("iptables", "base"); !isInit {
		return nil
	}
	return SyncIptablesPortWhiteList(list, true)
}

func buildIptablesPortRuleArgs(port FireInfo) ([]string, error) {
	portSpec, err := normalizePortSpec(port.Port)
	if err != nil {
		return nil, err
	}
	protocol := port.Protocol
	if protocol == "" {
		protocol = "tcp"
	}
	action := "ACCEPT"
	if port.Strategy == "drop" {
		action = "DROP"
	}
	return []string{"-p", protocol, "--dport", portSpec, "-j", action}, nil
}

func buildIptablesRichRuleArgs(rule FireInfo) ([]string, error) {
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
		ruleArgs = append(ruleArgs, "-s", address)
	}
	protocol := strings.TrimSpace(rule.Protocol)
	if rule.Port != "" && protocol == "" {
		protocol = "tcp"
	}
	if protocol != "" {
		ruleArgs = append(ruleArgs, "-p", protocol)
	}
	if rule.Port != "" {
		portSegment, err := normalizePortSpec(rule.Port)
		if err != nil {
			return nil, err
		}
		if protocol == "" {
			return nil, fmt.Errorf("protocol is required when specifying a port")
		}
		ruleArgs = append(ruleArgs, "--dport", portSegment)
	}
	ruleArgs = append(ruleArgs, "-j", action)
	return ruleArgs, nil
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
