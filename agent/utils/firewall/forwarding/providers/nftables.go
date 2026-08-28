package providers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
)

const (
	nftForwardFamily = "ip"
	nftForwardTable  = "nft_1panel_forward"
	nftForwardFile   = "1panel_forward.nft"
	nftForwardMarker = "1panel-forward:"
)

type nftablesAdapter struct{ system forwardingSystem }

func newNftablesAdapter() *nftablesAdapter {
	return &nftablesAdapter{system: defaultForwardingSystem{}}
}

func (n *nftablesAdapter) Name() string { return "nftables" }

func (n *nftablesAdapter) List() ([]forwarding.Rule, error) {
	rules := make([]forwarding.Rule, 0)
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		stdout, err := nftRun("-a", "list", "chain", nftTableFamily(family), nftForwardTable, nftForwardChain(forwarding.ChainPreRouting))
		if err != nil {
			return nil, fmt.Errorf("failed to list nftables %s forwarding rules: %w", family, err)
		}
		rules = append(rules, parseNftForwardRules(stdout)...)
	}
	return rules, nil
}

func (n *nftablesAdapter) Reconcile(rules []forwarding.Rule) error {
	if err := ensureNftForwardTables(); err != nil {
		return fmt.Errorf("initialize nftables forwarding table: %w", err)
	}
	commands, err := rebuildNftForwardCommands(rules)
	if err != nil {
		return err
	}
	return nftRunCommands(commands)
}

func (n *nftablesAdapter) Enable() error {
	if err := n.system.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}
	if err := n.system.WriteFile("/proc/sys/net/ipv6/conf/all/forwarding", []byte("1"), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to enable IPv6 forwarding: %w", err)
	}
	data, err := n.system.ReadFile("/etc/sysctl.conf")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read /etc/sysctl.conf: %w", err)
	}
	if err := n.system.WriteFile("/etc/sysctl.conf", []byte(enableForwardingSysctls(string(data), true)), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to persist IP forwarding: %w", err)
	}
	if err := n.system.RunWithOptionalSudo("sysctl", "-p"); err != nil {
		return fmt.Errorf("failed to apply IP forwarding: %w", err)
	}
	if err := ensureNftForwardTables(); err != nil {
		return fmt.Errorf("initialize nftables forwarding table: %w", err)
	}
	return nil
}

func (n *nftablesAdapter) Cleanup() error {
	commands := make([][]string, 0, 2)
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		if _, err := nftRun("list", "table", tableFamily, nftForwardTable); err != nil {
			continue
		}
		commands = append(commands, []string{"delete", "table", tableFamily, nftForwardTable})
	}
	if len(commands) > 0 {
		if err := nftRunCommands(commands); err != nil {
			return err
		}
	}
	file := filepath.Join(global.Dir.FirewallDir, nftForwardFile)
	if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (n *nftablesAdapter) InitStatus() (bool, bool, error) {
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		initialized, bound, err := n.FamilyStatus(family)
		if err != nil || !initialized || !bound {
			return initialized, bound, err
		}
	}
	return true, true, nil
}

func (n *nftablesAdapter) FamilyStatus(family string) (bool, bool, error) {
	sysctlPath := "/proc/sys/net/ipv4/ip_forward"
	if family == forwarding.FamilyIPv6 {
		sysctlPath = "/proc/sys/net/ipv6/conf/all/forwarding"
	}
	data, err := n.system.ReadFile(sysctlPath)
	if err != nil {
		return false, false, fmt.Errorf("read %s forwarding status: %w", family, err)
	}
	for _, chain := range []string{forwarding.ChainPreRouting, forwarding.ChainPostRouting, forwarding.ChainForward} {
		if _, err := nftRun("list", "chain", nftTableFamily(family), nftForwardTable, nftForwardChain(chain)); err != nil {
			return false, false, nil
		}
	}
	return true, strings.TrimSpace(string(data)) != "0", nil
}

func (n *nftablesAdapter) Replay() error {
	file := filepath.Join(global.Dir.FirewallDir, nftForwardFile)
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	allPresent := true
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
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
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		tableExists := true
		if _, err := nftRun("list", "table", tableFamily, nftForwardTable); err != nil {
			tableExists = false
			commands = append(commands, []string{"add", "table", tableFamily, nftForwardTable})
		}
		chains := []struct {
			name, chainType, hook, priority string
		}{
			{nftForwardChain(forwarding.ChainPreRouting), "nat", "prerouting", "-100"},
			{nftForwardChain(forwarding.ChainPostRouting), "nat", "postrouting", "100"},
			{nftForwardChain(forwarding.ChainForward), "filter", "forward", "0"},
		}
		for _, chain := range chains {
			if tableExists {
				if _, err := nftRun("list", "chain", tableFamily, nftForwardTable, chain.name); err == nil {
					continue
				}
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
	return nftRunCommands(commands)
}

func rebuildNftForwardCommands(rules []forwarding.Rule) ([][]string, error) {
	commands := make([][]string, 0, 6+len(rules)*4)
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		for _, chain := range []string{forwarding.ChainPreRouting, forwarding.ChainPostRouting, forwarding.ChainForward} {
			commands = append(commands, []string{"flush", "chain", nftTableFamily(family), nftForwardTable, nftForwardChain(chain)})
		}
	}
	for _, rule := range rules {
		normalized, err := forwarding.NormalizeRule(rule)
		if err != nil {
			return nil, err
		}
		rule = normalized
		tableFamily := nftTableFamily(rule.Family)
		addressKeyword := nftAddressKeyword(rule.Family)
		comment := strconv.Quote(encodeNftForwardRule(rule))
		interfaceMatch := make([]string, 0, 2)
		if rule.Interface != "" {
			interfaceMatch = append(interfaceMatch, "iifname", strconv.Quote(rule.Interface))
		}
		if isRemoteTarget(rule.Family, rule.TargetIP) {
			preRouting := []string{"add", "rule", tableFamily, nftForwardTable, nftForwardChain(forwarding.ChainPreRouting)}
			preRouting = append(preRouting, interfaceMatch...)
			preRouting = append(preRouting, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.Port, "dnat", "to", forwardingTarget(rule), "comment", comment)
			commands = append(commands,
				preRouting,
				[]string{"add", "rule", tableFamily, nftForwardTable, nftForwardChain(forwarding.ChainPostRouting), addressKeyword, "daddr", rule.TargetIP, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.TargetPort, "masquerade", "comment", comment},
				[]string{"add", "rule", tableFamily, nftForwardTable, nftForwardChain(forwarding.ChainForward), addressKeyword, "daddr", rule.TargetIP, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.TargetPort, "accept", "comment", comment},
				[]string{"add", "rule", tableFamily, nftForwardTable, nftForwardChain(forwarding.ChainForward), addressKeyword, "saddr", rule.TargetIP, "meta", "l4proto", rule.Protocol, rule.Protocol, "sport", rule.TargetPort, "accept", "comment", comment},
			)
			continue
		}
		preRouting := []string{"add", "rule", tableFamily, nftForwardTable, nftForwardChain(forwarding.ChainPreRouting)}
		preRouting = append(preRouting, interfaceMatch...)
		preRouting = append(preRouting, "meta", "l4proto", rule.Protocol, rule.Protocol, "dport", rule.Port, "redirect", "to", ":"+rule.TargetPort, "comment", comment)
		commands = append(commands, preRouting)
	}
	return commands, nil
}

func nftTableFamily(family string) string {
	if family == forwarding.FamilyIPv6 {
		return "ip6"
	}
	return nftForwardFamily
}

func nftAddressKeyword(family string) string {
	if family == forwarding.FamilyIPv6 {
		return "ip6"
	}
	return "ip"
}

func encodeNftForwardRule(rule forwarding.Rule) string {
	family, protocol := "4", "t"
	if rule.Family == forwarding.FamilyIPv6 {
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

func decodeNftForwardRule(value string) (forwarding.Rule, bool) {
	if !strings.HasPrefix(value, nftForwardMarker) {
		return forwarding.Rule{}, false
	}
	value = strings.TrimPrefix(value, nftForwardMarker)
	if strings.HasPrefix(value, "v2|") {
		return decodeCompactNftForwardRule(value)
	}
	return decodeLegacyNftForwardRule(value)
}

func decodeCompactNftForwardRule(value string) (forwarding.Rule, bool) {
	parts := strings.Split(value, "|")
	if len(parts) != 7 || parts[0] != "v2" {
		return forwarding.Rule{}, false
	}
	family, protocol := "", ""
	switch parts[1] {
	case "4":
		family = forwarding.FamilyIPv4
	case "6":
		family = forwarding.FamilyIPv6
	default:
		return forwarding.Rule{}, false
	}
	switch parts[2] {
	case "t":
		protocol = "tcp"
	case "u":
		protocol = "udp"
	default:
		return forwarding.Rule{}, false
	}
	return forwarding.Rule{
		Family: family, Protocol: protocol, Port: parts[3], TargetIP: parts[4], TargetPort: parts[5], Interface: parts[6],
	}, true
}

func decodeLegacyNftForwardRule(value string) (forwarding.Rule, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 6 {
		return forwarding.Rule{}, false
	}
	decoded := make([]string, len(parts))
	for index, part := range parts {
		data, err := base64.RawURLEncoding.DecodeString(part)
		if err != nil {
			return forwarding.Rule{}, false
		}
		decoded[index] = string(data)
	}
	return forwarding.Rule{Family: decoded[0], Protocol: decoded[1], Port: decoded[2], TargetIP: decoded[3], TargetPort: decoded[4], Interface: decoded[5]}, true
}

func parseNftForwardRules(stdout string) []forwarding.Rule {
	result := make([]forwarding.Rule, 0)
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

func nftForwardChain(logical string) string {
	return "NFT_" + logical
}

func nftRun(args ...string) (string, error) {
	return cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout("nft", args...)
}

func nftRunCommand(args ...string) error {
	return cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudo("nft", args...)
}

func nftRunCommands(commands [][]string) error {
	script, err := nftCommandsScript(commands)
	if err != nil {
		return err
	}
	return nftables_helper.RunScript(script)
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
