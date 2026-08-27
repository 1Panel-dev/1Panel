package lifecycle

import (
	"errors"
	"fmt"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle/providers"
)

const (
	ProviderFirewalld = constant.FirewallProviderFirewalld
	ProviderUFW       = constant.FirewallProviderUFW
	ProviderIptables  = constant.FirewallProviderIptables
	ProviderNftables  = constant.FirewallProviderNftables
)

var ErrNotInstalled = errors.New("is not installed")

type IptablesCommands struct {
	IPv4     string
	IPv6     string
	Restore4 string
	Restore6 string
}

func (c IptablesCommands) IPv6Available() bool {
	return c.IPv6 != "" && c.Restore6 != ""
}

type Runtime struct {
	Provider string
	Iptables IptablesCommands
}

var which = cmd.Which

func DetectRuntime() (Runtime, error) {
	hasFirewalld := which("firewalld")
	hasUFW := which("ufw")
	if hasFirewalld && hasUFW {
		return Runtime{}, errors.New("it is detected that the system has both firewalld and ufw services. To avoid conflicts, please uninstall and try again")
	}
	if hasFirewalld {
		return Runtime{Provider: ProviderFirewalld}, nil
	}
	if hasUFW {
		return Runtime{Provider: ProviderUFW}, nil
	}
	if commands, ok := detectIptablesCommands(""); ok {
		return Runtime{Provider: ProviderIptables, Iptables: commands}, nil
	}
	if commands, ok := detectIptablesCommands("-nft"); ok {
		return Runtime{Provider: ProviderIptables, Iptables: commands}, nil
	}
	if which("nft") {
		return Runtime{Provider: ProviderNftables}, nil
	}
	return Runtime{}, errors.New("no system firewall service detected (firewalld/ufw/iptables/iptables-nft/nft), please check and try again")
}

func detectIptablesCommands(suffix string) (IptablesCommands, bool) {
	ipv4 := "iptables" + suffix
	restore4 := "iptables" + suffix + "-restore"
	if !which(ipv4) || !which(restore4) {
		return IptablesCommands{}, false
	}
	commands := IptablesCommands{IPv4: ipv4, Restore4: restore4}
	ipv6 := "ip6tables" + suffix
	restore6 := "ip6tables" + suffix + "-restore"
	if which(ipv6) && which(restore6) {
		commands.IPv6 = ipv6
		commands.Restore6 = restore6
	}
	return commands, true
}

func ResolveIptablesCommands() (IptablesCommands, error) {
	if commands, ok := detectIptablesCommands(""); ok {
		return commands, nil
	}
	if commands, ok := detectIptablesCommands("-nft"); ok {
		return commands, nil
	}
	return IptablesCommands{}, fmt.Errorf("no complete iptables command family is available")
}

type Client interface {
	Name() string
	Start() error
	Stop() error
	Restart() error
	Status() (bool, error)
	Version() (string, error)
}

// Resetter restores a service-backed firewall to its installation defaults.
// Implementations must leave the firewall disabled after a successful reset.
type Resetter interface {
	Reset() error
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
			return nil, fmt.Errorf("firewalld %w", ErrNotInstalled)
		}
		return providers.NewFirewalld()
	case "ufw":
		if !which("ufw") {
			return nil, fmt.Errorf("ufw %w", ErrNotInstalled)
		}
		return providers.NewUFW()
	case "iptables":
		commands, err := ResolveIptablesCommands()
		if err != nil {
			return nil, fmt.Errorf("iptables %w: %v", ErrNotInstalled, err)
		}
		return providers.NewIptables(commands.IPv4)
	case "nftables":
		if !which("nft") {
			return nil, fmt.Errorf("nftables %w", ErrNotInstalled)
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

type State struct {
	Name     string
	IsActive bool
}

type Status struct {
	State
	Version string
}

func LoadState(client Client) (State, error) {
	state := State{Name: client.Name()}
	var err error
	state.IsActive, err = client.Status()
	return state, err
}

func LoadStatus(client Client) (Status, error) {
	status := Status{Version: "-"}
	var state State
	var version string
	var stateErr, versionErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		state, stateErr = LoadState(client)
	}()
	go func() {
		defer wg.Done()
		version, versionErr = client.Version()
	}()
	wg.Wait()
	status.State = state
	if stateErr != nil {
		return status, errors.Join(stateErr, versionErr)
	}
	if !status.IsActive {
		return status, nil
	}
	status.Version = version
	return status, versionErr
}
