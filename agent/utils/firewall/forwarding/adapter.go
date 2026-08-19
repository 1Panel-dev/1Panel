package forwarding

import "github.com/1Panel-dev/1Panel/agent/constant"

const (
	FamilyIPv4 = constant.FirewallFamilyIPv4
	FamilyIPv6 = constant.FirewallFamilyIPv6

	ChainPreRouting  = "1PANEL_PREROUTING"
	ChainPostRouting = "1PANEL_POSTROUTING"
	ChainForward     = "1PANEL_FORWARD"

	ForwardFile     = "1panel_forward.rules"
	PreRoutingFile  = "1panel_forward_pre.rules"
	PostRoutingFile = "1panel_forward_post.rules"
)

type Rule struct {
	Num        string
	Family     string
	Protocol   string
	Port       string
	TargetIP   string
	TargetPort string
	Interface  string
}

type OperationType string

const (
	OperationAdd    OperationType = "add"
	OperationRemove OperationType = "remove"
)

type Adapter interface {
	Name() string
	List() ([]Rule, error)
	Reconcile(rules []Rule) error
	Enable() error
	Cleanup() error
	InitStatus() (bool, bool, error)
	FamilyStatus(family string) (bool, bool, error)
	Replay() error
}
