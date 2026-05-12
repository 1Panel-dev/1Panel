package service

import (
	"context"
	"fmt"
	stdnet "net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
	"github.com/docker/docker/api/types/container"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type portKey struct {
	port     uint32
	protocol string
}

type dockerPortInfo struct {
	containerName string
	hostIP        string
}

type firewallRuleEntry struct {
	strategy string
	portStr  string
	protocol string
	address  string
}

// GetPortSecurityOverview scans all listening ports and cross-references them with
// Docker container bindings and firewall rules to produce a security status overview.
func (u *FirewallService) GetPortSecurityOverview(req dto.PortSecuritySearch) (*dto.PortSecurityOverview, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		wg             sync.WaitGroup
		connections    []net.ConnectionStat
		containers     []container.Summary
		firewallRules  []fireClient.FireInfo
		firewallActive bool
		dockerExists   bool
		connectionErr  error
		dockerErr      error
		firewallErr    error
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		connections, connectionErr = net.ConnectionsMaxWithContext(ctx, "inet", 32768)
	}()
	go func() {
		defer wg.Done()
		cli, err := docker.NewDockerClient()
		if err != nil {
			dockerErr = err
			return
		}
		defer cli.Close()
		dockerExists = true
		containers, dockerErr = cli.ContainerList(ctx, container.ListOptions{All: false})
	}()
	go func() {
		defer wg.Done()
		fwClient, err := firewall.NewFirewallClient()
		if err != nil {
			firewallErr = err
			return
		}
		firewallActive, _ = fwClient.Status()
		firewallRules, firewallErr = fwClient.ListPort()
	}()
	wg.Wait()

	if connectionErr != nil {
		return nil, fmt.Errorf("failed to get listening ports: %w", connectionErr)
	}

	dockerPortMap := buildDockerPortMap(containers, dockerErr)
	ruleIndex := buildFirewallRuleIndex(firewallRules, firewallErr)
	appPortMap, systemPort := buildAppPortMaps(ctx)

	seenIndex := make(map[portKey]int)
	items := make([]dto.PortSecurityItem, 0)

	for _, conn := range connections {
		if conn.Pid == 0 {
			continue
		}
		isListen := conn.Status == "LISTEN" && conn.Type == syscall.SOCK_STREAM
		isUDP := conn.Type == syscall.SOCK_DGRAM && conn.Raddr.Port == 0
		if !isListen && !isUDP {
			continue
		}

		proto := "tcp"
		if conn.Type == syscall.SOCK_DGRAM {
			proto = "udp"
		}

		bindAddr := conn.Laddr.IP
		if bindAddr == "" {
			bindAddr = "0.0.0.0"
		}

		key := portKey{port: conn.Laddr.Port, protocol: proto}
		if idx, exists := seenIndex[key]; exists {
			if isWildcardAddress(bindAddr) && !isWildcardAddress(items[idx].BindAddress) {
				items[idx].BindAddress = bindAddr
			}
			continue
		}
		seenIndex[key] = len(items)

		item := dto.PortSecurityItem{
			Port:        conn.Laddr.Port,
			Protocol:    proto,
			BindAddress: bindAddr,
			ProcessName: getProcessNameByPID(conn.Pid),
			PID:         conn.Pid,
			SourceType:  "host",
		}

		if dInfo, ok := dockerPortMap[key]; ok {
			item.SourceType = "docker"
			item.ContainerName = dInfo.containerName
			if dInfo.hostIP != "" {
				item.BindAddress = dInfo.hostIP
			}
		}

		// Only tag with App Store appName when the port is actually owned by a Docker
		// container — prevents mis-labelling a host process that happens to listen on
		// a port matching a stopped app's record.
		if item.SourceType == "docker" {
			if appName, ok := appPortMap[conn.Laddr.Port]; ok {
				item.AppName = appName
				item.SourceType = "appStore"
			}
		}
		if systemPort != 0 && conn.Laddr.Port == systemPort {
			item.AppName = "1panel"
		}

		ruleStrategy, hasRule, ruleAddress := matchFirewallRule(ruleIndex, conn.Laddr.Port, proto)
		item.HasRule = hasRule
		item.RuleStrategy = ruleStrategy
		item.RuleAddress = ruleAddress

		items = append(items, item)
	}

	// Add synthetic entries for Docker container ports that have no host-side LISTEN
	// socket — happens when Docker daemon runs with `userland-proxy: false`, where
	// traffic is forwarded purely via iptables DNAT and no docker-proxy process binds
	// the host port. Without this, those ports silently disappear from the overview.
	for key, dInfo := range dockerPortMap {
		if _, exists := seenIndex[key]; exists {
			continue
		}
		bindAddr := dInfo.hostIP
		if bindAddr == "" {
			bindAddr = "0.0.0.0"
		}
		item := dto.PortSecurityItem{
			Port:          key.port,
			Protocol:      key.protocol,
			BindAddress:   bindAddr,
			ContainerName: dInfo.containerName,
			SourceType:    "docker",
		}
		if appName, ok := appPortMap[key.port]; ok {
			item.AppName = appName
			item.SourceType = "appStore"
		}
		ruleStrategy, hasRule, ruleAddress := matchFirewallRule(ruleIndex, key.port, key.protocol)
		item.HasRule = hasRule
		item.RuleStrategy = ruleStrategy
		item.RuleAddress = ruleAddress
		seenIndex[key] = len(items)
		items = append(items, item)
	}

	for i := range items {
		items[i].Status = determineStatus(items[i].BindAddress, items[i].SourceType, firewallActive, items[i].HasRule, items[i].RuleStrategy)
	}

	// Sort by status priority > protocol > port number
	sort.Slice(items, func(i, j int) bool {
		pi, pj := statusSortPriority(items[i].Status), statusSortPriority(items[j].Status)
		if pi != pj {
			return pi < pj
		}
		if items[i].Protocol != items[j].Protocol {
			return items[i].Protocol < items[j].Protocol
		}
		return items[i].Port < items[j].Port
	})

	// Summary is computed before filtering so it reflects overall status
	summary := dto.PortSecuritySummary{Total: len(items)}
	for _, item := range items {
		switch item.Status {
		case "protected", "blocked":
			summary.Protected++
		case "noRule":
			summary.Unprotected++
		case "dockerBypass":
			summary.DockerBypassed++
		case "localOnly":
			summary.LocalOnly++
		case "firewallInactive":
			summary.Unprotected++
		}
	}

	// Apply filters after summary (so summary reflects totals, filters narrow the list)
	if req.Info != "" || req.Status != "" {
		filtered := make([]dto.PortSecurityItem, 0)
		keyword := strings.ToLower(req.Info)
		for _, item := range items {
			if req.Status != "" && item.Status != req.Status {
				continue
			}
			if keyword != "" {
				portStr := strconv.FormatUint(uint64(item.Port), 10)
				if !strings.Contains(portStr, keyword) &&
					!strings.Contains(strings.ToLower(item.ProcessName), keyword) &&
					!strings.Contains(strings.ToLower(item.ContainerName), keyword) &&
					!strings.Contains(strings.ToLower(item.AppName), keyword) {
					continue
				}
			}
			filtered = append(filtered, item)
		}
		items = filtered
	}

	return &dto.PortSecurityOverview{
		Items:       items,
		Summary:     summary,
		FireActive:  firewallActive,
		DockerExist: dockerExists,
	}, nil
}

// buildDockerPortMap creates a lookup map from host port to container info.
func buildDockerPortMap(containers []container.Summary, err error) map[portKey]dockerPortInfo {
	result := make(map[portKey]dockerPortInfo)
	if err != nil || containers == nil {
		return result
	}
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		for _, p := range c.Ports {
			if p.PublicPort == 0 {
				continue
			}
			key := portKey{port: uint32(p.PublicPort), protocol: p.Type}
			result[key] = dockerPortInfo{
				containerName: name,
				hostIP:        p.IP,
			}
		}
	}
	return result
}

// buildFirewallRuleIndex converts firewall rules into a list for port matching.
func buildFirewallRuleIndex(rules []fireClient.FireInfo, err error) []firewallRuleEntry {
	if err != nil || rules == nil {
		return nil
	}
	var entries []firewallRuleEntry
	for _, r := range rules {
		entries = append(entries, firewallRuleEntry{
			strategy: r.Strategy,
			portStr:  r.Port,
			protocol: r.Protocol,
			address:  r.Address,
		})
	}
	return entries
}

// buildAppPortMaps returns two lookup tables: ports owned by App Store installed
// apps (keyed by host port → app key), and the 1Panel system port. The system
// port is returned separately because it's served by a host process (1panel-core),
// not a Docker container, so it must not go through the docker→appStore promotion.
func buildAppPortMaps(ctx context.Context) (map[uint32]string, uint32) {
	appPorts := make(map[uint32]string)
	apps, err := appInstallRepo.ListBy(ctx)
	if err == nil {
		for _, app := range apps {
			if app.HttpPort > 0 {
				appPorts[uint32(app.HttpPort)] = app.App.Key
			}
			if app.HttpsPort > 0 {
				appPorts[uint32(app.HttpsPort)] = app.App.Key
			}
		}
	}
	var systemPort uint32
	if setting, err := settingRepo.Get(settingRepo.WithByKey("ServerPort")); err == nil && setting.Value != "" {
		if port, e := strconv.ParseUint(setting.Value, 10, 32); e == nil {
			systemPort = uint32(port)
		}
	}
	return appPorts, systemPort
}

// matchFirewallRule checks if a port has a matching firewall rule, respecting protocol.
// When multiple rules match (e.g. an accept and a drop on the same port), drop/reject
// takes precedence over accept. This is security-conservative — if any deny rule applies,
// the port is reported as denied. It also keeps the result deterministic regardless of
// the underlying backend's listing order (firewalld lists --list-ports accepts before
// rich-rule drops, iptables lists by line number, ufw by rule index).
func matchFirewallRule(rules []firewallRuleEntry, port uint32, proto string) (string, bool, string) {
	portStr := strconv.FormatUint(uint64(port), 10)
	acceptStrategy := ""
	acceptAddress := ""
	foundAccept := false
	for _, r := range rules {
		if r.protocol != "" && r.protocol != "tcp/udp" && r.protocol != proto {
			continue
		}
		if !portMatchesRule(portStr, port, r.portStr) {
			continue
		}
		if r.strategy == "drop" || r.strategy == "reject" {
			return r.strategy, true, r.address
		}
		if !foundAccept {
			acceptStrategy = r.strategy
			acceptAddress = r.address
			foundAccept = true
		}
	}
	return acceptStrategy, foundAccept, acceptAddress
}

func portMatchesRule(portStr string, portNum uint32, rulePort string) bool {
	if rulePort == portStr {
		return true
	}
	if strings.Contains(rulePort, ",") {
		for _, p := range strings.Split(rulePort, ",") {
			if strings.TrimSpace(p) == portStr {
				return true
			}
		}
		return false
	}
	sep := ""
	if strings.Contains(rulePort, "-") {
		sep = "-"
	} else if strings.Contains(rulePort, ":") {
		sep = ":"
	}
	if sep != "" {
		parts := strings.Split(rulePort, sep)
		if len(parts) == 2 {
			start, err1 := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
			end, err2 := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 32)
			if err1 == nil && err2 == nil && uint64(portNum) >= start && uint64(portNum) <= end {
				return true
			}
		}
	}
	return false
}

// determineStatus assigns a security status to a port based on its bind address,
// source type, firewall state, and rule coverage. Priority order:
// 1. Loopback or link-local bind (127.0.0.0/8, ::1, fe80::/10) → localOnly (unreachable from outside)
// 2. Docker/appStore source with wildcard bind → dockerBypass (rule applies to INPUT
//    chain but Docker traffic goes through FORWARD, so the rule is silently ineffective)
// 3. Firewall inactive → firewallInactive
// 4. Matching rule with drop/reject strategy → blocked
// 5. Matching rule with accept strategy → protected
// 6. Otherwise → noRule
//
// Note: a port bound to a specific non-loopback host IP (e.g. the server's public NIC IP
// or the docker bridge gateway 172.17.0.1) is reachable on that interface and therefore
// goes through the firewall-rule check rather than being labelled localOnly.
func determineStatus(bindAddr, sourceType string, firewallActive, hasRule bool, ruleStrategy string) string {
	if isLoopbackOrLinkLocal(bindAddr) {
		return "localOnly"
	}
	if sourceType == "docker" || sourceType == "appStore" {
		return "dockerBypass"
	}
	if !firewallActive {
		return "firewallInactive"
	}
	if hasRule {
		if ruleStrategy == "drop" || ruleStrategy == "reject" {
			return "blocked"
		}
		return "protected"
	}
	return "noRule"
}

func statusSortPriority(status string) int {
	switch status {
	case "firewallInactive":
		return 0
	case "dockerBypass":
		return 1
	case "noRule":
		return 2
	case "protected":
		return 3
	case "blocked":
		return 4
	case "localOnly":
		return 5
	default:
		return 6
	}
}

// isWildcardAddress returns true if the address binds to all interfaces.
func isWildcardAddress(addr string) bool {
	return addr == "0.0.0.0" || addr == "::" || addr == ""
}

// isLoopbackOrLinkLocal returns true if the address is provably unreachable from
// outside the host: loopback (127.0.0.0/8, ::1) or link-local (169.254.0.0/16,
// fe80::/10). Wildcard addresses (0.0.0.0, ::) are NOT loopback — they bind every
// interface including public ones. A specific non-loopback IP (a public NIC IP, a
// docker bridge gateway like 172.17.0.1, etc.) is reachable on that interface and
// must go through firewall-rule evaluation, not be labelled localOnly outright.
func isLoopbackOrLinkLocal(addr string) bool {
	if addr == "" {
		return false
	}
	ip := stdnet.ParseIP(addr)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func getProcessNameByPID(pid int32) string {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, _ := proc.Name()
	return name
}
