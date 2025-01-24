package app

import (
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
)

func Init() {
	go func() {
		_ = docker.CreateDefaultDockerNetwork()

		if f, err := firewall.NewFirewallClient(); err == nil {
			_ = f.EnableForward()
		}
	}()
}
