package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
