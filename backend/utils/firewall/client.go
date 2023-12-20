package firewall

import (
	"os"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/firewall/client"
)

type FirewallClient interface {
	Name() string // ufw firewalld
	Start() error
	Stop() error
	Restart() error
	Reload() error
	Status() (string, error) // running not running
	Version() (string, error)

	ListPort() ([]client.FireInfo, error)
	ListAddress() ([]client.FireInfo, error)

	Port(port client.FireInfo, operation string) error
	RichRules(rule client.FireInfo, operation string) error
	PortForward(info client.Forward, operation string) error
}

func NewFirewallClient() (FirewallClient, error) {
	if _, err := os.Stat("/usr/sbin/firewalld"); err == nil {
		return client.NewFirewalld()
	}
	if _, err := os.Stat("/usr/sbin/ufw"); err == nil {
		return client.NewUfw()
	}
	return nil, buserr.New(constant.ErrFirewall)
}
