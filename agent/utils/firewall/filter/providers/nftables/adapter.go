package nftables

import (
	"context"
	"errors"
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
	filter.CommentRuleReader
	ListChain(context.Context, filter.Scope) (string, error)
	Run(context.Context, filter.NativeCommand) error
	Save(context.Context) error
}

type TableReader interface {
	ListTable(context.Context, filter.Scope) (string, bool, error)
}

type Adapter struct{ backend Backend }

func NewAdapter() *Adapter { return &Adapter{backend: systemBackend{}} }

func NewAdapterWithBackend(backend Backend) *Adapter { return &Adapter{backend: backend} }

func (a *Adapter) Provider() filter.Provider { return filter.ProviderNftables }

func (a *Adapter) Capabilities(context.Context) (filter.Capabilities, error) {
	return filter.Capabilities{
		Marker: true, OwnedChains: true, ExplicitPosition: true,
	}, nil
}

func (a *Adapter) AppendUnverified(ctx context.Context, rule filter.FirewallRule, comment string) error {
	rule, err := filter.NormalizeRule(rule)
	if err != nil {
		return err
	}
	if err := validateScope(rule.Scope); err != nil {
		return err
	}
	script := fmt.Sprintf("add rule %s %s %s %s\n", nftables_helper.TableFamily(rule.Scope.Family), nftables_helper.TableName, nativeChainName(rule.Scope), strings.Join(compileExpressionArgs(rule, comment), " "))
	return a.backend.Run(ctx, filter.NativeCommand{Executable: "nft", Stdin: script})
}

func (a *Adapter) ListRulesByComment(ctx context.Context, scopes []filter.Scope, comment string) ([]filter.ObservedRule, error) {
	var rules []filter.ObservedRule
	for _, scope := range scopes {
		scope = scope.Normalize()
		if err := validateScope(scope); err != nil {
			return nil, err
		}
		output, err := a.backend.ReadRulesByComment(ctx, scope, comment)
		if err != nil {
			return nil, err
		}
		rules = append(rules, parseChain(scope, output)...)
	}
	return rules, nil
}

func (a *Adapter) ListRules(ctx context.Context, scope filter.Scope) (filter.RuleSet, error) {
	scope = scope.Normalize()
	if err := validateScope(scope); err != nil {
		return filter.RuleSet{}, err
	}
	if a.backend == nil {
		return filter.RuleSet{}, fmt.Errorf("nftables backend is required")
	}
	output, err := a.backend.ListChain(ctx, scope)
	if errors.Is(err, nftables_helper.ErrChainNotFound) {
		snapshot, snapshotErr := filter.NewRuleSet(scope, nil)
		if snapshotErr != nil {
			return filter.RuleSet{}, snapshotErr
		}
		snapshot.Notices = []filter.ScopeNotice{{
			Code: filter.ScopeNoticeManagedScopeMissing, Values: []string{string(scope.Family), scope.Chain},
		}}
		return snapshot, nil
	}
	if err != nil {
		return filter.RuleSet{}, err
	}
	return filter.NewRuleSet(scope, parseChain(scope, output))
}

func (a *Adapter) ListRuleScopes(ctx context.Context, scopes []filter.Scope) ([]filter.RuleSet, error) {
	normalized := make([]filter.Scope, len(scopes))
	for index, scope := range scopes {
		scope = scope.Normalize()
		if err := validateScope(scope); err != nil {
			return nil, err
		}
		normalized[index] = scope
	}
	if a.backend == nil {
		return nil, fmt.Errorf("nftables backend is required")
	}
	snapshots := make([]filter.RuleSet, len(scopes))
	reader, readsTable := a.backend.(TableReader)
	for index, scope := range normalized {
		if snapshots[index].Scope.Provider != "" {
			continue
		}
		if !readsTable {
			snapshot, err := a.ListRules(ctx, scope)
			if err != nil {
				return nil, err
			}
			snapshots[index] = snapshot
			continue
		}
		output, _, err := reader.ListTable(ctx, scope)
		if err != nil {
			return nil, err
		}
		chains := nftables_helper.ParseTableChains(output)
		for target := index; target < len(normalized); target++ {
			current := normalized[target]
			if current.Family != scope.Family || current.Table != scope.Table {
				continue
			}
			chain, exists := chains[nativeChainName(current)]
			snapshot, err := filter.NewRuleSet(current, parseChain(current, chain))
			if err != nil {
				return nil, err
			}
			if !exists {
				snapshot.Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeMissing, Values: []string{string(current.Family), current.Chain}}}
			}
			snapshots[target] = snapshot
		}
	}
	return snapshots, nil
}

func (a *Adapter) BuildCommands(snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	if err := validateScope(snapshot.Scope); err != nil {
		return filter.CommandBatch{}, err
	}
	if len(changes) == 0 {
		return filter.CommandBatch{}, fmt.Errorf("%w: nftables plan requires at least one change", filter.ErrInvalidRule)
	}
	createOnly, deleteOnly := true, true
	for _, change := range changes {
		createOnly = createOnly && change.Operation == filter.ChangeCreate && change.CommandOnly
		deleteOnly = deleteOnly && change.Operation == filter.ChangeDelete && change.CommandOnly
	}
	if createOnly {
		return compileCreateBatch(snapshot, changes)
	}
	if deleteOnly {
		return compileDeleteBatch(snapshot, changes)
	}
	if len(changes) != 1 {
		return filter.CommandBatch{}, fmt.Errorf("%w: nftables mutation requires exactly one change", filter.ErrInvalidRule)
	}
	change := changes[0]
	if change.Operation != filter.ChangeUpdate && change.Operation != filter.ChangeAdopt && change.Operation != filter.ChangeReorder {
		return filter.CommandBatch{}, fmt.Errorf("%w: unsupported nftables mutation %s", filter.ErrInvalidRule, change.Operation)
	}
	expected, previous, err := compileChange(snapshot, change)
	if err != nil {
		return filter.CommandBatch{}, err
	}
	handle := previous.Locator.NativeID
	if _, err := strconv.ParseUint(handle, 10, 64); err != nil || handle != change.Locator.NativeID {
		return filter.CommandBatch{}, filter.ErrRuleStale
	}
	chain := strings.Join([]string{nftables_helper.TableFamily(snapshot.Scope.Family), nftables_helper.TableName, nativeChainName(snapshot.Scope)}, " ")
	rulePlan := filter.RuleCommands{RuleUUID: ruleUUID(change), Operation: change.Operation, Previous: previous, Expected: expected}
	target := *expected.Locator.Position
	var script string
	if target == *previous.Locator.Position {
		if change.Operation != filter.ChangeReorder {
			script = fmt.Sprintf("replace rule %s handle %s %s\n", chain, handle, expected.Raw)
			if strings.ContainsAny(previous.Raw, "\r\n") || previous.Raw == "" {
				return filter.CommandBatch{}, fmt.Errorf("%w: invalid native nftables rule", filter.ErrInvalidRule)
			}
			rulePlan.RollbackCommands = []filter.NativeCommand{{Executable: "nft", Stdin: fmt.Sprintf("replace rule %s handle %s %s\n", chain, handle, previous.Raw)}}
		}
	} else {
		script = fmt.Sprintf("delete rule %s handle %s\n", chain, handle)
		if target == len(snapshot.Rules) {
			script += fmt.Sprintf("add rule %s %s\n", chain, expected.Raw)
		} else {
			anchorIndex := target - 1
			if target > *previous.Locator.Position {
				anchorIndex++
			}
			anchor := snapshot.Rules[anchorIndex].Locator.NativeID
			if _, err := strconv.ParseUint(anchor, 10, 64); err != nil {
				return filter.CommandBatch{}, filter.ErrRuleStale
			}
			script += fmt.Sprintf("insert rule %s position %s %s\n", chain, anchor, expected.Raw)
		}
	}
	if script != "" {
		rulePlan.Commands = []filter.NativeCommand{{Executable: "nft", Stdin: script}}
	}
	return filter.CommandBatch{Provider: filter.ProviderNftables, Scope: snapshot.Scope, CommandOnly: change.CommandOnly, Rules: []filter.RuleCommands{rulePlan}}, nil
}

func compileDeleteBatch(snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	plan := filter.CommandBatch{Provider: filter.ProviderNftables, Scope: snapshot.Scope, CommandOnly: true}
	var script strings.Builder
	for _, change := range changes {
		expected, previous, err := compileChange(snapshot, change)
		if err != nil {
			return filter.CommandBatch{}, err
		}
		handle := previous.Locator.NativeID
		if _, err := strconv.ParseUint(handle, 10, 64); err != nil || change.Locator.NativeID != handle {
			return filter.CommandBatch{}, filter.ErrRuleStale
		}
		fmt.Fprintf(&script, "delete rule %s %s %s handle %s\n", nftables_helper.TableFamily(snapshot.Scope.Family), nftables_helper.TableName, nativeChainName(snapshot.Scope), handle)
		plan.Rules = append(plan.Rules, filter.RuleCommands{
			RuleUUID: ruleUUID(change), Operation: filter.ChangeDelete, Previous: previous, Expected: expected,
		})
	}
	plan.Rules[0].Commands = []filter.NativeCommand{{Executable: "nft", Stdin: script.String()}}
	return plan, nil
}

func compileCreateBatch(snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	plan := filter.CommandBatch{Provider: filter.ProviderNftables, Scope: snapshot.Scope, CommandOnly: true}
	var script strings.Builder
	for _, change := range changes {
		if change.After == nil {
			return filter.CommandBatch{}, fmt.Errorf("%w: create rule is required", filter.ErrInvalidRule)
		}
		rule, err := filter.NormalizeRule(*change.After)
		if err != nil {
			return filter.CommandBatch{}, err
		}
		if rule.Scope.Key() != snapshot.Scope.Key() || rule.UUID == "" {
			return filter.CommandBatch{}, fmt.Errorf("%w: invalid nftables creation rule", filter.ErrInvalidRule)
		}
		marker := "1panel-rule:" + rule.UUID
		verb := "add"
		if !change.Append {
			if rule.OrderIndex == nil || *rule.OrderIndex != 1 {
				return filter.CommandBatch{}, fmt.Errorf("%w: batch insertion requires the first position", filter.ErrInvalidRule)
			}
			verb = "insert"
		}
		fmt.Fprintf(&script, "%s rule %s %s %s %s\n", verb, nftables_helper.TableFamily(rule.Scope.Family), nftables_helper.TableName, nativeChainName(rule.Scope), strings.Join(compileExpressionArgs(rule, marker), " "))
		plan.Rules = append(plan.Rules, filter.RuleCommands{
			RuleUUID: rule.UUID, Operation: filter.ChangeCreate,
			Expected: filter.ObservedRule{Rule: rule, Marker: marker, ParseStatus: filter.ParseStatusSupported},
		})
	}
	plan.Rules[0].Commands = []filter.NativeCommand{{Executable: "nft", Stdin: script.String()}}
	return plan, nil
}

func (a *Adapter) RunCommands(ctx context.Context, plan filter.CommandBatch) error {
	if err := validatePlan(plan); err != nil {
		return err
	}
	for _, rulePlan := range plan.Rules {
		for _, command := range rulePlan.Commands {
			if err := validateNativeCommand(command); err != nil {
				return err
			}
		}
	}
	for _, rulePlan := range plan.Rules {
		for _, command := range rulePlan.Commands {
			if err := a.backend.Run(ctx, command); err != nil {
				if plan.CommandOnly {
					return err
				}
				return a.compensate(ctx, plan, err)
			}
		}
	}

	return nil
}

func (a *Adapter) Rollback(ctx context.Context, plan filter.CommandBatch) error {
	if err := validatePlan(plan); err != nil {
		return err
	}
	for _, rulePlan := range plan.Rules {
		for _, command := range rulePlan.RollbackCommands {
			if err := validateNativeCommand(command); err != nil {
				return err
			}
		}
	}
	for _, rulePlan := range plan.Rules {
		for _, command := range rulePlan.RollbackCommands {
			if err := a.backend.Run(ctx, command); err != nil {
				return err
			}
		}
	}
	return a.backend.Save(ctx)
}

func (a *Adapter) compensate(ctx context.Context, plan filter.CommandBatch, cause error) error {
	if plan.CreatesOnly() {
		return cause
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
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
	case filter.BasicBeforeChain:
		return nftables_helper.BasicBeforeChain
	case filter.IptablesInputChain:
		return nftables_helper.BasicChain
	case filter.BasicAfterChain:
		return nftables_helper.BasicAfterChain
	default:
		return ""
	}
}

func validatePlan(plan filter.CommandBatch) error {
	if plan.Provider != filter.ProviderNftables || len(plan.Rules) == 0 {
		return fmt.Errorf("%w: invalid nftables plan", filter.ErrInvalidRule)
	}
	return validateScope(plan.Scope)
}

func validateNativeCommand(command filter.NativeCommand) error {
	if command.Executable != "nft" || len(command.Args) != 0 || command.Stdin == "" {
		return fmt.Errorf("%w: invalid nftables native command", filter.ErrInvalidRule)
	}
	return nil
}

func compileChange(snapshot filter.RuleSet, change filter.RuleChange) (filter.ObservedRule, *filter.ObservedRule, error) {
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	if rule == nil {
		return filter.ObservedRule{}, nil, fmt.Errorf("%w: %s rule is required", filter.ErrInvalidRule, change.Operation)
	}
	normalized, err := filter.NormalizeRule(*rule)
	if err != nil {
		return filter.ObservedRule{}, nil, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Key() || normalized.UUID == "" {
		return filter.ObservedRule{}, nil, fmt.Errorf("%w: invalid nftables mutation rule", filter.ErrInvalidRule)
	}
	if change.Locator == nil || change.Locator.Position == nil {
		return filter.ObservedRule{}, nil, fmt.Errorf("%w: mutation requires a position locator", filter.ErrInvalidRule)
	}
	position := *change.Locator.Position
	if position < 1 || position > len(snapshot.Rules) {
		return filter.ObservedRule{}, nil, filter.ErrRuleStale
	}
	previous := snapshot.Rules[position-1]
	if previous.Protected {
		return filter.ObservedRule{}, nil, filter.ErrProtectedRule
	}
	target := position
	if normalized.OrderIndex != nil && (change.Operation == filter.ChangeUpdate || change.Operation == filter.ChangeReorder) {
		target = int(*normalized.OrderIndex)
	}
	if target < 1 || target > len(snapshot.Rules) {
		return filter.ObservedRule{}, nil, fmt.Errorf("%w: target position is out of range", filter.ErrInvalidRule)
	}
	marker := "1panel-rule:" + normalized.UUID
	expected := observedRule(normalized, marker, target, strings.Join(compileExpressionArgs(normalized, marker), " "))
	return expected, &previous, nil
}

func observedRule(rule filter.FirewallRule, marker string, position int, raw string) filter.ObservedRule {
	return filter.ObservedRule{
		Rule: rule, Marker: marker, Raw: raw, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: filter.ProviderNftables, ScopeKey: rule.Scope.Key(), Position: &position},
	}
}

func ruleUUID(change filter.RuleChange) string {
	if change.After != nil {
		return change.After.UUID
	}
	if change.Before != nil {
		return change.Before.UUID
	}
	return ""
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
		if strings.HasPrefix(raw, "chain ") {
			continue
		}
		handle := strings.TrimSpace(line[handleIndex+len("# handle "):])
		position := len(rules) + 1
		rules = append(rules, parseRule(scope, raw, handle, position))
	}
	return rules
}

func parseRule(scope filter.Scope, raw, handle string, position int) filter.ObservedRule {
	locator := filter.Locator{Provider: filter.ProviderNftables, ScopeKey: scope.Key(), NativeID: handle, Canonical: raw, Position: &position}
	opaque := func() filter.ObservedRule {
		return filter.ObservedRule{
			Rule:    filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindOpaque},
			Locator: locator, ParseStatus: filter.ParseStatusOpaque, Raw: raw,
			Protected: scope.Chain == filter.BasicAfterChain,
		}
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
				want := string(filter.FamilyIPv4)
				if scope.Family == filter.FamilyIPv6 {
					want = string(filter.FamilyIPv6)
				}
				if tokens[index+2] != want {
					return opaque()
				}
			} else if tokens[index+1] == "l4proto" {
				rule.Protocol = parseProtocol(tokens[index+2])
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
			values := tokens[index]
			if values == "{" {
				index++
				start := index
				for index < len(tokens) && tokens[index] != "}" {
					index++
				}
				if index == len(tokens) {
					return opaque()
				}
				values = strings.Join(tokens[start:index], " ")
			}
			for _, state := range strings.Split(values, ",") {
				state = strings.TrimSpace(state)
				if state == "" {
					return opaque()
				}
				rule.ConnectionStates = append(rule.ConnectionStates, parseConnectionState(state))
			}
			index++
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
	return filter.ObservedRule{Rule: normalized, Locator: locator, Marker: marker, ParseStatus: filter.ParseStatusSupported, Raw: raw, Protected: filter.IsBuiltinProtectedRule(normalized)}
}

func parseProtocol(value string) string {
	switch numericSymbol(value) {
	case "1":
		return "icmp"
	case "6":
		return "tcp"
	case "17":
		return "udp"
	case "58", "ipv6-icmp":
		return "icmpv6"
	default:
		return value
	}
}

func parseConnectionState(value string) string {
	switch numericSymbol(value) {
	case "1":
		return "invalid"
	case "2":
		return "established"
	case "4":
		return "related"
	case "8":
		return "new"
	case "64":
		return "untracked"
	default:
		return value
	}
}

func numericSymbol(value string) string {
	base := 10
	digits := value
	if strings.HasPrefix(strings.ToLower(value), "0x") {
		base, digits = 16, value[2:]
	}
	if number, err := strconv.ParseUint(digits, base, 32); err == nil {
		return strconv.FormatUint(number, 10)
	}
	return value
}

type systemBackend struct{}

func (systemBackend) ReadRulesByComment(ctx context.Context, scope filter.Scope, comment string) (string, error) {
	return filter.ReadRulesByComment(ctx, "nft", []string{"-a", "-n", "-n", "list", "chain", nftables_helper.TableFamily(scope.Family), nftables_helper.TableName, nativeChainName(scope)}, comment)
}

func (systemBackend) ListChain(ctx context.Context, scope filter.Scope) (string, error) {
	run := func(args ...string) (string, error) {
		return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout(
			"nft", append([]string{"-n", "-n"}, args...)...,
		)
	}
	return nftables_helper.ReadChain(run, nftables_helper.TableFamily(scope.Family), nftables_helper.TableName, nativeChainName(scope))
}

func (systemBackend) ListTable(ctx context.Context, scope filter.Scope) (string, bool, error) {
	run := func(args ...string) (string, error) {
		return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout("nft", append([]string{"-n", "-n"}, args...)...)
	}
	return nftables_helper.ReadTable(run, nftables_helper.TableFamily(scope.Family), nftables_helper.TableName)
}

func (systemBackend) Run(ctx context.Context, command filter.NativeCommand) error {
	if command.Executable != "nft" {
		return fmt.Errorf("unexpected nftables executable %q", command.Executable)
	}
	if command.Stdin != "" {
		return nftables_helper.RunScriptContext(ctx, command.Stdin)
	}
	return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).
		RunWithOptionalSudo(command.Executable, command.Args...)
}

func (systemBackend) Save(ctx context.Context) error {
	return nftables_helper.PersistRuleset(ctx)
}

func (a *Adapter) SaveRules(ctx context.Context, scope filter.Scope) error {
	return a.backend.Save(ctx)
}
