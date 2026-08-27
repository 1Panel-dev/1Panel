package forwarding

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

var ErrRuleExists = errors.New("forwarding rule already exists")

const (
	FamilyIPv4 = constant.FirewallFamilyIPv4
	FamilyIPv6 = constant.FirewallFamilyIPv6

	ChainPreRouting  = "1PANEL_PREROUTING"
	ChainPostRouting = "1PANEL_POSTROUTING"
	ChainForward     = "1PANEL_FORWARD"

	ForwardFile     = "1panel_forward.rules"
	PreRoutingFile  = "1panel_forward_pre.rules"
	PostRoutingFile = "1panel_forward_post.rules"
)

type Rule struct {
	Num        string
	Family     string
	Protocol   string
	Port       string
	TargetIP   string
	TargetPort string
	Interface  string
}

func (r Rule) Identity() string {
	return strings.Join([]string{r.Family, r.Protocol, r.Port, r.TargetIP, r.TargetPort, r.Interface}, "\x00")
}

type OperationType string

const (
	OperationAdd    OperationType = "add"
	OperationRemove OperationType = "remove"
)

type Adapter interface {
	Name() string
	List() ([]Rule, error)
	Reconcile(rules []Rule) error
	Enable() error
	Cleanup() error
	InitStatus() (bool, bool, error)
	FamilyStatus(family string) (bool, bool, error)
	Replay() error
}

type Status struct {
	Name    string
	Version string
	IsInit  bool
	IsBind  bool
}

type RuntimeClient interface {
	Version() (string, error)
}

type Manager struct {
	adapter Adapter
	runtime RuntimeClient
}

func NewManager(adapter Adapter, runtime RuntimeClient) *Manager {
	return &Manager{adapter: adapter, runtime: runtime}
}

func (m *Manager) Status() (Status, error) {
	status := Status{Name: m.adapter.Name(), Version: "-"}
	var versionErr error
	var initErr error
	var wg sync.WaitGroup
	wg.Add(1)
	if m.runtime != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status.Version, versionErr = m.runtime.Version()
		}()
	}
	go func() {
		defer wg.Done()
		status.IsInit, status.IsBind, initErr = m.adapter.InitStatus()
	}()
	wg.Wait()
	return status, errors.Join(versionErr, initErr)
}

func (m *Manager) List(info, strategy string) ([]Rule, error) {
	rules, err := m.adapter.List()
	if err != nil {
		return nil, err
	}
	if strategy != "" {
		return []Rule{}, nil
	}
	filtered := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if info != "" && !strings.Contains(rule.Port, info) &&
			!strings.Contains(rule.TargetPort, info) && !strings.Contains(rule.TargetIP, info) {
			continue
		}
		filtered = append(filtered, rule)
	}
	return filtered, nil
}

func (m *Manager) Enable() error { return m.adapter.Enable() }

func (m *Manager) Reconcile(rules []Rule) error { return m.adapter.Reconcile(rules) }

func (m *Manager) Cleanup() error { return m.adapter.Cleanup() }

func (m *Manager) FamilyStatus(family string) (bool, bool, error) {
	return m.adapter.FamilyStatus(family)
}

func (m *Manager) Replay() error { return m.adapter.Replay() }

func (m *Manager) Name() string { return m.adapter.Name() }

func NormalizeRule(rule Rule) (Rule, error) {
	rule.Family = strings.ToLower(strings.TrimSpace(rule.Family))
	if rule.Family == "" {
		rule.Family = FamilyIPv4
	}
	if rule.Family != FamilyIPv4 && rule.Family != FamilyIPv6 {
		return Rule{}, fmt.Errorf("unsupported forwarding family %q", rule.Family)
	}
	rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
	if rule.Protocol != "tcp" && rule.Protocol != "udp" {
		return Rule{}, fmt.Errorf("unsupported forwarding protocol %q", rule.Protocol)
	}
	var err error
	if rule.Port, err = normalizeForwardPort(rule.Port); err != nil {
		return Rule{}, fmt.Errorf("invalid forwarding port: %w", err)
	}
	if rule.TargetPort, err = normalizeForwardPort(rule.TargetPort); err != nil {
		return Rule{}, fmt.Errorf("invalid forwarding target port: %w", err)
	}
	rule.TargetIP = strings.TrimSpace(rule.TargetIP)
	if rule.TargetIP == "" || strings.EqualFold(rule.TargetIP, "localhost") {
		if rule.Family == FamilyIPv6 {
			rule.TargetIP = "::1"
		} else {
			rule.TargetIP = "127.0.0.1"
		}
	}
	address, err := netip.ParseAddr(rule.TargetIP)
	if err == nil {
		address = address.Unmap()
	}
	if err != nil || (rule.Family == FamilyIPv4) != address.Is4() {
		return Rule{}, fmt.Errorf("invalid %s forwarding target %q", rule.Family, rule.TargetIP)
	}
	rule.TargetIP = address.String()
	rule.Interface = strings.TrimSpace(rule.Interface)
	if rule.Interface == "all" || rule.Interface == "*" {
		rule.Interface = ""
	}
	if rule.Interface != "" && !re.ForwardInterfaceRegex.MatchString(rule.Interface) {
		return Rule{}, fmt.Errorf("invalid forwarding interface %q", rule.Interface)
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
