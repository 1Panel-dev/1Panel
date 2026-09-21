package filter

import (
	"github.com/1Panel-dev/1Panel/agent/constant"
)

type RuleOrigin string

const (
	RuleOriginCreated RuleOrigin = constant.FirewallRuleOriginCreated
	RuleOriginAdopted RuleOrigin = constant.FirewallRuleOriginAdopted
)

type InventoryState string

const (
	InventoryStateManaged   InventoryState = "managed"
	InventoryStateAdopted   InventoryState = "adopted"
	InventoryStateExternal  InventoryState = "external"
	InventoryStateDrifted   InventoryState = "drifted"
	InventoryStateProtected InventoryState = "protected"
)

type InventoryMatch string

const (
	InventoryMatchNone      InventoryMatch = "none"
	InventoryMatchExact     InventoryMatch = "exact"
	InventoryMatchChanged   InventoryMatch = "changed"
	InventoryMatchMissing   InventoryMatch = "missing"
	InventoryMatchAmbiguous InventoryMatch = "ambiguous"
	InventoryMatchOpaque    InventoryMatch = "opaque"
)

type DesiredRule struct {
	UUID                string       `json:"uuid"`
	Rule                FirewallRule `json:"rule"`
	RuleKey             string       `json:"ruleKey"`
	Origin              RuleOrigin   `json:"origin"`
	Protected           bool         `json:"protected,omitempty"`
	Expanded            bool         `json:"expanded,omitempty"`
	Marker              string       `json:"marker,omitempty"`
	ObservedInstanceKey string       `json:"observedInstanceKey,omitempty"`
}

type PositionRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type InventoryItem struct {
	Incompatible bool           `json:"incompatible,omitempty"`
	Error        string         `json:"error,omitempty"`
	Rule         FirewallRule   `json:"rule"`
	Observed     *ObservedRule  `json:"observed,omitempty"`
	Desired      *DesiredRule   `json:"desired,omitempty"`
	State        InventoryState `json:"state"`
	Match        InventoryMatch `json:"match"`
}

type Inventory struct {
	Items   []InventoryItem `json:"items"`
	Notices []ScopeNotice   `json:"notices,omitempty"`
}

type InventoryMergeInput struct {
	Observed              []ObservedRule
	Desired               []DesiredRule
	ProtectedObservedKeys map[string]struct{}
}
