package firewall

import "testing"

func TestParsePortWhitelistLegacyAndStructured(t *testing.T) {
	legacy, err := ParsePortWhitelist("80/tcp,53/udp")
	if err != nil {
		t.Fatalf("parse legacy whitelist: %v", err)
	}
	if len(legacy) != 2 || legacy[0] != (PortWhitelist{Family: "ipv4", Port: "80", Protocol: "tcp"}) {
		t.Fatalf("unexpected legacy whitelist: %#v", legacy)
	}

	structured, err := ParsePortWhitelist(`[
		{"family":"ipv6","protocol":"TCP","port":"8080:8090"},
		{"family":"ipv4","protocol":"udp","port":"53"}
	]`)
	if err != nil {
		t.Fatalf("parse structured whitelist: %v", err)
	}
	want := []PortWhitelist{
		{Family: "ipv6", Port: "8080-8090", Protocol: "tcp"},
		{Family: "ipv4", Port: "53", Protocol: "udp"},
	}
	if len(structured) != len(want) {
		t.Fatalf("unexpected structured whitelist: %#v", structured)
	}
	for index := range want {
		if structured[index] != want[index] {
			t.Fatalf("rule %d = %#v, want %#v", index, structured[index], want[index])
		}
	}
}

func TestParsePortWhitelistRejectsInvalidRule(t *testing.T) {
	for _, value := range []string{
		`[{"family":"inet","protocol":"tcp","port":"80"}]`,
		`[{"family":"ipv4","protocol":"icmp","port":"80"}]`,
		`[{"family":"ipv4","protocol":"tcp","port":"9000-8000"}]`,
		`[{"family":"ipv6","protocol":"udp","port":"65536"}]`,
		`[{"family":"ipv4","protocol":"tcp","port":"8000-8100"},{"family":"ipv4","protocol":"tcp","port":"8080"}]`,
	} {
		if _, err := ParsePortWhitelist(value); err == nil {
			t.Fatalf("expected %q to fail", value)
		}
	}
}

func TestParsePortWhitelistKeepsFamiliesDistinct(t *testing.T) {
	rules, err := ParsePortWhitelist(`[
		{"family":"ipv4","protocol":"tcp","port":"443"},
		{"family":"ipv6","protocol":"tcp","port":"443"},
		{"family":"ipv4","protocol":"tcp","port":"443"}
	]`)
	if err != nil {
		t.Fatalf("parse whitelist: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected family-specific deduplication, got %#v", rules)
	}
}

func TestNormalizePortWhitelistPrefersFamilyNeutralRequiredRule(t *testing.T) {
	rules := NormalizePortWhitelist([]PortWhitelist{
		{Family: "ipv4", Port: "22", Protocol: "tcp"},
		{Family: "ipv6", Port: "22", Protocol: "tcp"},
		{Port: "22", Protocol: "tcp"},
	})
	if len(rules) != 1 || rules[0].Family != "" {
		t.Fatalf("family-neutral rule should replace family-specific duplicates: %#v", rules)
	}
}
