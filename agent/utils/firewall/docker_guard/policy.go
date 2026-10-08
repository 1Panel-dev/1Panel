package docker_guard

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/netip"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-shellwords"
)

type observedPolicy struct {
	policy         Policy
	orders         []int64
	dropAll        bool
	droppedSource  []string
	allowedSource  []string
	acceptedSource []string
	acceptAll      bool
}

func ConvertPolicyBackend(policy Policy, sourceBackend, targetBackend string) (Policy, error) {
	compiled := make(map[string][][]string, 2)
	for _, backend := range []string{sourceBackend, targetBackend} {
		switch backend {
		case "iptables":
			compiled[backend] = compilePolicy(policy)
		case "nftables":
			compiled[backend] = compileNftPolicy(policy)
			for i := range compiled[backend] {
				compiled[backend][i] = compiled[backend][i][5:]
			}
		default:
			return Policy{}, fmt.Errorf("unsupported Docker firewall backend %q", backend)
		}
	}
	if len(policy.NativeRules) == 0 {
		return policy, nil
	}
	sourceRules, targetRules := compiled[sourceBackend], compiled[targetBackend]
	if len(policy.NativeRules) != len(sourceRules) {
		return Policy{}, fmt.Errorf("Docker policy %s contains native rules that cannot be converted from %s to %s", policy.UUID, sourceBackend, targetBackend)
	}
	byRule := make(map[string]int, len(sourceRules))
	for index, rule := range sourceRules {
		byRule[dockerPolicyRuleKey(rule, sourceBackend)] = index
	}
	converted := make([]NativeRule, 0, len(policy.NativeRules))
	for _, native := range policy.NativeRules {
		key := dockerPolicyRuleKey(native.Tokens, sourceBackend)
		index, exists := byRule[key]
		if !exists || native.Family != policy.Family {
			return Policy{}, fmt.Errorf("Docker policy %s contains native conditions that cannot be converted from %s to %s", policy.UUID, sourceBackend, targetBackend)
		}
		delete(byRule, key)
		tokens := append([]string(nil), targetRules[index]...)
		for i, token := range tokens {
			if unquoted, err := strconv.Unquote(token); err == nil {
				tokens[i] = unquoted
			}
		}
		converted = append(converted, NativeRule{Family: native.Family, Order: native.Order, Tokens: tokens})
	}
	policy.NativeRules = converted
	return policy, nil
}

func dockerPolicyRuleKey(tokens []string, backend string) string {
	tokens = nativeRuleTokens(tokens)
	if backend == "iptables" {
		if len(tokens)%2 != 0 {
			return ""
		}
		protocol := ""
		if index := slices.Index(tokens, "-p"); index >= 0 && index+1 < len(tokens) {
			protocol = tokens[index+1]
		}
		parts := make([]string, 0, len(tokens)/2)
		for i := 0; i < len(tokens); i += 2 {
			option, value := tokens[i], tokens[i+1]
			if option == "-m" && value == protocol && (value == "tcp" || value == "udp") {
				continue
			}
			if option == "--ctorigdst" {
				value = normalizeObservedHost(value)
			}
			if option == "-s" {
				if address, err := netip.ParseAddr(value); err == nil {
					value = netip.PrefixFrom(address, address.BitLen()).String()
				} else if prefix, err := netip.ParsePrefix(value); err == nil {
					value = prefix.Masked().String()
				}
			}
			parts = append(parts, option+"\x00"+value)
		}
		sort.Strings(parts)
		return strings.Join(parts, "\x01")
	}
	parts := make([]string, 0, len(tokens))
	for i, token := range tokens {
		if token == "counter" {
			continue
		}
		if unquoted, err := strconv.Unquote(token); err == nil {
			token = unquoted
		}
		if i > 0 && tokens[i-1] == "daddr" {
			token = normalizeObservedHost(token)
		}
		if i > 0 && tokens[i-1] == "saddr" {
			if address, err := netip.ParseAddr(token); err == nil {
				token = netip.PrefixFrom(address, address.BitLen()).String()
			} else if prefix, err := netip.ParsePrefix(token); err == nil {
				token = prefix.Masked().String()
			}
		}
		parts = append(parts, token)
	}
	return strings.Join(parts, "\x00")
}

func parseDockerGuardPolicies(output, family string) (PolicyInventory, error) {
	groups := make(map[string]*observedPolicy)
	order := make([]string, 0)
	sequence := int64(0)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tokens, err := shellwords.Parse(line)
		if err != nil {
			return PolicyInventory{}, fmt.Errorf("parse Docker guard rule: %w", err)
		}
		sequence++
		fragment, source, action, err := parseDockerGuardRuleTokens(tokens, family)
		if err != nil {
			return PolicyInventory{}, err
		}
		if action == "" || (fragment.HostPort == 0 && action == "return") {
			continue
		}
		key := strings.Join([]string{fragment.UUID, fragment.Family, fragment.HostIP, strconv.Itoa(int(fragment.HostPort)), fragment.Protocol}, "|")
		group, exists := groups[key]
		if !exists {
			group = &observedPolicy{policy: fragment}
			groups[key] = group
			order = append(order, key)
		}
		switch {
		case action == "accept" && source != "":
			group.acceptedSource = append(group.acceptedSource, source)
		case action == "accept":
			group.acceptAll = true
		case action == "return" && source != "":
			group.allowedSource = append(group.allowedSource, source)
		case action == "drop" && source != "":
			group.droppedSource = append(group.droppedSource, source)
		case action == "drop":
			group.dropAll = true
		default:
			return PolicyInventory{}, fmt.Errorf("unsupported Docker guard rule action %q", action)
		}
		group.policy.NativeRules = append(group.policy.NativeRules, NativeRule{Family: family, Order: sequence, Tokens: nativeRuleTokens(tokens)})
		group.orders = append(group.orders, sequence)
	}
	inventory := PolicyInventory{Policies: make([]Policy, 0, len(order)), RuleOrders: make(map[string][]int64)}
	for _, key := range order {
		group := groups[key]
		switch {
		case group.acceptAll:
			group.policy.Mode = ModeAcceptAll
			group.policy.Sources = []string{}
		case len(group.acceptedSource) > 0:
			group.policy.Mode = ModeAcceptSources
			group.policy.Sources = uniqueSortedStrings(group.acceptedSource)
		case len(group.allowedSource) > 0:
			group.policy.Mode = ModeAllow
			group.policy.Sources = uniqueSortedStrings(group.allowedSource)
		case len(group.droppedSource) > 0:
			group.policy.Mode = ModeSources
			group.policy.Sources = uniqueSortedStrings(group.droppedSource)
		case group.dropAll:
			group.policy.Mode = ModeAll
		default:
			return PolicyInventory{}, fmt.Errorf("Docker guard policy %s has no effective rules", group.policy.UUID)
		}
		if _, err := uuid.Parse(group.policy.UUID); err != nil {
			rules := make([][]string, 0, len(group.policy.NativeRules))
			for _, rule := range group.policy.NativeRules {
				rules = append(rules, rule.Tokens)
			}
			fingerprint, _ := json.Marshal(rules)
			group.policy.UUID = uuid.NewSHA1(uuid.NameSpaceOID, append([]byte(family+"\x00"), fingerprint...)).String()
		}
		inventory.Policies = append(inventory.Policies, group.policy)
		inventory.RuleOrders[group.policy.Family+"\x00"+group.policy.UUID] = append([]int64(nil), group.orders...)
	}
	return inventory, nil
}

func nativeRuleTokens(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	for index, token := range tokens {
		if token == "#" {
			tokens = tokens[:index]
			break
		}
	}
	if len(tokens) >= 2 && tokens[len(tokens)-2] == "handle" {
		tokens = tokens[:len(tokens)-2]
	}
	for index := 0; index < len(tokens); index++ {
		result = append(result, tokens[index])
		if tokens[index] == "counter" && index+4 < len(tokens) && tokens[index+1] == "packets" && tokens[index+3] == "bytes" {
			index += 4
		}
	}
	return result
}

func parseDockerGuardRuleTokens(tokens []string, family string) (Policy, string, string, error) {
	policy := Policy{Family: family, HostIP: "0.0.0.0"}
	if family == FamilyIPv6 {
		policy.HostIP = "::"
	}
	source, action := "", ""
	for index := 0; index < len(tokens); index++ {
		switch tokens[index] {
		case "-p":
			policy.Protocol = nextPolicyToken(tokens, index)
		case "--ctorigdst":
			policy.HostIP = normalizeObservedHost(nextPolicyToken(tokens, index))
		case "-d":
			policy.HostIP = normalizeObservedHost(nextPolicyToken(tokens, index))
		case "--ctorigdstport":
			policy.HostPort = parsePolicyPort(nextPolicyToken(tokens, index))
		case "--dport":
			policy.HostPort = parsePolicyPort(nextPolicyToken(tokens, index))
		case "-s":
			source = nextPolicyToken(tokens, index)
		case "--comment", "comment":
			marker := nextPolicyToken(tokens, index)
			if strings.HasPrefix(marker, "1panel-docker:") {
				policy.UUID = strings.TrimPrefix(marker, "1panel-docker:")
			}
		case "-j":
			action = strings.ToLower(nextPolicyToken(tokens, index))
		case "meta":
			if nextPolicyToken(tokens, index) == "l4proto" {
				policy.Protocol = nextPolicyToken(tokens, index+1)
			}
		case "ct":
			if nextPolicyToken(tokens, index) != "original" {
				continue
			}
			switch nextPolicyToken(tokens, index+1) {
			case "proto-dst":
				policy.HostPort = parsePolicyPort(nextPolicyToken(tokens, index+2))
			case "ip", "ip6":
				if nextPolicyToken(tokens, index+2) == "daddr" {
					policy.HostIP = normalizeObservedHost(nextPolicyToken(tokens, index+3))
				}
			}
		case "ip", "ip6":
			switch nextPolicyToken(tokens, index) {
			case "saddr":
				source = nextPolicyToken(tokens, index+1)
			case "daddr":
				policy.HostIP = normalizeObservedHost(nextPolicyToken(tokens, index+1))
			}
		case "tcp", "udp":
			if nextPolicyToken(tokens, index) == "dport" {
				policy.Protocol = tokens[index]
				policy.HostPort = parsePolicyPort(nextPolicyToken(tokens, index+1))
			}
		case "accept", "drop", "return":
			if index > 0 && (tokens[index-1] == "comment" || tokens[index-1] == "--comment") {
				continue
			}
			action = tokens[index]
		}
	}
	if action != "accept" && action != "drop" && action != "return" {
		return policy, "", "", nil
	}
	if action == "drop" && (policy.Protocol == "" || policy.HostPort == 0) {
		return Policy{}, "", "", fmt.Errorf("incomplete 1Panel Docker guard rule")
	}
	if action == "accept" && policy.Protocol == "" {
		policy.Protocol = "all"
	}
	return policy, source, action, nil
}

func normalizeObservedHost(value string) string {
	if prefix, err := netip.ParsePrefix(value); err == nil && prefix.Bits() == prefix.Addr().BitLen() {
		return prefix.Addr().String()
	}
	return value
}

func nextPolicyToken(tokens []string, index int) string {
	if index+1 >= len(tokens) {
		return ""
	}
	return tokens[index+1]
}

func parsePolicyPort(value string) uint16 {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil || port == 0 {
		return 0
	}
	return uint16(port)
}

func uniqueSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

type orderedPolicyRule struct {
	order int64
	rule  []string
}

func sortPolicyRules(segments []orderedPolicyRule) [][]string {
	sort.SliceStable(segments, func(i, j int) bool { return segments[i].order < segments[j].order })
	rules := make([][]string, 0, len(segments))
	for _, segment := range segments {
		rules = append(rules, segment.rule)
	}
	return rules
}
