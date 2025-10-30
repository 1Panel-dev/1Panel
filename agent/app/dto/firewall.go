package dto

type FirewallBaseInfo struct {
	Name       string `json:"name"`
	IsExist    bool   `json:"isExist"`
	IsActive   bool   `json:"isActive"`
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
	Operation         string `json:"operation" validate:"required,oneof=start stop restart disablePing enablePing"`
	WithDockerRestart bool   `json:"withDockerRestart"`
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
	ForceDelete bool `json:"forceDelete"`
	Rules       []struct {
		Operation  string `json:"operation" validate:"required,oneof=add remove"`
		Num        string `json:"num"`
		Protocol   string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
		Interface  string `json:"interface"`
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

// Iptables Filter DTO

type IptablesFilterRuleSearch struct {
	Chains []string `json:"chains"`
}

type IptablesChainInfo struct {
	Version       string                   `json:"version"`
	Name          string                   `json:"name"`
	DefaultPolicy string                   `json:"defaultPolicy"`
	Rules         []IptablesFilterRuleInfo `json:"rules"`
	IsApplied     bool                     `json:"isApplied"`
}

type IptablesFilterRuleInfo struct {
	ID          uint   `json:"id"`
	Protocol    string `json:"protocol"`
	SourceIP    string `json:"sourceIP"`
	SourcePort  uint16 `json:"sourcePort"`
	DestIP      string `json:"destIP"`
	DestPort    uint16 `json:"destPort"`
	Action      string `json:"action"`
	Comment     string `json:"comment"`
	Description string `json:"description"`
	RuleOrder   int    `json:"ruleOrder"`
	IsActive    bool   `json:"isActive"`
}

type IptablesFilterRuleOperate struct {
	Operation   string `json:"operation" validate:"required,oneof=add remove"`
	ID          uint   `json:"id"`
	Chain       string `json:"chain" validate:"required,oneof=1PANEL_INPUT 1PANEL_OUTPUT"`
	Protocol    string `json:"protocol"`
	SourceIP    string `json:"sourceIP"`
	SourcePort  uint16 `json:"sourcePort"`
	DestIP      string `json:"destIP"`
	DestPort    uint16 `json:"destPort"`
	Action      string `json:"action" validate:"required,oneof=ACCEPT DROP REJECT"`
	Description string `json:"description"`
}

type IptablesFilterApply struct {
	Operation string `json:"operation" validate:"required,oneof=apply unload init"`
}

type IptablesFilterBatchOperate struct {
	Rules []IptablesFilterRuleOperate `json:"rules"`
}
