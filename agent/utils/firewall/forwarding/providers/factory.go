package providers

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

func New(provider string) (forwarding.Adapter, error) {
	switch provider {
	case "firewalld":
		return newFirewalldAdapter(), nil
	case "ufw", "iptables":
		return newIptablesNATAdapter(provider), nil
	default:
		return nil, errors.New("unsupported forwarding provider: " + provider)
	}
}
