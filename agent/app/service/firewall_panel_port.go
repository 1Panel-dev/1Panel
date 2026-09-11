package service

import (
	"context"
	"errors"
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
	managedChains := supportsManagedFilterChains(provider)
	if managedChains {
		firewallRuleMutationMu.Lock()
		defer firewallRuleMutationMu.Unlock()
	}
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	protected := firewall.NormalizePortWhitelist(append(configured, required...))
	if managedChains {
		prepared := append([]firewall.PortWhitelist{{Port: strconv.Itoa(int(oldPort)), Protocol: "tcp"}}, required...)
		if err := syncPanelRequiredPorts(provider, prepared); err != nil {
			return err
		}
		if err := syncPanelRequiredPorts(provider, required); err != nil {
			warnPanelPortCleanupFailure(oldPort, err)
			return nil
		}
		warnPanelPortCleanupFailure(oldPort, s.cleanupPanelPortLocked(ctx, provider, oldPort, protected))
		return nil
	}
	ports := systemPorts([]firewall.PortWhitelist{{Port: strconv.Itoa(int(port)), Protocol: "tcp"}})
	for _, port := range ports {
		if err := s.ensureSystemPort(ctx, port); err != nil {
			return err
		}
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	warnPanelPortCleanupFailure(oldPort, s.cleanupPanelPortLocked(ctx, provider, oldPort, protected))
	return nil
}

// cleanupPanelPortLocked removes the old system-owned policy as well as any
// remaining managed runtime rule. The caller must hold firewallRuleMutationMu.
func (s *FirewallService) cleanupPanelPortLocked(ctx context.Context, provider string, oldPort uint, protected []firewall.PortWhitelist) error {
	ctx = context.WithValue(ctx, panelPortWhitelistKey{}, protected)
	ports := systemPorts([]firewall.PortWhitelist{{Port: strconv.Itoa(int(oldPort)), Protocol: "tcp"}})
	if provider == constant.FirewallProviderUFW {
		for _, port := range ports {
			port.Protocol = "all"
			ports = append(ports, port)
		}
	}
	// Include family-neutral records left by firewalld or older versions, even
	// when the selected backend has since changed.
	ports = append(ports, dto.FirewallSystemPort{Port: strconv.Itoa(int(oldPort)), Protocol: "tcp"})
	var cleanupErrors []error
	for _, port := range ports {
		if panelPortStillRequired(port, protected) {
			continue
		}
		records, err := s.systemPortRecords(ctx, port)
		if err != nil {
			cleanupErrors = append(cleanupErrors, err)
			continue
		}
		for _, record := range records {
			if err := s.deleteRule(ctx, record.UUID, true); err != nil && !errors.Is(err, filter.ErrProtectedRule) {
				cleanupErrors = append(cleanupErrors, err)
			}
		}
	}
	return errors.Join(cleanupErrors...)
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
