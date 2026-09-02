package docker_guard

import (
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-shellwords"
)

type observedPolicy struct {
	policy         Policy
	sequence       int64
	nativeRules    []NativeRule
	managedOrders  []int64
	dropAll        bool
	droppedSource  []string
	allowedSource  []string
	acceptedSource []string
	acceptAll      bool
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
		managed := strings.Contains(line, "1panel-docker:")
		if !managed && !hasAcceptAction(tokens) {
			continue
		}
		sequence++
		fragment, source, action, err := parseDockerGuardRuleTokens(tokens, family)
		if err != nil {
			return PolicyInventory{}, err
		}
		identity := fragment.UUID
		if action == "accept" {
			identity = action
		}
		key := strings.Join([]string{identity, fragment.Family, fragment.HostIP, strconv.Itoa(int(fragment.HostPort)), fragment.Protocol}, "|")
		group, exists := groups[key]
		if !exists {
			group = &observedPolicy{policy: fragment, sequence: sequence}
			groups[key] = group
			order = append(order, key)
		}
		switch {
		case action == "accept" && source != "":
			group.acceptedSource = append(group.acceptedSource, source)
			group.nativeRules = append(group.nativeRules, NativeRule{Family: family, Order: sequence, Tokens: nativeRuleTokens(tokens)})
		case action == "accept":
			group.acceptAll = true
			group.nativeRules = append(group.nativeRules, NativeRule{Family: family, Order: sequence, Tokens: nativeRuleTokens(tokens)})
		case action == "return" && source != "":
			group.allowedSource = append(group.allowedSource, source)
		case action == "drop" && source != "":
			group.droppedSource = append(group.droppedSource, source)
		case action == "drop":
			group.dropAll = true
		default:
			return PolicyInventory{}, fmt.Errorf("unsupported Docker guard rule action %q", action)
		}
		if action != "accept" {
			group.managedOrders = append(group.managedOrders, sequence)
		}
	}
	inventory := PolicyInventory{Policies: make([]Policy, 0, len(order)), ManagedRuleOrders: make(map[string][]int64)}
	for _, key := range order {
		group := groups[key]
		if group.acceptAll || len(group.acceptedSource) > 0 {
			group.policy.Sources = uniqueSortedStrings(group.acceptedSource)
			inventory.ReadOnly = append(inventory.ReadOnly, ReadOnlyPolicy{
				Policy: group.policy, Action: "accept", Sequence: group.sequence, NativeRules: group.nativeRules,
			})
			continue
		}
		switch {
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
		inventory.Policies = append(inventory.Policies, group.policy)
		inventory.ManagedRuleOrders[managedOrderKey(group.policy.Family, group.policy.UUID)] = append([]int64(nil), group.managedOrders...)
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

func managedOrderKey(family, policyUUID string) string {
	return family + "\x00" + policyUUID
}

func parseDockerGuardRuleTokens(tokens []string, family string) (Policy, string, string, error) {
	policy := Policy{Family: family, HostIP: wildcardHost(family)}
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
			if isCommentValue(tokens, index) {
				continue
			}
			action = tokens[index]
		}
	}
	if action == "" || (action != "accept" && (policy.UUID == "" || policy.Protocol == "" || policy.HostPort == 0)) {
		return Policy{}, "", "", fmt.Errorf("incomplete 1Panel Docker guard rule")
	}
	if action == "accept" && policy.Protocol == "" {
		policy.Protocol = "all"
	}
	return policy, source, action, nil
}

func hasAcceptAction(tokens []string) bool {
	for index, token := range tokens {
		if token == "-j" && strings.EqualFold(nextPolicyToken(tokens, index), "accept") {
			return true
		}
		if strings.EqualFold(token, "accept") && !isCommentValue(tokens, index) {
			return true
		}
	}
	return false
}

func isCommentValue(tokens []string, index int) bool {
	if index == 0 {
		return false
	}
	return tokens[index-1] == "comment" || tokens[index-1] == "--comment"
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

func wildcardHost(family string) string {
	if family == FamilyIPv6 {
		return "::"
	}
	return "0.0.0.0"
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
