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
	policy        Policy
	dropAll       bool
	droppedSource []string
	allowedSource []string
}

func parseDockerGuardPolicies(output, family string) ([]Policy, error) {
	groups := make(map[string]*observedPolicy)
	order := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "1panel-docker:") {
			continue
		}
		tokens, err := shellwords.Parse(strings.TrimSpace(line))
		if err != nil {
			return nil, fmt.Errorf("parse Docker guard rule: %w", err)
		}
		fragment, source, action, err := parseDockerGuardRuleTokens(tokens, family)
		if err != nil {
			return nil, err
		}
		key := strings.Join([]string{fragment.UUID, fragment.Family, fragment.HostIP, strconv.Itoa(int(fragment.HostPort)), fragment.Protocol}, "|")
		group, exists := groups[key]
		if !exists {
			group = &observedPolicy{policy: fragment}
			groups[key] = group
			order = append(order, key)
		}
		switch {
		case action == "return" && source != "":
			group.allowedSource = append(group.allowedSource, source)
		case action == "drop" && source != "":
			group.droppedSource = append(group.droppedSource, source)
		case action == "drop":
			group.dropAll = true
		default:
			return nil, fmt.Errorf("unsupported Docker guard rule action %q", action)
		}
	}
	policies := make([]Policy, 0, len(order))
	for _, key := range order {
		group := groups[key]
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
			return nil, fmt.Errorf("Docker guard policy %s has no effective rules", group.policy.UUID)
		}
		policies = append(policies, group.policy)
	}
	return policies, nil
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
		case "--ctorigdstport":
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
			if nextPolicyToken(tokens, index) == "saddr" {
				source = nextPolicyToken(tokens, index+1)
			}
		case "drop", "return":
			action = tokens[index]
		}
	}
	if policy.UUID == "" || policy.Protocol == "" || policy.HostPort == 0 || action == "" {
		return Policy{}, "", "", fmt.Errorf("incomplete 1Panel Docker guard rule")
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
