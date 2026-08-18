package providers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

const (
	nftForwardFamily = "ip"
	nftForwardTable  = "nft_1panel_forward"
	nftForwardFile   = "1panel_forward.nft"
)

type nftablesAdapter struct{ system forwardingSystem }

var nftInterfacePattern = regexp.MustCompile(`^[A-Za-z0-9_.:@-]{1,15}$`)

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
			{nftForwardChain(forwarding.ChainPreRouting), "nat", "prerouting", "dstnat"},
			{nftForwardChain(forwarding.ChainPostRouting), "nat", "postrouting", "srcnat"},
			{nftForwardChain(forwarding.ChainForward), "filter", "forward", "filter"},
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
		normalized, err := normalizeNftForwardRule(rule)
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

func normalizeNftForwardRule(rule forwarding.Rule) (forwarding.Rule, error) {
	return NormalizeRule(rule)
}

func NormalizeRule(rule forwarding.Rule) (forwarding.Rule, error) {
	rule.Family = strings.ToLower(strings.TrimSpace(rule.Family))
	if rule.Family == "" {
		rule.Family = forwarding.FamilyIPv4
	}
	if rule.Family != forwarding.FamilyIPv4 && rule.Family != forwarding.FamilyIPv6 {
		return forwarding.Rule{}, fmt.Errorf("unsupported forwarding family %q", rule.Family)
	}
	rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
	if rule.Protocol != "tcp" && rule.Protocol != "udp" {
		return forwarding.Rule{}, fmt.Errorf("unsupported forwarding protocol %q", rule.Protocol)
	}
	var err error
	if rule.Port, err = normalizeForwardPort(rule.Port); err != nil {
		return forwarding.Rule{}, fmt.Errorf("invalid forwarding port: %w", err)
	}
	if rule.TargetPort, err = normalizeForwardPort(rule.TargetPort); err != nil {
		return forwarding.Rule{}, fmt.Errorf("invalid forwarding target port: %w", err)
	}
	rule.TargetIP = strings.TrimSpace(rule.TargetIP)
	if rule.TargetIP == "" || strings.EqualFold(rule.TargetIP, "localhost") {
		if rule.Family == forwarding.FamilyIPv6 {
			rule.TargetIP = "::1"
		} else {
			rule.TargetIP = "127.0.0.1"
		}
	}
	address, err := netip.ParseAddr(rule.TargetIP)
	if err == nil {
		address = address.Unmap()
	}
	if err != nil || (rule.Family == forwarding.FamilyIPv4) != address.Is4() {
		return forwarding.Rule{}, fmt.Errorf("invalid %s forwarding target %q", rule.Family, rule.TargetIP)
	}
	rule.TargetIP = address.String()
	rule.Interface = strings.TrimSpace(rule.Interface)
	if rule.Interface == "all" {
		rule.Interface = ""
	}
	if rule.Interface != "" && !nftInterfacePattern.MatchString(rule.Interface) {
		return forwarding.Rule{}, fmt.Errorf("invalid forwarding interface %q", rule.Interface)
	}
	return rule, nil
}

func normalizeForwardRule(rule forwarding.Rule) (forwarding.Rule, error) {
	return NormalizeRule(rule)
}

func normalizeForwardPort(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) < 1 || len(parts) > 2 {
		return "", fmt.Errorf("invalid port range %q", value)
	}
	ports := make([]int, len(parts))
	for index, part := range parts {
		port, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || port < 1 || port > 65535 {
			return "", fmt.Errorf("invalid port %q", part)
		}
		ports[index] = port
	}
	if len(ports) == 2 {
		if ports[0] > ports[1] {
			return "", fmt.Errorf("descending port range %q", value)
		}
		if ports[0] != ports[1] {
			return strconv.Itoa(ports[0]) + "-" + strconv.Itoa(ports[1]), nil
		}
	}
	return strconv.Itoa(ports[0]), nil
}

func encodeNftForwardRule(rule forwarding.Rule) string {
	fields := []string{rule.Family, rule.Protocol, rule.Port, rule.TargetIP, rule.TargetPort, rule.Interface}
	for index := range fields {
		fields[index] = base64.RawURLEncoding.EncodeToString([]byte(fields[index]))
	}
	return "1panel-forward:" + strings.Join(fields, ".")
}

func decodeNftForwardRule(value string) (forwarding.Rule, bool) {
	value = strings.TrimPrefix(value, "1panel-forward:")
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
		commentStart := strings.Index(line, `comment "1panel-forward:`)
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
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(script)))
	return manager.RunWithOptionalSudo("nft", "-f", "-")
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
