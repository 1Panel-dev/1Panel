package forwarding

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

const (
	FamilyIPv4       = constant.FirewallFamilyIPv4
	FamilyIPv6       = constant.FirewallFamilyIPv6
	ChainPreRouting  = "1PANEL_PREROUTING"
	ChainPostRouting = "1PANEL_POSTROUTING"
	ChainForward     = "1PANEL_FORWARD"
	ForwardFile      = "1panel_forward.rules"
	PreRoutingFile   = "1panel_forward_pre.rules"
	PostRoutingFile  = "1panel_forward_post.rules"
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
	CreateRules(context.Context, []Rule) error
	DeleteRules(context.Context, []Rule) error
	ReplaceRules(rules []Rule) error
	Enable() error
	Cleanup() error
	InitStatus() (bool, bool, error)
	FamilyStatus(family string) (bool, bool, error)
	Replay() error
}

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
