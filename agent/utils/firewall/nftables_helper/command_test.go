package nftables_helper

import (
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
