package forwarding

import (
	"errors"
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
