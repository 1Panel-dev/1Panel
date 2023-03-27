package dto

type RuleSearch struct {
	PageInfo
	Type string `json:"type" validate:"required"`
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
