package lifecycle

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle/providers"
)

type Client interface {
	Name() string
	Start() error
	Stop() error
	Restart() error
	Status() (bool, error)
	Version() (string, error)
}

func NewClient() (Client, error) {
	runtime, err := DetectRuntime()
	if err != nil {
		return nil, err
	}
	return NewClientFor(runtime.Provider)
}

func NewClientFor(provider string) (Client, error) {
	switch provider {
	case "firewalld":
		if !which("firewalld") {
			return nil, fmt.Errorf("firewalld is not installed")
		}
		return providers.NewFirewalld()
	case "ufw":
		if !which("ufw") {
			return nil, fmt.Errorf("ufw is not installed")
		}
		return providers.NewUFW()
	case "iptables":
		commands, err := ResolveIptablesCommands()
		if err != nil {
			return nil, err
		}
		return providers.NewIptables(commands.IPv4)
	case "nftables":
		if !which("nft") {
			return nil, fmt.Errorf("nftables is not installed")
		}
		return providers.NewNftables()
	default:
		return nil, fmt.Errorf("unsupported firewall provider: %s", provider)
	}
}

func InstalledProviders() []string {
	providers := make([]string, 0, 4)
	if which("firewalld") {
		providers = append(providers, ProviderFirewalld)
	}
	if which("ufw") {
		providers = append(providers, ProviderUFW)
	}
	if _, err := ResolveIptablesCommands(); err == nil {
		providers = append(providers, ProviderIptables)
	}
	if which("nft") {
		providers = append(providers, ProviderNftables)
	}
	return providers
}

// NewNetfilterClients resolves all forwarding backends independently of the
// host firewall service. Native nftables is listed first as the default for an
// uninitialized host; callers may still prefer an already initialized backend.
func NewNetfilterClients() ([]Client, error) {
	clients := make([]Client, 0, 2)
	if which("nft") {
		client, err := providers.NewNftables()
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if commands, err := ResolveIptablesCommands(); err == nil {
		client, err := providers.NewIptables(commands.IPv4)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("no supported forwarding backend detected (iptables/iptables-nft/nftables)")
	}
	return clients, nil
}

func DetectProvider() (string, error) {
	runtime, err := DetectRuntime()
	if err != nil {
		return "", err
	}
	return runtime.Provider, nil
}
