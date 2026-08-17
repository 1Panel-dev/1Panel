package providers

import (
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

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
	if len(commands) != 7 {
		t.Fatalf("unexpected command count: %d", len(commands))
	}
	joined := strings.Join(commands[3], " ")
	if !strings.Contains(joined, `iifname "eth0"`) || !strings.Contains(joined, "dport 8080") || !strings.Contains(joined, "dnat to 10.0.0.2:80") {
		t.Fatalf("unexpected prerouting command: %q", joined)
	}
}
