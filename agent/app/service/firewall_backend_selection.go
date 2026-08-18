package service

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

const (
	settingSystemFirewallBackend = "FirewallProvider"
	settingForwardingBackend     = "ForwardingBackend"
	settingDockerFirewallBackend = "DockerFirewallBackend"
	settingDockerPortGuardStatus = "DockerPortGuardStatus"
	// These persisted keys predate native nftables support. Keep their values
	// stable while using backend-neutral names inside the service.
	settingFilterInitialized     = "IptablesStatus"
	settingForwardingInitialized = "IptablesForwardStatus"
	settingPingStatus            = "BanPing"
)

func selectedDockerFirewallBackend(fallback string) string {
	selected := ""
	if global.DB != nil {
		selected, _ = settingRepo.GetValueByKey(settingDockerFirewallBackend)
	}
	selected = strings.ToLower(strings.TrimSpace(selected))
	if selected == "iptables" || selected == "nftables" {
		return selected
	}
	fallback = strings.ToLower(strings.TrimSpace(fallback))
	if fallback == "nftables" {
		return fallback
	}
	return "iptables"
}

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
