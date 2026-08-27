package model

import (
	"errors"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func TestFirewallRuleFromDomainUsesProviderNeutralIdentity(t *testing.T) {
	rule := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
			Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
		},
		Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	}
	iptables, err := FirewallRuleFromDomain(rule)
	if err != nil {
		t.Fatal(err)
	}
	rule.Scope.Provider = filter.ProviderNftables
	nftables, err := FirewallRuleFromDomain(rule)
	if err != nil {
		t.Fatal(err)
	}
	if iptables.PolicyKey() != nftables.PolicyKey() {
		t.Fatalf("provider leaked into desired-rule identity: iptables=%#v nftables=%#v", iptables, nftables)
	}
}

func TestFirewallRuleFromDomainRejectsProviderNativeOnlyRule(t *testing.T) {
	_, err := FirewallRuleFromDomain(filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
			Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
		},
		NativeKind: filter.NativeKindZoneService,
		Protocol:   "all",
		Action:     filter.ActionAccept,
	})
	if !errors.Is(err, filter.ErrUnsupportedScope) {
		t.Fatalf("provider-native rule error = %v, want unsupported scope", err)
	}
}

func TestFirewallRuleFromDomainPersistsOnlyFirewalldPriority(t *testing.T) {
	priority := -100
	rule := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
			Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
		},
		NativeKind: filter.NativeKindRichRule, Protocol: "tcp", DestinationPort: "443",
		Action: filter.ActionAccept, Priority: &priority,
	}
	firewalld, err := FirewallRuleFromDomain(rule)
	if err != nil {
		t.Fatal(err)
	}
	if firewalld.Priority == nil || *firewalld.Priority != priority || firewalld.Sequence != nil {
		t.Fatalf("firewalld placement was not persisted: %#v", firewalld)
	}

	rule.Scope = filter.Scope{
		Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	rule.NativeKind = filter.NativeKindRule
	rule.Priority = nil
	iptables, err := FirewallRuleFromDomain(rule)
	if err != nil {
		t.Fatal(err)
	}
	if iptables.Priority != nil || iptables.Sequence != nil {
		t.Fatalf("positional backend persisted provider priority: %#v", iptables)
	}
}
