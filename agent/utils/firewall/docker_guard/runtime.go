package docker_guard

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

type Runtime interface {
	Initialize([]Policy) error
	Bind() error
	Reconcile([]Policy) error
	Unbind() error
	Cleanup() error
	Initialized(string) (bool, error)
	Status(string) FamilyStatus
	ListPolicies() ([]Policy, error)
}

func NewRuntime(provider string) Runtime {
	if provider == constant.FirewallProviderNftables {
		return NewNftablesManager()
	}
	return NewManager()
}

func Verify(runtime Runtime, desired []Policy) error {
	actual, err := runtime.ListPolicies()
	if err != nil {
		return fmt.Errorf("verify synchronized Docker firewall policies: %w", err)
	}
	if !PolicyStatesEqual(actual, desired) {
		return fmt.Errorf("verify synchronized Docker firewall policies: target policies do not match the database")
	}
	return nil
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
