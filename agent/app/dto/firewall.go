package dto

type FirewallBaseInfo struct {
	Name       string `json:"name"`
	IsExist    bool   `json:"isExist"`
	IsActive   bool   `json:"isActive"`
	IsInit     bool   `json:"isInit"`
	IsBind     bool   `json:"isBind"`
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
	Operation         string `json:"operation" validate:"required,oneof=start stop restart disableBanPing enableBanPing"`
	WithDockerRestart bool   `json:"withDockerRestart"`
}

type PortRuleOperate struct {
	ID        uint   `json:"id"`
	Operation string `json:"operation" validate:"required,oneof=add remove"`
	Chain     string `json:"chain"`
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
	Chain    string `json:"chain"`
	SrcIP    string `json:"srcIP"`
	DstIP    string `json:"dstIP"`
	SrcPort  string `json:"srcPort"`
	DstPort  string `json:"dstPort"`
	Protocol string `json:"protocol"`
	Strategy string `json:"strategy" validate:"required,oneof=accept drop"`

	Description string `json:"description"`
}

type AddrRuleOperate struct {
	ID        uint   `json:"id"`
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

type IptablesOp struct {
	Name    string `json:"name" validate:"required,oneof=1PANEL_INPUT 1PANEL_OUTPUT 1PANEL_BASIC"`
	Operate string `json:"operate" validate:"required,oneof=init-base init-forward init-advance bind-base unbind-base bind unbind"`
}

type IptablesRuleOp struct {
	Operation   string `json:"operation" validate:"required,oneof=add remove"`
	ID          uint   `json:"id"`
	Chain       string `json:"chain" validate:"required,oneof=1PANEL_BASIC 1PANEL_BASIC_BEFORE 1PANEL_INPUT 1PANEL_OUTPUT"`
	Protocol    string `json:"protocol"`
	SrcIP       string `json:"srcIP"`
	SrcPort     uint   `json:"srcPort"`
	DstIP       string `json:"dstIP"`
	DstPort     uint   `json:"dstPort"`
	Strategy    string `json:"strategy" validate:"required,oneof=accept drop reject"`
	Description string `json:"description"`
}

type IptablesBatchOperate struct {
	Rules []IptablesRuleOp `json:"rules"`
}

type IptablesChainStatus struct {
	IsBind          bool   `json:"isBind"`
	DefaultStrategy string `json:"defaultStrategy"`
}
