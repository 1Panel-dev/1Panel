package filter

import (
	"errors"
	"fmt"
	"strings"
)

var ErrProviderUnavailable = errors.New("firewall provider is unavailable")

type Provider string

const (
	ProviderIptables  Provider = "iptables"
	ProviderFirewalld Provider = "firewalld"
	ProviderUFW       Provider = "ufw"
)

type Family string

const (
	FamilyIPv4 Family = "ipv4"
	FamilyIPv6 Family = "ipv6"
	FamilyInet Family = "inet"
)

type Direction string

const (
	DirectionInput Direction = "input"
)

type Action string

const (
	ActionAccept Action = "accept"
	ActionDrop   Action = "drop"
	ActionReject Action = "reject"
)

type NativeKind string

const (
	NativeKindRule        NativeKind = "rule"
	NativeKindZonePort    NativeKind = "zone_port"
	NativeKindRichRule    NativeKind = "rich_rule"
	NativeKindUFWRule     NativeKind = "ufw_rule"
	NativeKindOpaque      NativeKind = "opaque"
	NativeKindZoneService NativeKind = "zone_service"
)

const (
	OrderBucketRichPre            = "rich_pre"
	OrderBucketRichZeroDeny       = "rich_zero_deny"
	OrderBucketZonePrimitiveAllow = "zone_primitive_allow"
	OrderBucketRichZeroAllow      = "rich_zero_allow"
	OrderBucketRichPost           = "rich_post"
)

type ParseStatus string

const (
	ParseStatusSupported ParseStatus = "supported"
	ParseStatusOpaque    ParseStatus = "opaque"
)

type PersistenceStatus string

const (
	PersistenceStatusConverged     PersistenceStatus = "converged"
	PersistenceStatusRuntimeOnly   PersistenceStatus = "runtime_only"
	PersistenceStatusPermanentOnly PersistenceStatus = "permanent_only"
)

type ScopeNoticeCode string

const (
	ScopeNoticeDefaultScopeMismatch     ScopeNoticeCode = "default_scope_mismatch"
	ScopeNoticeManagedScopeInactive     ScopeNoticeCode = "managed_scope_inactive"
	ScopeNoticeUnmanagedActiveScopes    ScopeNoticeCode = "unmanaged_active_scopes"
	ScopeNoticeRuntimePermanentMismatch ScopeNoticeCode = "runtime_permanent_mismatch"
	ScopeNoticeDefaultPolicy            ScopeNoticeCode = "default_policy"
	ScopeNoticeManagedScopeMissing      ScopeNoticeCode = "managed_scope_missing"
)

const (
	IptablesInputChain = "1PANEL_BASIC"
	FirewalldInputZone = "public"
	UFWInputChain      = "incoming"
)

type ScopeNotice struct {
	Code   ScopeNoticeCode `json:"code"`
	Values []string        `json:"values,omitempty"`
}

var (
	ErrInvalidScope     = errors.New("invalid firewall scope")
	ErrUnsupportedScope = errors.New("unsupported firewall scope")
	ErrInvalidRule      = errors.New("invalid firewall rule")
	ErrProtectedRule    = errors.New("protected firewall rule cannot be modified")
	ErrLockoutRisk      = errors.New("firewall change may lock out management access")
	ErrCompositeRule    = errors.New("firewall rule must be atomic")
	ErrExpansionLimit   = errors.New("firewall rule expansion limit exceeded")
)

type Scope struct {
	Provider  Provider  `json:"provider"`
	Family    Family    `json:"family"`
	Table     string    `json:"table,omitempty"`
	Zone      string    `json:"zone,omitempty"`
	Chain     string    `json:"chain,omitempty"`
	Direction Direction `json:"direction"`
}

func (s Scope) Normalize() Scope {
	s.Provider = Provider(strings.ToLower(strings.TrimSpace(string(s.Provider))))
	s.Family = Family(strings.ToLower(strings.TrimSpace(string(s.Family))))
	s.Table = strings.ToLower(strings.TrimSpace(s.Table))
	s.Zone = strings.ToLower(strings.TrimSpace(s.Zone))
	s.Chain = strings.TrimSpace(s.Chain)
	s.Direction = Direction(strings.ToLower(strings.TrimSpace(string(s.Direction))))

	if s.Provider == ProviderIptables {
		s.Chain = strings.ToUpper(s.Chain)
		if s.Chain == "" {
			s.Chain = IptablesInputChain
		}
	}
	if s.Provider == ProviderUFW && s.Chain == "" {
		s.Chain = UFWInputChain
	}
	return s
}

func (s Scope) Key() string {
	s = s.Normalize()
	switch s.Provider {
	case ProviderIptables:
		return strings.Join([]string{string(s.Provider), string(s.Family), s.Table, s.Chain, string(s.Direction)}, ":")
	case ProviderFirewalld:
		return strings.Join([]string{string(s.Provider), s.Zone, string(s.Direction)}, ":")
	case ProviderUFW:
		return strings.Join([]string{string(s.Provider), s.Chain, string(s.Family)}, ":")
	default:
		return strings.Join([]string{string(s.Provider), string(s.Family), s.Table, s.Zone, s.Chain, string(s.Direction)}, ":")
	}
}

func (s Scope) ValidateMVP() error {
	s = s.Normalize()
	if !s.Provider.valid() || !s.Family.valid() || !s.Direction.valid() {
		return fmt.Errorf("%w: provider=%q family=%q direction=%q", ErrInvalidScope, s.Provider, s.Family, s.Direction)
	}

	switch s.Provider {
	case ProviderIptables:
		if s.Family != FamilyIPv4 && s.Family != FamilyIPv6 {
			return fmt.Errorf("%w: iptables family %q", ErrUnsupportedScope, s.Family)
		}
		if s.Table != "filter" || s.Zone != "" || s.Direction != DirectionInput || !isBasicChain(s.Chain) {
			return fmt.Errorf("%w: %s", ErrUnsupportedScope, s.Key())
		}
	case ProviderFirewalld:
		if s.Table != "" || s.Chain != "" || s.Direction != DirectionInput || s.Zone != FirewalldInputZone {
			return fmt.Errorf("%w: %s", ErrUnsupportedScope, s.Key())
		}
	case ProviderUFW:
		if (s.Family != FamilyIPv4 && s.Family != FamilyIPv6) || s.Table != "" || s.Zone != "" ||
			s.Direction != DirectionInput || s.Chain != UFWInputChain {
			return fmt.Errorf("%w: %s", ErrUnsupportedScope, s.Key())
		}
	default:
		return fmt.Errorf("%w: provider %q", ErrUnsupportedScope, s.Provider)
	}
	return nil
}

func (p Provider) valid() bool {
	switch p {
	case ProviderIptables, ProviderFirewalld, ProviderUFW:
		return true
	default:
		return false
	}
}

func (f Family) valid() bool {
	switch f {
	case FamilyIPv4, FamilyIPv6, FamilyInet:
		return true
	default:
		return false
	}
}

func (d Direction) valid() bool {
	return d == DirectionInput
}

func isBasicChain(chain string) bool {
	switch chain {
	case "1PANEL_BASIC_BEFORE", "1PANEL_BASIC", "1PANEL_BASIC_AFTER":
		return true
	default:
		return false
	}
}

type ScopePattern struct {
	Provider   Provider
	Families   []Family
	Table      string
	Zone       string
	Chains     []string
	Directions []Direction
}

func (p ScopePattern) Matches(scope Scope) bool {
	scope = scope.Normalize()
	return p.Provider == scope.Provider &&
		containsFamily(p.Families, scope.Family) &&
		strings.EqualFold(p.Table, scope.Table) &&
		strings.EqualFold(p.Zone, scope.Zone) &&
		containsFold(p.Chains, scope.Chain) &&
		containsDirection(p.Directions, scope.Direction)
}

func MVPScopePatterns() []ScopePattern {
	return []ScopePattern{
		{
			Provider:   ProviderIptables,
			Families:   []Family{FamilyIPv4, FamilyIPv6},
			Table:      "filter",
			Chains:     []string{"1PANEL_BASIC_BEFORE", "1PANEL_BASIC", "1PANEL_BASIC_AFTER"},
			Directions: []Direction{DirectionInput},
		},
		{
			Provider:   ProviderFirewalld,
			Families:   []Family{FamilyIPv4, FamilyIPv6, FamilyInet},
			Zone:       "public",
			Directions: []Direction{DirectionInput},
		},
		{
			Provider:   ProviderUFW,
			Families:   []Family{FamilyIPv4, FamilyIPv6},
			Chains:     []string{"incoming"},
			Directions: []Direction{DirectionInput},
		},
	}
}

type FirewallRule struct {
	UUID               string     `json:"uuid,omitempty"`
	Scope              Scope      `json:"scope"`
	NativeKind         NativeKind `json:"nativeKind"`
	Protocol           string     `json:"protocol"`
	SourceAddress      string     `json:"sourceAddress,omitempty"`
	SourcePort         string     `json:"sourcePort,omitempty"`
	DestinationAddress string     `json:"destinationAddress,omitempty"`
	DestinationPort    string     `json:"destinationPort,omitempty"`
	Interface          string     `json:"interface,omitempty"`
	ConnectionStates   []string   `json:"connectionStates,omitempty"`
	Action             Action     `json:"action"`
	Priority           *int       `json:"priority,omitempty"`
	OrderIndex         *int64     `json:"orderIndex,omitempty"`
	OrderBucket        string     `json:"orderBucket,omitempty"`
	Description        string     `json:"description,omitempty"`
}

type Locator struct {
	Provider  Provider `json:"provider"`
	ScopeKey  string   `json:"scopeKey"`
	NativeID  string   `json:"nativeId,omitempty"`
	Canonical string   `json:"canonical,omitempty"`
	Position  *int     `json:"position,omitempty"`
}

type ObservedRule struct {
	Rule        FirewallRule      `json:"rule"`
	Locator     Locator           `json:"locator"`
	InstanceKey string            `json:"instanceKey,omitempty"`
	Marker      string            `json:"marker,omitempty"`
	ParseStatus ParseStatus       `json:"parseStatus"`
	Raw         string            `json:"raw,omitempty"`
	Protected   bool              `json:"protected"`
	Persistence PersistenceStatus `json:"persistence,omitempty"`
}

type Snapshot struct {
	Scope    Scope          `json:"scope"`
	Revision string         `json:"revision"`
	Rules    []ObservedRule `json:"rules"`
	Notices  []ScopeNotice  `json:"notices,omitempty"`
}

type Capabilities struct {
	Scopes                []ScopePattern
	Marker                bool
	AtomicApply           bool
	TransactionalRollback bool
	OwnedChains           bool
	ExplicitPosition      bool
	ExplicitPriority      bool
	NativePort            bool
}

func (c Capabilities) SupportsScope(scope Scope) bool {
	for _, pattern := range c.Scopes {
		if pattern.Matches(scope) {
			return true
		}
	}
	return false
}

func containsFamily(values []Family, target Family) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsDirection(values []Direction, target Direction) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsFold(values []string, target string) bool {
	if len(values) == 0 {
		return target == ""
	}
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
