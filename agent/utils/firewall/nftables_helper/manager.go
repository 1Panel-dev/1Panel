package nftables_helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const requiredPortComment = "1Panel Port Whitelist"

func Cleanup() error {
	commands := make([][]string, 0, 2)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		_, exists, err := ReadTable(run, tableFamily, TableName)
		if err != nil {
			return err
		}
		if !exists {
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
	return nil
}

func Operate(operation firewall.BaseOperation, requiredPorts []firewall.PortWhitelist) error {
	switch operation {
	case firewall.BaseOperationInit, firewall.BaseOperationBind:
		return enableBase(true, requiredPorts)
	case firewall.BaseOperationBindWithoutInit:
		return enableBase(false, requiredPorts)
	case firewall.BaseOperationUnbind:
		return Unbind()
	default:
		return fmt.Errorf("unsupported nftables base operation %q", operation)
	}
}

func enableBase(prepare bool, requiredPorts []firewall.PortWhitelist) error {
	if prepare {
		if err := ensureBaseChains(); err != nil {
			return err
		}
		if err := initPreRules(requiredPorts); err != nil {
			return err
		}
	}
	if err := Bind(); err != nil {
		return err
	}
	return nil
}

func ensureBaseChains(families ...filter.Family) error {
	if len(families) == 0 {
		families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
	}
	commands := make([][]string, 0, 10)
	for _, family := range families {
		tableFamily := TableFamily(family)
		output, tableExists, err := ReadTable(run, tableFamily, TableName)
		if err != nil {
			return err
		}
		chains := ParseTableChains(output)
		if !tableExists {
			commands = append(commands, []string{"add", "table", tableFamily, TableName})
		}
		if _, exists := chains[InputChain]; !exists {
			commands = append(commands, []string{
				"add", "chain", tableFamily, TableName, InputChain,
				"{", "type", "filter", "hook", "input", "priority", "0", ";", "policy", "accept", ";", "}",
			})
		}
		for _, nativeChain := range BasicChains() {
			if _, exists := chains[nativeChain]; exists {
				continue
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

func initPreRules(requiredPorts []firewall.PortWhitelist, families ...filter.Family) error {
	if len(families) == 0 {
		families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
	}
	ports, err := firewall.NormalizeRequiredPorts(requiredPorts)
	if err != nil {
		return err
	}
	rules := firewall.ExpandPortWhitelist(ports)
	var commands [][]string
	for _, family := range families {
		tableFamily := TableFamily(family)
		output, _, err := ReadTable(run, tableFamily, TableName)
		if err != nil {
			return err
		}
		chains := ParseTableChains(output)
		existing := make(map[string]map[string]bool)
		for _, chain := range []string{BasicBeforeChain, BasicAfterChain} {
			existing[chain] = make(map[string]bool)
			for _, line := range strings.Split(chains[chain], "\n") {
				existing[chain][canonicalRequiredPortRule(line)] = true
			}
		}
		candidates := [][]string{
			{"add", "rule", tableFamily, TableName, BasicBeforeChain, "iifname", `"lo"`, "accept", "comment", `"Loopback Whitelist"`},
			{"add", "rule", tableFamily, TableName, BasicBeforeChain, "ct", "state", "{", "established,related", "}", "accept", "comment", `"ESTABLISHED Whitelist"`},
			{"add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "tcp", "drop"},
			{"add", "rule", tableFamily, TableName, BasicAfterChain, "meta", "l4proto", "udp", "drop"},
		}
		for _, rule := range rules {
			if rule.Family == string(family) {
				candidates = append(candidates, requiredPortCommand(tableFamily, rule))
			}
		}
		for _, command := range candidates {
			chain := command[4]
			expression := strings.Join(command[5:], " ")
			key := canonicalRequiredPortRule(expression)
			if !existing[chain][key] {
				existing[chain][key] = true
				commands = append(commands, command)
			}
		}
	}
	return runBatch(commands...)
}

func canonicalRequiredPortRule(line string) string {
	line, _, _ = strings.Cut(line, " comment ")
	line, _, _ = strings.Cut(line, " # handle ")
	for _, protocol := range []string{"tcp", "udp"} {
		line = strings.ReplaceAll(line, "meta l4proto "+protocol+" "+protocol+" ", protocol+" ")
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

func Bind(families ...filter.Family) error {
	if len(families) == 0 {
		families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
	}
	for _, family := range families {
		tableFamily := TableFamily(family)
		initialized, _, err := LoadFamilyInitStatus(family, "base")
		if err != nil {
			return err
		}
		if !initialized {
			return fmt.Errorf("1Panel nftables %s chains are not initialized", tableFamily)
		}
	}
	commands := make([][]string, 0, 8)
	for _, family := range families {
		tableFamily := TableFamily(family)
		commands = append(commands, []string{"add", "table", tableFamily, TableName})
		commands = append(commands, []string{"flush", "chain", tableFamily, TableName, InputChain})
		for _, chain := range BasicChains() {
			commands = append(commands, []string{"add", "rule", tableFamily, TableName, InputChain, "jump", chain})
		}
	}
	if err := runBatch(commands...); err != nil {
		cleanupErr := flushInputChains(families...)
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

func flushInputChains(families ...filter.Family) error {
	if len(families) == 0 {
		families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
	}
	commands := make([][]string, 0, 2)
	for _, family := range families {
		commands = append(commands, []string{"flush", "chain", TableFamily(family), TableName, InputChain})
	}
	return runBatch(commands...)
}

func OperateFamily(family filter.Family, initialize bool, ports []firewall.PortWhitelist) error {
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return fmt.Errorf("unsupported nftables family %q", family)
	}
	if initialize {
		if err := ensureBaseChains(family); err != nil {
			return err
		}
		if err := initPreRules(ports, family); err != nil {
			return err
		}
	}
	return Bind(family)
}
