package filter

import (
	"context"
	"errors"
)

var (
	ErrAdapterUnavailable   = errors.New("firewall rule adapter is unavailable")
	ErrInventoryUnavailable = errors.New("firewall rule inventory is unavailable")
	ErrFamilyUnavailable    = errors.New("firewall address family is unavailable")
	ErrRuleNotFound         = errors.New("firewall rule does not exist")
)

type ChangeOperation string

const (
	ChangeCreate  ChangeOperation = "create"
	ChangeUpdate  ChangeOperation = "update"
	ChangeDelete  ChangeOperation = "delete"
	ChangeReorder ChangeOperation = "reorder"
)

type RuleChange struct {
	Target         *ObservedRule   `json:"-"`
	Raw            string          `json:"raw,omitempty"`
	CommandOnly    bool            `json:"-"`
	Operation      ChangeOperation `json:"operation"`
	Before         *FirewallRule   `json:"before,omitempty"`
	After          *FirewallRule   `json:"after,omitempty"`
	Locator        *Locator        `json:"locator,omitempty"`
	PreviousMarker string          `json:"previousMarker,omitempty"`
	Append         bool            `json:"append,omitempty"`
	RestoreAtEnd   bool            `json:"restoreAtEnd,omitempty"`
}

type NativeCommand struct {
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
	Stdin      string   `json:"stdin,omitempty"`
}

type RuleCommands struct {
	RuleUUID         string          `json:"ruleUUID"`
	Operation        ChangeOperation `json:"operation"`
	Commands         []NativeCommand `json:"commands"`
	RollbackCommands []NativeCommand `json:"rollbackCommands,omitempty"`
	Previous         *ObservedRule   `json:"previous,omitempty"`
	Expected         ObservedRule    `json:"expected"`
}

type CommandBatch struct {
	CommandOnly bool           `json:"-"`
	Provider    Provider       `json:"provider"`
	Scope       Scope          `json:"scope"`
	Rules       []RuleCommands `json:"rules"`
}

func (p CommandBatch) CreatesOnly() bool {
	if len(p.Rules) == 0 {
		return false
	}
	for _, rule := range p.Rules {
		if rule.Operation != ChangeCreate {
			return false
		}
	}
	return true
}

type Adapter interface {
	Provider() Provider
	ListRules(context.Context, Scope) (RuleSet, error)
	BuildCommands(RuleSet, []RuleChange) (CommandBatch, error)
	RunCommands(context.Context, CommandBatch) error
	Rollback(context.Context, CommandBatch) error
}

type MultiScopeReader interface {
	ListRuleScopes(context.Context, []Scope) ([]RuleSet, error)
}

type RulePreparer interface {
	PrepareRule(FirewallRule) (FirewallRule, error)
}

type RuleChecker interface {
	CheckRule(context.Context, FirewallRule) error
}

type ExternalRuleAdapter interface {
	ListRulesByComment(context.Context, []Scope, string) ([]ObservedRule, error)
	AppendUnverified(context.Context, FirewallRule, string) error
}

type NativeDetailReader interface {
	NativeDetail(context.Context, string, bool) (string, error)
}

type RuleSaver interface {
	SaveRules(context.Context, Scope) error
}
