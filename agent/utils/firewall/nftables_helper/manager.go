package nftables_helper

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
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

func requiredPortCommand(tableFamily string, rule firewall.SystemPort) []string {
	command := []string{
		"insert", "rule", tableFamily, TableName, BasicBeforeChain,
	}
	if rule.SourceAddress != "" {
		command = append(command, tableFamily, "saddr", rule.SourceAddress)
	}
	return append(command, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.Port,
		"accept", "comment", `"`+requiredPortComment+`"`)
}

func (m *Manager) initPreRules() error {
	ports, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	ports, err = firewall.NormalizeRequiredPorts(ports)
	if err != nil {
		return err
	}
	rules := firewall.ExpandPortWhitelist(ports)
	var commands [][]string
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		output, _, err := readNftObject(run, "-n", "list", "chain", tableFamily, TableName, BasicBeforeChain)
		if err != nil {
			return err
		}
		candidates := [][]string{
			{"add", "rule", tableFamily, TableName, BasicBeforeChain, "iifname", `"lo"`, "accept", "comment", `"Loopback Whitelist"`},
			{"add", "rule", tableFamily, TableName, BasicBeforeChain, "ct", "state", "{", "established,related", "}", "accept", "comment", `"ESTABLISHED Whitelist"`},
		}
		for _, rule := range rules {
			if rule.Family == string(family) {
				candidates = append(candidates, requiredPortCommand(tableFamily, rule))
			}
		}
		for _, command := range candidates {
			expression := strings.Join(command[5:], " ")
			if !containsRequiredPortRule(output, expression) {
				commands = append(commands, command)
				output += "\n" + expression
			}
		}
		commands = append(commands,
			[]string{"flush", "chain", tableFamily, TableName, BasicAfterChain},
			[]string{"add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "tcp", "drop"},
			[]string{"add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "udp", "drop"},
		)
	}
	return runBatch(commands...)
}

func containsRequiredPortRule(output, expression string) bool {
	canonical := func(line string) string {
		line, _, _ = strings.Cut(line, " comment ")
		line, _, _ = strings.Cut(line, " # handle ")
		for _, protocol := range []string{"tcp", "udp"} {
			line = strings.ReplaceAll(line, "meta l4proto "+protocol+" ", "")
		}
		line = strings.NewReplacer("{", "", "}", "", ", ", ",", " ,", ",").Replace(line)
		fields := strings.Fields(line)
		for index, field := range fields {
			if index >= 2 && fields[index-2] == "ct" && fields[index-1] == "state" {
				states := strings.Split(field, ",")
				for i, state := range states {
					value, err := strconv.ParseUint(state, 0, 64)
					if err != nil {
						continue
					}
					switch value {
					case 1:
						states[i] = "invalid"
					case 2:
						states[i] = "established"
					case 4:
						states[i] = "related"
					case 8:
						states[i] = "new"
					case 64:
						states[i] = "untracked"
					}
				}
				slices.Sort(states)
				fields[index] = strings.Join(states, ",")
			}
			if prefix, err := netip.ParsePrefix(field); err == nil {
				prefix = prefix.Masked()
				fields[index] = prefix.String()
				if prefix.Bits() == prefix.Addr().BitLen() {
					fields[index] = prefix.Addr().String()
				}
			}
		}
		return strings.Join(fields, " ")
	}
	wanted := canonical(expression)
	for _, line := range strings.Split(output, "\n") {
		if canonical(line) == wanted {
			return true
		}
	}
	return false
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
