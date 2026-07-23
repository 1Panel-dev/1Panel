package client

import (
	"strings"
	"testing"
)

// UFW/firewalld filter operations remain thin native proxies and never emit managed filter chains.

func TestUfwPortCommandArgsContract(t *testing.T) {
	tests := []struct {
		name      string
		port      FireInfo
		operation string
		want      []string
	}{
		{
			name:      "add accept tcp",
			port:      FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept"},
			operation: "add",
			want:      []string{"allow", "80/tcp"},
		},
		{
			name:      "remove drop udp",
			port:      FireInfo{Port: "53", Protocol: "udp", Strategy: "drop"},
			operation: "remove",
			want:      []string{"delete", "deny", "53/udp"},
		},
		{
			name:      "add without protocol",
			port:      FireInfo{Port: "443", Strategy: "accept"},
			operation: "add",
			want:      []string{"allow", "443"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildUfwPortArgs(tt.port, tt.operation)
			if err != nil {
				t.Fatal(err)
			}
			assertNoManagedChainToken(t, got)
			assertStringSliceEqual(t, got, tt.want)
		})
	}
}

func TestUfwRichRuleCommandArgsContract(t *testing.T) {
	rule := FireInfo{
		Address:  "10.0.0.1",
		Port:     "22",
		Protocol: "tcp",
		Strategy: "accept",
	}
	var err error
	rule.Strategy, err = normalizeUfwStrategy(rule.Strategy)
	if err != nil {
		t.Fatal(err)
	}
	got := buildUfwRichRuleArgs(rule, "add", 1)
	want := []string{"insert", "1", "allow", "proto", "tcp", "from", "10.0.0.1", "to", "any", "port", "22"}
	assertNoManagedChainToken(t, got)
	assertStringSliceEqual(t, got, want)

	removeRule := FireInfo{
		Address:  "10.0.0.1",
		Protocol: "tcp",
		Strategy: "drop",
	}
	removeRule.Strategy, err = normalizeUfwStrategy(removeRule.Strategy)
	if err != nil {
		t.Fatal(err)
	}
	removeArgs := buildUfwRichRuleArgs(removeRule, "remove", 1)
	wantRemove := []string{"delete", "deny", "proto", "tcp", "from", "10.0.0.1"}
	assertNoManagedChainToken(t, removeArgs)
	assertStringSliceEqual(t, removeArgs, wantRemove)
}

func TestFirewalldPortCommandArgsContract(t *testing.T) {
	got := buildFirewalldPortArgs(FireInfo{Port: "80", Protocol: "tcp"}, "add")
	want := []string{"--zone=public", "--add-port=80/tcp", "--permanent"}
	assertNoManagedChainToken(t, got)
	assertStringSliceEqual(t, got, want)
}

func TestFirewalldRichRuleCommandContract(t *testing.T) {
	ruleStr := buildFirewalldRichRuleString(FireInfo{
		Address:  "1.2.3.4",
		Port:     "443",
		Protocol: "tcp",
		Strategy: "accept",
	})
	if strings.Contains(ruleStr, "1PANEL_") {
		t.Fatalf("firewalld rich rule must not contain managed chain: %s", ruleStr)
	}
	want := "rule family=ipv4 source address=1.2.3.4 port port=443 protocol=tcp accept"
	if ruleStr != want {
		t.Fatalf("got %q want %q", ruleStr, want)
	}

	args := buildFirewalldRichRuleArgs(ruleStr, "add")
	assertNoManagedChainToken(t, args)
	assertStringSliceEqual(t, args, []string{"--zone=public", "--add-rich-rule", ruleStr, "--permanent"})

	dualStack := buildFirewalldRichRuleStrings(FireInfo{Port: "53", Protocol: "udp", Strategy: "drop"})
	if len(dualStack) != 2 ||
		!strings.Contains(dualStack[0], "family=ipv4") ||
		!strings.Contains(dualStack[1], "family=ipv6") {
		t.Fatalf("empty-source rich rule must preserve dev-v2 dual-stack commands: %#v", dualStack)
	}
}

func TestExternalFilterCommandsNeverEmitManagedChains(t *testing.T) {
	samples := [][]string{
		mustUfwPortArgs(t, FireInfo{Port: "80", Protocol: "tcp", Strategy: "accept"}, "add"),
		mustUfwPortArgs(t, FireInfo{Port: "443", Protocol: "tcp", Strategy: "drop"}, "remove"),
		mustUfwRichRuleArgs(t, FireInfo{Address: "8.8.8.8", Port: "53", Protocol: "udp", Strategy: "accept"}, "add", 2),
		buildFirewalldPortArgs(FireInfo{Port: "22", Protocol: "tcp"}, "remove"),
		buildFirewalldRichRuleArgs(buildFirewalldRichRuleString(FireInfo{Address: "2001:db8::1", Strategy: "drop"}), "add"),
	}
	for i, args := range samples {
		assertNoManagedChainToken(t, args)
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "1PANEL_") {
			t.Fatalf("sample %d contains managed chain token: %s", i, joined)
		}
	}
}

func mustUfwPortArgs(t *testing.T, port FireInfo, operation string) []string {
	t.Helper()
	args, err := buildUfwPortArgs(port, operation)
	if err != nil {
		t.Fatal(err)
	}
	return args
}

func mustUfwRichRuleArgs(t *testing.T, rule FireInfo, operation string, insertNum int) []string {
	t.Helper()
	var err error
	rule.Strategy, err = normalizeUfwStrategy(rule.Strategy)
	if err != nil {
		t.Fatal(err)
	}
	return buildUfwRichRuleArgs(rule, operation, insertNum)
}

func assertNoManagedChainToken(t *testing.T, args []string) {
	t.Helper()
	for _, arg := range args {
		if strings.Contains(arg, "1PANEL_") {
			t.Fatalf("external command args must not contain managed chain token: %#v", args)
		}
	}
}
