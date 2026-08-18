package providers

import (
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

func TestNftForwardNamingContract(t *testing.T) {
	if nftForwardTable != "nft_1panel_forward" {
		t.Fatalf("unexpected nftables forwarding table %q", nftForwardTable)
	}
	if nftForwardFile != "1panel_forward.nft" {
		t.Fatalf("unexpected nftables forwarding rules file %q", nftForwardFile)
	}
	wantChains := map[string]string{
		forwarding.ChainPreRouting:  "NFT_1PANEL_PREROUTING",
		forwarding.ChainPostRouting: "NFT_1PANEL_POSTROUTING",
		forwarding.ChainForward:     "NFT_1PANEL_FORWARD",
	}
	for logical, want := range wantChains {
		if got := nftForwardChain(logical); got != want {
			t.Fatalf("nftForwardChain(%q) = %q, want %q", logical, got, want)
		}
	}
}

func TestNormalizeNftForwardRuleRejectsScriptTokens(t *testing.T) {
	base := forwarding.Rule{Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80"}
	tests := []struct {
		name   string
		mutate func(*forwarding.Rule)
	}{
		{name: "source port", mutate: func(rule *forwarding.Rule) { rule.Port = "8080\nflush ruleset" }},
		{name: "target address", mutate: func(rule *forwarding.Rule) { rule.TargetIP = "10.0.0.2; flush ruleset" }},
		{name: "target port", mutate: func(rule *forwarding.Rule) { rule.TargetPort = "80; flush ruleset" }},
		{name: "interface", mutate: func(rule *forwarding.Rule) { rule.Interface = "eth0;flush" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule := base
			test.mutate(&rule)
			if _, err := normalizeNftForwardRule(rule); err == nil {
				t.Fatal("expected invalid nftables forwarding rule")
			}
		})
	}
}

func TestRebuildNftForwardCommandsUseArguments(t *testing.T) {
	commands, err := rebuildNftForwardCommands([]forwarding.Rule{{
		Protocol: "tcp", Port: "08080", TargetIP: "10.0.0.2", TargetPort: "080", Interface: "eth0",
	}})
	if err != nil {
		t.Fatalf("build commands: %v", err)
	}
	if len(commands) != 10 {
		t.Fatalf("unexpected command count: %d", len(commands))
	}
	joined := strings.Join(commands[6], " ")
	if !strings.Contains(joined, "add rule ip nft_1panel_forward NFT_1PANEL_PREROUTING") ||
		!strings.Contains(joined, `iifname "eth0"`) || !strings.Contains(joined, "dport 8080") ||
		!strings.Contains(joined, "dnat to 10.0.0.2:80") {
		t.Fatalf("unexpected prerouting command: %q", joined)
	}
}

func TestNftForwardCommandsBuildSingleBatchScript(t *testing.T) {
	commands, err := rebuildNftForwardCommands([]forwarding.Rule{{
		Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80",
	}})
	if err != nil {
		t.Fatal(err)
	}
	script, err := nftCommandsScript(commands)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(script, "\n"); lines != len(commands) {
		t.Fatalf("batch script has %d lines, want %d:\n%s", lines, len(commands), script)
	}
	if !strings.Contains(script, "flush chain ip nft_1panel_forward NFT_1PANEL_PREROUTING\n") ||
		!strings.Contains(script, "add rule ip nft_1panel_forward NFT_1PANEL_PREROUTING meta l4proto tcp tcp dport 8080 redirect to :80") {
		t.Fatalf("unexpected nftables batch script:\n%s", script)
	}
	if _, err := nftCommandsScript([][]string{{"add", "rule\nflush ruleset"}}); err == nil {
		t.Fatal("batch script accepted a newline token")
	}
}

func TestRebuildNftIPv6ForwardCommands(t *testing.T) {
	commands, err := rebuildNftForwardCommands([]forwarding.Rule{{
		Family: forwarding.FamilyIPv6, Protocol: "tcp", Port: "8443", TargetIP: "2001:db8::20", TargetPort: "443", Interface: "eth0",
	}})
	if err != nil {
		t.Fatalf("build IPv6 commands: %v", err)
	}
	if len(commands) != 10 {
		t.Fatalf("unexpected command count: %d", len(commands))
	}
	preRouting := strings.Join(commands[6], " ")
	if !strings.Contains(preRouting, "add rule ip6 nft_1panel_forward NFT_1PANEL_PREROUTING") ||
		!strings.Contains(preRouting, "dnat to [2001:db8::20]:443") {
		t.Fatalf("unexpected IPv6 prerouting command: %q", preRouting)
	}
	forward := strings.Join(commands[8], " ")
	if !strings.Contains(forward, "ip6 daddr 2001:db8::20") {
		t.Fatalf("unexpected IPv6 forward command: %q", forward)
	}
}

func TestNormalizeForwardRuleEnforcesAddressFamily(t *testing.T) {
	ipv6, err := normalizeForwardRule(forwarding.Rule{Family: forwarding.FamilyIPv6, Protocol: "tcp", Port: "8080", TargetIP: "2001:db8::2", TargetPort: "80"})
	if err != nil || ipv6.TargetIP != "2001:db8::2" {
		t.Fatalf("normalize IPv6 rule = %#v, %v", ipv6, err)
	}
	if _, err := normalizeForwardRule(forwarding.Rule{Family: forwarding.FamilyIPv6, Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80"}); err == nil {
		t.Fatal("expected an IPv4 target to be rejected for an IPv6 rule")
	}
	loopback, err := normalizeForwardRule(forwarding.Rule{Family: forwarding.FamilyIPv6, Protocol: "udp", Port: "5353", TargetPort: "53"})
	if err != nil || loopback.TargetIP != "::1" {
		t.Fatalf("IPv6 loopback normalization = %#v, %v", loopback, err)
	}
}
