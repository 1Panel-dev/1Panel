package providers

import (
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

var forwardInterfacePattern = regexp.MustCompile(`^[A-Za-z0-9_.:@-]{1,15}$`)

func NormalizeRule(rule forwarding.Rule) (forwarding.Rule, error) {
	rule.Family = strings.ToLower(strings.TrimSpace(rule.Family))
	if rule.Family == "" {
		rule.Family = forwarding.FamilyIPv4
	}
	if rule.Family != forwarding.FamilyIPv4 && rule.Family != forwarding.FamilyIPv6 {
		return forwarding.Rule{}, fmt.Errorf("unsupported forwarding family %q", rule.Family)
	}
	rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
	if rule.Protocol != "tcp" && rule.Protocol != "udp" {
		return forwarding.Rule{}, fmt.Errorf("unsupported forwarding protocol %q", rule.Protocol)
	}
	var err error
	if rule.Port, err = normalizeForwardPort(rule.Port); err != nil {
		return forwarding.Rule{}, fmt.Errorf("invalid forwarding port: %w", err)
	}
	if rule.TargetPort, err = normalizeForwardPort(rule.TargetPort); err != nil {
		return forwarding.Rule{}, fmt.Errorf("invalid forwarding target port: %w", err)
	}
	rule.TargetIP = strings.TrimSpace(rule.TargetIP)
	if rule.TargetIP == "" || strings.EqualFold(rule.TargetIP, "localhost") {
		if rule.Family == forwarding.FamilyIPv6 {
			rule.TargetIP = "::1"
		} else {
			rule.TargetIP = "127.0.0.1"
		}
	}
	address, err := netip.ParseAddr(rule.TargetIP)
	if err == nil {
		address = address.Unmap()
	}
	if err != nil || (rule.Family == forwarding.FamilyIPv4) != address.Is4() {
		return forwarding.Rule{}, fmt.Errorf("invalid %s forwarding target %q", rule.Family, rule.TargetIP)
	}
	rule.TargetIP = address.String()
	rule.Interface = strings.TrimSpace(rule.Interface)
	if rule.Interface == "all" || rule.Interface == "*" {
		rule.Interface = ""
	}
	if rule.Interface != "" && !forwardInterfacePattern.MatchString(rule.Interface) {
		return forwarding.Rule{}, fmt.Errorf("invalid forwarding interface %q", rule.Interface)
	}
	return rule, nil
}

func normalizeForwardPort(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) < 1 || len(parts) > 2 {
		return "", fmt.Errorf("invalid port range %q", value)
	}
	ports := make([]int, len(parts))
	for index, part := range parts {
		port, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || port < 1 || port > 65535 {
			return "", fmt.Errorf("invalid port %q", part)
		}
		ports[index] = port
	}
	if len(ports) == 2 {
		if ports[0] > ports[1] {
			return "", fmt.Errorf("descending port range %q", value)
		}
		if ports[0] != ports[1] {
			return strconv.Itoa(ports[0]) + "-" + strconv.Itoa(ports[1]), nil
		}
	}
	return strconv.Itoa(ports[0]), nil
}
