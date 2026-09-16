package runtime

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func (e *Engine) NewCreatePlanner(snapshot filter.Snapshot) (filter.CreatePlanner, error) {
	factory, ok := e.adapter.(filter.CreatePlannerFactory)
	if !ok {
		return nil, filter.ErrAdapterUnavailable
	}
	return factory.NewCreatePlanner(snapshot), nil
}

func (e *Engine) ExecutePlannedCreate(ctx context.Context, planner filter.CreatePlanner, change filter.DesiredChange) (filter.ObservedRule, error) {
	if err := ctx.Err(); err != nil {
		return filter.ObservedRule{}, err
	}
	plan, err := planner.Compile(change)
	if err != nil {
		return filter.ObservedRule{}, err
	}
	if !plan.CreatesOnly() {
		return filter.ObservedRule{}, filter.ErrInvalidRule
	}
	InvalidateInventory()
	defer InvalidateInventory()
	plan.CommandOnly = change.CommandOnly
	result, err := e.adapter.Apply(ctx, plan)
	if err != nil {
		return filter.ObservedRule{}, err
	}
	if len(result.Applied) != 1 {
		return filter.ObservedRule{}, filter.ErrVerificationFailed
	}
	planner.Applied(result.Applied[0])
	return result.Applied[0], nil
}
