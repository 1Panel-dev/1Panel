package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

type panelPortWhitelistKey struct{}

func (s *FirewallService) UpdatePanelPort(ctx context.Context, oldPort, port uint) error {
	if oldPort == 0 || oldPort > 65535 || port == 0 || port > 65535 {
		return fmt.Errorf("invalid panel port transition %d -> %d", oldPort, port)
	}
	if LoadPanelPort() != strconv.Itoa(int(oldPort)) {
		return fmt.Errorf("panel port changed before firewall update")
	}
	if oldPort == port {
		return nil
	}
	client, err := s.baseClient()
	if err != nil {
		if configuredSystemFirewallBackend() == "" && len(lifecycle.InstalledProviders()) == 0 {
			return nil
		}
		return err
	}
	provider := client.Name()
	active, err := client.Status()
	if err != nil {
		return err
	}
	if !supportsManagedFilterChains(provider) && !active {
		return nil
	}
	required, err := loadRequiredFirewallPorts(strconv.Itoa(int(port)))
	if err != nil {
		return err
	}
	if supportsManagedFilterChains(provider) {
		firewallRuleMutationMu.Lock()
		defer firewallRuleMutationMu.Unlock()
		prepared := append([]firewall.PortWhitelist{{Port: strconv.Itoa(int(oldPort)), Protocol: "tcp"}}, required...)
		if err := syncPanelRequiredPorts(provider, prepared); err != nil {
			return err
		}
		warnPanelPortCleanupFailure(oldPort, syncPanelRequiredPorts(provider, required))
		return nil
	}
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	protected := firewall.NormalizePortWhitelist(append(configured, required...))
	ports := systemPorts([]firewall.PortWhitelist{{Port: strconv.Itoa(int(port)), Protocol: "tcp"}})
	for _, port := range ports {
		if err := s.ensureSystemPort(ctx, port); err != nil {
			return err
		}
	}
	ctx = context.WithValue(ctx, panelPortWhitelistKey{}, protected)
	for index := range ports {
		ports[index].Port = strconv.Itoa(int(oldPort))
	}
	if provider == constant.FirewallProviderUFW {
		for _, port := range ports {
			port.Protocol = "all"
			ports = append(ports, port)
		}
	}
	if provider == constant.FirewallProviderFirewalld {
		ports = append(ports, dto.FirewallSystemPort{Port: strconv.Itoa(int(oldPort)), Protocol: "tcp"})
	}
	for _, port := range ports {
		if panelPortStillRequired(port, protected) {
			continue
		}
		if err := s.deleteSystemPort(ctx, port); err != nil {
			warnPanelPortCleanupFailure(oldPort, err)
		}
	}
	return nil
}

func warnPanelPortCleanupFailure(port uint, err error) {
	if err != nil && global.LOG != nil {
		global.LOG.Warnf("clean up old panel firewall port %d failed: %v", port, err)
	}
}

func panelPortStillRequired(port dto.FirewallSystemPort, protected []firewall.PortWhitelist) bool {
	rule := systemPortRule(filter.ProviderUFW, port)
	for _, required := range protected {
		if port.Family != "" && required.Family != "" && port.Family != required.Family {
			continue
		}
		other := rule
		other.Protocol, other.DestinationPort = required.Protocol, required.Port
		if filter.RulesOverlap(rule, other) {
			return true
		}
	}
	return false
}

func panelRuleStillRequired(rule filter.FirewallRule, protected []firewall.PortWhitelist) bool {
	family := string(rule.Scope.Family)
	if rule.Scope.Family == filter.FamilyInet {
		family = ""
	}
	return panelPortStillRequired(dto.FirewallSystemPort{
		Family: family, Port: rule.DestinationPort, Protocol: rule.Protocol,
	}, protected)
}
