package client

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

// Expansion is the only place provider specific rule shapes are decided, so the
// expected sequences below are the dev-v2 command order written down literally.

type portExpander struct {
	name   string
	expand func(FireInfo) []PortUnit
	rich   func(FireInfo) bool
}

func portExpanders(t *testing.T) []portExpander {
	t.Helper()
	ufw, err := NewUfw()
	if err != nil {
		t.Fatal(err)
	}
	firewalld, err := NewFirewalld()
	if err != nil {
		t.Fatal(err)
	}
	iptablesClient, err := NewIptables()
	if err != nil {
		t.Fatal(err)
	}
	return []portExpander{
		{name: "ufw", expand: ufw.ExpandPortRule, rich: ufwNeedsRichRule},
		{name: "firewalld", expand: firewalld.ExpandPortRule, rich: needsRichRule},
		{name: "iptables", expand: iptablesClient.ExpandPortRule, rich: needsRichRule},
	}
}

func formatPortUnit(unit PortUnit, rich bool) string {
	return fmt.Sprintf("apply(port=%s proto=%s addr=%s) record(port=%s proto=%s addr=%s) chain=%s rich=%v",
		unit.Apply.Port, unit.Apply.Protocol, unit.Apply.Address,
		unit.Record.Port, unit.Record.Protocol, unit.Record.Address,
		unit.Chain, rich)
}

func TestExpandPortRuleGoldenSequence(t *testing.T) {
	tests := []struct {
		name string
		rule FireInfo
		want map[string][]string
	}{
		{
			name: "single port",
			rule: FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept"},
			want: map[string][]string{
				"ufw":       {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=Anywhere) chain= rich=false"},
				"firewalld": {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain= rich=false"},
				"iptables":  {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC rich=false"},
			},
		},
		{
			name: "port range",
			rule: FireInfo{Port: "8000-8010", Protocol: "tcp", Strategy: "accept"},
			want: map[string][]string{
				"ufw":       {"apply(port=8000:8010 proto=tcp addr=) record(port=8000-8010 proto=tcp addr=Anywhere) chain= rich=false"},
				"firewalld": {"apply(port=8000-8010 proto=tcp addr=) record(port=8000-8010 proto=tcp addr=) chain= rich=false"},
				"iptables":  {"apply(port=8000-8010 proto=tcp addr=) record(port=8000-8010 proto=tcp addr=) chain=1PANEL_BASIC rich=false"},
			},
		},
		{
			name: "port range already colon separated",
			rule: FireInfo{Port: "8000:8010", Protocol: "tcp", Strategy: "accept"},
			want: map[string][]string{
				"ufw":       {"apply(port=8000:8010 proto=tcp addr=) record(port=8000:8010 proto=tcp addr=Anywhere) chain= rich=false"},
				"firewalld": {"apply(port=8000:8010 proto=tcp addr=) record(port=8000:8010 proto=tcp addr=) chain= rich=false"},
				"iptables":  {"apply(port=8000:8010 proto=tcp addr=) record(port=8000:8010 proto=tcp addr=) chain=1PANEL_BASIC rich=false"},
			},
		},
		{
			name: "comma separated ports",
			rule: FireInfo{Port: "80,443", Protocol: "tcp", Strategy: "accept"},
			want: map[string][]string{
				"ufw": {"apply(port=80,443 proto=tcp addr=) record(port=80,443 proto=tcp addr=Anywhere) chain= rich=false"},
				"firewalld": {
					"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain= rich=false",
					"apply(port=443 proto=tcp addr=) record(port=443 proto=tcp addr=) chain= rich=false",
				},
				"iptables": {
					"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC rich=false",
					"apply(port=443 proto=tcp addr=) record(port=443 proto=tcp addr=) chain=1PANEL_BASIC rich=false",
				},
			},
		},
		{
			name: "dual protocol single port",
			rule: FireInfo{Port: "53", Protocol: "tcp/udp", Strategy: "accept"},
			want: map[string][]string{
				"ufw": {"apply(port=53 proto= addr=) record(port=53 proto=tcp/udp addr=Anywhere) chain= rich=false"},
				"firewalld": {
					"apply(port=53 proto=tcp addr=) record(port=53 proto=tcp addr=) chain= rich=false",
					"apply(port=53 proto=udp addr=) record(port=53 proto=udp addr=) chain= rich=false",
				},
				"iptables": {
					"apply(port=53 proto=tcp addr=) record(port=53 proto=tcp addr=) chain=1PANEL_BASIC rich=false",
					"apply(port=53 proto=udp addr=) record(port=53 proto=udp addr=) chain=1PANEL_BASIC rich=false",
				},
			},
		},
		{
			name: "dual protocol comma separated ports",
			rule: FireInfo{Port: "80,443", Protocol: "tcp/udp", Strategy: "accept"},
			want: map[string][]string{
				"ufw": {
					"apply(port=80,443 proto=tcp addr=) record(port=80,443 proto=tcp addr=Anywhere) chain= rich=false",
					"apply(port=80,443 proto=udp addr=) record(port=80,443 proto=udp addr=Anywhere) chain= rich=false",
				},
				"firewalld": {
					"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain= rich=false",
					"apply(port=443 proto=tcp addr=) record(port=443 proto=tcp addr=) chain= rich=false",
					"apply(port=80 proto=udp addr=) record(port=80 proto=udp addr=) chain= rich=false",
					"apply(port=443 proto=udp addr=) record(port=443 proto=udp addr=) chain= rich=false",
				},
				"iptables": {
					"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC rich=false",
					"apply(port=443 proto=tcp addr=) record(port=443 proto=tcp addr=) chain=1PANEL_BASIC rich=false",
					"apply(port=80 proto=udp addr=) record(port=80 proto=udp addr=) chain=1PANEL_BASIC rich=false",
					"apply(port=443 proto=udp addr=) record(port=443 proto=udp addr=) chain=1PANEL_BASIC rich=false",
				},
			},
		},
		{
			name: "single source",
			rule: FireInfo{Port: "22", Protocol: "tcp", Strategy: "accept", Address: "10.0.0.1"},
			want: map[string][]string{
				"ufw":       {"apply(port=22 proto=tcp addr=10.0.0.1) record(port=22 proto=tcp addr=10.0.0.1) chain= rich=true"},
				"firewalld": {"apply(port=22 proto=tcp addr=10.0.0.1) record(port=22 proto=tcp addr=10.0.0.1) chain= rich=true"},
				"iptables":  {"apply(port=22 proto=tcp addr=10.0.0.1) record(port=22 proto=tcp addr=10.0.0.1) chain=1PANEL_BASIC rich=true"},
			},
		},
		{
			name: "cidr source",
			rule: FireInfo{Port: "22", Protocol: "tcp", Strategy: "accept", Address: "10.0.0.0/24"},
			want: map[string][]string{
				"ufw":       {"apply(port=22 proto=tcp addr=10.0.0.0/24) record(port=22 proto=tcp addr=10.0.0.0/24) chain= rich=true"},
				"firewalld": {"apply(port=22 proto=tcp addr=10.0.0.0/24) record(port=22 proto=tcp addr=10.0.0.0/24) chain= rich=true"},
				"iptables":  {"apply(port=22 proto=tcp addr=10.0.0.0/24) record(port=22 proto=tcp addr=10.0.0.0/24) chain=1PANEL_BASIC rich=true"},
			},
		},
		{
			name: "multiple sources",
			rule: FireInfo{Port: "22", Protocol: "tcp", Strategy: "accept", Address: "1.1.1.1,2.2.2.2"},
			want: map[string][]string{
				"ufw": {
					"apply(port=22 proto=tcp addr=1.1.1.1) record(port=22 proto=tcp addr=1.1.1.1) chain= rich=true",
					"apply(port=22 proto=tcp addr=2.2.2.2) record(port=22 proto=tcp addr=2.2.2.2) chain= rich=true",
				},
				"firewalld": {
					"apply(port=22 proto=tcp addr=1.1.1.1) record(port=22 proto=tcp addr=1.1.1.1) chain= rich=true",
					"apply(port=22 proto=tcp addr=2.2.2.2) record(port=22 proto=tcp addr=2.2.2.2) chain= rich=true",
				},
				"iptables": {
					"apply(port=22 proto=tcp addr=1.1.1.1) record(port=22 proto=tcp addr=1.1.1.1) chain=1PANEL_BASIC rich=true",
					"apply(port=22 proto=tcp addr=2.2.2.2) record(port=22 proto=tcp addr=2.2.2.2) chain=1PANEL_BASIC rich=true",
				},
			},
		},
		{
			name: "trailing comma source",
			rule: FireInfo{Port: "22", Protocol: "tcp", Strategy: "accept", Address: "1.1.1.1,"},
			want: map[string][]string{
				"ufw":       {"apply(port=22 proto=tcp addr=1.1.1.1) record(port=22 proto=tcp addr=1.1.1.1) chain= rich=true"},
				"firewalld": {"apply(port=22 proto=tcp addr=1.1.1.1) record(port=22 proto=tcp addr=1.1.1.1) chain= rich=true"},
				"iptables":  {"apply(port=22 proto=tcp addr=1.1.1.1) record(port=22 proto=tcp addr=1.1.1.1) chain=1PANEL_BASIC rich=true"},
			},
		},
		{
			name: "anywhere source",
			rule: FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept", Address: "Anywhere"},
			want: map[string][]string{
				"ufw":       {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=Anywhere) chain= rich=false"},
				"firewalld": {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain= rich=false"},
				"iptables":  {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC rich=false"},
			},
		},
		{
			name: "drop without source",
			rule: FireInfo{Port: "3306", Protocol: "tcp", Strategy: "drop"},
			want: map[string][]string{
				"ufw":       {"apply(port=3306 proto=tcp addr=) record(port=3306 proto=tcp addr=Anywhere) chain= rich=false"},
				"firewalld": {"apply(port=3306 proto=tcp addr=) record(port=3306 proto=tcp addr=) chain= rich=true"},
				"iptables":  {"apply(port=3306 proto=tcp addr=) record(port=3306 proto=tcp addr=) chain=1PANEL_BASIC rich=true"},
			},
		},
		{
			name: "explicit chain is preserved",
			rule: FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept", Chain: iptables.Chain1PanelBasicBefore},
			want: map[string][]string{
				"ufw":       {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=Anywhere) chain=1PANEL_BASIC_BEFORE rich=false"},
				"firewalld": {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC_BEFORE rich=false"},
				"iptables":  {"apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC_BEFORE rich=false"},
			},
		},
	}

	for _, tt := range tests {
		for _, expander := range portExpanders(t) {
			t.Run(tt.name+"/"+expander.name, func(t *testing.T) {
				units := expander.expand(tt.rule)
				got := make([]string, 0, len(units))
				for _, unit := range units {
					got = append(got, formatPortUnit(unit, expander.rich(unit.Apply)))
				}
				want := tt.want[expander.name]
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("expansion changed\ngot  %#v\nwant %#v", got, want)
				}
			})
		}
	}
}

func TestExpandAddressRuleGoldenSequence(t *testing.T) {
	ufw, err := NewUfw()
	if err != nil {
		t.Fatal(err)
	}
	firewalld, err := NewFirewalld()
	if err != nil {
		t.Fatal(err)
	}
	iptablesClient, err := NewIptables()
	if err != nil {
		t.Fatal(err)
	}
	expanders := map[string]func(FireInfo) []AddressUnit{
		"ufw":       ufw.ExpandAddressRule,
		"firewalld": firewalld.ExpandAddressRule,
		"iptables":  iptablesClient.ExpandAddressRule,
	}
	chains := map[string]string{"ufw": "", "firewalld": "", "iptables": iptables.Chain1PanelBasic}

	tests := []struct {
		name      string
		rule      FireInfo
		addresses []string
	}{
		{name: "single source", rule: FireInfo{Address: "10.0.0.1", Strategy: "drop"}, addresses: []string{"10.0.0.1"}},
		{name: "cidr source", rule: FireInfo{Address: "10.0.0.0/24", Strategy: "drop"}, addresses: []string{"10.0.0.0/24"}},
		{name: "multiple sources", rule: FireInfo{Address: "1.1.1.1,2.2.2.2", Strategy: "accept"}, addresses: []string{"1.1.1.1", "2.2.2.2"}},
		{name: "empty entries dropped", rule: FireInfo{Address: "1.1.1.1,,2.2.2.2", Strategy: "accept"}, addresses: []string{"1.1.1.1", "2.2.2.2"}},
		{name: "trailing comma", rule: FireInfo{Address: "1.1.1.1,", Strategy: "accept"}, addresses: []string{"1.1.1.1"}},
		{name: "empty source expands to nothing", rule: FireInfo{Address: "", Strategy: "accept"}},
		// address rules are not normalized, unlike port rules
		{name: "anywhere is kept verbatim", rule: FireInfo{Address: "Anywhere", Strategy: "drop"}, addresses: []string{"Anywhere"}},
	}

	for _, tt := range tests {
		for name, expand := range expanders {
			t.Run(tt.name+"/"+name, func(t *testing.T) {
				units := expand(tt.rule)
				if len(units) != len(tt.addresses) {
					t.Fatalf("got %d units want %d: %#v", len(units), len(tt.addresses), units)
				}
				for i, unit := range units {
					if unit.Apply.Address != tt.addresses[i] {
						t.Fatalf("unit %d address %q want %q", i, unit.Apply.Address, tt.addresses[i])
					}
					if unit.Apply.Strategy != tt.rule.Strategy {
						t.Fatalf("unit %d strategy %q want %q", i, unit.Apply.Strategy, tt.rule.Strategy)
					}
					if unit.Chain != chains[name] {
						t.Fatalf("unit %d chain %q want %q", i, unit.Chain, chains[name])
					}
				}
			})
		}
	}
}

func TestNormalizePortWhiteListContract(t *testing.T) {
	got := normalizePortWhiteList([]PortWhiteListEntry{
		{Port: "22", Protocol: "tcp"},
		{Port: "22", Protocol: "tcp"},
		{Port: "", Protocol: "tcp"},
		{Port: "80", Protocol: "tcp"},
	})
	want := []PortWhiteListEntry{
		{Port: "22", Protocol: "tcp"},
		{Port: "80", Protocol: "tcp"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}
