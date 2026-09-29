package docker_guard

import (
	"github.com/1Panel-dev/1Panel/agent/buserr"

	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

const (
	Chain                    = "1PANEL_DOCKER"
	DockerChain              = "DOCKER-USER"
	FamilyIPv4               = constant.FirewallFamilyIPv4
	FamilyIPv6               = constant.FirewallFamilyIPv6
	ModeSources              = "deny_sources"
	ModeAllow                = "allow_sources"
	ModeAll                  = "deny_all"
	StatusEffective          = "effective"
	StatusDisabled           = "disabled"
	StatusNotEffective       = "not_effective"
	ReasonCommandMissing     = "command_missing"
	ReasonDockerChainMissing = "docker_chain_missing"
	ReasonGuardChainMissing  = "guard_chain_missing"
	ReasonJumpMissing        = "jump_missing"
	ReasonJumpNotFirst       = "jump_not_first"
	ReasonJumpDuplicate      = "jump_duplicate"
	ReasonInspectFailed      = "inspect_failed"
)

type FamilyError struct {
	Family string
	Err    error
}

func (e *FamilyError) Error() string { return fmt.Sprintf("%s Docker port guard: %v", e.Family, e.Err) }
func (e *FamilyError) Unwrap() error { return e.Err }

type ProxyEndpoint struct {
	Protocol string
	HostIP   string
	HostPort uint16
}

type ProxyEndpoints struct {
	Items     []ProxyEndpoint
	Inspected bool
}

type DNATRules struct {
	Output    string
	Inspected bool
}

type Policy struct {
	UUID     string
	Family   string
	HostIP   string
	HostPort uint16
	Protocol string
	Mode     string
	Sources  []string
}

type FamilyStatus struct {
	State       string
	Reason      string
	Initialized bool
	Bound       bool
	Effective   bool
}

type NativeRule struct {
	Family string   `json:"family"`
	Order  int64    `json:"order"`
	Tokens []string `json:"tokens"`
}

type ReadOnlyPolicy struct {
	Policy      Policy
	Action      string
	Sequence    int64
	NativeRules []NativeRule
}

type PolicyInventory struct {
	Policies          []Policy
	ReadOnly          []ReadOnlyPolicy
	ManagedRuleOrders map[string][]int64
}

type Runtime interface {
	Initialize([]Policy, PolicyInventory) error
	Bind() error
	ReplacePolicies([]Policy, PolicyInventory) error
	Unbind() error
	Cleanup() error
	Initialized(string) (bool, error)
	Status(string) FamilyStatus
	ListPolicies() (PolicyInventory, error)
}

const ipv4ForwardingPath = "/proc/sys/net/ipv4/ip_forward"

func CheckIPv4Forwarding() error {
	return checkIPv4Forwarding(os.ReadFile)
}

func checkIPv4Forwarding(readFile func(string) ([]byte, error)) error {
	value, err := readFile(ipv4ForwardingPath)
	if err != nil {
		return fmt.Errorf("inspect IPv4 forwarding: %w", err)
	}
	if strings.TrimSpace(string(value)) != "1" {
		return buserr.New("ErrDockerIPv4ForwardingDisabled")
	}
	return nil
}
