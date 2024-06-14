package dto

type FirewallBaseInfo struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Version    string `json:"version"`
	PingStatus string `json:"pingStatus"`
}

type RuleSearch struct {
	PageInfo
	Info     string `json:"info"`
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
	Type     string `json:"type" validate:"required"`
}

type FirewallOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start stop restart disablePing enablePing"`
}

type PortRuleOperate struct {
	Operation string `json:"operation" validate:"required,oneof=add remove"`
	Address   string `json:"address"`
	Port      string `json:"port" validate:"required"`
	Protocol  string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
	Strategy  string `json:"strategy" validate:"required,oneof=accept drop"`

	Description string `json:"description"`
}

type ForwardRuleOperate struct {
	Rules []struct {
		Operation  string `json:"operation" validate:"required,oneof=add remove"`
		Num        string `json:"num"`
		Protocol   string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
		Port       string `json:"port" validate:"required"`
		TargetIP   string `json:"targetIP"`
		TargetPort string `json:"targetPort" validate:"required"`
	} `json:"rules"`
}

type UpdateFirewallDescription struct {
	Type     string `json:"type"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Strategy string `json:"strategy" validate:"required,oneof=accept drop"`

	Description string `json:"description"`
}

type AddrRuleOperate struct {
	Operation string `json:"operation" validate:"required,oneof=add remove"`
	Address   string `json:"address"  validate:"required"`
	Strategy  string `json:"strategy" validate:"required,oneof=accept drop"`

	Description string `json:"description"`
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
