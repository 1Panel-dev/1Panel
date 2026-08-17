package nftables_helper

import (
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
)

func TestActiveLegacyChainRequiresJump(t *testing.T) {
	if got := activeLegacyChain(`chain 1PANEL_BASIC { type filter hook input priority filter; }`); got != "" {
		t.Fatalf("unbound legacy chain was treated as active: %q", got)
	}
	if got := activeLegacyChain(`-A INPUT -j 1PANEL_BASIC`); got != "1PANEL_BASIC" {
		t.Fatalf("iptables jump was not detected: %q", got)
	}
	if got := activeLegacyChain(`jump 1PANEL_BASIC_BEFORE`); got != "1PANEL_BASIC_BEFORE" {
		t.Fatalf("nftables jump was not detected: %q", got)
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
