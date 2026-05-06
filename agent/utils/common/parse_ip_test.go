package common

import "testing"

// Regression test for 1Panel-dev/1Panel#12646: panel SSL self-signed flow
// rejected IPv6 hosts because net.ParseIP does not accept the bracketed
// form (e.g. "[::1]") and the upstream caller passes the host portion of
// a URL rather than a bare IP.
func TestParseIPLoose(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"bare ipv4", "127.0.0.1", true},
		{"bare ipv6", "::1", true},
		{"bracketed ipv6", "[::1]", true},
		{"bracketed full ipv6", "[2001:db8::1]", true},
		{"trimmed bare ipv6", "  ::1  ", true},
		{"trimmed bracketed ipv6", "  [::1]  ", true},
		{"empty", "", false},
		{"only brackets", "[]", false},
		{"hostname", "example.com", false},
		{"bracketed garbage", "[notanip]", false},
		{"unbalanced bracket left", "[::1", false},
		{"unbalanced bracket right", "::1]", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseIPLoose(tc.in) != nil
			if got != tc.want {
				t.Errorf("ParseIPLoose(%q) ok = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
