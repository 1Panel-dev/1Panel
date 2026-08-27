package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	filteriptables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/iptables"
	filternftables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/nftables"
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
)

type SnapshotPolicy func(context.Context, filter.Snapshot) (filter.Snapshot, error)

type Engine struct {
	adapter filter.Adapter
	policy  SnapshotPolicy
}

type Registry map[filter.Provider]*Engine

func NewRegistry(policy SnapshotPolicy) Registry {
	return Registry{
		filter.ProviderIptables:  New(filteriptables.NewAdapter(), policy),
		filter.ProviderNftables:  New(filternftables.NewAdapter(), policy),
		filter.ProviderFirewalld: New(filterfirewalld.NewAdapter(), policy),
		filter.ProviderUFW:       New(filterufw.NewAdapter(), policy),
	}
}

func New(adapter filter.Adapter, policy SnapshotPolicy) *Engine {
	return &Engine{adapter: adapter, policy: policy}
}

func (r Registry) Resolve(provider filter.Provider) (*Engine, error) {
	engine, exists := r[provider]
	if !exists || engine == nil || engine.adapter == nil {
		return nil, fmt.Errorf("%w: %s", filter.ErrAdapterUnavailable, provider)
	}
	return engine, nil
}

func (r Registry) Providers() []filter.Provider {
	providers := make([]filter.Provider, 0, len(r))
	for provider := range r {
		providers = append(providers, provider)
	}
	return providers
}

func (e *Engine) Provider() filter.Provider {
	if e == nil || e.adapter == nil {
		return ""
	}
	return e.adapter.Provider()
}

func (e *Engine) Observe(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	snapshot, err := e.adapter.Observe(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	if e.policy == nil {
		return snapshot, nil
	}
	return e.policy(ctx, snapshot)
}

func (e *Engine) ObserveScopes(ctx context.Context, scopes []filter.Scope) ([]filter.Snapshot, error) {
	observer, ok := e.adapter.(filter.MultiScopeObserver)
	if !ok {
		return nil, fmt.Errorf("%w: %s multi-scope inventory", filter.ErrAdapterUnavailable, e.adapter.Provider())
	}
	snapshots, err := observer.ObserveScopes(ctx, scopes)
	if err != nil {
		return nil, err
	}
	if e.policy == nil {
		return snapshots, nil
	}
	for index := range snapshots {
		snapshots[index], err = e.policy(ctx, snapshots[index])
		if err != nil {
			return nil, err
		}
	}
	return snapshots, nil
}

func (e *Engine) ObserveMutation(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	snapshot, err := e.Observe(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	for _, notice := range snapshot.Notices {
		if notice.Code == filter.ScopeNoticeManagedScopeInactive || notice.Code == filter.ScopeNoticeManagedScopeMissing {
			return filter.Snapshot{}, fmt.Errorf("%w: managed firewall scope is unavailable", filter.ErrProviderUnavailable)
		}
	}
	return snapshot, nil
}

func (e *Engine) Prepare(rule filter.FirewallRule) (filter.FirewallRule, error) {
	preparer, ok := e.adapter.(filter.RulePreparer)
	if !ok {
		return rule, nil
	}
	return preparer.PrepareRule(rule)
}

func (e *Engine) CheckRule(ctx context.Context, rule filter.FirewallRule) error {
	checker, ok := e.adapter.(filter.RuleChecker)
	if !ok {
		return nil
	}
	return checker.CheckRule(ctx, rule)
}

func (e *Engine) CompileDesired(
	ctx context.Context,
	policyUUID string,
	origin filter.RuleOrigin,
	rules []filter.FirewallRule,
) ([]filter.DesiredRule, error) {
	result := make([]filter.DesiredRule, 0, len(rules))
	scopeOrdinals := make(map[string]int)
	for _, rule := range rules {
		prepared, err := e.Prepare(rule)
		if err != nil {
			return nil, err
		}
		if err = e.CheckRule(ctx, prepared); err != nil {
			return nil, err
		}
		ruleKey, err := filter.RuleKey(prepared)
		if err != nil {
			return nil, err
		}
		scopeKey := prepared.Scope.Key()
		ordinal := scopeOrdinals[scopeKey]
		scopeOrdinals[scopeKey] = ordinal + 1
		prepared.UUID = compiledRuleUUID(policyUUID, ruleKey, ordinal)
		result = append(result, filter.DesiredRule{
			UUID: policyUUID, Rule: prepared, RuleKey: ruleKey, Origin: origin,
			Marker: "1panel-rule:" + prepared.UUID,
		})
	}
	return result, nil
}

func (e *Engine) ValidatePosition(
	ctx context.Context,
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
	target int64,
) error {
	if rule.Scope.Provider == filter.ProviderUFW {
		minimum, maximum := positionBounds(snapshot)
		if target < minimum || target > maximum {
			return fmt.Errorf(
				"%w: target position %d is outside the %s range %d-%d",
				filter.ErrInvalidRule, target, rule.Scope.Family, minimum, maximum,
			)
		}
		return nil
	}
	maximum, err := e.MaxPosition(ctx, snapshot, rule)
	if err != nil {
		return err
	}
	if target > maximum {
		return fmt.Errorf("%w: target position %d is out of range 1-%d", filter.ErrInvalidRule, target, maximum)
	}
	return nil
}

func (e *Engine) AppendPosition(ctx context.Context, snapshot filter.Snapshot, rule filter.FirewallRule) (int64, error) {
	if rule.Scope.Family == filter.FamilyIPv4 {
		return snapshotMaxPosition(snapshot) + 1, nil
	}
	maximum, err := e.MaxPosition(ctx, snapshot, rule)
	if err != nil {
		return 0, err
	}
	return maximum + 1, nil
}

func (e *Engine) MaxPosition(
	ctx context.Context,
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
) (int64, error) {
	maximum := snapshotMaxPosition(snapshot)
	if rule.Scope.Provider != filter.ProviderUFW {
		return maximum, nil
	}
	relatedScope := rule.Scope
	if relatedScope.Family == filter.FamilyIPv4 {
		relatedScope.Family = filter.FamilyIPv6
	} else {
		relatedScope.Family = filter.FamilyIPv4
	}
	relatedSnapshot, err := e.ObserveMutation(ctx, relatedScope)
	if err != nil {
		return 0, err
	}
	if relatedMaximum := snapshotMaxPosition(relatedSnapshot); relatedMaximum > maximum {
		maximum = relatedMaximum
	}
	return maximum, nil
}

func (e *Engine) NativeDetail(ctx context.Context, name string, permanent bool) (string, error) {
	reader, ok := e.adapter.(filter.NativeDetailReader)
	if !ok {
		return "", fmt.Errorf("%w: native details for %s", filter.ErrAdapterUnavailable, e.Provider())
	}
	return reader.NativeDetail(ctx, name, permanent)
}

func (e *Engine) Capabilities(ctx context.Context) (filter.Capabilities, error) {
	return e.adapter.Capabilities(ctx)
}

func (e *Engine) Execute(ctx context.Context, snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, filter.VerifyResult, error) {
	plan, err := e.adapter.Compile(snapshot, changes)
	if err != nil {
		return filter.BackendPlan{}, filter.VerifyResult{}, err
	}
	result, err := e.adapter.Apply(ctx, plan)
	if err != nil {
		return plan, filter.VerifyResult{}, err
	}
	if result.Verification != nil {
		if !result.Verification.Matched {
			if rollbackErr := e.Rollback(ctx, plan); rollbackErr != nil {
				return plan, *result.Verification, errors.Join(filter.ErrVerificationFailed, rollbackErr)
			}
		}
		return plan, *result.Verification, nil
	}
	verification, err := e.adapter.Verify(ctx, plan)
	if err != nil {
		return plan, verification, e.rollback(ctx, plan, err)
	}
	if !verification.Matched {
		if rollbackErr := e.Rollback(ctx, plan); rollbackErr != nil {
			return plan, verification, errors.Join(filter.ErrVerificationFailed, rollbackErr)
		}
	}
	return plan, verification, nil
}

func (e *Engine) Rollback(ctx context.Context, plan filter.BackendPlan) error {
	rollbacker, ok := e.adapter.(filter.PlanRollbacker)
	if !ok {
		return fmt.Errorf("%w: provider %s does not support applied-plan rollback", filter.ErrAdapterUnavailable, e.adapter.Provider())
	}
	return rollbacker.Rollback(ctx, plan)
}

func (e *Engine) rollback(ctx context.Context, plan filter.BackendPlan, cause error) error {
	if err := e.Rollback(ctx, plan); err != nil {
		return errors.Join(cause, fmt.Errorf("rollback applied firewall plan: %w", err))
	}
	return cause
}

func positionBounds(snapshot filter.Snapshot) (int64, int64) {
	minimum, maximum := int64(0), int64(0)
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position == nil {
			continue
		}
		position := int64(*observed.Locator.Position)
		if minimum == 0 || position < minimum {
			minimum = position
		}
		if position > maximum {
			maximum = position
		}
	}
	return minimum, maximum
}

func snapshotMaxPosition(snapshot filter.Snapshot) int64 {
	var maximum int64
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position != nil && int64(*observed.Locator.Position) > maximum {
			maximum = int64(*observed.Locator.Position)
		}
	}
	return maximum
}

func compiledRuleUUID(policyUUID, ruleKey string, scopeOrdinal int) string {
	if scopeOrdinal == 0 {
		return policyUUID
	}
	const suffixLength = 12
	if len(ruleKey) > suffixLength {
		ruleKey = ruleKey[:suffixLength]
	}
	return fmt.Sprintf("%s-%d-%s", policyUUID, scopeOrdinal+1, ruleKey)
}
