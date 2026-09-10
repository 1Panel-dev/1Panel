package dto

type ForwardRuleSearch struct {
	PageInfo
	All      bool   `json:"all,omitempty"`
	Info     string `json:"info"`
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
}

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

	IsDesired  bool   `json:"isDesired"`
	IsRuntime  bool   `json:"isRuntime"`
	SyncStatus string `json:"syncStatus"`
}

type ForwardRuleOperate struct {
	ForceDelete bool                   `json:"forceDelete"`
	Rules       []ForwardRuleOperation `json:"rules" validate:"required,min=1,dive"`
}

type ForwardRuleOperation struct {
	Operation  string `json:"operation" validate:"required,oneof=add remove"`
	Num        string `json:"num"`
	Family     string `json:"family" validate:"omitempty,oneof=ipv4 ipv6"`
	Protocol   string `json:"protocol" validate:"required,oneof=tcp udp tcp/udp"`
	Interface  string `json:"interface"`
	Port       string `json:"port" validate:"required"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort" validate:"required"`
}
