package forwarding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/network"
)

func ensureFamilyForwardingSysctls(system forwardingSystem, family string) error {
	withIPv6 := family == FamilyIPv6
	if withIPv6 {
		interfaces, err := IPv6RAInterfaces(system.ReadFile, LoadDockerIPv4BridgePorts)
		if err != nil {
			global.LOG.Warnf("IPv6 forwarding blocked: failed to check Router Advertisement: %v", err)
			return buserr.New("ErrIPv6RACheckFailed")
		}
		if len(interfaces) > 0 {
			global.LOG.Warnf("IPv6 forwarding blocked: interfaces %s may depend on RA/SLAAC with accept_ra=1; review the network configuration and persist accept_ra=2 on interfaces that require RA before retrying", strings.Join(interfaces, ", "))
			return buserr.New("ErrIPv6RARisk")
		}
	}
	paths := []string{"/proc/sys/net/ipv4/ip_forward"}
	if withIPv6 {
		paths = []string{"/proc/sys/net/ipv6/conf/all/forwarding"}
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
	content := enableFamilyForwardingSysctls(string(data), family)
	if err := system.WriteFile("/etc/sysctl.conf", []byte(content), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to persist IP forwarding: %w", err)
	}
	return nil
}

func IPv6RAInterfaces(readFile func(string) ([]byte, error), loadDockerPorts func() (map[string]bool, error)) ([]string, error) {
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
		globalAddresses[name] = globalAddresses[name] || fields[3] == "00"
		if fields[3] == "00" && flags&permanentAddress == 0 {
			raInterfaces[name] = true
		}
	}
	routes, err := readFile("/proc/net/ipv6_route")
	if err != nil {
		return nil, err
	}
	const raRouteFlags = 0x00010000 | 0x00080000 | 0x00800000
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
	uncertainInterfaces := make(map[string]bool)
	for name, hasGlobalAddress := range globalAddresses {
		if !hasGlobalAddress && !raInterfaces[name] {
			raInterfaces[name] = true
			uncertainInterfaces[name] = true
		}
	}
	var interfaces []string
	checkDockerPorts := false
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
			checkDockerPorts = checkDockerPorts || uncertainInterfaces[name]
		case "0", "2":
		default:
			return nil, fmt.Errorf("invalid accept_ra value for %s", name)
		}
	}
	if checkDockerPorts {
		ports, err := loadDockerPorts()
		if err != nil {
			global.LOG.Warnf("IPv6 forwarding RA check: cannot exclude Docker bridge ports: %v", err)
		} else {
			kept := interfaces[:0]
			for _, name := range interfaces {
				if !uncertainInterfaces[name] || !ports[name] {
					kept = append(kept, name)
				}
			}
			interfaces = kept
		}
	}
	sort.Strings(interfaces)
	return interfaces, nil
}

func LoadDockerIPv4BridgePorts() (map[string]bool, error) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	if !strings.HasPrefix(cli.DaemonHost(), "unix://") {
		return nil, errors.New("cannot verify local bridge ports using a remote Docker daemon")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	info, err := cli.Info(ctx)
	if err != nil {
		return nil, err
	}
	for _, option := range info.SecurityOptions {
		if option == "name=rootless" {
			return nil, errors.New("cannot verify host bridge ports using a rootless Docker daemon")
		}
	}
	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}
	output, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second)).RunWithOptionalSudoAndStdout("ip", "-details", "-json", "link", "show")
	if err != nil {
		return nil, err
	}
	return dockerIPv4BridgePorts(networks, output)
}

func dockerIPv4BridgePorts(networks []network.Summary, linkJSON string) (map[string]bool, error) {
	bridges := make(map[string]bool)
	for _, item := range networks {
		if item.Driver != "bridge" || item.EnableIPv6 || item.ConfigOnly {
			continue
		}
		name := item.Options["com.docker.network.bridge.name"]
		if name == "" {
			if item.Options["com.docker.network.bridge.default_bridge"] == "true" {
				name = "docker0"
			} else if len(item.ID) >= 12 {
				name = "br-" + item.ID[:12]
			}
		}
		if name != "" {
			bridges[name] = true
		}
	}
	var links []struct {
		Name     string `json:"ifname"`
		Master   string `json:"master"`
		LinkInfo struct {
			Kind string `json:"info_kind"`
		} `json:"linkinfo"`
	}
	if err := json.Unmarshal([]byte(linkJSON), &links); err != nil {
		return nil, err
	}
	confirmedBridges := make(map[string]bool)
	for _, link := range links {
		if link.LinkInfo.Kind == "bridge" && bridges[link.Name] {
			confirmedBridges[link.Name] = true
		}
	}
	ports := make(map[string]bool)
	for _, link := range links {
		if link.LinkInfo.Kind == "veth" && confirmedBridges[link.Master] {
			ports[link.Name] = true
		}
	}
	return ports, nil
}

func enableFamilyForwardingSysctls(content, family string) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	wanted := map[string]string{"net.ipv4.ip_forward": "net.ipv4.ip_forward = 1"}
	if family == FamilyIPv6 {
		wanted = map[string]string{"net.ipv6.conf.all.forwarding": "net.ipv6.conf.all.forwarding = 1"}
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
