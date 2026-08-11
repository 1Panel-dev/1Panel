package lifecycle

import (
	"errors"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
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
	provider, err := DetectProvider()
	if err != nil {
		return nil, err
	}
	switch provider {
	case "firewalld":
		return providers.NewFirewalld()
	case "ufw":
		return providers.NewUFW()
	case "iptables":
		return providers.NewIptables()
	default:
		return nil, fmt.Errorf("unsupported firewall provider: %s", provider)
	}
}

func DetectProvider() (string, error) {
	hasFirewalld := cmd.Which("firewalld")
	hasUFW := cmd.Which("ufw")
	if hasFirewalld && hasUFW {
		return "", errors.New("it is detected that the system has both firewalld and ufw services. To avoid conflicts, please uninstall and try again")
	}
	if hasFirewalld {
		return "firewalld", nil
	}
	if hasUFW {
		return "ufw", nil
	}
	if cmd.Which("iptables") {
		return "iptables", nil
	}
	return "", errors.New("no system firewall service detected (firewalld/ufw/iptables), please check and try again")
}
