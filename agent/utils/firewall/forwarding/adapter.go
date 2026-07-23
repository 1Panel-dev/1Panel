package forwarding

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

const (
	ChainPreRouting  = "1PANEL_PREROUTING"
	ChainPostRouting = "1PANEL_POSTROUTING"
	ChainForward     = "1PANEL_FORWARD"

	ForwardFile     = "1panel_forward.rules"
	PreRoutingFile  = "1panel_forward_pre.rules"
	PostRoutingFile = "1panel_forward_post.rules"
)

type Rule struct {
	Num        string
	Protocol   string
	Port       string
	TargetIP   string
	TargetPort string
	Interface  string
}

// Adapter is the complete provider-specific forwarding surface. Filter
// clients intentionally do not implement any of these methods.
type Adapter interface {
	Name() string
	List() ([]Rule, error)
	Operate(rule Rule, operation string) error
	Enable() error
	InitStatus() (bool, bool)
	Replay() error
	CleanupOwned() error
}

func NewAdapter(provider string) (Adapter, error) {
	switch provider {
	case "firewalld":
		return newFirewalldAdapter(), nil
	case "ufw", "iptables":
		return newLegacyNATAdapter(provider), nil
	default:
		return nil, errors.New("unsupported forwarding provider: " + provider)
	}
}

// IsOwnedChain is deliberately exact. Cleanup must never select a built-in or
// third-party NAT/filter chain.
func IsOwnedChain(table, chain string) bool {
	switch table {
	case iptables.NatTab:
		return chain == ChainPreRouting || chain == ChainPostRouting
	case iptables.FilterTab:
		return chain == ChainForward
	default:
		return false
	}
}
