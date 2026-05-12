package service

import (
	"testing"
)

func TestPortMatchesRule(t *testing.T) {
	tests := []struct {
		name    string
		portStr string
		portNum uint32
		rule    string
		want    bool
	}{
		{"exact match", "80", 80, "80", true},
		{"no match", "80", 80, "443", false},
		{"dash range match", "8080", 8080, "8000-9000", true},
		{"dash range lower bound", "8000", 8000, "8000-9000", true},
		{"dash range upper bound", "9000", 9000, "8000-9000", true},
		{"dash range miss below", "7999", 7999, "8000-9000", false},
		{"dash range miss above", "9001", 9001, "8000-9000", false},
		{"colon range match", "8500", 8500, "8000:9000", true},
		{"colon range miss", "7999", 7999, "8000:9000", false},
		{"comma list first", "80", 80, "80,443,8080", true},
		{"comma list middle", "443", 443, "80,443,8080", true},
		{"comma list last", "8080", 8080, "80,443,8080", true},
		{"comma list miss", "22", 22, "80,443,8080", false},
		{"comma with spaces", "443", 443, "80, 443, 8080", true},
		{"empty rule", "80", 80, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portMatchesRule(tt.portStr, tt.portNum, tt.rule)
			if got != tt.want {
				t.Errorf("portMatchesRule(%q, %d, %q) = %v, want %v", tt.portStr, tt.portNum, tt.rule, got, tt.want)
			}
		})
	}
}

func TestMatchFirewallRule(t *testing.T) {
	rules := []firewallRuleEntry{
		{strategy: "accept", portStr: "22", protocol: "tcp", address: ""},
		{strategy: "accept", portStr: "3306", protocol: "tcp", address: "10.0.0.0/8"},
		{strategy: "drop", portStr: "3306", protocol: "tcp", address: "1.2.3.4"},
		{strategy: "accept", portStr: "8000-9000", protocol: "tcp", address: ""},
		{strategy: "accept", portStr: "53", protocol: "tcp", address: ""},
		{strategy: "accept", portStr: "123", protocol: "tcp/udp", address: ""},
		{strategy: "accept", portStr: "80,443", protocol: "tcp", address: ""},
	}

	tests := []struct {
		name         string
		port         uint32
		proto        string
		wantStrategy string
		wantFound    bool
		wantAddress  string
	}{
		{"exact match tcp/22", 22, "tcp", "accept", true, ""},
		{"no match udp/22", 22, "udp", "", false, ""},
		{"drop wins over accept for 3306", 3306, "tcp", "drop", true, "1.2.3.4"},
		{"range match 8500", 8500, "tcp", "accept", true, ""},
		{"protocol mismatch udp/8500", 8500, "udp", "", false, ""},
		{"source-scoped accept 3306 returns address", 3306, "tcp", "drop", true, "1.2.3.4"},
		{"no match port 9999", 9999, "tcp", "", false, ""},
		{"comma list match 443", 443, "tcp", "accept", true, ""},
		{"comma list match 80", 80, "tcp", "accept", true, ""},
		{"tcp/udp protocol matches tcp", 123, "tcp", "accept", true, ""},
		{"tcp/udp protocol matches udp", 123, "udp", "accept", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, found, address := matchFirewallRule(rules, tt.port, tt.proto)
			if strategy != tt.wantStrategy || found != tt.wantFound || address != tt.wantAddress {
				t.Errorf("matchFirewallRule(port=%d, proto=%s) = (%q, %v, %q), want (%q, %v, %q)",
					tt.port, tt.proto, strategy, found, address, tt.wantStrategy, tt.wantFound, tt.wantAddress)
			}
		})
	}
}

func TestDetermineStatus(t *testing.T) {
	tests := []struct {
		name           string
		bindAddr       string
		sourceType     string
		firewallActive bool
		hasRule        bool
		ruleStrategy   string
		want           string
	}{
		{"loopback ipv4", "127.0.0.1", "host", true, false, "", "localOnly"},
		{"loopback ipv4 alt", "127.0.0.53", "host", true, false, "", "localOnly"},
		{"loopback ipv6", "::1", "host", true, false, "", "localOnly"},
		{"link-local ipv6", "fe80::1", "host", true, false, "", "localOnly"},
		{"wildcard ipv4", "0.0.0.0", "host", true, false, "", "noRule"},
		{"wildcard ipv6", "::", "host", true, false, "", "noRule"},
		{"empty addr", "", "host", true, false, "", "noRule"},
		{"public ip host", "114.212.85.247", "host", true, false, "", "noRule"},
		{"bridge gateway host", "172.17.0.1", "host", true, false, "", "noRule"},
		{"docker wildcard", "0.0.0.0", "docker", true, false, "", "dockerBypass"},
		{"docker loopback", "127.0.0.1", "docker", true, false, "", "localOnly"},
		{"docker specific ip", "172.17.0.1", "docker", true, false, "", "dockerBypass"},
		{"appStore wildcard", "0.0.0.0", "appStore", true, false, "", "dockerBypass"},
		{"firewall inactive host", "0.0.0.0", "host", false, false, "", "firewallInactive"},
		{"firewall inactive docker", "0.0.0.0", "docker", false, false, "", "dockerBypass"},
		{"accept rule", "0.0.0.0", "host", true, true, "accept", "protected"},
		{"drop rule", "0.0.0.0", "host", true, true, "drop", "blocked"},
		{"reject rule", "0.0.0.0", "host", true, true, "reject", "blocked"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineStatus(tt.bindAddr, tt.sourceType, tt.firewallActive, tt.hasRule, tt.ruleStrategy)
			if got != tt.want {
				t.Errorf("determineStatus(%q, %q, active=%v, rule=%v, %q) = %q, want %q",
					tt.bindAddr, tt.sourceType, tt.firewallActive, tt.hasRule, tt.ruleStrategy, got, tt.want)
			}
		})
	}
}

func TestIsLoopbackOrLinkLocal(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.53", true},
		{"127.0.0.54", true},
		{"::1", true},
		{"fe80::1", true},
		{"fe80::be24:11ff:fe31:1da9", true},
		{"0.0.0.0", false},
		{"::", false},
		{"", false},
		{"114.212.85.247", false},
		{"172.17.0.1", false},
		{"10.0.0.1", false},
		{"192.168.1.1", false},
		{"2001:db8::1", false},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := isLoopbackOrLinkLocal(tt.addr)
			if got != tt.want {
				t.Errorf("isLoopbackOrLinkLocal(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

func TestIsWildcardAddress(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"0.0.0.0", true},
		{"::", true},
		{"", true},
		{"127.0.0.1", false},
		{"172.17.0.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := isWildcardAddress(tt.addr)
			if got != tt.want {
				t.Errorf("isWildcardAddress(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}
