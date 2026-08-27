package firewall

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

type PortWhitelist = filter.PortWhitelist

func ParsePortWhitelist(value string) ([]PortWhitelist, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []PortWhitelist{}, nil
	}
	if strings.HasPrefix(value, "[") {
		var rules []PortWhitelist
		if err := json.Unmarshal([]byte(value), &rules); err != nil {
			return nil, fmt.Errorf("invalid firewall port whitelist JSON: %w", err)
		}
		return validatePortWhitelist(rules)
	}

	items := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';' || r == ' '
	})
	rules := make([]PortWhitelist, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.Split(item, "/")
		rule := PortWhitelist{Family: constant.FirewallFamilyIPv4, Protocol: "tcp"}
		switch len(parts) {
		case 1:
			rule.Port = parts[0]
		case 2:
			rule.Port, rule.Protocol = parts[0], parts[1]
		case 3:
			rule.Family, rule.Port, rule.Protocol = parts[0], parts[1], parts[2]
		default:
			return nil, fmt.Errorf("invalid firewall port whitelist: %s", item)
		}
		rules = append(rules, rule)
	}
	return validatePortWhitelist(rules)
}

func validatePortWhitelist(rules []PortWhitelist) ([]PortWhitelist, error) {
	result := make([]PortWhitelist, 0, len(rules))
	exists := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		rule.Family = strings.ToLower(strings.TrimSpace(rule.Family))
		if rule.Family == "" {
			rule.Family = constant.FirewallFamilyIPv4
		}
		if rule.Family != constant.FirewallFamilyIPv4 && rule.Family != constant.FirewallFamilyIPv6 {
			return nil, fmt.Errorf("invalid firewall port whitelist family: %s", rule.Family)
		}
		rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
		if rule.Protocol == "" {
			rule.Protocol = "tcp"
		}
		if rule.Protocol != "tcp" && rule.Protocol != "udp" {
			return nil, fmt.Errorf("invalid firewall port whitelist protocol: %s", rule.Protocol)
		}
		port, err := normalizeWhitelistPort(rule.Port)
		if err != nil {
			return nil, err
		}
		rule.Port = port
		key := PortWhitelistKey(rule)
		if _, ok := exists[key]; ok {
			continue
		}
		for _, current := range result {
			if current.Family == rule.Family && current.Protocol == rule.Protocol && whitelistPortsOverlap(current.Port, rule.Port) {
				return nil, fmt.Errorf("overlapping firewall port whitelist rules: %s and %s", current.Port, rule.Port)
			}
		}
		exists[key] = struct{}{}
		result = append(result, rule)
	}
	return result, nil
}

func normalizeWhitelistPort(value string) (string, error) {
	value = strings.TrimSpace(value)
	separator := ""
	if strings.Contains(value, "-") {
		separator = "-"
	} else if strings.Contains(value, ":") {
		separator = ":"
	}
	if separator == "" {
		port, err := parseWhitelistPort(value)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(port), nil
	}
	parts := strings.Split(value, separator)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid firewall port whitelist range: %s", value)
	}
	start, err := parseWhitelistPort(parts[0])
	if err != nil {
		return "", err
	}
	end, err := parseWhitelistPort(parts[1])
	if err != nil || start > end {
		return "", fmt.Errorf("invalid firewall port whitelist range: %s", value)
	}
	if start == end {
		return strconv.Itoa(start), nil
	}
	return fmt.Sprintf("%d-%d", start, end), nil
}

func parseWhitelistPort(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid firewall port whitelist: %s", value)
	}
	return port, nil
}

func whitelistPortsOverlap(left, right string) bool {
	parseRange := func(value string) (int, int) {
		parts := strings.Split(value, "-")
		start, _ := strconv.Atoi(parts[0])
		if len(parts) == 1 {
			return start, start
		}
		end, _ := strconv.Atoi(parts[1])
		return start, end
	}
	leftStart, leftEnd := parseRange(left)
	rightStart, rightEnd := parseRange(right)
	return leftStart <= rightEnd && rightStart <= leftEnd
}

func NormalizePortWhitelist(items []PortWhitelist) []PortWhitelist {
	ports := make([]PortWhitelist, 0, len(items))
	for _, item := range items {
		if item.Port == "" {
			continue
		}
		baseKey := item.Port + "/" + strings.ToLower(strings.TrimSpace(item.Protocol))
		duplicate := false
		for _, current := range ports {
			currentBaseKey := current.Port + "/" + strings.ToLower(strings.TrimSpace(current.Protocol))
			if currentBaseKey == baseKey && (current.Family == "" || current.Family == item.Family) {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		if item.Family == "" {
			filtered := ports[:0]
			for _, current := range ports {
				currentBaseKey := current.Port + "/" + strings.ToLower(strings.TrimSpace(current.Protocol))
				if currentBaseKey != baseKey {
					filtered = append(filtered, current)
				}
			}
			ports = filtered
		}
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
	key := item.Port + "/" + strings.ToLower(strings.TrimSpace(item.Protocol))
	if family := strings.ToLower(strings.TrimSpace(item.Family)); family != "" {
		return family + "/" + key
	}
	return key
}
