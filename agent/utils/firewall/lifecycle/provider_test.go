package lifecycle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name        string
		executables []string
		want        string
		wantErr     bool
	}{
		{name: "none", wantErr: true},
		{name: "iptables", executables: []string{"iptables"}, want: "iptables"},
		{name: "ufw", executables: []string{"iptables", "ufw"}, want: "ufw"},
		{name: "firewalld", executables: []string{"iptables", "firewalld"}, want: "firewalld"},
		{name: "conflict", executables: []string{"firewalld", "ufw"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, executable := range test.executables {
				name := filepath.Join(directory, executable)
				if err := os.WriteFile(name, []byte("#!/bin/sh\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", directory)
			got, err := DetectProvider()
			if (err != nil) != test.wantErr {
				t.Fatalf("DetectProvider() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("DetectProvider() = %q, want %q", got, test.want)
			}
		})
	}
}
