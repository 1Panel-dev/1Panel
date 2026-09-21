package docker_guard

import (
	"net/netip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func ReadDNATRules(backend, family string) DNATRules {
	manager := cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second), cmd.WithEnv("LC_ALL=C"))
	if backend == constant.FirewallProviderNftables {
		tableFamily := "ip"
		if family == constant.FirewallFamilyIPv6 {
			tableFamily = "ip6"
		}
		tables, err := manager.RunWithOptionalSudoAndStdout("nft", "list", "tables")
		if err != nil {
			return DNATRules{}
		}
		if !strings.Contains(tables, "table "+tableFamily+" docker-bridges") {
			return DNATRules{Inspected: true}
		}
		output, err := manager.RunWithOptionalSudoAndStdout("nft", "list", "table", tableFamily, "docker-bridges")
		return DNATRules{Output: output, Inspected: err == nil}
	}
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return DNATRules{}
	}
	executable := commands.IPv4
	if family == constant.FirewallFamilyIPv6 {
		executable = commands.IPv6
	}
	if executable == "" {
		return DNATRules{}
	}
	output, err := manager.RunWithOptionalSudoAndStdout(executable, "-w", "-t", "nat", "-S")
	return DNATRules{Output: output, Inspected: err == nil}
}

func ReadProxyEndpoints() ProxyEndpoints {
	manager := cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second), cmd.WithEnv("LC_ALL=C"))
	output, err := manager.RunWithStdout("ps", "-ww", "-eo", "args=")
	if err != nil {
		return ProxyEndpoints{}
	}
	return ProxyEndpoints{Items: parseDockerProxyEndpoints(output), Inspected: true}
}

func parseDockerProxyEndpoints(output string) []ProxyEndpoint {
	result := make([]ProxyEndpoint, 0)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		isProxy := slices.ContainsFunc(fields, func(field string) bool { return filepath.Base(field) == "docker-proxy" })
		if !isProxy {
			continue
		}
		protocol := commandFlagValue(fields, "-proto")
		hostIP := commandFlagValue(fields, "-host-ip")
		hostPortValue := commandFlagValue(fields, "-host-port")
		hostPort, err := strconv.ParseUint(hostPortValue, 10, 16)
		if err != nil || (protocol != "tcp" && protocol != "udp") || hostIP == "" {
			continue
		}
		hostIP = strings.TrimSpace(hostIP)
		if address, err := netip.ParseAddr(hostIP); err == nil {
			hostIP = address.String()
		}
		result = append(result, ProxyEndpoint{Protocol: protocol, HostIP: hostIP, HostPort: uint16(hostPort)})
	}
	return result
}

func commandFlagValue(fields []string, name string) string {
	for i := 0; i < len(fields); i++ {
		if fields[i] == name && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(fields[i], name+"=") {
			return strings.TrimPrefix(fields[i], name+"=")
		}
	}
	return ""
}

func ProxyEndpointMatches(proxies []ProxyEndpoint, family, hostIP string, hostPort uint16, protocol string) bool {
	for _, proxy := range proxies {
		if proxy.Protocol == protocol && proxy.HostPort == hostPort && hostAddressMatches(proxy.HostIP, hostIP, family) {
			return true
		}
	}
	return false
}

func DNATRuleMatches(backend, output string, family, hostIP string, hostPort uint16, protocol string) bool {
	return InspectEndpoints(backend, family, DNATRules{Output: output}, ProxyEndpoints{}).DNATMatches(hostIP, hostPort, protocol)
}

func DNATIngressReachable(backend, output string) bool {
	if backend == constant.FirewallProviderNftables {
		return strings.Contains(output, "hook prerouting")
	}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "-A" && fields[1] == "PREROUTING" && commandFlagValue(fields, "-j") == "DOCKER" {
			return true
		}
	}
	return false
}

type EndpointInspection struct {
	DNATInspected, ProxyInspected, IngressReachable bool
	family                                          string
	dnat, proxies                                   map[ProxyEndpoint]bool
}

func InspectEndpoints(backend, family string, rules DNATRules, proxies ProxyEndpoints) EndpointInspection {
	inspection := EndpointInspection{
		DNATInspected: rules.Inspected, ProxyInspected: proxies.Inspected,
		IngressReachable: DNATIngressReachable(backend, rules.Output), family: family,
		dnat: make(map[ProxyEndpoint]bool), proxies: make(map[ProxyEndpoint]bool),
	}
	for _, proxy := range proxies.Items {
		if isWildcardHostAddress(proxy.HostIP, family) {
			wildcard := proxy
			wildcard.HostIP = ""
			inspection.proxies[wildcard] = true
		}
		proxy.HostIP = normalizedEndpointAddress(proxy.HostIP)
		inspection.proxies[proxy] = true
	}
	replacer := strings.NewReplacer("{", " ", "}", " ", ",", " ", ";", " ")
	addressToken := "ip"
	if family == constant.FirewallFamilyIPv6 {
		addressToken = "ip6"
	}
	for line := range strings.SplitSeq(rules.Output, "\n") {
		if backend != constant.FirewallProviderNftables {
			fields := strings.Fields(line)
			if commandFlagValue(fields, "-j") != "DNAT" {
				continue
			}
			port, err := strconv.ParseUint(commandFlagValue(fields, "--dport"), 10, 16)
			if err != nil || strconv.FormatUint(port, 10) != commandFlagValue(fields, "--dport") {
				continue
			}
			address, _, _ := strings.Cut(commandFlagValue(fields, "-d"), "/")
			inspection.dnat[ProxyEndpoint{Protocol: commandFlagValue(fields, "-p"), HostIP: normalizedEndpointAddress(address), HostPort: uint16(port)}] = true
			continue
		}
		fields := strings.Fields(replacer.Replace(line))
		if !slices.Contains(fields, "dnat") {
			continue
		}
		destination := ""
		protocols := make([]string, 0, 1)
		for i := 0; i+2 < len(fields); i++ {
			if fields[i] == "meta" && fields[i+1] == "l4proto" {
				protocols = append(protocols, fields[i+2])
			}
			if destination == "" && fields[i] == addressToken && fields[i+1] == "daddr" {
				destination = fields[i+2]
			}
		}
		destination, _, _ = strings.Cut(destination, "/")
		destination = normalizedEndpointAddress(destination)
		for i := 0; i+2 < len(fields); i++ {
			if fields[i+1] != "dport" {
				continue
			}
			port, err := strconv.ParseUint(fields[i+2], 10, 16)
			if err != nil || strconv.FormatUint(port, 10) != fields[i+2] {
				continue
			}
			matches := []string{fields[i]}
			if fields[i] == "th" {
				matches = protocols
			}
			for _, protocol := range matches {
				inspection.dnat[ProxyEndpoint{Protocol: protocol, HostIP: destination, HostPort: uint16(port)}] = true
			}
		}
	}
	return inspection
}

func normalizedEndpointAddress(address string) string {
	address = strings.TrimSpace(address)
	if parsed, err := netip.ParseAddr(address); err == nil {
		return parsed.String()
	}
	return address
}

func (inspection EndpointInspection) DNATMatches(hostIP string, hostPort uint16, protocol string) bool {
	key := ProxyEndpoint{Protocol: protocol, HostPort: hostPort}
	if inspection.dnat[key] {
		return true
	}
	if isWildcardHostAddress(hostIP, inspection.family) {
		return false
	}
	key.HostIP = normalizedEndpointAddress(hostIP)
	return inspection.dnat[key]
}

func (inspection EndpointInspection) ProxyMatches(hostIP string, hostPort uint16, protocol string) bool {
	key := ProxyEndpoint{Protocol: protocol, HostPort: hostPort, HostIP: normalizedEndpointAddress(hostIP)}
	if inspection.proxies[key] {
		return true
	}
	if !isWildcardHostAddress(hostIP, inspection.family) {
		return false
	}
	key.HostIP = ""
	return inspection.proxies[key]
}

func hostAddressMatches(left, right, family string) bool {
	if isWildcardHostAddress(left, family) && isWildcardHostAddress(right, family) {
		return true
	}
	left, right = strings.TrimSpace(left), strings.TrimSpace(right)
	if address, err := netip.ParseAddr(left); err == nil {
		left = address.String()
	}
	if address, err := netip.ParseAddr(right); err == nil {
		right = address.String()
	}
	return left == right
}

func isWildcardHostAddress(value, family string) bool {
	value = strings.TrimSpace(value)
	if family == constant.FirewallFamilyIPv6 {
		return value == "" || value == "::"
	}
	return value == "" || value == "0.0.0.0"
}
