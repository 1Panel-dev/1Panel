package firewalld

import "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"

type createPlanner struct {
	adapter     *Adapter
	snapshot    filter.Snapshot
	byCanonical map[string][]filter.ObservedRule
}

func (a *Adapter) NewCreatePlanner(snapshot filter.Snapshot) filter.CreatePlanner {
	byCanonical := make(map[string][]filter.ObservedRule, len(snapshot.Rules))
	for _, observed := range snapshot.Rules {
		byCanonical[observed.Locator.Canonical] = append(byCanonical[observed.Locator.Canonical], observed)
	}
	snapshot.Rules = nil
	return &createPlanner{adapter: a, snapshot: snapshot, byCanonical: byCanonical}
}

func (p *createPlanner) Compile(change filter.DesiredChange) (filter.BackendPlan, error) {
	if change.Operation != filter.ChangeCreate || change.After == nil {
		return filter.BackendPlan{}, filter.ErrInvalidRule
	}
	rule, err := p.adapter.PrepareRule(*change.After)
	if err != nil {
		return filter.BackendPlan{}, err
	}
	snapshot := p.snapshot
	snapshot.Rules = p.byCanonical[nativeCanonical(rule)]
	return p.adapter.Compile(snapshot, []filter.DesiredChange{change})
}

func (p *createPlanner) Applied(rule filter.ObservedRule) {
	p.byCanonical[rule.Locator.Canonical] = []filter.ObservedRule{rule}
}
