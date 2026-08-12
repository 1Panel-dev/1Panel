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

func TestNormalizeNativeDestinationPortSets(t *testing.T) {
	for _, provider := range []Provider{ProviderIptables, ProviderUFW} {
		scope := Scope{Provider: provider, Family: FamilyIPv4, Direction: DirectionInput}
		if provider == ProviderIptables {
			scope.Table = "filter"
			scope.Chain = IptablesInputChain
		}
		rule, err := NormalizeRule(FirewallRule{
			Scope: scope, Protocol: "tcp", DestinationPort: "080,443,8080:8090,443", Action: ActionAccept,
		})
		if err != nil {
			t.Fatalf("normalize %s port set: %v", provider, err)
		}
		if rule.DestinationPort != "80,443,8080-8090" {
			t.Fatalf("unexpected %s port set: %q", provider, rule.DestinationPort)
		}
	}

	_, err := NormalizeRule(FirewallRule{
		Scope:    Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: FirewalldInputZone, Direction: DirectionInput},
		Protocol: "tcp", DestinationPort: "80,443", Action: ActionAccept,
	})
	if !errors.Is(err, ErrCompositeRule) {
		t.Fatalf("expected firewalld port set expansion, got %v", err)
	}

	tooMany := "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16"
	_, err = NormalizeRule(FirewallRule{
		Scope:    Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: DirectionInput},
		Protocol: "tcp", DestinationPort: tooMany, Action: ActionAccept,
	})
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("expected native port-set limit error, got %v", err)
	}
}

func TestNormalizeUFWAllowsAllProtocolsForDestinationPort(t *testing.T) {
	rule, err := NormalizeRule(FirewallRule{
		Scope:    Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: DirectionInput},
		Protocol: "all", DestinationPort: "53", Action: ActionAccept,
	})
	if err != nil {
		t.Fatalf("normalize UFW all-protocol port rule: %v", err)
	}
	if rule.Protocol != "all" || rule.DestinationPort != "53" {
		t.Fatalf("unexpected normalized rule: %#v", rule)
	}

	for _, provider := range []Provider{ProviderIptables, ProviderFirewalld} {
		testRule := rule
		testRule.Scope.Provider = provider
		switch provider {
		case ProviderIptables:
			testRule.Scope.Table = "filter"
			testRule.Scope.Chain = IptablesInputChain
		case ProviderFirewalld:
			testRule.Scope.Family = FamilyInet
			testRule.Scope.Zone = FirewalldInputZone
			testRule.Scope.Chain = ""
		}
		if _, err = NormalizeRule(testRule); !errors.Is(err, ErrInvalidRule) {
			t.Fatalf("expected %s all-protocol port rejection, got %v", provider, err)
		}
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
	if len(rules) != 4 {
		t.Fatalf("expected 4 rules with native iptables port sets, got %d", len(rules))
	}
	for _, rule := range rules {
		if rule.Protocol == "tcp/udp" || rule.SourceAddress == "" || rule.DestinationPort != "80,443" {
			t.Fatalf("rule was not expanded correctly: %#v", rule)
		}
	}
}

func TestExpandAtomicRulesKeepsNativePortSetsAndSplitsFirewalld(t *testing.T) {
	iptablesRules, err := ExpandAtomicRules(FirewallRule{
		Scope:    Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: IptablesInputChain, Direction: DirectionInput},
		Protocol: "tcp", DestinationPort: "80,443", Action: ActionAccept,
	})
	if err != nil || len(iptablesRules) != 1 || iptablesRules[0].DestinationPort != "80,443" {
		t.Fatalf("iptables port set was expanded: rules=%#v err=%v", iptablesRules, err)
	}
	firewalldRules, err := ExpandAtomicRules(FirewallRule{
		Scope:    Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: FirewalldInputZone, Direction: DirectionInput},
		Protocol: "tcp", DestinationPort: "80,443", Action: ActionDrop,
	})
	if err != nil || len(firewalldRules) != 2 || firewalldRules[0].DestinationPort != "80" || firewalldRules[1].DestinationPort != "443" {
		t.Fatalf("firewalld port set was not expanded: rules=%#v err=%v", firewalldRules, err)
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
