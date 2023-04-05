package firewall

import (
	"github.com/1Panel-dev/1Panel/backend/utils/firewall/client"
)

type FirewallClient interface {
	Start() error
	Stop() error
	Reload() error
	Status() (string, error)
	ListPort() ([]client.FireInfo, error)
	ListAddress() ([]client.FireInfo, error)

	Port(port client.FireInfo, operation string) error
	RichRules(rule client.FireInfo, operation string) error
	PortForward(info client.Forward, operation string) error
}

func NewFirewallClient() (FirewallClient, error) {
	// if _, err := os.Stat("/usr/sbin/firewalld"); err == nil {
	return client.NewFirewalld()
	// }
	// if _, err := os.Stat("/usr/sbin/ufw"); err == nil {
	// 	return client.NewUfw()
	// }
	// return nil, errors.New("no such type")
}
