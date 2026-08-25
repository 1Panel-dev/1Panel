package nftables_helper

import (
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
)

func TestHasBaseChainBinding(t *testing.T) {
	if hasBaseChainBinding(`chain NFT_1PANEL_INPUT { policy accept; }`) {
		t.Fatal("empty input chain was treated as bound")
	}
	if !hasBaseChainBinding(`jump NFT_1PANEL_BASIC`) {
		t.Fatal("partial nftables binding was not detected")
	}
}

func TestRequiredPortChangesPreserveExistingRules(t *testing.T) {
	existing := []requiredPortRule{
		{Key: "22/tcp", Handle: "12"},
		{Key: "22/tcp", Handle: "13"},
		{Key: "80/tcp", Handle: "14"},
	}
	desired := []firewall.PortWhitelist{{Port: "22", Protocol: "tcp"}, {Port: "53", Protocol: "udp"}}
	missing, stale := requiredPortChanges(existing, desired)
	if want := []firewall.PortWhitelist{{Port: "53", Protocol: "udp"}}; !reflect.DeepEqual(missing, want) {
		t.Fatalf("missing=%#v, want %#v", missing, want)
	}
	if want := []string{"13", "14"}; !reflect.DeepEqual(stale, want) {
		t.Fatalf("stale=%#v, want %#v", stale, want)
	}
}

func TestRequiredPortRules(t *testing.T) {
	output := `
		tcp dport 22 accept comment "1Panel Port Whitelist" # handle 12
		tcp dport 80 accept comment "external" # handle 13
		udp dport 53 accept comment "1Panel Port Whitelist" # handle 14
	`
	want := []requiredPortRule{{Key: "22/tcp", Handle: "12"}, {Key: "53/udp", Handle: "14"}}
	if got := requiredPortRules(output); !reflect.DeepEqual(got, want) {
		t.Fatalf("rules=%#v, want %#v", got, want)
	}
}
