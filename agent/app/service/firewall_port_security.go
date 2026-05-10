package service

import (
	"context"
	"fmt"
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
	appPortMap := buildAppPortMap(ctx)

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

		if appName, ok := appPortMap[conn.Laddr.Port]; ok {
			item.AppName = appName
			if item.SourceType == "docker" {
				item.SourceType = "appStore"
			}
		}

		ruleStrategy, hasRule := matchFirewallRule(ruleIndex, conn.Laddr.Port, proto)
		item.HasRule = hasRule
		item.RuleStrategy = ruleStrategy

		items = append(items, item)
	}

	for i := range items {
		items[i].Status = determineStatus(items[i].BindAddress, items[i].SourceType, firewallActive, items[i].HasRule)
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
		case "protected":
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
					!strings.Contains(strings.ToLower(item.ContainerName), keyword) {
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
		})
	}
	return entries
}

// buildAppPortMap creates a lookup map from port number to app name.
func buildAppPortMap(ctx context.Context) map[uint32]string {
	result := make(map[uint32]string)
	apps, err := appInstallRepo.ListBy(ctx)
	if err != nil {
		return result
	}
	for _, app := range apps {
		if app.HttpPort > 0 {
			result[uint32(app.HttpPort)] = app.App.Key
		}
		if app.HttpsPort > 0 {
			result[uint32(app.HttpsPort)] = app.App.Key
		}
	}
	systemPort, err := settingRepo.Get(settingRepo.WithByKey("ServerPort"))
	if err == nil && systemPort.Value != "" {
		if port, e := strconv.ParseUint(systemPort.Value, 10, 32); e == nil {
			result[uint32(port)] = "1panel"
		}
	}
	return result
}

// matchFirewallRule checks if a port has a matching firewall rule, respecting protocol.
func matchFirewallRule(rules []firewallRuleEntry, port uint32, proto string) (string, bool) {
	portStr := strconv.FormatUint(uint64(port), 10)
	for _, r := range rules {
		if r.protocol != "" && r.protocol != "tcp/udp" && r.protocol != proto {
			continue
		}
		if portMatchesRule(portStr, port, r.portStr) {
			return r.strategy, true
		}
	}
	return "", false
}

func portMatchesRule(portStr string, portNum uint32, rulePort string) bool {
	if rulePort == portStr {
		return true
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
// 1. Non-wildcard bind address → localOnly
// 2. Docker/appStore source → dockerBypass
// 3. Firewall inactive → firewallInactive
// 4. Has matching rule → protected
// 5. Otherwise → noRule
func determineStatus(bindAddr, sourceType string, firewallActive, hasRule bool) string {
	if !isWildcardAddress(bindAddr) {
		return "localOnly"
	}
	if sourceType == "docker" || sourceType == "appStore" {
		return "dockerBypass"
	}
	if !firewallActive {
		return "firewallInactive"
	}
	if hasRule {
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
	case "localOnly":
		return 4
	default:
		return 5
	}
}

// isWildcardAddress returns true if the address binds to all interfaces.
func isWildcardAddress(addr string) bool {
	return addr == "0.0.0.0" || addr == "::" || addr == ""
}

func getProcessNameByPID(pid int32) string {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}
	name, _ := proc.Name()
	return name
}
