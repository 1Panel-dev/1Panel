package dto

type FirewallBaseInfo struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Version    string `json:"version"`
	PingStatus string `json:"pingStatus"`
}

type RuleSearch struct {
	PageInfo
	Info string `json:"info"`
	Type string `json:"type" validate:"required"`
}

type FirewallOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start stop disablePing enablePing"`
}

type PortRuleOperate struct {
	Operation string `json:"operation" validate:"required,oneof=add remove"`
	Address   string `json:"address"`
	Port      string `json:"port" validate:"required"`
	Protocol  string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
	Strategy  string `json:"strategy" validate:"required,oneof=accept drop"`
}

type AddrRuleOperate struct {
	Operation string `json:"operation" validate:"required,oneof=add remove"`
	Address   string `json:"address"  validate:"required"`
	Strategy  string `json:"strategy" validate:"required,oneof=accept drop"`
}

type PortRuleUpdate struct {
	OldRule PortRuleOperate `json:"oldRule"`
	NewRule PortRuleOperate `json:"newRule"`
}

type AddrRuleUpdate struct {
	OldRule AddrRuleOperate `json:"oldRule"`
	NewRule AddrRuleOperate `json:"newRule"`
}

type BatchRuleOperate struct {
	Type  string            `json:"type" validate:"required"`
	Rules []PortRuleOperate `json:"rules"`
}
