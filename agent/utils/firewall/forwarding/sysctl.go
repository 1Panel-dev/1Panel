package forwarding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

func ensureForwardingSysctls(system forwardingSystem, withIPv6 bool) error {
	if withIPv6 {
		interfaces, err := IPv6RAInterfaces(system.ReadFile)
		if err != nil {
			return fmt.Errorf("check IPv6 Router Advertisement: %w", err)
		}
		if len(interfaces) > 0 {
			return fmt.Errorf("IPv6 forwarding blocked: interfaces %s may depend on RA/SLAAC with accept_ra=1; persist accept_ra=2 on interfaces that require RA before retrying", strings.Join(interfaces, ", "))
		}
	}
	paths := []string{"/proc/sys/net/ipv4/ip_forward"}
	if withIPv6 {
		paths = append(paths, "/proc/sys/net/ipv6/conf/all/forwarding")
	}
	for _, path := range paths {
		if err := system.WriteFile(path, []byte("1"), constant.FilePerm); err != nil {
			return fmt.Errorf("failed to enable IP forwarding at %s: %w", path, err)
		}
	}
	data, err := system.ReadFile("/etc/sysctl.conf")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read /etc/sysctl.conf: %w", err)
	}
	content := enableForwardingSysctls(string(data), withIPv6)
	if err := system.WriteFile("/etc/sysctl.conf", []byte(content), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to persist IP forwarding: %w", err)
	}
	return nil
}

func IPv6RAInterfaces(readFile func(string) ([]byte, error)) ([]string, error) {
	addresses, err := readFile("/proc/net/if_inet6")
	if err != nil {
		return nil, err
	}
	const permanentAddress = 0x80
	globalAddresses := make(map[string]bool)
	raInterfaces := make(map[string]bool)
	for _, line := range strings.Split(string(addresses), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 6 {
			return nil, fmt.Errorf("invalid IPv6 address entry: %q", line)
		}
		name := fields[5]
		if name == "lo" {
			continue
		}
		flags, err := strconv.ParseUint(fields[4], 16, 32)
		if err != nil {
			return nil, err
		}
		if _, exists := globalAddresses[name]; !exists {
			globalAddresses[name] = false
		}
		if fields[3] == "00" {
			globalAddresses[name] = true
			if flags&permanentAddress == 0 {
				raInterfaces[name] = true
			}
		}
	}
	routes, err := readFile("/proc/net/ipv6_route")
	if err != nil {
		return nil, err
	}
	const raRouteFlags = 0x00010000 | 0x00800000
	for _, line := range strings.Split(string(routes), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 10 {
			return nil, fmt.Errorf("invalid IPv6 route entry: %q", line)
		}
		flags, err := strconv.ParseUint(fields[8], 16, 32)
		if err != nil {
			return nil, err
		}
		if flags&raRouteFlags != 0 {
			raInterfaces[fields[9]] = true
		}
	}
	for name, hasGlobalAddress := range globalAddresses {
		if !hasGlobalAddress {
			raInterfaces[name] = true
		}
	}
	var interfaces []string
	for name := range raInterfaces {
		if name == "lo" {
			continue
		}
		if name == "." || name == ".." || filepath.Base(name) != name {
			return nil, fmt.Errorf("invalid IPv6 interface %q", name)
		}
		value, err := readFile("/proc/sys/net/ipv6/conf/" + name + "/accept_ra")
		if err != nil {
			return nil, err
		}
		switch strings.TrimSpace(string(value)) {
		case "1":
			interfaces = append(interfaces, name)
		case "0", "2":
		default:
			return nil, fmt.Errorf("invalid accept_ra value for %s", name)
		}
	}
	sort.Strings(interfaces)
	return interfaces, nil
}

func enableForwardingSysctls(content string, withIPv6 bool) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	wanted := map[string]string{"net.ipv4.ip_forward": "net.ipv4.ip_forward = 1"}
	if withIPv6 {
		wanted["net.ipv6.conf.all.forwarding"] = "net.ipv6.conf.all.forwarding = 1"
	}
	found := make(map[string]bool, len(wanted))
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if replacement, ok := wanted[key]; ok {
			lines[index] = replacement
			found[key] = true
		}
	}
	for _, key := range []string{"net.ipv4.ip_forward", "net.ipv6.conf.all.forwarding"} {
		if replacement, ok := wanted[key]; ok && !found[key] {
			lines = append(lines, replacement)
		}
	}
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n") + "\n"
}
