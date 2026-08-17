package nftables_helper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const requiredPortComment = "1Panel Port Whitelist"

type Manager struct {
	UpdateSetting     func(key, value string) error
	LoadRequiredPorts func() ([]firewall.PortWhitelist, error)
}

func (m *Manager) Operate(operation firewall.BaseOperation) error {
	switch operation {
	case firewall.BaseOperationInit, firewall.BaseOperationBind:
		return m.enableBase(true)
	case firewall.BaseOperationBindWithoutInit:
		return m.enableBase(false)
	case firewall.BaseOperationUnbind:
		return m.disableBase()
	default:
		return fmt.Errorf("unsupported nftables base operation %q", operation)
	}
}

func (m *Manager) enableBase(prepare bool) error {
	if prepare {
		if err := m.ensureBaseChains(); err != nil {
			return err
		}
		if err := m.initPreRules(); err != nil {
			return err
		}
	}
	if err := Bind(); err != nil {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusEnable)
}

func (m *Manager) disableBase() error {
	if err := Unbind(); err != nil {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusDisable)
}

func (m *Manager) ensureBaseChains() error {
	hasTable := false
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		if _, err := run("list", "table", TableFamily(family), TableName); err == nil {
			hasTable = true
			break
		}
	}
	if !hasTable {
		if err := rejectLegacyOnePanelChains(); err != nil {
			return err
		}
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		if _, err := run("list", "table", tableFamily, TableName); err != nil {
			if _, createErr := run("add", "table", tableFamily, TableName); createErr != nil {
				return fmt.Errorf("create 1Panel nftables %s table: %w", tableFamily, createErr)
			}
		}
		if _, err := run("list", "chain", tableFamily, TableName, InputChain); err != nil {
			if err := runCommand(
				"add", "chain", tableFamily, TableName, InputChain,
				"{", "type", "filter", "hook", "input", "priority", "filter", ";", "policy", "accept", ";", "}",
			); err != nil {
				return fmt.Errorf("create 1Panel nftables %s input chain: %w", tableFamily, err)
			}
		}
		for _, nativeChain := range BasicChains() {
			if _, err := run("list", "chain", tableFamily, TableName, nativeChain); err == nil {
				continue
			}
			if _, err := run("add", "chain", tableFamily, TableName, nativeChain); err != nil {
				return fmt.Errorf("create nftables %s chain %s: %w", tableFamily, nativeChain, err)
			}
		}
	}
	return nil
}

func rejectLegacyOnePanelChains() error {
	for _, family := range []string{"ip", "ip6"} {
		stdout, err := run("list", "chain", family, "filter", "INPUT")
		if err == nil {
			if chain := activeLegacyChain(stdout); chain != "" {
				return fmt.Errorf("legacy 1Panel firewall chain %s is still bound to %s filter INPUT; unbind or migrate it before initializing native nftables", chain, family)
			}
		}
	}
	for _, executable := range []string{"iptables", "iptables-nft", "ip6tables", "ip6tables-nft"} {
		if !cmd.Which(executable) {
			continue
		}
		stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithOptionalSudoAndStdout(executable, "-t", "filter", "-S", "INPUT")
		if err != nil {
			continue
		}
		if chain := activeLegacyChain(stdout); chain != "" {
			return fmt.Errorf("legacy 1Panel firewall chain %s is still bound through %s INPUT; unbind or migrate it before initializing native nftables", chain, executable)
		}
	}
	return nil
}

func activeLegacyChain(output string) string {
	legacy := map[string]struct{}{
		"1PANEL_BASIC_BEFORE": {}, "1PANEL_BASIC": {}, "1PANEL_BASIC_AFTER": {},
	}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		for index := 0; index+1 < len(fields); index++ {
			if fields[index] != "jump" && fields[index] != "-j" {
				continue
			}
			chain := strings.Trim(fields[index+1], `"`)
			if _, ok := legacy[chain]; ok {
				return chain
			}
		}
	}
	return ""
}

func (m *Manager) initPreRules() error {
	ports, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	ports, err = validateRequiredPorts(ports)
	if err != nil {
		return err
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		if err := runCommand("flush", "chain", tableFamily, TableName, BasicBeforeChain); err != nil {
			return err
		}
		if err := runCommand("add", "rule", tableFamily, TableName, BasicBeforeChain, "iifname", `"lo"`, "accept", "comment", `"Loopback Whitelist"`); err != nil {
			return err
		}
		if err := runCommand("add", "rule", tableFamily, TableName, BasicBeforeChain, "ct", "state", "{", "established,related", "}", "accept", "comment", `"ESTABLISHED Whitelist"`); err != nil {
			return err
		}
		for _, port := range ports {
			if err := addRequiredPortRule(tableFamily, port); err != nil {
				return err
			}
		}
		if err := runCommand("flush", "chain", tableFamily, TableName, BasicAfterChain); err != nil {
			return err
		}
		if err := runCommand("add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "tcp", "drop"); err != nil {
			return err
		}
		if err := runCommand("add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "udp", "drop"); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) SyncRequiredPorts() error {
	ports, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	ports, err = validateRequiredPorts(ports)
	if err != nil {
		return err
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		stdout, err := run("-n", "-a", "list", "chain", tableFamily, TableName, BasicBeforeChain)
		if err != nil {
			return err
		}
		existing := requiredPortRules(stdout)
		missing, staleHandles := requiredPortChanges(existing, ports)
		for _, port := range missing {
			if err := addRequiredPortRule(tableFamily, port); err != nil {
				return err
			}
		}
		for _, handle := range staleHandles {
			if err := runCommand("delete", "rule", tableFamily, TableName, BasicBeforeChain, "handle", handle); err != nil {
				return err
			}
		}
	}
	if err := PersistRuleset(context.Background()); err != nil {
		return err
	}
	return nil
}

func addRequiredPortRule(tableFamily string, port firewall.PortWhitelist) error {
	return runCommand(
		"add", "rule", tableFamily, TableName, BasicBeforeChain,
		"meta", "l4proto", port.Protocol, port.Protocol, "dport", port.Port,
		"accept", "comment", `"`+requiredPortComment+`"`,
	)
}

func validateRequiredPorts(ports []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
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

type requiredPortRule struct {
	Key    string
	Handle string
}

func requiredPortChanges(existing []requiredPortRule, desiredPorts []firewall.PortWhitelist) ([]firewall.PortWhitelist, []string) {
	desired := firewall.PortWhitelistMap(desiredPorts)
	existingKeys := make(map[string]struct{}, len(existing))
	for _, rule := range existing {
		existingKeys[rule.Key] = struct{}{}
	}
	missing := make([]firewall.PortWhitelist, 0)
	for _, port := range desiredPorts {
		if _, exists := existingKeys[firewall.PortWhitelistKey(port)]; !exists {
			missing = append(missing, port)
		}
	}
	kept := make(map[string]struct{}, len(existing))
	staleHandles := make([]string, 0)
	for _, rule := range existing {
		if _, wanted := desired[rule.Key]; wanted {
			if _, alreadyKept := kept[rule.Key]; !alreadyKept {
				kept[rule.Key] = struct{}{}
				continue
			}
		}
		staleHandles = append(staleHandles, rule.Handle)
	}
	return missing, staleHandles
}

func requiredPortRules(output string) []requiredPortRule {
	rules := make([]requiredPortRule, 0)
	marker := `comment "` + requiredPortComment + `"`
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, marker) {
			continue
		}
		handleIndex := strings.LastIndex(line, "# handle ")
		if handleIndex < 0 {
			continue
		}
		handle := strings.TrimSpace(line[handleIndex+len("# handle "):])
		if _, err := strconv.ParseUint(handle, 10, 64); err != nil {
			continue
		}
		fields := strings.Fields(line[:handleIndex])
		for index := 0; index+2 < len(fields); index++ {
			protocol := fields[index]
			if (protocol != "tcp" && protocol != "udp") || fields[index+1] != "dport" {
				continue
			}
			port, err := strconv.Atoi(fields[index+2])
			if err != nil || port < 1 || port > 65535 {
				break
			}
			rules = append(rules, requiredPortRule{
				Key: firewall.PortWhitelistKey(firewall.PortWhitelist{Protocol: protocol, Port: strconv.Itoa(port)}), Handle: handle,
			})
			break
		}
	}
	return rules
}

func (m *Manager) updateSetting(key, value string) error {
	if m != nil && m.UpdateSetting != nil {
		return m.UpdateSetting(key, value)
	}
	return nil
}

func (m *Manager) loadRequiredPorts() ([]firewall.PortWhitelist, error) {
	if m != nil && m.LoadRequiredPorts != nil {
		return m.LoadRequiredPorts()
	}
	return nil, fmt.Errorf("load required firewall ports is not configured")
}

func Bind() error {
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		if _, err := run("list", "chain", tableFamily, TableName, InputChain); err != nil {
			return fmt.Errorf("1Panel nftables %s input chain is not initialized: %w", tableFamily, err)
		}
	}
	if err := flushInputChains(); err != nil {
		return err
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		for _, chain := range BasicChains() {
			if err := runCommand("add", "rule", tableFamily, TableName, InputChain, "jump", chain); err != nil {
				cleanupErr := flushInputChains()
				return errors.Join(err, cleanupErr)
			}
		}
	}
	return PersistRuleset(context.Background())
}

func Unbind() error {
	if err := flushInputChains(); err != nil {
		return err
	}
	return PersistRuleset(context.Background())
}

func flushInputChains() error {
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		if err := runCommand("flush", "chain", TableFamily(family), TableName, InputChain); err != nil {
			return err
		}
	}
	return nil
}
