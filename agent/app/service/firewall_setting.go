package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
)

func loadConfiguredFirewallPortWhiteList() ([]fireClient.PortWhiteListEntry, error) {
	value, err := settingRepo.GetValueByKey(constant.FirewallPortWhiteList)
	if err != nil {
		value = constant.FirewallPortWhiteListValue
		if err := settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, value); err != nil {
			return nil, err
		}
	}
	return parseFirewallPortWhiteList(value)
}

func loadRequiredFirewallPortWhiteList() ([]fireClient.PortWhiteListEntry, error) {
	panelPort := LoadPanelPort()
	if panelPort == "" {
		return nil, fmt.Errorf("find 1panel service port failed")
	}
	return []fireClient.PortWhiteListEntry{
		{Port: panelPort, Protocol: "tcp"},
		{Port: loadSSHPort(), Protocol: "tcp"},
	}, nil
}

// loadFirewallPortWhiteList resolves the whitelist state for a provider. oldValue
// is the raw setting value replaced by this change, empty when nothing is replaced.
func loadFirewallPortWhiteList(oldValue string) (fireClient.PortWhiteList, error) {
	var list fireClient.PortWhiteList
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return list, err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return list, err
	}
	previous, err := parseFirewallPortWhiteList(oldValue)
	if err != nil {
		return list, err
	}
	list.Configured, list.Required, list.Previous = configured, required, previous
	return list, nil
}

func parseFirewallPortWhiteList(value string) ([]fireClient.PortWhiteListEntry, error) {
	items := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';' || r == ' '
	})
	ports := make([]fireClient.PortWhiteListEntry, 0, len(items))
	exists := make(map[string]struct{})
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		port, protocol, ok := strings.Cut(item, "/")
		if !ok {
			protocol = "tcp"
		}
		port = strings.TrimSpace(port)
		protocol = strings.ToLower(strings.TrimSpace(protocol))
		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("invalid firewall port whitelist protocol: %s", item)
		}
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return nil, fmt.Errorf("invalid firewall port whitelist: %s", item)
		}
		key := fmt.Sprintf("%d/%s", portNum, protocol)
		if _, ok := exists[key]; ok {
			continue
		}
		exists[key] = struct{}{}
		ports = append(ports, fireClient.PortWhiteListEntry{Port: strconv.Itoa(portNum), Protocol: protocol})
	}
	return ports, nil
}

func syncFirewallPortWhiteListAfterUpdate(oldValue string) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	return syncFirewallPortWhiteListAfterUpdateWithClient(client, oldValue)
}

func syncFirewallPortWhiteListAfterUpdateWithClient(client firewall.FilterClient, oldValue string) error {
	list, err := loadFirewallPortWhiteList(oldValue)
	if err != nil {
		return err
	}
	return client.SyncPortWhiteList(list)
}
