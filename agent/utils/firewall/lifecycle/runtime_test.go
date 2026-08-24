package lifecycle

import "testing"

func TestDetectRuntimePriority(t *testing.T) {
	original := which
	t.Cleanup(func() { which = original })

	tests := []struct {
		name       string
		commands   map[string]bool
		provider   string
		executable string
	}{
		{name: "default iptables", commands: map[string]bool{"iptables": true, "iptables-restore": true, "iptables-nft": true, "iptables-nft-restore": true, "nft": true}, provider: ProviderIptables, executable: "iptables"},
		{name: "explicit iptables nft", commands: map[string]bool{"iptables-nft": true, "iptables-nft-restore": true, "nft": true}, provider: ProviderIptables, executable: "iptables-nft"},
		{name: "native nft", commands: map[string]bool{"nft": true}, provider: ProviderNftables},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			which = func(name string) bool { return test.commands[name] }
			runtime, err := DetectRuntime()
			if err != nil {
				t.Fatalf("detect: %v", err)
			}
			if runtime.Provider != test.provider || runtime.Iptables.IPv4 != test.executable {
				t.Fatalf("unexpected runtime: %#v", runtime)
			}
		})
	}
}

func TestDetectRuntimeRequiresRestoreCommand(t *testing.T) {
	original := which
	t.Cleanup(func() { which = original })
	which = func(name string) bool { return name == "iptables" || name == "nft" }
	runtime, err := DetectRuntime()
	if err != nil || runtime.Provider != ProviderNftables {
		t.Fatalf("incomplete iptables family should fall back to nft: runtime=%#v err=%v", runtime, err)
	}
}
