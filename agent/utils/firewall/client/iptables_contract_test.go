package client

import (
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

func TestNormalizePortSpec(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "single", input: "80", want: "80"},
		{name: "range dash", input: "8000-8010", want: "8000:8010"},
		{name: "range colon", input: "8000:8010", want: "8000:8010"},
		{name: "trim", input: " 443 ", want: "443"},
		{name: "empty", input: "", wantErr: true},
		{name: "invalid", input: "abc", wantErr: true},
		{name: "bad range order", input: "90-80", wantErr: true},
		{name: "out of range", input: "70000", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePortSpec(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizePortSpec(%q)=%q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	if _, err := parsePort("0"); err == nil {
		t.Fatal("expected error for port 0")
	}
	got, err := parsePort("22")
	if err != nil {
		t.Fatal(err)
	}
	if got != 22 {
		t.Fatalf("got %d", got)
	}
}

func TestIptablesPortRuleArgsContract(t *testing.T) {
	tests := []struct {
		name string
		port FireInfo
		want []string
	}{
		{
			name: "accept tcp default chain",
			port: FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept"},
			want: []string{"-p", "tcp", "--dport", "80", "-j", "ACCEPT"},
		},
		{
			name: "drop udp",
			port: FireInfo{Port: "53", Protocol: "udp", Strategy: "drop"},
			want: []string{"-p", "udp", "--dport", "53", "-j", "DROP"},
		},
		{
			name: "range",
			port: FireInfo{Port: "8000-8010", Protocol: "tcp", Strategy: "accept"},
			want: []string{"-p", "tcp", "--dport", "8000:8010", "-j", "ACCEPT"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildIptablesPortRuleArgs(tt.port)
			if err != nil {
				t.Fatal(err)
			}
			assertStringSliceEqual(t, got, tt.want)
		})
	}
}

func TestIptablesRichRuleArgsContract(t *testing.T) {
	got, err := buildIptablesRichRuleArgs(FireInfo{
		Address:  "1.2.3.4",
		Port:     "443",
		Protocol: "tcp",
		Strategy: "drop",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-s", "1.2.3.4", "-p", "tcp", "--dport", "443", "-j", "DROP"}
	assertStringSliceEqual(t, got, want)
}

func TestIptablesDefaultChainIsPanelOwned(t *testing.T) {
	if iptables.Chain1PanelBasic != "1PANEL_BASIC" {
		t.Fatalf("unexpected basic chain: %s", iptables.Chain1PanelBasic)
	}
	for _, chain := range []string{
		iptables.Chain1PanelBasicBefore,
		iptables.Chain1PanelBasic,
		iptables.Chain1PanelBasicAfter,
		iptables.Chain1PanelInput,
		iptables.Chain1PanelOutput,
	} {
		if !strings.HasPrefix(chain, "1PANEL_") {
			t.Fatalf("legacy chain %q must be 1PANEL_ owned", chain)
		}
	}
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d\ngot  %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("index %d: got %q want %q\nfull got %#v\nfull want %#v", i, got[i], want[i], got, want)
		}
	}
}
