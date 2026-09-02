package docker_guard

import (
	"fmt"
	"slices"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

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
	Initialize([]Policy) error
	Bind() error
	Reconcile([]Policy) error
	Unbind() error
	Cleanup() error
	Initialized(string) (bool, error)
	Status(string) FamilyStatus
	ListPolicies() (PolicyInventory, error)
}

func NewRuntime(provider string) Runtime {
	if provider == constant.FirewallProviderNftables {
		return NewNftablesManager()
	}
	return NewManager()
}

func Verify(runtime Runtime, desired []Policy, preserved []ReadOnlyPolicy) error {
	inventory, err := runtime.ListPolicies()
	if err != nil {
		return fmt.Errorf("verify synchronized Docker firewall policies: %w", err)
	}
	if !PolicyStatesEqual(inventory.Policies, desired) {
		return fmt.Errorf("verify synchronized Docker firewall policies: target policies do not match the database")
	}
	if !readOnlyStatesEqual(inventory.ReadOnly, preserved) {
		return fmt.Errorf("verify synchronized Docker firewall policies: read-only runtime rules changed")
	}
	return nil
}

func readOnlyStatesEqual(left, right []ReadOnlyPolicy) bool {
	if len(left) != len(right) {
		return false
	}
	leftRules := flattenNativeRules(left)
	rightRules := flattenNativeRules(right)
	if len(leftRules) != len(rightRules) {
		return false
	}
	for index := range leftRules {
		if leftRules[index].Family != rightRules[index].Family || !slices.Equal(leftRules[index].Tokens, rightRules[index].Tokens) {
			return false
		}
	}
	return true
}

func flattenNativeRules(policies []ReadOnlyPolicy) []NativeRule {
	rules := make([]NativeRule, 0)
	for _, policy := range policies {
		rules = append(rules, policy.NativeRules...)
	}
	slices.SortStableFunc(rules, func(left, right NativeRule) int {
		if left.Family < right.Family {
			return -1
		}
		if left.Family > right.Family {
			return 1
		}
		if left.Order < right.Order {
			return -1
		}
		if left.Order > right.Order {
			return 1
		}
		return 0
	})
	return rules
}

func ReconcileTarget(backend string, policies []Policy, runtime Runtime) error {
	families := make(map[string]struct{}, len(policies))
	needsInitialize, needsBind := false, false
	for _, policy := range policies {
		families[policy.Family] = struct{}{}
	}
	if len(families) == 0 {
		initialized := false
		for _, family := range []string{FamilyIPv4, FamilyIPv6} {
			status := runtime.Status(family)
			if status.Reason == ReasonInspectFailed {
				return fmt.Errorf("inspect Docker firewall target %s for %s failed", backend, family)
			}
			initialized = initialized || status.Initialized
		}
		if initialized {
			return runtime.Reconcile(nil)
		}
		return nil
	}
	for family := range families {
		status := runtime.Status(family)
		needsInitialize = needsInitialize || !status.Initialized
		needsBind = needsBind || !status.Bound || !status.Effective
	}
	var err error
	if needsInitialize {
		err = runtime.Initialize(policies)
	} else {
		if needsBind {
			err = runtime.Bind()
		}
		if err == nil {
			err = runtime.Reconcile(policies)
		}
	}
	if err != nil {
		return err
	}
	for family := range families {
		if !runtime.Status(family).Effective {
			return fmt.Errorf("Docker firewall target %s is not effective for %s", backend, family)
		}
	}
	return nil
}
