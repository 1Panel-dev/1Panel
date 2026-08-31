package nftables_helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const requiredPortComment = "1Panel Port Whitelist"

type Manager struct {
	UpdateSetting     func(key, value string) error
	LoadRequiredPorts func() ([]firewall.PortWhitelist, error)
}

func (m *Manager) Cleanup() error {
	commands := make([][]string, 0, 2)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		if _, err := run("list", "table", tableFamily, TableName); err != nil {
			continue
		}
		commands = append(commands, []string{"delete", "table", tableFamily, TableName})
	}
	if err := runBatch(commands...); err != nil {
		return err
	}
	file := filepath.Join(global.Dir.FirewallDir, RulesFile)
	if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusDisable)
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
	commands := make([][]string, 0, 10)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		tableExists := true
		if _, err := run("list", "table", tableFamily, TableName); err != nil {
			tableExists = false
			commands = append(commands, []string{"add", "table", tableFamily, TableName})
		}
		if !tableExists {
			commands = append(commands, []string{
				"add", "chain", tableFamily, TableName, InputChain,
				"{", "type", "filter", "hook", "input", "priority", "0", ";", "policy", "accept", ";", "}",
			})
		} else if _, err := run("list", "chain", tableFamily, TableName, InputChain); err != nil {
			commands = append(commands, []string{
				"add", "chain", tableFamily, TableName, InputChain,
				"{", "type", "filter", "hook", "input", "priority", "0", ";", "policy", "accept", ";", "}",
			})
		}
		for _, nativeChain := range BasicChains() {
			if tableExists {
				if _, err := run("list", "chain", tableFamily, TableName, nativeChain); err == nil {
					continue
				}
			}
			commands = append(commands, []string{"add", "chain", tableFamily, TableName, nativeChain})
		}
	}
	if err := runBatch(commands...); err != nil {
		return fmt.Errorf("batch create 1Panel nftables base chains: %w", err)
	}
	return nil
}

func requiredPortCommand(tableFamily string, port firewall.PortWhitelist) []string {
	return []string{
		"add", "rule", tableFamily, TableName, BasicBeforeChain,
		"meta", "l4proto", port.Protocol, port.Protocol, "dport", port.Port,
		"accept", "comment", `"` + requiredPortComment + `"`,
	}
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
	commands := make([][]string, 0, 12+len(ports)*2)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		commands = append(commands,
			[]string{"flush", "chain", tableFamily, TableName, BasicBeforeChain},
			[]string{"add", "rule", tableFamily, TableName, BasicBeforeChain, "iifname", `"lo"`, "accept", "comment", `"Loopback Whitelist"`},
			[]string{"add", "rule", tableFamily, TableName, BasicBeforeChain, "ct", "state", "{", "established,related", "}", "accept", "comment", `"ESTABLISHED Whitelist"`},
		)
		for _, port := range ports {
			commands = append(commands, requiredPortCommand(tableFamily, port))
		}
		commands = append(commands,
			[]string{"flush", "chain", tableFamily, TableName, BasicAfterChain},
			[]string{"add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "tcp", "drop"},
			[]string{"add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "udp", "drop"},
		)
	}
	return runBatch(commands...)
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
	commands := make([][]string, 0)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		stdout, err := run("-n", "-a", "list", "chain", tableFamily, TableName, BasicBeforeChain)
		if err != nil {
			return err
		}
		existing := requiredPortRules(stdout)
		missing, staleHandles := requiredPortChanges(existing, ports)
		for _, port := range missing {
			commands = append(commands, requiredPortCommand(tableFamily, port))
		}
		for _, handle := range staleHandles {
			commands = append(commands, []string{"delete", "rule", tableFamily, TableName, BasicBeforeChain, "handle", handle})
		}
	}
	if err := runBatch(commands...); err != nil {
		return err
	}
	if err := PersistRuleset(context.Background()); err != nil {
		return err
	}
	return nil
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
	commands := make([][]string, 0, 8)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		commands = append(commands, []string{"flush", "chain", tableFamily, TableName, InputChain})
		for _, chain := range BasicChains() {
			commands = append(commands, []string{"add", "rule", tableFamily, TableName, InputChain, "jump", chain})
		}
	}
	if err := runBatch(commands...); err != nil {
		cleanupErr := flushInputChains()
		return errors.Join(err, cleanupErr)
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
	commands := make([][]string, 0, 2)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		commands = append(commands, []string{"flush", "chain", TableFamily(family), TableName, InputChain})
	}
	return runBatch(commands...)
}
