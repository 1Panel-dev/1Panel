package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const FirewallRuleSequenceStep int64 = 1 << 32

type DockerPortGuardPolicy struct {
	BaseModel

	UUID        string `gorm:"size:64;not null;uniqueIndex" json:"uuid"`
	Family      string `gorm:"size:16;not null;uniqueIndex:idx_docker_port_guard_endpoint" json:"family"`
	HostIP      string `gorm:"size:64;not null;uniqueIndex:idx_docker_port_guard_endpoint" json:"hostIP"`
	HostPort    uint16 `gorm:"not null;uniqueIndex:idx_docker_port_guard_endpoint" json:"hostPort"`
	Protocol    string `gorm:"size:8;not null;uniqueIndex:idx_docker_port_guard_endpoint" json:"protocol"`
	Mode        string `gorm:"size:32;not null" json:"mode"`
	Sources     string `gorm:"type:text" json:"-"`
	Description string `gorm:"type:text" json:"description"`
}

type ForwardingRule struct {
	BaseModel

	Family     string `gorm:"size:16;not null;uniqueIndex:idx_forwarding_rule_identity" json:"family"`
	Protocol   string `gorm:"size:8;not null;uniqueIndex:idx_forwarding_rule_identity" json:"protocol"`
	Port       string `gorm:"size:32;not null;uniqueIndex:idx_forwarding_rule_identity" json:"port"`
	TargetIP   string `gorm:"size:64;not null;uniqueIndex:idx_forwarding_rule_identity" json:"targetIP"`
	TargetPort string `gorm:"size:32;not null;uniqueIndex:idx_forwarding_rule_identity" json:"targetPort"`
	Interface  string `gorm:"size:32;not null;default:'';uniqueIndex:idx_forwarding_rule_identity" json:"interface"`
}

type FirewallRule struct {
	UUID   string `gorm:"size:64;primaryKey" json:"uuid"`
	Family string `gorm:"size:16;not null" json:"family"`

	Protocol           string `gorm:"size:32;not null" json:"protocol"`
	SourceAddress      string `gorm:"size:255" json:"sourceAddress"`
	SourcePort         string `gorm:"size:64" json:"sourcePort"`
	DestinationAddress string `gorm:"size:255" json:"destinationAddress"`
	DestinationPort    string `gorm:"size:64" json:"destinationPort"`
	Interface          string `gorm:"size:128" json:"interface"`
	ConnectionStates   string `gorm:"type:text" json:"connectionStates"`
	Action             string `gorm:"size:32;not null" json:"action"`
	Description        string `gorm:"type:text" json:"description"`
	CompatibilityError string `gorm:"type:text" json:"compatibilityError,omitempty"`
	Priority           *int   `json:"priority,omitempty"`
	Sequence           *int64 `gorm:"index" json:"sequence,omitempty"`

	Origin   string `gorm:"size:32;not null" json:"origin"`
	Owner    string `gorm:"size:320;not null" json:"owner"`
	Revision uint   `gorm:"not null;default:1" json:"revision"`
}

func FirewallRuleOwner(sourceKind, sourceID string) string {
	sourceKind = strings.TrimSpace(sourceKind)
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		return sourceKind
	}
	return sourceKind + ":" + sourceID
}

func FirewallRulesRevision(rules []FirewallRule) (string, error) {
	ordered := append([]FirewallRule(nil), rules...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].UUID < ordered[j].UUID
	})
	payload, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func FirewallRuleFromDomain(rule filter.FirewallRule) (FirewallRule, error) {
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return FirewallRule{}, err
	}
	switch normalized.NativeKind {
	case "", filter.NativeKindRule, filter.NativeKindZonePort, filter.NativeKindRichRule, filter.NativeKindUFWRule:
	default:
		return FirewallRule{}, fmt.Errorf("%w: native rule %q cannot be stored as a provider-neutral policy", filter.ErrUnsupportedScope, normalized.NativeKind)
	}
	record := FirewallRule{
		Family:             string(normalized.Scope.Family),
		Protocol:           normalized.Protocol,
		SourceAddress:      normalized.SourceAddress,
		SourcePort:         normalized.SourcePort,
		DestinationAddress: normalized.DestinationAddress,
		DestinationPort:    normalized.DestinationPort,
		Interface:          normalized.Interface,
		ConnectionStates:   strings.Join(normalized.ConnectionStates, ","),
		Action:             string(normalized.Action),
		Description:        normalized.Description,
	}
	if normalized.Scope.Provider == filter.ProviderFirewalld {
		record.Priority = normalized.Priority
	}
	return record, nil
}

func (rule FirewallRule) PolicyKey() string {
	payload, _ := json.Marshal(struct {
		Family             string `json:"family"`
		Protocol           string `json:"protocol"`
		SourceAddress      string `json:"sourceAddress,omitempty"`
		SourcePort         string `json:"sourcePort,omitempty"`
		DestinationAddress string `json:"destinationAddress,omitempty"`
		DestinationPort    string `json:"destinationPort,omitempty"`
		Interface          string `json:"interface,omitempty"`
		ConnectionStates   string `json:"connectionStates,omitempty"`
		Action             string `json:"action"`
	}{
		Family: rule.Family, Protocol: rule.Protocol,
		SourceAddress: rule.SourceAddress, SourcePort: rule.SourcePort,
		DestinationAddress: rule.DestinationAddress, DestinationPort: rule.DestinationPort,
		Interface: rule.Interface, ConnectionStates: rule.ConnectionStates, Action: rule.Action,
	})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func (rule FirewallRule) RulesForProvider(provider filter.Provider) ([]filter.FirewallRule, error) {
	if rule.CompatibilityError != "" {
		return nil, fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, rule.CompatibilityError)
	}
	connectionStates := make([]string, 0)
	if rule.ConnectionStates != "" {
		connectionStates = strings.Split(rule.ConnectionStates, ",")
	}
	base := filter.FirewallRule{
		Protocol: rule.Protocol, SourceAddress: rule.SourceAddress, SourcePort: rule.SourcePort,
		DestinationAddress: rule.DestinationAddress, DestinationPort: rule.DestinationPort,
		Interface: rule.Interface, ConnectionStates: connectionStates,
		Action: filter.Action(rule.Action), Description: rule.Description,
	}
	if provider == filter.ProviderFirewalld {
		base.Priority = rule.Priority
	}
	families := []filter.Family{filter.Family(rule.Family)}
	if provider != filter.ProviderFirewalld && families[0] == filter.FamilyInet {
		hasIPv4, hasIPv6 := ruleAddressFamilies(base)
		switch {
		case hasIPv4 && hasIPv6:
			return nil, fmt.Errorf("%w: inet policy contains both IPv4 and IPv6 addresses", filter.ErrUnsupportedScope)
		case hasIPv6 || strings.EqualFold(base.Protocol, "icmpv6"):
			families = []filter.Family{filter.FamilyIPv6}
		case hasIPv4:
			families = []filter.Family{filter.FamilyIPv4}
		default:
			families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
		}
	}
	result := make([]filter.FirewallRule, 0, len(families))
	for _, family := range families {
		compiled := base
		compiled.Scope = filter.Scope{Provider: provider, Family: family, Direction: filter.DirectionInput}
		switch provider {
		case filter.ProviderIptables, filter.ProviderNftables:
			compiled.Scope.Table, compiled.Scope.Chain = "filter", filter.IptablesInputChain
		case filter.ProviderFirewalld:
			compiled.Scope.Zone = filter.FirewalldInputZone
		case filter.ProviderUFW:
			compiled.Scope.Chain = filter.UFWInputChain
		default:
			return nil, fmt.Errorf("%w: unsupported firewall provider %q", filter.ErrProviderUnavailable, provider)
		}
		expanded, err := filter.ExpandAtomicRules(compiled)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

func SortFirewallRules(rules []FirewallRule, provider filter.Provider) {
	sort.SliceStable(rules, func(i, j int) bool {
		left, right := rules[i], rules[j]
		if provider == filter.ProviderFirewalld {
			switch {
			case left.Priority == nil && right.Priority != nil:
				return false
			case left.Priority != nil && right.Priority == nil:
				return true
			case left.Priority != nil && right.Priority != nil && *left.Priority != *right.Priority:
				return *left.Priority < *right.Priority
			}
		} else {
			switch {
			case left.Sequence == nil && right.Sequence != nil:
				return false
			case left.Sequence != nil && right.Sequence == nil:
				return true
			case left.Sequence != nil && right.Sequence != nil && *left.Sequence != *right.Sequence:
				return *left.Sequence < *right.Sequence
			}
		}
		return left.UUID < right.UUID
	})
}

func ruleAddressFamilies(rule filter.FirewallRule) (bool, bool) {
	hasIPv4, hasIPv6 := false, false
	for _, address := range []string{rule.SourceAddress, rule.DestinationAddress} {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		if strings.Contains(address, ":") {
			hasIPv6 = true
		} else {
			hasIPv4 = true
		}
	}
	return hasIPv4, hasIPv6
}
