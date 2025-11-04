package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

var portRuleRegex = regexp.MustCompile(`-A\s+INPUT\s+-p\s+(\w+)(?:\s+-m\s+\w+)*\s+--dport\s+(\d+(?::\d+)?)\s+-j\s+(\w+)`)
var addressRuleRegex = regexp.MustCompile(`-A\s+(INPUT|OUTPUT)\s+-s\s+(\S+)\s+-j\s+(\w+)`)

// IptablesFirewall wraps Iptables to implement FirewallClient interface
type IptablesFirewall struct {
	iptables *Iptables
}

// NewIptablesFirewall creates a new IptablesFirewall instance
func NewIptablesFirewall() (*IptablesFirewall, error) {
	iptables, err := NewIptables()
	if err != nil {
		return nil, err
	}
	return &IptablesFirewall{
		iptables: iptables,
	}, nil
}

func (ifw *IptablesFirewall) Name() string {
	return "iptables"
}

func (ifw *IptablesFirewall) Status() (bool, error) {
	// Check if iptables has any rules (indicates it's running)
	stdout, err := cmd.RunDefaultWithStdoutBashC("iptables -L -n | head -1")
	if err != nil {
		return false, err
	}
	// If we can list rules, iptables is functional
	return strings.Contains(stdout, "Chain"), nil
}

func (ifw *IptablesFirewall) Start() error {
	// Try different service names
	services := []string{"iptables", "iptables-services"}
	var lastErr error

	for _, svc := range services {
		if err := controller.HandleStart(svc); err == nil {
			break
		} else {
			lastErr = err
		}
	}

	// If systemd service doesn't work, try loading kernel modules
	if err := cmd.RunDefaultBashC("modprobe ip_tables"); err != nil {
		global.LOG.Warnf("failed to load ip_tables module: %v", err)
	}

	if lastErr != nil {
		global.LOG.Warnf("iptables service management not available: %v", lastErr)
	}

	// Use unified chain setup method
	if err := ifw.iptables.Setup1PanelFirewallChains("input"); err != nil {
		return fmt.Errorf("failed to setup 1Panel firewall chains: %w", err)
	}

	global.LOG.Infof("1Panel firewall chains setup completed")
	return nil
}

func (ifw *IptablesFirewall) Stop() error {
	global.LOG.Info("Stopping 1Panel firewall - removing chains from INPUT")

	// Remove jump rule from INPUT to 1PANEL_INPUT
	err := ifw.iptables.run(FilterTab, fmt.Sprintf("-D INPUT -j %s", Chain1PanelInput))
	if err != nil {
		global.LOG.Warnf("failed to detach 1PANEL_INPUT from INPUT: %v", err)
	}

	// Remove jump rule from INPUT to 1PANEL_BASIC
	err = ifw.iptables.run(FilterTab, fmt.Sprintf("-D INPUT -j %s", Chain1PanelBasic))
	if err != nil {
		global.LOG.Warnf("failed to detach 1PANEL_BASIC from INPUT: %v", err)
	}

	// Flush the 1PANEL_INPUT chain (clear all rules but keep chain)
	err = ifw.iptables.run(FilterTab, fmt.Sprintf("-F %s", Chain1PanelInput))
	if err != nil {
		global.LOG.Warnf("failed to flush 1PANEL_INPUT chain: %v", err)
	}

	// Flush the 1PANEL_BASIC chain (clear all rules but keep chain)
	err = ifw.iptables.run(FilterTab, fmt.Sprintf("-F %s", Chain1PanelBasic))
	if err != nil {
		global.LOG.Warnf("failed to flush 1PANEL_BASIC chain: %v", err)
	}

	// Keep chains (do not delete) so data can still be read from them
	global.LOG.Info("1Panel firewall chains flushed and detached")

	return nil
}

func (ifw *IptablesFirewall) Restart() error {
	// Restart is disabled - frontend will remove this functionality
	// No operation needed for iptables
	global.LOG.Info("Restart is a no-op for iptables")
	return nil
}

func (ifw *IptablesFirewall) Reload() error {
	return ifw.iptables.Reload()
}

func (ifw *IptablesFirewall) Version() (string, error) {
	return ifw.iptables.GetVersion()
}

func (ifw *IptablesFirewall) ListPort() ([]FireInfo, error) {
	stdout, err := ifw.iptables.out(FilterTab, fmt.Sprintf("-S %s", Chain1PanelBasic))
	if err != nil {
		return nil, fmt.Errorf("failed to list 1PANEL_BASIC rules: %w", err)
	}

	var datas []FireInfo
	lines := strings.Split(stdout, "\n")

	// Updated regex to match 1PANEL_BASIC chain
	chainPortRegex := regexp.MustCompile(fmt.Sprintf(`-A\s+%s\s+(?:-s\s+(\S+)\s+)?-p\s+(\w+)(?:\s+-m\s+\w+)*\s+--dport\s+(\d+(?::\d+)?)\s+-j\s+(\w+)`, Chain1PanelBasic))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, fmt.Sprintf("-A %s", Chain1PanelBasic)) {
			continue
		}

		// Parse rules like: -A 1PANEL_BASIC -p tcp --dport 80 -j ACCEPT
		if matches := chainPortRegex.FindStringSubmatch(line); len(matches) == 5 {
			address := matches[1]
			protocol := matches[2]
			port := strings.ReplaceAll(matches[3], ":", "-")
			action := strings.ToLower(matches[4])

			strategy := "accept"
			if action == "drop" || action == "reject" {
				strategy = "drop"
			}

			datas = append(datas, FireInfo{
				Address:  address,
				Protocol: protocol,
				Port:     port,
				Strategy: strategy,
				Family:   "ipv4",
			})
		}
	}

	return datas, nil
}

func (ifw *IptablesFirewall) ListForward() ([]FireInfo, error) {
	natList, err := ifw.iptables.NatList(PreRoutingChain)
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

func (ifw *IptablesFirewall) ListAddress() ([]FireInfo, error) {
	stdout, err := ifw.iptables.out(FilterTab, fmt.Sprintf("-S %s", Chain1PanelBasic))
	if err != nil {
		return nil, fmt.Errorf("failed to list 1PANEL_BASIC address rules: %w", err)
	}

	var datas []FireInfo
	lines := strings.Split(stdout, "\n")
	addressMap := make(map[string]FireInfo) // Deduplicate by address

	// Updated regex to match both source and destination addresses in 1PANEL_BASIC chain
	chainAddressRegex := regexp.MustCompile(fmt.Sprintf(`-A\s+%s\s+(?:-s\s+(\S+)|(?:-d\s+(\S+)))?\s+-j\s+(\w+)`, Chain1PanelBasic))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, fmt.Sprintf("-A %s", Chain1PanelBasic)) {
			continue
		}

		// Parse rules like: -A 1PANEL_BASIC -s 192.168.1.1 -j DROP
		if matches := chainAddressRegex.FindStringSubmatch(line); len(matches) >= 4 {
			// matches[1] is source address (-s), matches[2] is dest address (-d), matches[3] is action
			address := matches[1]
			if address == "" {
				address = matches[2]
			}
			if address == "" {
				continue
			}
			action := strings.ToLower(matches[3])

			strategy := "accept"
			if action == "drop" || action == "reject" {
				strategy = "drop"
			}

			// Use address as key to deduplicate
			if _, exists := addressMap[address]; !exists {
				addressMap[address] = FireInfo{
					Address:  address,
					Strategy: strategy,
					Family:   "ipv4",
				}
			}
		}
	}

	// Convert map to slice
	for _, info := range addressMap {
		datas = append(datas, info)
	}

	return datas, nil
}

func (ifw *IptablesFirewall) Port(port FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}

	// Validate port
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

	opFlag := "-A"
	if operation == "remove" {
		opFlag = "-D"
	}

	args := []string{fmt.Sprintf("%s %s", opFlag, Chain1PanelBasic), fmt.Sprintf("-p %s", protocol)}
	if protocol == "tcp" || protocol == "udp" {
		args = append(args, fmt.Sprintf("-m %s", protocol))
	}
	args = append(args, fmt.Sprintf("--dport %s", portSpec), fmt.Sprintf("-j %s", action))

	return ifw.iptables.run(FilterTab, strings.Join(args, " "))
}

func (ifw *IptablesFirewall) RichRules(rule FireInfo, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}

	address := strings.TrimSpace(rule.Address)
	if strings.EqualFold(address, "Anywhere") {
		address = ""
	}

	action := "ACCEPT"
	if rule.Strategy == "drop" {
		action = "DROP"
	}

	var ruleStr string
	opFlag := "-A"
	if operation == "remove" {
		opFlag = "-D"
	}

	args := []string{fmt.Sprintf("%s %s", opFlag, Chain1PanelBasic)}
	if address != "" {
		args = append(args, fmt.Sprintf("-s %s", address))
	}

	protocol := strings.TrimSpace(rule.Protocol)
	if rule.Port != "" && protocol == "" {
		protocol = "tcp"
	}

	if protocol != "" {
		args = append(args, fmt.Sprintf("-p %s", protocol))
	}

	if rule.Port != "" {
		portSegment, err := normalizePortSpec(rule.Port)
		if err != nil {
			return err
		}
		if protocol == "" {
			return fmt.Errorf("protocol is required when specifying a port")
		}
		if protocol == "tcp" || protocol == "udp" {
			args = append(args, fmt.Sprintf("-m %s", protocol))
		}
		args = append(args, fmt.Sprintf("--dport %s", portSegment))
	}

	args = append(args, fmt.Sprintf("-j %s", action))
	ruleStr = strings.Join(args, " ")
	return ifw.iptables.run(FilterTab, ruleStr)
}

func (ifw *IptablesFirewall) PortForward(info Forward, operation string) error {
	if operation != "add" && operation != "remove" {
		return buserr.New("ErrCmdIllegal")
	}

	// Validate inputs
	if info.Protocol == "" || info.Port == "" || info.TargetPort == "" {
		return fmt.Errorf("protocol, port, and target port are required")
	}

	if operation == "add" {
		// Use existing NatAdd method
		return ifw.iptables.NatAdd(info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface, true)
	} else {
		// For remove, we need to find the rule number first
		natList, err := ifw.iptables.NatList(PreRoutingChain)
		if err != nil {
			return fmt.Errorf("failed to list NAT rules: %w", err)
		}

		// Find matching rule
		for _, nat := range natList {
			if nat.Protocol == info.Protocol &&
				strings.TrimPrefix(nat.SrcPort, ":") == info.Port &&
				strings.TrimPrefix(nat.DestPort, ":") == info.TargetPort {

				targetIP := info.TargetIP
				if targetIP == "" {
					targetIP = "127.0.0.1"
				}

				return ifw.iptables.NatRemove(nat.Num, info.Protocol, info.Port, targetIP, info.TargetPort, info.Interface)
			}
		}

		return fmt.Errorf("forward rule not found")
	}
}

func (ifw *IptablesFirewall) EnableForward() error {
	// Enable IP forwarding
	if err := cmd.RunDefaultBashC("echo 1 > /proc/sys/net/ipv4/ip_forward"); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	// Make it persistent (if sysctl.conf exists)
	_ = cmd.RunDefaultBashC("grep -q '^net.ipv4.ip_forward' /etc/sysctl.conf || echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf")
	_ = cmd.RunDefaultBashC("sysctl -p")

	// Setup NAT and Forward chains
	if err := ifw.iptables.ensureChainWithJump(NatTab, "PREROUTING", PreRoutingChain); err != nil {
		return err
	}
	if err := ifw.iptables.ensureChainWithJump(NatTab, "POSTROUTING", PostRoutingChain); err != nil {
		return err
	}
	if err := ifw.iptables.ensureChainWithJump(FilterTab, "FORWARD", ForwardChain); err != nil {
		return err
	}

	return nil
}

// Helper function to parse port number
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
