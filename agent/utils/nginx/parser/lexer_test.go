package parser

import (
	"slices"
	"testing"
)

func TestParseBracedVariables(t *testing.T) {
	config, err := NewStringParser(`server {
    listen 80;
    return 301 https://${host}$request_uri;
    add_header X-Scheme ${scheme}://${host};
}`).Parse()
	if err != nil {
		t.Fatal(err)
	}
	servers := config.FindServers()
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	tests := map[string][]string{
		"return":     {"301", "https://${host}$request_uri"},
		"add_header": {"X-Scheme", "${scheme}://${host}"},
	}
	for name, want := range tests {
		directives := servers[0].FindDirectives(name)
		if len(directives) != 1 {
			t.Fatalf("expected 1 %s directive, got %d", name, len(directives))
		}
		if got := directives[0].GetParameters(); !slices.Equal(got, want) {
			t.Fatalf("%s: expected parameters %q, got %q", name, want, got)
		}
	}
}
