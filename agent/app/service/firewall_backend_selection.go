package service

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

const (
	settingSystemFirewallBackend = "FirewallProvider"
	settingForwardingBackend     = "ForwardingBackend"
	settingDockerFirewallBackend = "DockerFirewallBackend"
)

func selectedSystemFirewallClient() (lifecycle.Client, error) {
	if provider, _ := settingRepo.GetValueByKey(settingSystemFirewallBackend); strings.TrimSpace(provider) != "" {
		return lifecycle.NewClientFor(strings.TrimSpace(provider))
	}
	client, err := lifecycle.NewClient()
	if err != nil {
		return nil, err
	}
	_ = settingRepo.UpdateOrCreate(settingSystemFirewallBackend, client.Name())
	return client, nil
}

func NewSelectedSystemFirewallClient() (lifecycle.Client, error) {
	return selectedSystemFirewallClient()
}

func selectedSystemFirewallProvider() (string, error) {
	client, err := selectedSystemFirewallClient()
	if err != nil {
		return "", err
	}
	return client.Name(), nil
}
