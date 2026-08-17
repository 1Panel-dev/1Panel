package firewall

import (
	"fmt"
	"strconv"
	"strings"
)

type PortWhitelist struct {
	Port     string
	Protocol string
}

func ParsePortWhitelist(value string) ([]PortWhitelist, error) {
	items := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';' || r == ' '
	})
	ports := make([]PortWhitelist, 0, len(items))
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
		entry := PortWhitelist{Port: strconv.Itoa(portNum), Protocol: protocol}
		key := PortWhitelistKey(entry)
		if _, ok := exists[key]; ok {
			continue
		}
		exists[key] = struct{}{}
		ports = append(ports, entry)
	}
	return ports, nil
}

func NormalizePortWhitelist(items []PortWhitelist) []PortWhitelist {
	ports := make([]PortWhitelist, 0, len(items))
	exists := make(map[string]struct{})
	for _, item := range items {
		if item.Port == "" {
			continue
		}
		key := PortWhitelistKey(item)
		if _, ok := exists[key]; ok {
			continue
		}
		exists[key] = struct{}{}
		ports = append(ports, item)
	}
	return ports
}

func PortWhitelistMap(items []PortWhitelist) map[string]struct{} {
	ports := make(map[string]struct{}, len(items))
	for _, item := range items {
		ports[PortWhitelistKey(item)] = struct{}{}
	}
	return ports
}

func PortWhitelistKey(item PortWhitelist) string {
	return item.Port + "/" + item.Protocol
}
