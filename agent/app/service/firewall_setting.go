package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

type firewallPortWhitelist struct {
	Port     string
	Protocol string
}

func loadFirewallPortWhiteList() ([]firewallPortWhitelist, error) {
	value, err := settingRepo.GetValueByKey(constant.FirewallPortWhiteList)
	if err != nil {
		value = constant.FirewallPortWhiteListValue
		if err := settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, value); err != nil {
			return nil, err
		}
	}
	return parseFullFirewallPortWhiteList(value)
}

func parseFullFirewallPortWhiteList(value string) ([]firewallPortWhitelist, error) {
	portWhiteList, err := parseFirewallPortWhiteList(value)
	if err != nil {
		return nil, err
	}
	panelPort := LoadPanelPort()
	if panelPort == "" {
		return nil, fmt.Errorf("find 1panel service port failed")
	}
	portWhiteList = append(portWhiteList, firewallPortWhitelist{Port: panelPort, Protocol: "tcp"})
	portWhiteList = append(portWhiteList, firewallPortWhitelist{Port: loadSSHPort(), Protocol: "tcp"})
	return normalizeFirewallPortWhiteList(portWhiteList), nil
}

func parseFirewallPortWhiteList(value string) ([]firewallPortWhitelist, error) {
	items := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';' || r == ' '
	})
	ports := make([]firewallPortWhitelist, 0, len(items))
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
		ports = append(ports, firewallPortWhitelist{Port: strconv.Itoa(portNum), Protocol: protocol})
	}
	return ports, nil
}

func normalizeFirewallPortWhiteList(portWhiteList []firewallPortWhitelist) []firewallPortWhitelist {
	ports := make([]firewallPortWhitelist, 0, len(portWhiteList))
	exists := make(map[string]struct{})
	for _, item := range portWhiteList {
		if item.Port == "" {
			continue
		}
		key := fmt.Sprintf("%s/%s", item.Port, item.Protocol)
		if _, ok := exists[key]; ok {
			continue
		}
		exists[key] = struct{}{}
		ports = append(ports, item)
	}
	return ports
}

func syncFirewallPortWhiteListAfterUpdate(oldValue string) error {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return err
	}
	portWhiteList, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	if client.Name() != "iptables" {
		oldPortWhiteList, err := parseFullFirewallPortWhiteList(oldValue)
		if err != nil {
			return err
		}
		return syncFirewallClientPortWhiteList(client, oldPortWhiteList, portWhiteList)
	}
	isInit, _ := iptables.LoadInitStatus("iptables", "base")
	if !isInit {
		return nil
	}
	return applyFirewallPortWhiteListRules(portWhiteList, true)
}

func syncFirewallClientPortWhiteList(client firewall.FirewallClient, oldPortWhiteList, portWhiteList []firewallPortWhitelist) error {
	oldPorts := firewallPortWhiteListMap(oldPortWhiteList)
	newPorts := firewallPortWhiteListMap(portWhiteList)
	for _, item := range oldPortWhiteList {
		key := firewallPortWhiteListKey(item)
		if _, ok := newPorts[key]; ok {
			continue
		}
		if err := client.Port(fireClient.FireInfo{Port: item.Port, Protocol: item.Protocol, Strategy: "accept"}, "remove"); err != nil {
			return err
		}
	}
	for _, item := range portWhiteList {
		key := firewallPortWhiteListKey(item)
		if _, ok := oldPorts[key]; ok {
			continue
		}
		if err := client.Port(fireClient.FireInfo{Port: item.Port, Protocol: item.Protocol, Strategy: "accept"}, "add"); err != nil {
			return err
		}
	}
	return client.Reload()
}

func firewallPortWhiteListMap(portWhiteList []firewallPortWhitelist) map[string]struct{} {
	ports := make(map[string]struct{})
	for _, item := range portWhiteList {
		ports[firewallPortWhiteListKey(item)] = struct{}{}
	}
	return ports
}

func firewallPortWhiteListKey(item firewallPortWhitelist) string {
	return item.Port + "/" + item.Protocol
}
