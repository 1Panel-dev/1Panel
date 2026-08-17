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
	switch runtime.Provider {
	case "firewalld":
		return providers.NewFirewalld()
	case "ufw":
		return providers.NewUFW()
	case "iptables":
		return providers.NewIptables(runtime.Iptables.IPv4)
	case "nftables":
		return providers.NewNftables()
	default:
		return nil, fmt.Errorf("unsupported firewall provider: %s", runtime.Provider)
	}
}

func DetectProvider() (string, error) {
	runtime, err := DetectRuntime()
	if err != nil {
		return "", err
	}
	return runtime.Provider, nil
}
