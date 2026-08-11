package dto

import "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"

type FirewallBaseInfo struct {
	Name       string `json:"name"`
	IsExist    bool   `json:"isExist"`
	IsActive   bool   `json:"isActive"`
	IsInit     bool   `json:"isInit"`
	IsBind     bool   `json:"isBind"`
	Version    string `json:"version"`
	PingStatus string `json:"pingStatus"`
}

type FirewallOperation struct {
	Operation         string `json:"operation" validate:"required,oneof=start stop restart disableBanPing enableBanPing"`
	WithDockerRestart bool   `json:"withDockerRestart"`
}

type IptablesOp struct {
	Name    string `json:"name" validate:"required,eq=1PANEL_BASIC"`
	Operate string `json:"operate" validate:"required,oneof=init-base bind-base unbind-base"`
}

type FirewallSystemPort struct {
	Port     string
	Protocol string
}

type FirewallRuleInventoryResponse struct {
	Items   []filter.InventoryItem `json:"items"`
	Notices []filter.ScopeNotice   `json:"notices,omitempty"`
}

type FirewallRuleCheckResponse struct {
	Decision         filter.CheckDecision       `json:"decision"`
	Classification   filter.CheckClassification `json:"classification"`
	Reason           string                     `json:"reason"`
	RequestedRule    filter.FirewallRule        `json:"requestedRule"`
	RequestedRuleKey string                     `json:"requestedRuleKey"`
	ExistingRuleUUID string                     `json:"existingRuleUUID,omitempty"`
	Candidates       []filter.ObservedRule      `json:"candidates,omitempty"`
	AllowedActions   []filter.CheckAction       `json:"allowedActions,omitempty"`
	CheckFlag        string                     `json:"checkFlag"`
}

type FirewallRuleInventory struct {
	Scope filter.Scope `json:"scope" validate:"required"`
}

type FirewallNativeDetail struct {
	Provider   filter.Provider   `json:"provider" validate:"required,oneof=firewalld ufw"`
	NativeKind filter.NativeKind `json:"nativeKind" validate:"required,oneof=zone_service ufw_application"`
	Name       string            `json:"name" validate:"required"`
	Permanent  bool              `json:"permanent"`
}

type FirewallRuleCheck struct {
	UUID string              `json:"uuid"`
	Rule filter.FirewallRule `json:"rule" validate:"required"`
}

type FirewallRuleBatchCheck struct {
	Rules []filter.FirewallRule `json:"rules" validate:"required,min=1,max=256,dive"`
}

type FirewallRuleBatchCheckResponse struct {
	Items []FirewallRuleCheckResponse `json:"items"`
}

type FirewallRuleCreate struct {
	Rule             filter.FirewallRule `json:"rule" validate:"required"`
	CheckFlag        string              `json:"checkFlag"`
	Action           filter.CheckAction  `json:"action"`
	AdoptInstanceKey string              `json:"adoptInstanceKey"`
	SourceKind       string              `json:"sourceKind" validate:"omitempty,oneof=user imported"`
	SourceID         string              `json:"sourceID"`
}

type FirewallRuleBatchCreate struct {
	Items []FirewallRuleCreate `json:"items" validate:"required,min=1,max=256,dive"`
}

type FirewallRuleBatchCreateResponse struct {
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type FirewallRuleDelete struct {
	UUID string `json:"uuid" validate:"required"`
}

type FirewallRuleUpdate struct {
	UUID string              `json:"uuid"`
	Rule filter.FirewallRule `json:"rule" validate:"required"`
}

type FirewallRuleReorder struct {
	UUID           string `json:"uuid"`
	TargetPosition *int64 `json:"targetPosition"`
	Priority       *int   `json:"priority"`
}
