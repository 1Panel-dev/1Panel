package providers

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

func New(provider string) (forwarding.Adapter, error) {
	switch provider {
	case "iptables":
		return newIptablesNATAdapter(provider), nil
	case "nftables":
		return newNftablesAdapter(), nil
	default:
		return nil, errors.New("unsupported forwarding provider: " + provider)
	}
}
