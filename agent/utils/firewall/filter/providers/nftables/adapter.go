package nftables

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/mattn/go-shellwords"
)

type Backend interface {
	ListChain(context.Context, filter.Scope) (string, error)
	Run(context.Context, filter.NativeCommand) error
	Save(context.Context) error
}

type Adapter struct{ backend Backend }

func NewAdapter() *Adapter { return &Adapter{backend: systemBackend{}} }

func NewAdapterWithBackend(backend Backend) *Adapter { return &Adapter{backend: backend} }

func (a *Adapter) Provider() filter.Provider { return filter.ProviderNftables }

func (a *Adapter) Capabilities(context.Context) (filter.Capabilities, error) {
	return filter.Capabilities{
		Scopes: []filter.ScopePattern{{
			Provider: filter.ProviderNftables, Families: []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}, Table: "filter",
			Chains: []string{"1PANEL_BASIC_BEFORE", "1PANEL_BASIC", "1PANEL_BASIC_AFTER"}, Directions: []filter.Direction{filter.DirectionInput},
		}}, Marker: true, AtomicApply: false, TransactionalRollback: true, OwnedChains: true, ExplicitPosition: true,
	}, nil
}

func (a *Adapter) Observe(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	scope = scope.Normalize()
	if err := validateScope(scope); err != nil {
		return filter.Snapshot{}, err
	}
	if a.backend == nil {
		return filter.Snapshot{}, fmt.Errorf("nftables backend is required")
	}
	output, err := a.backend.ListChain(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	return filter.NewSnapshot(scope, parseChain(scope, output))
}

func (a *Adapter) Compile(snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, error) {
	if snapshot.Revision == "" {
		return filter.BackendPlan{}, filter.ErrRuleStale
	}
	if err := validateScope(snapshot.Scope); err != nil {
		return filter.BackendPlan{}, err
	}
	if len(changes) != 1 {
		return filter.BackendPlan{}, fmt.Errorf("%w: nftables plans currently require exactly one change", filter.ErrInvalidRule)
	}
	rules, expected, previous, err := applyChange(snapshot, changes[0])
	if err != nil {
		return filter.BackendPlan{}, err
	}
	chain := nativeChainName(snapshot.Scope)
	commands := rebuildCommands(snapshot.Scope, chain, rules)
	rollbackCommands := rebuildCommands(snapshot.Scope, chain, snapshot.Rules)
	plan := filter.BackendPlan{Provider: filter.ProviderNftables, Scope: snapshot.Scope, SnapshotRevision: snapshot.Revision}
	plan.Rules = []filter.NativeRulePlan{{
		RuleUUID: ruleUUID(changes[0]), Operation: changes[0].Operation, Previous: previous, Expected: expected,
		Commands:         commands,
		RollbackCommands: rollbackCommands,
	}}
	return plan, nil
}

func (a *Adapter) Apply(ctx context.Context, plan filter.BackendPlan) (filter.ApplyResult, error) {
	if err := validatePlan(plan); err != nil {
		return filter.ApplyResult{}, err
	}
	for _, command := range plan.Rules[0].Commands {
		if err := validateNativeCommand(command); err != nil {
			return filter.ApplyResult{}, err
		}
	}
	for _, command := range plan.Rules[0].Commands {
		if err := a.backend.Run(ctx, command); err != nil {
			return filter.ApplyResult{}, a.compensate(ctx, plan, err)
		}
	}
	if err := a.backend.Save(ctx); err != nil {
		return filter.ApplyResult{}, a.compensate(ctx, plan, err)
	}
	return filter.ApplyResult{Applied: []filter.ObservedRule{plan.Rules[0].Expected}}, nil
}

func (a *Adapter) Verify(ctx context.Context, plan filter.BackendPlan) (filter.VerifyResult, error) {
	if err := validatePlan(plan); err != nil {
		return filter.VerifyResult{}, err
	}
	snapshot, err := a.Observe(ctx, plan.Scope)
	if err != nil {
		return filter.VerifyResult{}, err
	}
	expected := plan.Rules[0]
	matches := 0
	for _, observed := range snapshot.Rules {
		if observed.Marker == expected.Expected.Marker {
			if expected.Operation == filter.ChangeDelete {
				matches++
				continue
			}
			want, wantErr := filter.RuleKey(expected.Expected.Rule)
			got, gotErr := filter.RuleKey(observed.Rule)
			if wantErr == nil && gotErr == nil && want == got {
				matches++
			}
		}
	}
	matched := matches == 1
	if expected.Operation == filter.ChangeDelete {
		matched = matches == 0
	}
	return filter.VerifyResult{Snapshot: snapshot, Matched: matched}, nil
}

func (a *Adapter) Rollback(ctx context.Context, plan filter.BackendPlan) error {
	if err := validatePlan(plan); err != nil {
		return err
	}
	for _, command := range plan.Rules[0].RollbackCommands {
		if err := validateNativeCommand(command); err != nil {
			return err
		}
	}
	for _, command := range plan.Rules[0].RollbackCommands {
		if err := a.backend.Run(ctx, command); err != nil {
			return err
		}
	}
	return a.backend.Save(ctx)
}

func (a *Adapter) compensate(ctx context.Context, plan filter.BackendPlan, cause error) error {
	if err := a.Rollback(ctx, plan); err != nil {
		return fmt.Errorf("nftables apply failed: %w; compensation failed: %v", cause, err)
	}
	return cause
}

func validateScope(scope filter.Scope) error {
	scope = scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return err
	}
	if scope.Provider != filter.ProviderNftables {
		return fmt.Errorf("%w: nftables adapter scope %s", filter.ErrUnsupportedScope, scope.Key())
	}
	return nil
}

func nativeChainName(scope filter.Scope) string {
	switch scope.Normalize().Chain {
	case "1PANEL_BASIC_BEFORE":
		return nftables_helper.BasicBeforeChain
	case "1PANEL_BASIC":
		return nftables_helper.BasicChain
	case "1PANEL_BASIC_AFTER":
		return nftables_helper.BasicAfterChain
	default:
		return ""
	}
}

func validatePlan(plan filter.BackendPlan) error {
	if plan.Provider != filter.ProviderNftables || len(plan.Rules) != 1 {
		return fmt.Errorf("%w: invalid nftables plan", filter.ErrInvalidRule)
	}
	return validateScope(plan.Scope)
}

func validateNativeCommand(command filter.NativeCommand) error {
	if command.Executable != "nft" || len(command.Args) == 0 {
		return fmt.Errorf("%w: invalid nftables native command", filter.ErrInvalidRule)
	}
	return nil
}

func applyChange(snapshot filter.Snapshot, change filter.DesiredChange) ([]filter.ObservedRule, filter.ObservedRule, *filter.ObservedRule, error) {
	rules := append([]filter.ObservedRule(nil), snapshot.Rules...)
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	if rule == nil {
		return nil, filter.ObservedRule{}, nil, fmt.Errorf("%w: %s rule is required", filter.ErrInvalidRule, change.Operation)
	}
	normalized, err := filter.NormalizeRule(*rule)
	if err != nil {
		return nil, filter.ObservedRule{}, nil, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Key() || normalized.UUID == "" {
		return nil, filter.ObservedRule{}, nil, fmt.Errorf("%w: invalid nftables mutation rule", filter.ErrInvalidRule)
	}
	if (change.Operation == filter.ChangeCreate || change.Operation == filter.ChangeUpdate) && broadDeny(normalized) {
		return nil, filter.ObservedRule{}, nil, filter.ErrLockoutRisk
	}
	position := len(rules) + 1
	marker := "1panel-rule:" + normalized.UUID
	var previous *filter.ObservedRule

	if change.Operation != filter.ChangeCreate {
		if change.Locator == nil || change.Locator.Position == nil {
			return nil, filter.ObservedRule{}, nil, fmt.Errorf("%w: mutation requires a position locator", filter.ErrInvalidRule)
		}
		position = *change.Locator.Position
		if position < 1 || position > len(rules) {
			return nil, filter.ObservedRule{}, nil, filter.ErrRuleStale
		}
		selected := rules[position-1]
		if selected.Protected {
			return nil, filter.ObservedRule{}, nil, filter.ErrProtectedRule
		}
		previousCopy := selected
		previous = &previousCopy
		rules = append(rules[:position-1], rules[position:]...)
	}

	target := position
	if normalized.OrderIndex != nil {
		target = int(*normalized.OrderIndex)
	}
	if change.Operation == filter.ChangeCreate && normalized.OrderIndex == nil {
		target = len(rules) + 1
	}
	if change.Operation == filter.ChangeDelete {
		expected := observedRule(normalized, marker, position, "")
		return rules, expected, previous, nil
	}
	if target < 1 || target > len(rules)+1 {
		return nil, filter.ObservedRule{}, nil, fmt.Errorf("%w: target position is out of range", filter.ErrInvalidRule)
	}
	if change.Operation == filter.ChangeReorder || change.Operation == filter.ChangeUpdate {
		start, end := position, target
		if start > end {
			start, end = end, start
		}
		for index := start; index <= end && index <= len(snapshot.Rules); index++ {
			candidate := snapshot.Rules[index-1]
			if candidate.Protected || candidate.ParseStatus == filter.ParseStatusOpaque || candidate.Marker == "" {
				return nil, filter.ObservedRule{}, nil, fmt.Errorf("%w: reorder cannot cross external or opaque rules", filter.ErrUnsupportedScope)
			}
		}
	}
	expected := observedRule(normalized, marker, target, strings.Join(compileExpressionArgs(normalized, marker), " "))
	rules = append(rules, filter.ObservedRule{})
	copy(rules[target:], rules[target-1:])
	rules[target-1] = expected
	for index := range rules {
		position := index + 1
		rules[index].Locator.Position = &position
	}
	return rules, expected, previous, nil
}

func observedRule(rule filter.FirewallRule, marker string, position int, raw string) filter.ObservedRule {
	return filter.ObservedRule{
		Rule: rule, Marker: marker, Raw: raw, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: filter.ProviderNftables, ScopeKey: rule.Scope.Key(), Position: &position},
	}
}

func ruleUUID(change filter.DesiredChange) string {
	if change.After != nil {
		return change.After.UUID
	}
	if change.Before != nil {
		return change.Before.UUID
	}
	return ""
}

func broadDeny(rule filter.FirewallRule) bool {
	return (rule.Action == filter.ActionDrop || rule.Action == filter.ActionReject) && rule.SourceAddress == "" &&
		rule.DestinationAddress == "" && rule.SourcePort == "" && rule.DestinationPort == ""
}

func rebuildCommands(scope filter.Scope, chain string, rules []filter.ObservedRule) []filter.NativeCommand {
	tableFamily := nftables_helper.TableFamily(scope.Family)
	commands := []filter.NativeCommand{{
		Executable: "nft", Args: []string{"flush", "chain", tableFamily, nftables_helper.TableName, chain},
	}}
	for _, rule := range rules {
		raw := strings.TrimSpace(rule.Raw)
		if rule.ParseStatus == filter.ParseStatusSupported && rule.Marker != "" {
			args := []string{"add", "rule", tableFamily, nftables_helper.TableName, chain}
			args = append(args, compileExpressionArgs(rule.Rule, rule.Marker)...)
			commands = append(commands, filter.NativeCommand{Executable: "nft", Args: args})
			continue
		}
		if raw != "" {
			commands = append(commands, filter.NativeCommand{
				Executable: "nft", Args: []string{"add", "rule", tableFamily, nftables_helper.TableName, chain, raw},
			})
		}
	}
	return commands
}

func compileExpressionArgs(rule filter.FirewallRule, marker string) []string {
	parts := make([]string, 0, 24)
	if rule.Protocol != "all" {
		protocol := rule.Protocol
		if protocol == "icmpv6" {
			protocol = "ipv6-icmp"
		}
		parts = append(parts, "meta", "l4proto", protocol)
	}
	prefix := "ip"
	if rule.Scope.Family == filter.FamilyIPv6 {
		prefix = "ip6"
	}
	if rule.SourceAddress != "" {
		parts = append(parts, prefix, "saddr", rule.SourceAddress)
	}
	if rule.DestinationAddress != "" {
		parts = append(parts, prefix, "daddr", rule.DestinationAddress)
	}
	if rule.Interface != "" {
		parts = append(parts, "iifname", strconv.Quote(rule.Interface))
	}
	if rule.SourcePort != "" {
		parts = append(parts, rule.Protocol, "sport", rule.SourcePort)
	}
	if rule.DestinationPort != "" {
		parts = append(parts, rule.Protocol, "dport", rule.DestinationPort)
	}
	if len(rule.ConnectionStates) != 0 {
		parts = append(parts, "ct", "state", "{", strings.Join(rule.ConnectionStates, ","), "}")
	}
	parts = append(parts, string(rule.Action), "comment", strconv.Quote(marker))
	return parts
}

func parseChain(scope filter.Scope, output string) []filter.ObservedRule {
	rules := make([]filter.ObservedRule, 0)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		handleIndex := strings.LastIndex(line, "# handle ")
		if handleIndex < 0 {
			continue
		}
		raw := strings.TrimSpace(line[:handleIndex])
		handle := strings.TrimSpace(line[handleIndex+len("# handle "):])
		position := len(rules) + 1
		rules = append(rules, parseRule(scope, raw, handle, position))
	}
	return rules
}

func parseRule(scope filter.Scope, raw, handle string, position int) filter.ObservedRule {
	locator := filter.Locator{Provider: filter.ProviderNftables, ScopeKey: scope.Key(), NativeID: handle, Canonical: raw, Position: &position}
	opaque := func() filter.ObservedRule {
		return filter.ObservedRule{Rule: filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindOpaque}, Locator: locator, ParseStatus: filter.ParseStatusOpaque, Raw: raw}
	}
	tokens, err := shellwords.Parse(raw)
	if err != nil {
		return opaque()
	}
	rule := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "all"}
	marker := ""
	for index := 0; index < len(tokens); {
		switch tokens[index] {
		case "meta":
			if index+2 >= len(tokens) {
				return opaque()
			}
			if tokens[index+1] == "nfproto" {
				want := "ipv4"
				if scope.Family == filter.FamilyIPv6 {
					want = "ipv6"
				}
				if tokens[index+2] != want {
					return opaque()
				}
			} else if tokens[index+1] == "l4proto" {
				rule.Protocol = tokens[index+2]
				if rule.Protocol == "ipv6-icmp" {
					rule.Protocol = "icmpv6"
				}
			} else {
				return opaque()
			}
			index += 3
		case "ip", "ip6":
			if index+2 >= len(tokens) {
				return opaque()
			}
			if tokens[index+1] == "saddr" {
				rule.SourceAddress = tokens[index+2]
			} else if tokens[index+1] == "daddr" {
				rule.DestinationAddress = tokens[index+2]
			} else {
				return opaque()
			}
			index += 3
		case "tcp", "udp":
			if index+2 >= len(tokens) {
				return opaque()
			}
			if rule.Protocol == "all" {
				rule.Protocol = tokens[index]
			}
			value := strings.ReplaceAll(tokens[index+2], ":", "-")
			if tokens[index+1] == "sport" {
				rule.SourcePort = value
			} else if tokens[index+1] == "dport" {
				rule.DestinationPort = value
			} else {
				return opaque()
			}
			index += 3
		case "iifname":
			if index+1 >= len(tokens) {
				return opaque()
			}
			rule.Interface = tokens[index+1]
			index += 2
		case "ct":
			if index+2 >= len(tokens) || tokens[index+1] != "state" {
				return opaque()
			}
			index += 2
			if tokens[index] != "{" {
				for _, state := range strings.Split(tokens[index], ",") {
					if state = strings.TrimSpace(state); state != "" {
						rule.ConnectionStates = append(rule.ConnectionStates, state)
					}
				}
				index++
				continue
			}
			index++
			for index < len(tokens) && tokens[index] != "}" {
				state := strings.Trim(tokens[index], ",")
				if state != "" {
					rule.ConnectionStates = append(rule.ConnectionStates, state)
				}
				index++
			}
			if index < len(tokens) && tokens[index] == "}" {
				index++
			}
		case "accept", "drop", "reject":
			rule.Action = filter.Action(tokens[index])
			index++
		case "comment":
			if index+1 >= len(tokens) {
				return opaque()
			}
			if strings.HasPrefix(tokens[index+1], "1panel-rule:") {
				marker = tokens[index+1]
			} else {
				rule.Description = tokens[index+1]
			}
			index += 2
		case "counter":
			index++
			if index+3 < len(tokens) && tokens[index] == "packets" {
				index += 4
			}
		default:
			return opaque()
		}
	}
	if rule.Action == "" {
		return opaque()
	}
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return opaque()
	}
	return filter.ObservedRule{Rule: normalized, Locator: locator, Marker: marker, ParseStatus: filter.ParseStatusSupported, Raw: raw, Protected: scope.Chain != "1PANEL_BASIC"}
}

type systemBackend struct{}

func (systemBackend) ListChain(ctx context.Context, scope filter.Scope) (string, error) {
	return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout(
		"nft", "-n", "-a", "list", "chain", nftables_helper.TableFamily(scope.Family), nftables_helper.TableName, nativeChainName(scope),
	)
}

func (systemBackend) Run(ctx context.Context, command filter.NativeCommand) error {
	if command.Executable != "nft" {
		return fmt.Errorf("unexpected nftables executable %q", command.Executable)
	}
	return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).RunWithOptionalSudo(command.Executable, command.Args...)
}

func (systemBackend) Save(ctx context.Context) error {
	return nftables_helper.PersistRuleset(ctx)
}
