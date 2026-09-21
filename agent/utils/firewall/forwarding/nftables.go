package forwarding

import (
	"context"
	"encoding/base64"
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
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
)

const (
	nftForwardFamily = "ip"
	nftForwardTable  = "nft_1panel_forward"
	nftForwardFile   = "1panel_forward.nft"
	nftForwardMarker = "1panel-forward:"
)

type Nftables struct{ system forwardingSystem }

func NewNftables() *Nftables {
	return &Nftables{system: defaultForwardingSystem{}}
}

func (n *Nftables) Name() string { return "nftables" }

func (n *Nftables) List() ([]Rule, error) {
	rules := make([]Rule, 0)
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		stdout, err := nftables_helper.ReadChain(nftRun, nftTableFamily(family), nftForwardTable, "NFT_"+ChainPreRouting)
		if errors.Is(err, nftables_helper.ErrChainNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list nftables %s forwarding rules: %w", family, err)
		}
		rules = append(rules, parseNftForwardRules(stdout)...)
	}
	return rules, nil
}

func (n *Nftables) ReplaceRules(rules []Rule) error {
	if err := ensureNftForwardTables(); err != nil {
		return fmt.Errorf("initialize nftables forwarding table: %w", err)
	}
	commands, err := rebuildNftForwardCommands(rules)
	if err != nil {
		return err
	}
	return nftRunCommands(context.Background(), commands)
}

func (n *Nftables) CreateRules(ctx context.Context, rules []Rule) error {
	if len(rules) == 0 {
		return nil
	}
	commands, err := createNftForwardCommands(rules)
	if err != nil {
		return err
	}
	return nftRunCommands(ctx, commands)
}

func (n *Nftables) DeleteRules(ctx context.Context, rules []Rule) error {
	wanted := make(map[string]map[string]bool)
	for _, rule := range rules {
		normalized, err := NormalizeRule(rule)
		if err != nil {
			return err
		}
		family := nftTableFamily(normalized.Family)
		if wanted[family] == nil {
			wanted[family] = make(map[string]bool)
		}
		wanted[family][normalized.Identity()] = true
	}
	var commands [][]string
	for _, family := range []string{"ip", "ip6"} {
		if len(wanted[family]) == 0 {
			continue
		}
		output, _, err := nftables_helper.ReadTable(func(args ...string) (string, error) {
			return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout("nft", args...)
		}, family, nftForwardTable)
		if err != nil {
			return err
		}
		chains := nftables_helper.ParseTableChains(output)
		for _, chain := range []string{ChainPreRouting, ChainPostRouting, ChainForward} {
			for _, rule := range parseNftForwardRules(chains["NFT_"+chain]) {
				if !wanted[family][rule.Identity()] {
					continue
				}
				if _, err := strconv.ParseUint(rule.Num, 10, 64); err != nil {
					return fmt.Errorf("invalid nftables forwarding handle %q", rule.Num)
				}
				commands = append(commands, []string{"delete", "rule", family, nftForwardTable, "NFT_" + chain, "handle", rule.Num})
			}
		}
	}
	if len(commands) == 0 {
		return nil
	}
	return nftRunCommands(ctx, commands)
}

func (n *Nftables) Enable() error {
	if err := ensureForwardingSysctls(n.system, true); err != nil {
		return err
	}
	if err := ensureNftForwardTables(); err != nil {
		return fmt.Errorf("initialize nftables forwarding table: %w", err)
	}
	return nil
}

func (n *Nftables) Cleanup() error {
	commands := make([][]string, 0, 2)
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		if _, err := nftRun("list", "table", tableFamily, nftForwardTable); err != nil {
			continue
		}
		commands = append(commands, []string{"delete", "table", tableFamily, nftForwardTable})
	}
	if len(commands) > 0 {
		if err := nftRunCommands(context.Background(), commands); err != nil {
			return err
		}
	}
	file := filepath.Join(global.Dir.FirewallDir, nftForwardFile)
	if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (n *Nftables) InitStatus() (bool, bool, error) {
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		initialized, bound, err := n.FamilyStatus(family)
		if err != nil || !initialized || !bound {
			return initialized, bound, err
		}
	}
	return true, true, nil
}

func (n *Nftables) FamilyStatus(family string) (bool, bool, error) {
	sysctlPath := "/proc/sys/net/ipv4/ip_forward"
	if family == FamilyIPv6 {
		sysctlPath = "/proc/sys/net/ipv6/conf/all/forwarding"
	}
	data, err := n.system.ReadFile(sysctlPath)
	if err != nil {
		return false, false, fmt.Errorf("read %s forwarding status: %w", family, err)
	}
	output, exists, err := nftables_helper.ReadTable(nftRun, nftTableFamily(family), nftForwardTable)
	if err != nil {
		return false, false, err
	}
	if !exists {
		return false, false, nil
	}
	chains := nftables_helper.ParseTableChains(output)
	for _, chain := range []string{ChainPreRouting, ChainPostRouting, ChainForward} {
		if _, exists := chains["NFT_"+chain]; !exists {
			return false, false, nil
		}
	}
	return true, strings.TrimSpace(string(data)) != "0", nil
}

func (n *Nftables) Replay() error {
	file := filepath.Join(global.Dir.FirewallDir, nftForwardFile)
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	allPresent := true
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		if _, err := nftRun("list", "table", nftTableFamily(family), nftForwardTable); err != nil {
			allPresent = false
		}
	}
	if allPresent {
		return nil
	}
	return nftRunCommand("-f", file)
}

func ensureNftForwardTables() error {
	commands := make([][]string, 0, 8)
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		output, tableExists, err := nftables_helper.ReadTable(nftRun, tableFamily, nftForwardTable)
		if err != nil {
			return err
		}
		existingChains := nftables_helper.ParseTableChains(output)
		if !tableExists {
			commands = append(commands, []string{"add", "table", tableFamily, nftForwardTable})
		}
		chains := []struct {
			name, chainType, hook, priority string
		}{
			{"NFT_" + ChainPreRouting, "nat", "prerouting", "-100"},
			{"NFT_" + ChainPostRouting, "nat", "postrouting", "100"},
			{"NFT_" + ChainForward, "filter", "forward", "0"},
		}
		for _, chain := range chains {
			if _, exists := existingChains[chain.name]; exists {
				continue
			}
			commands = append(commands, []string{
				"add", "chain", tableFamily, nftForwardTable, chain.name,
				"{", "type", chain.chainType, "hook", chain.hook, "priority", chain.priority, ";", "policy", "accept", ";", "}",
			})
		}
	}
	if len(commands) == 0 {
		return nil
	}
	return nftRunCommands(context.Background(), commands)
}

func rebuildNftForwardCommands(rules []Rule) ([][]string, error) {
	commands := make([][]string, 0, 6+len(rules)*4)
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		for _, chain := range []string{ChainPreRouting, ChainPostRouting, ChainForward} {
			commands = append(commands, []string{"flush", "chain", nftTableFamily(family), nftForwardTable, "NFT_" + chain})
		}
	}
	additions, err := createNftForwardCommands(rules)
	return append(commands, additions...), err
}

func createNftForwardCommands(rules []Rule) ([][]string, error) {
	commands := make([][]string, 0, len(rules)*4)
	for _, rule := range rules {
		normalized, err := NormalizeRule(rule)
		if err != nil {
			return nil, err
		}
		rule = normalized
		tableFamily := nftTableFamily(rule.Family)
		addressKeyword := tableFamily
		comment := strconv.Quote(encodeNftForwardRule(rule))
		interfaceMatch := make([]string, 0, 2)
		if rule.Interface != "" {
			interfaceMatch = append(interfaceMatch, "iifname", strconv.Quote(rule.Interface))
		}
		if isRemoteTarget(rule.Family, rule.TargetIP) {
			preRouting := []string{"add", "rule", tableFamily, nftForwardTable, "NFT_" + ChainPreRouting}
			preRouting = append(preRouting, interfaceMatch...)
			preRouting = append(preRouting, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.Port, "dnat", "to", forwardingTarget(rule), "comment", comment)
			commands = append(commands,
				preRouting,
				[]string{"add", "rule", tableFamily, nftForwardTable, "NFT_" + ChainPostRouting, addressKeyword, "daddr", rule.TargetIP, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.TargetPort, "masquerade", "comment", comment},
				[]string{"add", "rule", tableFamily, nftForwardTable, "NFT_" + ChainForward, addressKeyword, "daddr", rule.TargetIP, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.TargetPort, "accept", "comment", comment},
				[]string{"add", "rule", tableFamily, nftForwardTable, "NFT_" + ChainForward, addressKeyword, "saddr", rule.TargetIP, "meta", "l4proto", rule.Protocol, rule.Protocol, "sport", rule.TargetPort, "accept", "comment", comment},
			)
			continue
		}
		preRouting := []string{"add", "rule", tableFamily, nftForwardTable, "NFT_" + ChainPreRouting}
		preRouting = append(preRouting, interfaceMatch...)
		preRouting = append(preRouting, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.Port, "redirect", "to", ":"+rule.TargetPort, "comment", comment)
		commands = append(commands, preRouting)
	}
	return commands, nil
}

func nftTableFamily(family string) string {
	if family == FamilyIPv6 {
		return "ip6"
	}
	return nftForwardFamily
}

func encodeNftForwardRule(rule Rule) string {
	family, protocol := "4", "t"
	if rule.Family == FamilyIPv6 {
		family = "6"
	}
	if rule.Protocol == "udp" {
		protocol = "u"
	}
	return nftForwardMarker + "v2|" + strings.Join(
		[]string{family, protocol, rule.Port, rule.TargetIP, rule.TargetPort, rule.Interface},
		"|",
	)
}

func decodeNftForwardRule(value string) (Rule, bool) {
	if !strings.HasPrefix(value, nftForwardMarker) {
		return Rule{}, false
	}
	value = strings.TrimPrefix(value, nftForwardMarker)
	if strings.HasPrefix(value, "v2|") {
		return decodeCompactNftForwardRule(value)
	}
	return decodeLegacyNftForwardRule(value)
}

func decodeCompactNftForwardRule(value string) (Rule, bool) {
	parts := strings.Split(value, "|")
	if len(parts) != 7 || parts[0] != "v2" {
		return Rule{}, false
	}
	family, protocol := "", ""
	switch parts[1] {
	case "4":
		family = FamilyIPv4
	case "6":
		family = FamilyIPv6
	default:
		return Rule{}, false
	}
	switch parts[2] {
	case "t":
		protocol = "tcp"
	case "u":
		protocol = "udp"
	default:
		return Rule{}, false
	}
	return Rule{
		Family: family, Protocol: protocol, Port: parts[3], TargetIP: parts[4], TargetPort: parts[5], Interface: parts[6],
	}, true
}

func decodeLegacyNftForwardRule(value string) (Rule, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 6 {
		return Rule{}, false
	}
	decoded := make([]string, len(parts))
	for index, part := range parts {
		data, err := base64.RawURLEncoding.DecodeString(part)
		if err != nil {
			return Rule{}, false
		}
		decoded[index] = string(data)
	}
	return Rule{Family: decoded[0], Protocol: decoded[1], Port: decoded[2], TargetIP: decoded[3], TargetPort: decoded[4], Interface: decoded[5]}, true
}

func parseNftForwardRules(stdout string) []Rule {
	result := make([]Rule, 0)
	for _, line := range strings.Split(stdout, "\n") {
		commentStart := strings.Index(line, `comment "`+nftForwardMarker)
		handleStart := strings.LastIndex(line, "# handle ")
		if commentStart < 0 || handleStart < 0 {
			continue
		}
		encodedStart := commentStart + len(`comment "`)
		encodedEnd := strings.Index(line[encodedStart:], `"`)
		if encodedEnd < 0 {
			continue
		}
		rule, ok := decodeNftForwardRule(line[encodedStart : encodedStart+encodedEnd])
		if !ok {
			continue
		}
		rule.Num = strings.TrimSpace(line[handleStart+len("# handle "):])
		result = append(result, rule)
	}
	return result
}

func nftRun(args ...string) (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout("nft", args...)
	if err != nil {
		return stdout, fmt.Errorf("command=nft %s failed: %w", strings.Join(args, " "), err)
	}
	return stdout, nil
}

func nftRunCommand(args ...string) error {
	err := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudo("nft", args...)
	if err != nil {
		return fmt.Errorf("command=nft %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

func nftRunCommands(ctx context.Context, commands [][]string) error {
	script, err := nftCommandsScript(commands)
	if err != nil {
		return err
	}
	var stderr strings.Builder
	err = cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(script)), cmd.WithStderr(&stderr)).RunWithOptionalSudo("nft", "-f", "-")
	if err == nil && strings.TrimSpace(stderr.String()) != "" {
		err = fmt.Errorf("firewall command warning: %s", strings.TrimSpace(stderr.String()))
	}
	return firewallutil.WrapBatchCommandError("nft -f -", script, err)
}

func nftCommandsScript(commands [][]string) (string, error) {
	var script strings.Builder
	for _, args := range commands {
		if len(args) == 0 {
			return "", fmt.Errorf("empty nftables command")
		}
		for _, token := range args {
			if strings.ContainsAny(token, "\r\n") {
				return "", fmt.Errorf("invalid newline in nftables command token")
			}
		}
		script.WriteString(strings.Join(args, " "))
		script.WriteByte('\n')
	}
	return script.String(), nil
}
