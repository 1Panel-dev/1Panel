package firewall

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
)

type FirewallClient interface {
	Name() string // ufw firewalld
	Start() error
	Stop() error
	Restart() error
	Reload() error
	Status() (bool, error)
	Version() (string, error)

	ListPort() ([]client.FireInfo, error)
	ListForward() ([]client.FireInfo, error)
	ListAddress() ([]client.FireInfo, error)

	Port(port client.FireInfo, operation string) error
	RichRules(rule client.FireInfo, operation string) error
	PortForward(info client.Forward, operation string) error

	EnableForward() error
}

func NewFirewallClient() (FirewallClient, error) {
	firewalld := cmd.Which("firewalld")
	ufw := cmd.Which("ufw")

	if firewalld && ufw {
		return nil, errors.New("It is detected that the system has both firewalld and ufw services. To avoid conflicts, please uninstall and try again!")
	}
	if firewalld {
		return client.NewFirewalld()
	}
	if ufw {
		return client.NewUfw()
	}

	iptables := cmd.Which("iptables")
	if iptables {
		return client.NewIptables()
	}
	return nil, errors.New("No system firewall service detected (firewalld/ufw/iptables), please check and try again!")
}
