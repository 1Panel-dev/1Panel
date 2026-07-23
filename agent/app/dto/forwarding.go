package dto

// ForwardRule preserves the existing firewall search response shape while
// keeping forwarding data separate from the filter client model.
type ForwardRule struct {
	ID       uint   `json:"id"`
	Chain    string `json:"chain"`
	Family   string `json:"family"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Strategy string `json:"strategy"`

	Num        string `json:"num"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort"`
	Interface  string `json:"interface"`

	UsedStatus  string `json:"usedStatus"`
	Description string `json:"description"`
}

type ForwardRuleOperate struct {
	ForceDelete bool                   `json:"forceDelete"`
	Rules       []ForwardRuleOperation `json:"rules"`
}

type ForwardRuleOperation struct {
	Operation  string `json:"operation" validate:"required,oneof=add remove"`
	Num        string `json:"num"`
	Protocol   string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
	Interface  string `json:"interface"`
	Port       string `json:"port" validate:"required"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort" validate:"required"`
}
