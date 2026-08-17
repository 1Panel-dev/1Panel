package lifecycle

import "testing"

func TestNewNetfilterClientsIgnoreHostFirewallService(t *testing.T) {
	original := which
	t.Cleanup(func() { which = original })

	tests := []struct {
		name     string
		commands map[string]bool
		want     []string
		wantErr  bool
	}{
		{
			name: "firewalld host with iptables",
			commands: map[string]bool{
				"firewalld": true, "iptables": true, "iptables-restore": true,
			},
			want: []string{ProviderIptables},
		},
		{
			name: "ufw host with iptables nft",
			commands: map[string]bool{
				"ufw": true, "iptables-nft": true, "iptables-nft-restore": true, "nft": true,
			},
			want: []string{ProviderNftables, ProviderIptables},
		},
		{
			name:     "firewalld host with native nft",
			commands: map[string]bool{"firewalld": true, "nft": true},
			want:     []string{ProviderNftables},
		},
		{
			name:     "service without netfilter command backend",
			commands: map[string]bool{"firewalld": true},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			which = func(name string) bool { return test.commands[name] }
			clients, err := NewNetfilterClients()
			if (err != nil) != test.wantErr {
				t.Fatalf("NewNetfilterClients() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			if len(clients) != len(test.want) {
				t.Fatalf("NewNetfilterClients() returned %d clients, want %d", len(clients), len(test.want))
			}
			for index, client := range clients {
				if client.Name() != test.want[index] {
					t.Fatalf("NewNetfilterClients()[%d] = %q, want %q", index, client.Name(), test.want[index])
				}
			}
		})
	}
}
