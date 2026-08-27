package service

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func selectedDockerFirewallBackend(fallback string) string {
	selected := configuredDockerFirewallBackend()
	if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
		return selected
	}
	fallback = strings.ToLower(strings.TrimSpace(fallback))
	if fallback == constant.FirewallProviderNftables {
		return fallback
	}
	return constant.FirewallProviderIptables
}

func configuredDockerFirewallBackend() string {
	if global.DB == nil {
		return ""
	}
	selected, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	selected = strings.ToLower(strings.TrimSpace(selected))
	if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
		return selected
	}
	return ""
}

func selectedSystemFirewallClient() (lifecycle.Client, error) {
	if provider := configuredSystemFirewallBackend(); provider != "" {
		return lifecycle.NewClientFor(provider)
	}
	client, err := lifecycle.NewClient()
	if err != nil {
		return nil, err
	}
	_ = settingRepo.UpdateOrCreate(constant.FirewallSystemBackendKey, client.Name())
	return client, nil
}

func configuredSystemFirewallBackend() string {
	if global.DB == nil {
		return ""
	}
	provider, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	return strings.TrimSpace(provider)
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
