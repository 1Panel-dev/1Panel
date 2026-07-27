package service

import (
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

const fail2BanRestoreWithFirewallMarker = "/run/1panel_fail2ban_restore_with_firewall"

type firewallFail2BanState struct {
	markerPath string
	isExist    func(string) bool
	isActive   func(string) bool
	start      func(string) error
}

func newFirewallFail2BanState() *firewallFail2BanState {
	return &firewallFail2BanState{
		markerPath: fail2BanRestoreWithFirewallMarker,
		isExist: func(serviceName string) bool {
			exists, err := controller.CheckExist(serviceName)
			if err != nil {
				global.LOG.Warnf("check %s installation before stopping the firewall failed: %v", serviceName, err)
			}
			return exists
		},
		isActive: func(serviceName string) bool {
			active, err := controller.CheckActive(serviceName)
			if err != nil {
				global.LOG.Warnf("check %s status before stopping the firewall failed: %v", serviceName, err)
			}
			return active
		},
		start: controller.HandleStart,
	}
}

func (s *firewallFail2BanState) rememberBeforeFirewallStop() error {
	if !s.isExist("fail2ban.service") {
		return nil
	}
	if !s.isActive("fail2ban.service") {
		return nil
	}
	return s.markForRestore()
}

func (s *firewallFail2BanState) markForRestore() error {
	if err := os.WriteFile(s.markerPath, nil, 0600); err != nil {
		return fmt.Errorf("mark Fail2Ban for restoration with the firewall: %w", err)
	}
	return nil
}

func (s *firewallFail2BanState) restoreAfterFirewallStart() error {
	_, err := os.Stat(s.markerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load Fail2Ban restore marker after starting the firewall: %w", err)
	}

	if err := s.start("fail2ban.service"); err != nil {
		return fmt.Errorf("restore Fail2Ban after starting the firewall: %w", err)
	}
	return s.clearRestoreMarker()
}

func (s *firewallFail2BanState) clearRestoreMarker() error {
	if err := os.Remove(s.markerPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear Fail2Ban firewall restore status: %w", err)
	}
	return nil
}
