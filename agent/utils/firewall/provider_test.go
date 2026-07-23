package firewall

import (
	"strings"
	"testing"
)

func TestResolveProviderFromPresence(t *testing.T) {
	tests := []struct {
		name      string
		firewalld bool
		ufw       bool
		iptables  bool
		want      string
		wantErr   string
	}{
		{
			name:      "firewalld preferred",
			firewalld: true,
			want:      "firewalld",
		},
		{
			name: "ufw",
			ufw:  true,
			want: "ufw",
		},
		{
			name:     "iptables",
			iptables: true,
			want:     "iptables",
		},
		{
			name:      "conflict",
			firewalld: true,
			ufw:       true,
			wantErr:   "both firewalld and ufw",
		},
		{
			name:    "none",
			wantErr: "No system firewall service detected",
		},
		{
			name:      "firewalld wins over iptables",
			firewalld: true,
			iptables:  true,
			want:      "firewalld",
		},
		{
			name:     "ufw wins over iptables",
			ufw:      true,
			iptables: true,
			want:     "ufw",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveProviderFromPresence(tt.firewalld, tt.ufw, tt.iptables)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				if got != "" {
					t.Fatalf("failed detection must not name a provider, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestNewClientByNameRejectsUnknownProvider(t *testing.T) {
	if _, err := newClientByName("unknown"); err == nil {
		t.Fatal("unknown provider must not build a filter client")
	}
	for _, name := range []string{"ufw", "firewalld", "iptables"} {
		client, err := newClientByName(name)
		if err != nil {
			t.Fatal(err)
		}
		if client.Name() != name {
			t.Fatalf("got client %q want %q", client.Name(), name)
		}
	}
}
