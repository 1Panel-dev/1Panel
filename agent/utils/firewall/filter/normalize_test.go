package filter

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNormalizeRule(t *testing.T) {
	rule, err := NormalizeRule(FirewallRule{
		Scope: Scope{
			Provider:  ProviderUFW,
			Family:    FamilyIPv4,
			Direction: DirectionInput,
		},
		Protocol:           " TCP ",
		SourceAddress:      "172.16.10.111",
		DestinationAddress: "0.0.0.0/0",
		DestinationPort:    "080:080",
		Action:             "ALLOW",
		ConnectionStates:   []string{"NEW", "established", "new"},
	})
	if err != nil {
		t.Fatalf("normalize rule: %v", err)
	}

	if rule.Scope.Key() != "ufw:incoming:ipv4" {
		t.Fatalf("unexpected scope key %q", rule.Scope.Key())
	}
	if rule.NativeKind != NativeKindUFWRule {
		t.Fatalf("unexpected native kind %q", rule.NativeKind)
	}
	if rule.Protocol != "tcp" || rule.Action != ActionAccept {
		t.Fatalf("unexpected protocol/action: %s/%s", rule.Protocol, rule.Action)
	}
	if rule.SourceAddress != "172.16.10.111/32" || rule.DestinationAddress != "" {
		t.Fatalf("unexpected normalized addresses: %q -> %q", rule.SourceAddress, rule.DestinationAddress)
	}
	if rule.DestinationPort != "80" {
		t.Fatalf("unexpected destination port %q", rule.DestinationPort)
	}
	if len(rule.ConnectionStates) != 2 || rule.ConnectionStates[0] != "established" || rule.ConnectionStates[1] != "new" {
		t.Fatalf("unexpected states: %#v", rule.ConnectionStates)
	}
}

func TestNormalizeRuleRejectsCompositeAndFamilyMismatch(t *testing.T) {
	base := FirewallRule{
		Scope:           Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: DirectionInput},
		Protocol:        "tcp/udp",
		DestinationPort: "80",
		Action:          ActionAccept,
	}
	if _, err := NormalizeRule(base); !errors.Is(err, ErrCompositeRule) {
		t.Fatalf("expected composite rule error, got %v", err)
	}

	base.Protocol = "tcp"
	base.SourceAddress = "2001:db8::1"
	if _, err := NormalizeRule(base); !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("expected family validation error, got %v", err)
	}
}

func TestExpandAtomicRules(t *testing.T) {
	rules, err := ExpandAtomicRules(FirewallRule{
		Scope:           Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: DirectionInput},
		Protocol:        "tcp/udp",
		SourceAddress:   "172.16.10.111, 172.16.10.112",
		DestinationPort: "80,443,80",
		Action:          ActionAccept,
	})
	if err != nil {
		t.Fatalf("expand atomic rules: %v", err)
	}
	if len(rules) != 8 {
		t.Fatalf("expected 8 atomic rules, got %d", len(rules))
	}
	for _, rule := range rules {
		if rule.Protocol == "tcp/udp" || rule.SourceAddress == "" || rule.DestinationPort == "" {
			t.Fatalf("rule was not atomic: %#v", rule)
		}
	}
}

func TestExpandUFWInetByFamily(t *testing.T) {
	rules, err := ExpandAtomicRules(FirewallRule{
		Scope:           Scope{Provider: ProviderUFW, Family: FamilyInet, Direction: DirectionInput},
		Protocol:        "tcp",
		DestinationPort: "22",
		Action:          ActionAccept,
	})
	if err != nil {
		t.Fatalf("expand UFW families: %v", err)
	}
	if len(rules) != 2 || rules[0].Scope.Family != FamilyIPv4 || rules[1].Scope.Family != FamilyIPv6 {
		t.Fatalf("unexpected UFW family expansion: %#v", rules)
	}

	rules, err = ExpandAtomicRules(FirewallRule{
		Scope:         Scope{Provider: ProviderUFW, Family: FamilyInet, Direction: DirectionInput},
		Protocol:      "all",
		SourceAddress: "172.16.10.111",
		Action:        ActionDrop,
	})
	if err != nil {
		t.Fatalf("expand family-specific UFW rule: %v", err)
	}
	if len(rules) != 1 || rules[0].Scope.Family != FamilyIPv4 {
		t.Fatalf("expected one IPv4 rule, got %#v", rules)
	}

	rules, err = ExpandAtomicRules(FirewallRule{
		Scope:         Scope{Provider: ProviderUFW, Family: FamilyInet, Direction: DirectionInput},
		Protocol:      "all",
		SourceAddress: "::/0",
		Action:        ActionDrop,
	})
	if err != nil {
		t.Fatalf("expand IPv6-any UFW rule: %v", err)
	}
	if len(rules) != 1 || rules[0].Scope.Family != FamilyIPv6 {
		t.Fatalf("expected one IPv6 rule, got %#v", rules)
	}
}

func TestNormalizeFirewalldNativeKindSetsExecutionBucket(t *testing.T) {
	priority := -100
	rich, err := NormalizeRule(FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyIPv4, Zone: "public", Direction: DirectionInput},
		NativeKind: NativeKindRichRule, Protocol: "tcp", DestinationPort: "3306", Action: ActionDrop, Priority: &priority,
	})
	if err != nil {
		t.Fatalf("normalize firewalld rich rule: %v", err)
	}
	if rich.OrderBucket != OrderBucketRichPre || rich.Priority == nil || *rich.Priority != -100 {
		t.Fatalf("unexpected rich rule placement: %#v", rich)
	}
	zonePort, err := NormalizeRule(FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "public", Direction: DirectionInput},
		NativeKind: NativeKindZonePort, Protocol: "tcp", DestinationPort: "3306", Action: ActionAccept, Priority: &priority,
	})
	if err != nil {
		t.Fatalf("normalize firewalld zone port: %v", err)
	}
	if zonePort.OrderBucket != OrderBucketZonePrimitiveAllow || zonePort.Priority != nil {
		t.Fatalf("native port exposed fake priority: %#v", zonePort)
	}
}

func TestExpandAtomicRulesLimit(t *testing.T) {
	var addresses strings.Builder
	for i := 1; i <= 65; i++ {
		if i > 1 {
			addresses.WriteString(",")
		}
		addresses.WriteString(fmt.Sprintf("10.0.0.%d", i))
	}
	_, err := ExpandAtomicRules(FirewallRule{
		Scope:         Scope{Provider: ProviderUFW, Family: FamilyInet, Direction: DirectionInput},
		Protocol:      "tcp/udp",
		SourceAddress: addresses.String(),
		Action:        ActionAccept,
	})
	if !errors.Is(err, ErrExpansionLimit) {
		t.Fatalf("expected expansion limit error, got %v", err)
	}
}
