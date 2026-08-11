package filter

import (
	"context"
	"errors"
)

var ErrAdapterUnavailable = errors.New("firewall rule adapter is unavailable")

type ChangeOperation string

const (
	ChangeCreate  ChangeOperation = "create"
	ChangeAdopt   ChangeOperation = "adopt"
	ChangeUpdate  ChangeOperation = "update"
	ChangeDelete  ChangeOperation = "delete"
	ChangeReorder ChangeOperation = "reorder"
)

type DesiredChange struct {
	Operation ChangeOperation `json:"operation"`
	Before    *FirewallRule   `json:"before,omitempty"`
	After     *FirewallRule   `json:"after,omitempty"`
	Locator   *Locator        `json:"locator,omitempty"`
}

type NativeCommand struct {
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
}

type NativeRulePlan struct {
	RuleUUID         string          `json:"ruleUUID"`
	Operation        ChangeOperation `json:"operation"`
	Commands         []NativeCommand `json:"commands"`
	RollbackCommands []NativeCommand `json:"rollbackCommands,omitempty"`
	Previous         *ObservedRule   `json:"previous,omitempty"`
	Expected         ObservedRule    `json:"expected"`
}

type BackendPlan struct {
	Provider         Provider         `json:"provider"`
	Scope            Scope            `json:"scope"`
	SnapshotRevision string           `json:"snapshotRevision"`
	Rules            []NativeRulePlan `json:"rules"`
}

type ApplyResult struct {
	Applied      []ObservedRule `json:"applied"`
	Verification *VerifyResult  `json:"verification,omitempty"`
}

type VerifyResult struct {
	Snapshot Snapshot `json:"snapshot"`
	Matched  bool     `json:"matched"`
}

type Adapter interface {
	Provider() Provider
	Capabilities(context.Context) (Capabilities, error)
	Observe(context.Context, Scope) (Snapshot, error)
	Compile(Snapshot, []DesiredChange) (BackendPlan, error)
	Apply(context.Context, BackendPlan) (ApplyResult, error)
	Verify(context.Context, BackendPlan) (VerifyResult, error)
}

type RulePreparer interface {
	PrepareRule(FirewallRule) (FirewallRule, error)
}

// NativeDetailReader loads provider-specific details for an opaque named
// object, such as a firewalld service or UFW application profile.
type NativeDetailReader interface {
	NativeDetail(context.Context, string, bool) (string, error)
}

// PlanRollbacker reverses a backend plan that was applied completely but
// could not be committed, for example after verification or DB persistence
// fails. Apply implementations remain responsible for compensating partially
// executed plans.
type PlanRollbacker interface {
	Rollback(context.Context, BackendPlan) error
}
