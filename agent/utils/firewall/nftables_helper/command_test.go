package nftables_helper

import (
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func TestNativeNames(t *testing.T) {
	tests := []struct {
		family      filter.Family
		tableFamily string
	}{
		{family: filter.FamilyIPv4, tableFamily: "ip"},
		{family: filter.FamilyIPv6, tableFamily: "ip6"},
	}
	for _, test := range tests {
		if got := TableFamily(test.family); got != test.tableFamily {
			t.Fatalf("TableFamily(%s) = %q, want %q", test.family, got, test.tableFamily)
		}
	}

	wantChains := []string{"NFT_1PANEL_BASIC_BEFORE", "NFT_1PANEL_BASIC", "NFT_1PANEL_BASIC_AFTER"}
	for index, got := range BasicChains() {
		if got != wantChains[index] {
			t.Fatalf("BasicChains()[%d] = %q, want %q", index, got, wantChains[index])
		}
	}
}

func TestBuildBatchScriptCombinesCommands(t *testing.T) {
	script, err := buildBatchScript(
		[]string{"flush", "chain", "ip", TableName, BasicBeforeChain},
		[]string{"add", "rule", "ip", TableName, BasicBeforeChain, "tcp", "dport", "443", "accept"},
		[]string{"delete", "rule", "ip6", TableName, BasicBeforeChain, "handle", "12"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(script, "\n") != 3 || !strings.Contains(script, "tcp dport 443 accept\n") ||
		!strings.Contains(script, "delete rule ip6 "+TableName+" "+BasicBeforeChain+" handle 12\n") {
		t.Fatalf("unexpected nftables batch script:\n%s", script)
	}
}

func TestBuildBatchScriptRejectsNewline(t *testing.T) {
	if _, err := buildBatchScript([]string{"add", "rule", "ip", TableName, BasicChain, "unsafe\nrule"}); err == nil {
		t.Fatal("expected newline validation error")
	}
}
