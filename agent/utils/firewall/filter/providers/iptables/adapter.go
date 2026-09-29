package iptables

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	native "github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/mattn/go-shellwords"
)

type RuleReader interface {
	filter.CommentRuleReader
	ListChain(context.Context, filter.Scope) (string, error)
}

type TableReader interface {
	ListTable(context.Context, filter.Scope) (string, error)
}

type RuleWriter interface {
	Run(context.Context, filter.NativeCommand) error
	Save(context.Context, filter.Scope) error
}

type Adapter struct {
	reader RuleReader
	writer RuleWriter
}

func NewAdapter() *Adapter {
	backend := systemBackend{}
	return &Adapter{reader: backend, writer: backend}
}

func NewAdapterWithReader(reader RuleReader) *Adapter {
	return &Adapter{reader: reader}
}

func NewAdapterWithBackend(reader RuleReader, writer RuleWriter) *Adapter {
	return &Adapter{reader: reader, writer: writer}
}

func (a *Adapter) Provider() filter.Provider { return filter.ProviderIptables }

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
	if err := validateAdapterScope(rule.Scope); err != nil {
		return err
	}
	args := []string{"-w", "-t", rule.Scope.Table, "-A", rule.Scope.Chain}
	args = append(args, compileRuleArgs(rule, comment)...)
	return a.writer.Run(ctx, filter.NativeCommand{Executable: executableForFamily(rule.Scope.Family), Args: args})
}

func (a *Adapter) ListRulesByComment(ctx context.Context, scopes []filter.Scope, comment string) ([]filter.ObservedRule, error) {
	var rules []filter.ObservedRule
	for _, scope := range scopes {
		scope = scope.Normalize()
		if err := validateAdapterScope(scope); err != nil {
			return nil, err
		}
		output, err := a.reader.ReadRulesByComment(ctx, scope, comment)
		if err != nil {
			return nil, err
		}
		rules = append(rules, parseChainRules(scope, output)...)
	}
	return rules, nil
}

func (a *Adapter) ListRules(ctx context.Context, scope filter.Scope) (filter.RuleSet, error) {
	snapshots, err := a.ListRuleScopes(ctx, []filter.Scope{scope})
	if err != nil {
		return filter.RuleSet{}, err
	}
	return snapshots[0], nil
}

func (a *Adapter) ListRuleScopes(ctx context.Context, scopes []filter.Scope) ([]filter.RuleSet, error) {
	normalized := make([]filter.Scope, len(scopes))
	for index, scope := range scopes {
		scope = scope.Normalize()
		if err := validateAdapterScope(scope); err != nil {
			return nil, err
		}
		normalized[index] = scope
	}
	if a.reader == nil {
		return nil, fmt.Errorf("iptables reader is required")
	}
	tableReader, readsTable := a.reader.(TableReader)
	snapshots := make([]filter.RuleSet, len(scopes))
	for index, scope := range normalized {
		if snapshots[index].Scope.Provider != "" {
			continue
		}
		var output string
		var err error
		if readsTable {
			output, err = tableReader.ListTable(ctx, scope)
		} else {
			output, err = a.reader.ListChain(ctx, scope)
		}
		if err != nil && (readsTable || !errors.Is(err, filter.ErrProviderUnavailable)) {
			return nil, err
		}
		for target := index; target < len(normalized); target++ {
			current := normalized[target]
			if readsTable {
				if current.Family != scope.Family || current.Table != scope.Table {
					continue
				}
			} else if target != index {
				continue
			}
			missing := errors.Is(err, filter.ErrProviderUnavailable) || readsTable && !containsChainDeclaration(output, current.Chain)
			var rules []filter.ObservedRule
			if !missing {
				rules = parseChainRules(current, output)
			}
			snapshot, buildErr := filter.NewRuleSet(current, rules)
			if buildErr != nil {
				return nil, buildErr
			}
			if missing {
				snapshot.Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeMissing, Values: []string{string(current.Family), current.Chain}}}
			}
			snapshots[target] = snapshot
		}
	}
	return snapshots, nil
}

func (a *Adapter) BuildCommands(snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	snapshot.Scope = snapshot.Scope.Normalize()
	if err := validateAdapterScope(snapshot.Scope); err != nil {
		return filter.CommandBatch{}, err
	}
	if len(changes) == 0 {
		return filter.CommandBatch{}, fmt.Errorf("%w: iptables plan requires at least one change", filter.ErrInvalidRule)
	}
	createOnly, deleteOnly := true, true
	for _, change := range changes {
		createOnly = createOnly && change.Operation == filter.ChangeCreate && change.CommandOnly
		deleteOnly = deleteOnly && change.Operation == filter.ChangeDelete && change.CommandOnly
	}
	if createOnly {
		return compileCreateBatch(snapshot, changes)
	}
	externalDelete := len(changes) == 1 && changes[0].Locator == nil && (changes[0].UnmarkedAdopted || changes[0].PreviousMarker != "")
	if deleteOnly && !externalDelete {
		return compileDeleteBatch(snapshot, changes)
	}
	if len(changes) != 1 {
		return filter.CommandBatch{}, fmt.Errorf("%w: iptables mutation requires exactly one change", filter.ErrInvalidRule)
	}
	rulePlan, err := compileChange(snapshot, changes[0])
	if err != nil {
		return filter.CommandBatch{}, err
	}
	return filter.CommandBatch{
		Provider: filter.ProviderIptables, Scope: snapshot.Scope, CommandOnly: changes[0].CommandOnly,
		Rules: []filter.RuleCommands{rulePlan},
	}, nil
}

func (a *Adapter) RunCommands(ctx context.Context, plan filter.CommandBatch) error {
	if a.writer == nil {
		return fmt.Errorf("iptables writer is required")
	}
	if plan.Provider != filter.ProviderIptables {
		return fmt.Errorf("%w: backend plan provider %q", filter.ErrUnsupportedScope, plan.Provider)
	}
	if err := validateAdapterScope(plan.Scope); err != nil {
		return err
	}
	if len(plan.Rules) == 0 {
		return fmt.Errorf("%w: iptables plan requires at least one rule", filter.ErrInvalidRule)
	}
	for _, rulePlan := range plan.Rules {
		for _, command := range rulePlan.Commands {
			if err := validateNativeCommand(plan.Scope, command); err != nil {
				return err
			}
		}
	}
	for ruleIndex, rulePlan := range plan.Rules {
		executed := 0
		for _, command := range rulePlan.Commands {
			if err := a.writer.Run(ctx, command); err != nil {
				if plan.CommandOnly {
					return err
				}
				return a.compensate(ctx, plan, ruleIndex, executed, err)
			}
			executed++
		}
	}

	return nil
}

func compileDeleteBatch(snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	plan := filter.CommandBatch{Provider: filter.ProviderIptables, Scope: snapshot.Scope, CommandOnly: true}
	var script strings.Builder
	fmt.Fprintf(&script, "*%s\n", snapshot.Scope.Table)
	for _, change := range changes {
		rulePlan, err := compileChange(snapshot, change)
		if err != nil {
			return filter.CommandBatch{}, err
		}
		line, err := restoreRuleLine(snapshot.Scope, *rulePlan.Previous)
		if err != nil {
			return filter.CommandBatch{}, err
		}
		script.WriteString(strings.Replace(line, "-A ", "-D ", 1))
		script.WriteByte('\n')
		rulePlan.Commands, rulePlan.RollbackCommands = nil, nil
		plan.Rules = append(plan.Rules, rulePlan)
	}
	script.WriteString("COMMIT\n")
	plan.Rules[0].Commands = []filter.NativeCommand{{
		Executable: restoreExecutableForFamily(snapshot.Scope.Family), Args: []string{"--noflush", "--wait"}, Stdin: script.String(),
	}}
	return plan, nil
}

func compileCreateBatch(snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	plan := filter.CommandBatch{Provider: filter.ProviderIptables, Scope: snapshot.Scope, CommandOnly: true}
	var script strings.Builder
	fmt.Fprintf(&script, "*%s\n", snapshot.Scope.Table)
	for _, change := range changes {
		rulePlan, err := compileChange(snapshot, change)
		if err != nil {
			return filter.CommandBatch{}, err
		}
		line, err := restoreRuleLine(snapshot.Scope, rulePlan.Expected)
		if err != nil {
			return filter.CommandBatch{}, err
		}
		if !change.Append {
			line = strings.Replace(line, "-A "+snapshot.Scope.Chain+" ", fmt.Sprintf("-I %s %d ", snapshot.Scope.Chain, *rulePlan.Expected.Locator.Position), 1)
		}
		rulePlan.Expected.Locator.Position = nil
		script.WriteString(line)
		script.WriteByte('\n')
		rulePlan.Commands, rulePlan.RollbackCommands = nil, nil
		plan.Rules = append(plan.Rules, rulePlan)
	}
	script.WriteString("COMMIT\n")
	plan.Rules[0].Commands = []filter.NativeCommand{{
		Executable: restoreExecutableForFamily(snapshot.Scope.Family), Args: []string{"--noflush", "--wait"}, Stdin: script.String(),
	}}
	return plan, nil
}

func restoreRuleLine(scope filter.Scope, observed filter.ObservedRule) (string, error) {
	if observed.Raw != "" {
		if strings.ContainsAny(observed.Raw, "\r\n") {
			return "", fmt.Errorf("%w: invalid newline in native iptables rule", filter.ErrInvalidRule)
		}
		args, err := shellwords.Parse(observed.Raw)
		if err != nil || len(args) < 3 || args[0] != "-A" || args[1] != scope.Chain {
			return "", fmt.Errorf("%w: invalid native iptables rule", filter.ErrInvalidRule)
		}
		return observed.Raw, nil
	}
	tokens := append([]string{"-A", scope.Chain}, compileRuleArgs(observed.Rule, observed.Marker)...)
	for _, token := range tokens {
		if token == "" || strings.ContainsAny(token, " \t\r\n\\\"'") {
			return "", fmt.Errorf("%w: invalid iptables-restore token %q", filter.ErrInvalidRule, token)
		}
	}
	return strings.Join(tokens, " "), nil
}

func (a *Adapter) Rollback(ctx context.Context, plan filter.CommandBatch) error {
	if a.writer == nil {
		return fmt.Errorf("iptables writer is required")
	}
	if plan.Provider != filter.ProviderIptables {
		return fmt.Errorf("%w: backend plan provider %q", filter.ErrUnsupportedScope, plan.Provider)
	}
	if err := validateAdapterScope(plan.Scope); err != nil {
		return err
	}
	if len(plan.Rules) == 0 {
		return nil
	}
	return a.rollback(ctx, plan, len(plan.Rules)-1, -1)
}

func (a *Adapter) compensate(ctx context.Context, plan filter.CommandBatch, lastRule, lastCommandCount int, cause error) error {
	if plan.CreatesOnly() {
		return cause
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	rollbackErr := a.rollback(ctx, plan, lastRule, lastCommandCount)
	if rollbackErr != nil {
		return fmt.Errorf("iptables apply failed: %w; compensation failed: %v", cause, rollbackErr)
	}
	return cause
}

func (a *Adapter) rollback(ctx context.Context, plan filter.CommandBatch, lastRule, lastCommandCount int) error {
	var rollbackErr error
	for index := lastRule; index >= 0; index-- {
		commands := plan.Rules[index].RollbackCommands
		limit := len(commands)
		if index == lastRule && lastCommandCount >= 0 && lastCommandCount < limit {
			limit = lastCommandCount
		}
		for commandIndex := limit - 1; commandIndex >= 0; commandIndex-- {
			if err := validateNativeCommand(plan.Scope, commands[commandIndex]); err != nil {
				if rollbackErr == nil {
					rollbackErr = err
				}
				continue
			}
			if err := a.writer.Run(ctx, commands[commandIndex]); err != nil && rollbackErr == nil {
				rollbackErr = err
			}
		}
	}
	if err := a.writer.Save(ctx, plan.Scope); err != nil && rollbackErr == nil {
		rollbackErr = err
	}
	return rollbackErr
}

func validateAdapterScope(scope filter.Scope) error {
	scope = scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return err
	}
	if scope.Provider != filter.ProviderIptables || (scope.Family != filter.FamilyIPv4 && scope.Family != filter.FamilyIPv6) {
		return fmt.Errorf("%w: iptables adapter scope %s", filter.ErrUnsupportedScope, scope.Key())
	}
	return nil
}

func compileChange(snapshot filter.RuleSet, change filter.RuleChange) (filter.RuleCommands, error) {
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	if rule == nil {
		return filter.RuleCommands{}, fmt.Errorf("%w: %s rule is required", filter.ErrInvalidRule, change.Operation)
	}
	normalized, err := filter.NormalizeRule(*rule)
	if err != nil {
		return filter.RuleCommands{}, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Key() {
		return filter.RuleCommands{}, fmt.Errorf("%w: change scope %s", filter.ErrUnsupportedScope, normalized.Scope.Key())
	}
	if normalized.UUID == "" {
		return filter.RuleCommands{}, fmt.Errorf("%w: rule UUID is required", filter.ErrInvalidRule)
	}
	if (normalized.Scope.Family == filter.FamilyIPv4 && normalized.Protocol == "icmpv6") ||
		(normalized.Scope.Family == filter.FamilyIPv6 && normalized.Protocol == "icmp") {
		return filter.RuleCommands{}, fmt.Errorf("%w: protocol %q does not match %s", filter.ErrInvalidRule, normalized.Protocol, normalized.Scope.Family)
	}
	marker := "1panel-rule:" + normalized.UUID
	if change.Operation == filter.ChangeDelete && change.CommandOnly && change.Locator == nil {
		previous := filter.ObservedRule{Rule: normalized, Marker: marker, ParseStatus: filter.ParseStatusSupported}
		var commands []filter.NativeCommand
		if change.UnmarkedAdopted || change.PreviousMarker != "" {
			previous.Marker = change.PreviousMarker
			args := []string{"-w", "-t", snapshot.Scope.Table, "-D", snapshot.Scope.Chain}
			args = append(args, compileObservedRuleArgs(previous)...)
			commands = []filter.NativeCommand{{Executable: executableForFamily(snapshot.Scope.Family), Args: args}}
		}
		return filter.RuleCommands{RuleUUID: normalized.UUID, Operation: change.Operation, Previous: &previous, Expected: previous, Commands: commands}, nil
	}
	position := len(snapshot.Rules) + 1
	verb := "-I"
	var target filter.ObservedRule
	switch change.Operation {
	case filter.ChangeCreate:
		hasSnapshot := !change.CommandOnly || snapshot.Rules != nil
		if normalized.OrderIndex != nil && (*normalized.OrderIndex < 1 || hasSnapshot && *normalized.OrderIndex > int64(len(snapshot.Rules)+1)) {
			return filter.RuleCommands{}, fmt.Errorf("%w: create target is out of range", filter.ErrInvalidRule)
		}
		position = insertionPosition(snapshot, normalized)
	case filter.ChangeAdopt:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.RuleCommands{}, err
		}
		verb = "-R"
	case filter.ChangeUpdate:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.RuleCommands{}, err
		}
		targetPosition := position
		if normalized.OrderIndex != nil {
			if *normalized.OrderIndex < 1 || *normalized.OrderIndex > int64(len(snapshot.Rules)) {
				return filter.RuleCommands{}, fmt.Errorf("%w: update target is out of range", filter.ErrInvalidRule)
			}
			targetPosition = int(*normalized.OrderIndex)
		}
		if position != targetPosition {
			return positionalMutationPlan(snapshot, normalized, target, marker, position, targetPosition, change.Operation), nil
		}
		verb = "-R"
	case filter.ChangeDelete:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.RuleCommands{}, err
		}
		verb = "-D"
	case filter.ChangeReorder:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.RuleCommands{}, err
		}
		if normalized.OrderIndex == nil || *normalized.OrderIndex < 1 || *normalized.OrderIndex > int64(len(snapshot.Rules)) {
			return filter.RuleCommands{}, fmt.Errorf("%w: reorder target is out of range", filter.ErrInvalidRule)
		}
		targetPosition := int(*normalized.OrderIndex)
		return positionalMutationPlan(snapshot, normalized, target, marker, position, targetPosition, change.Operation), nil
	default:
		return filter.RuleCommands{}, fmt.Errorf("%w: unsupported operation %s", filter.ErrInvalidRule, change.Operation)
	}
	args := []string{"-w", "-t", snapshot.Scope.Table, verb, snapshot.Scope.Chain, strconv.Itoa(position)}
	if change.Operation != filter.ChangeDelete {
		args = append(args, compileRuleArgs(normalized, marker)...)
	}
	rollbackArgs := []string{"-w", "-t", snapshot.Scope.Table}
	switch change.Operation {
	case filter.ChangeCreate:
		rollbackArgs = append(rollbackArgs, "-D", snapshot.Scope.Chain)
		rollbackArgs = append(rollbackArgs, compileRuleArgs(normalized, marker)...)
	case filter.ChangeAdopt, filter.ChangeUpdate:
		rollbackArgs = append(rollbackArgs, "-R", snapshot.Scope.Chain, strconv.Itoa(position))
		rollbackArgs = append(rollbackArgs, compileObservedRuleArgs(target)...)
	case filter.ChangeDelete:
		rollbackArgs = append(rollbackArgs, "-I", snapshot.Scope.Chain, strconv.Itoa(position))
		rollbackArgs = append(rollbackArgs, compileObservedRuleArgs(target)...)
	}
	expected := filter.ObservedRule{
		Rule: normalized, Marker: marker, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: snapshot.Scope.Key(), Position: &position},
	}
	var previous *filter.ObservedRule
	if change.Operation != filter.ChangeCreate {
		previous = &target
	}
	return filter.RuleCommands{
		RuleUUID: normalized.UUID, Operation: change.Operation,
		Commands:         []filter.NativeCommand{{Executable: executableForFamily(snapshot.Scope.Family), Args: args}},
		RollbackCommands: []filter.NativeCommand{{Executable: executableForFamily(snapshot.Scope.Family), Args: rollbackArgs}},
		Previous:         previous,
		Expected:         expected,
	}, nil
}

func positionalMutationPlan(snapshot filter.RuleSet, rule filter.FirewallRule, previous filter.ObservedRule, marker string, position int, targetPosition int, operation filter.ChangeOperation) filter.RuleCommands {
	expected := filter.ObservedRule{
		Rule: rule, Marker: marker, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: snapshot.Scope.Key(), Position: &targetPosition},
	}
	plan := filter.RuleCommands{
		RuleUUID: rule.UUID, Operation: operation, Previous: &previous, Expected: expected,
	}
	if position == targetPosition {
		return plan
	}
	executable := executableForFamily(snapshot.Scope.Family)
	deleteArgs := append([]string{"-w", "-t", snapshot.Scope.Table, "-D", snapshot.Scope.Chain}, compileObservedRuleArgs(previous)...)
	insertArgs := []string{"-w", "-t", snapshot.Scope.Table, "-I", snapshot.Scope.Chain, strconv.Itoa(targetPosition)}
	insertArgs = append(insertArgs, compileRuleArgs(rule, marker)...)
	restoreArgs := []string{"-w", "-t", snapshot.Scope.Table, "-I", snapshot.Scope.Chain, strconv.Itoa(position)}
	restoreArgs = append(restoreArgs, compileObservedRuleArgs(previous)...)
	deleteInsertedArgs := []string{"-w", "-t", snapshot.Scope.Table, "-D", snapshot.Scope.Chain}
	deleteInsertedArgs = append(deleteInsertedArgs, compileRuleArgs(rule, marker)...)
	plan.Commands = []filter.NativeCommand{
		{Executable: executable, Args: deleteArgs},
		{Executable: executable, Args: insertArgs},
	}
	plan.RollbackCommands = []filter.NativeCommand{
		{Executable: executable, Args: restoreArgs},
		{Executable: executable, Args: deleteInsertedArgs},
	}
	return plan
}

func compileRuleArgs(rule filter.FirewallRule, marker string) []string {
	args := make([]string, 0, 24)
	if rule.Protocol != "all" {
		protocol := rule.Protocol
		if protocol == "icmpv6" {
			protocol = "ipv6-icmp"
		}
		args = append(args, "-p", protocol)
	}
	if rule.Interface != "" {
		args = append(args, "-i", rule.Interface)
	}
	if rule.SourceAddress != "" {
		args = append(args, "-s", rule.SourceAddress)
	}
	if rule.DestinationAddress != "" {
		args = append(args, "-d", rule.DestinationAddress)
	}
	if rule.SourcePort != "" {
		if strings.Contains(rule.SourcePort, ",") {
			args = append(args, "-m", "multiport", "--sports", strings.ReplaceAll(rule.SourcePort, "-", ":"))
		} else {
			args = append(args, "--sport", strings.ReplaceAll(rule.SourcePort, "-", ":"))
		}
	}
	if rule.DestinationPort != "" {
		if strings.Contains(rule.DestinationPort, ",") {
			args = append(args, "-m", "multiport", "--dports", strings.ReplaceAll(rule.DestinationPort, "-", ":"))
		} else {
			args = append(args, "--dport", strings.ReplaceAll(rule.DestinationPort, "-", ":"))
		}
	}
	if len(rule.ConnectionStates) != 0 {
		args = append(args, "-m", "conntrack", "--ctstate", strings.ToUpper(strings.Join(rule.ConnectionStates, ",")))
	}
	if marker != "" {
		args = append(args, "-m", "comment", "--comment", marker)
	}
	args = append(args, "-j", strings.ToUpper(string(rule.Action)))
	return args
}

func compileObservedRuleArgs(observed filter.ObservedRule) []string {
	comment := observed.Marker
	if comment == "" {
		comment = observed.Rule.Description
	}
	return compileRuleArgs(observed.Rule, comment)
}

func validateMutationTarget(snapshot filter.RuleSet, change filter.RuleChange, after filter.FirewallRule, marker string) (int, filter.ObservedRule, error) {
	if change.Locator == nil || change.Locator.Position == nil {
		return 0, filter.ObservedRule{}, fmt.Errorf("%w: mutation requires a position locator", filter.ErrInvalidRule)
	}
	if change.Locator.Provider != "" && change.Locator.Provider != filter.ProviderIptables {
		return 0, filter.ObservedRule{}, fmt.Errorf("%w: locator provider mismatch", filter.ErrInvalidRule)
	}
	if change.Locator.ScopeKey != "" && change.Locator.ScopeKey != snapshot.Scope.Key() {
		return 0, filter.ObservedRule{}, fmt.Errorf("%w: locator scope mismatch", filter.ErrInvalidRule)
	}
	position := *change.Locator.Position
	if position < 1 || position > len(snapshot.Rules) {
		return 0, filter.ObservedRule{}, fmt.Errorf("%w: locator position %d is out of range", filter.ErrRuleStale, position)
	}
	observed := snapshot.Rules[position-1]
	if observed.Protected {
		return 0, filter.ObservedRule{}, filter.ErrProtectedRule
	}
	if observed.ParseStatus == filter.ParseStatusOpaque {
		return 0, filter.ObservedRule{}, fmt.Errorf("%w: target rule is opaque", filter.ErrUnsupportedScope)
	}
	want := after
	if change.Operation == filter.ChangeUpdate || change.Operation == filter.ChangeDelete {
		if change.Before == nil {
			return 0, filter.ObservedRule{}, fmt.Errorf("%w: previous rule is required", filter.ErrInvalidRule)
		}
		want = *change.Before
	}
	wantKey, wantErr := filter.RuleKey(want)
	observedKey, observedErr := filter.RuleKey(observed.Rule)
	if wantErr != nil || observedErr != nil || wantKey != observedKey {
		return 0, filter.ObservedRule{}, filter.ErrRuleStale
	}
	if change.Operation != filter.ChangeAdopt && observed.Marker != marker &&
		!(change.Operation == filter.ChangeDelete && change.UnmarkedAdopted && observed.Marker == "") {
		return 0, filter.ObservedRule{}, filter.ErrRuleStale
	}
	return position, observed, nil
}

func insertionPosition(snapshot filter.RuleSet, rule filter.FirewallRule) int {
	if rule.OrderIndex != nil {
		return int(*rule.OrderIndex)
	}
	if snapshot.Scope.Chain == native.BasicAfterChain && rule.Action == filter.ActionAccept {
		for index, observed := range snapshot.Rules {
			if observed.ParseStatus == filter.ParseStatusSupported && observed.Rule.Action == filter.ActionDrop &&
				observed.Rule.SourceAddress == "" && observed.Rule.DestinationAddress == "" &&
				observed.Rule.SourcePort == "" && observed.Rule.DestinationPort == "" {
				return index + 1
			}
		}
	}
	return len(snapshot.Rules) + 1
}

type systemBackend struct{}

func (systemBackend) ReadRulesByComment(ctx context.Context, scope filter.Scope, comment string) (string, error) {
	executable, err := runtimeExecutable(executableForFamily(scope.Family))
	if err != nil {
		return "", err
	}
	return filter.ReadRulesByComment(ctx, executable, []string{"-w", "-t", scope.Table, "-S", scope.Chain}, comment)
}

func (systemBackend) ListTable(ctx context.Context, scope filter.Scope) (string, error) {
	return native.ReadTable(ctx, scope.Table, scope.Family == filter.FamilyIPv6)
}

func (b systemBackend) ListChain(ctx context.Context, scope filter.Scope) (string, error) {
	output, err := b.ListTable(ctx, scope)
	if err != nil {
		return "", err
	}
	if !containsChainDeclaration(output, scope.Chain) {
		return "", fmt.Errorf("%w: iptables %s chain %s is not initialized", filter.ErrProviderUnavailable, scope.Family, scope.Chain)
	}
	return output, nil
}

func containsChainDeclaration(output, chain string) bool {
	declaration := "-N " + chain
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == declaration {
			return true
		}
	}
	return false
}

func (systemBackend) Run(ctx context.Context, command filter.NativeCommand) error {
	if command.Executable != "iptables" && command.Executable != "ip6tables" &&
		command.Executable != "iptables-restore" && command.Executable != "ip6tables-restore" {
		return fmt.Errorf("unexpected iptables executable %q", command.Executable)
	}
	executable, err := runtimeExecutable(command.Executable)
	if err != nil {
		return err
	}
	options := []cmd.Option{cmd.WithContext(ctx), cmd.WithTimeout(60 * time.Second)}
	if command.Stdin != "" {
		options = append(options, cmd.WithStdin(strings.NewReader(command.Stdin)))
	}
	return cmd.NewCommandMgr(options...).RunWithOptionalSudo(executable, command.Args...)
}

func (systemBackend) Save(ctx context.Context, scope filter.Scope) error {
	fileName := map[string]string{
		native.BasicBeforeChain: native.BasicBeforeFileName,
		native.BasicChain:       native.BasicFileName,
		native.BasicAfterChain:  native.BasicAfterFileName,
	}[scope.Chain]
	if fileName == "" {
		return fmt.Errorf("unsupported persistence chain %q", scope.Chain)
	}
	if scope.Family == filter.FamilyIPv6 {
		fileName = native.IPv6FileName(fileName)
		return native.SaveIPv6RulesToFileContext(ctx, scope.Table, scope.Chain, fileName)
	}
	return native.SaveRulesToFileContext(ctx, scope.Table, scope.Chain, fileName)
}

func executableForFamily(family filter.Family) string {
	if family == filter.FamilyIPv6 {
		return "ip6tables"
	}
	return "iptables"
}

func restoreExecutableForFamily(family filter.Family) string {
	if family == filter.FamilyIPv6 {
		return "ip6tables-restore"
	}
	return "iptables-restore"
}

func runtimeExecutable(logical string) (string, error) {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return "", err
	}
	switch logical {
	case "ip6tables":
		if !commands.IPv6Available() {
			return "", fmt.Errorf("%w: ip6tables command family is unavailable", filter.ErrFamilyUnavailable)
		}
		return commands.IPv6, nil
	case "iptables-restore":
		return commands.Restore4, nil
	case "ip6tables-restore":
		if commands.Restore6 == "" {
			return "", fmt.Errorf("%w: ip6tables-restore command family is unavailable", filter.ErrFamilyUnavailable)
		}
		return commands.Restore6, nil
	}
	return commands.IPv4, nil
}

func validateNativeCommand(scope filter.Scope, command filter.NativeCommand) error {
	expected := executableForFamily(scope.Family)
	if command.Stdin != "" {
		expected = restoreExecutableForFamily(scope.Family)
	}
	if command.Executable != expected {
		return fmt.Errorf("%w: %s scope requires %s, got %s", filter.ErrInvalidRule, scope.Family, expected, command.Executable)
	}
	if strings.HasSuffix(command.Executable, "-restore") && command.Stdin == "" {
		return fmt.Errorf("%w: iptables restore command requires input", filter.ErrInvalidRule)
	}
	return nil
}

func parseChainRules(scope filter.Scope, output string) []filter.ObservedRule {
	lines := strings.Split(output, "\n")
	rules := make([]filter.ObservedRule, 0, len(lines))
	position := 0
	for _, raw := range lines {
		raw = strings.TrimSpace(raw)
		if !strings.HasPrefix(raw, "-A "+scope.Chain+" ") {
			continue
		}
		position++
		observed := parseRule(scope, raw, position)
		rules = append(rules, observed)
	}
	return rules
}

func parseRule(scope filter.Scope, raw string, position int) filter.ObservedRule {
	locator := filter.Locator{
		Provider: filter.ProviderIptables, ScopeKey: scope.Key(), Canonical: raw, Position: &position,
	}
	opaque := func() filter.ObservedRule {
		return filter.ObservedRule{
			Rule:    filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindOpaque},
			Locator: locator, ParseStatus: filter.ParseStatusOpaque, Raw: raw,
		}
	}
	args, err := shellwords.Parse(raw)
	if err != nil || len(args) < 4 || args[0] != "-A" || args[1] != scope.Chain {
		return opaque()
	}
	rule := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "all"}
	comment := ""
	rejectWith := ""
	for index := 2; index < len(args); index++ {
		switch args[index] {
		case "-p", "--protocol":
			if !takeValue(args, &index, &rule.Protocol) {
				return opaque()
			}
			if rule.Protocol == "ipv6-icmp" {
				rule.Protocol = "icmpv6"
			}
		case "-s", "--source":
			if !takeValue(args, &index, &rule.SourceAddress) {
				return opaque()
			}
		case "-d", "--destination":
			if !takeValue(args, &index, &rule.DestinationAddress) {
				return opaque()
			}
		case "--sport", "--source-port":
			if !takeValue(args, &index, &rule.SourcePort) {
				return opaque()
			}
		case "--dport", "--destination-port":
			if !takeValue(args, &index, &rule.DestinationPort) {
				return opaque()
			}
		case "--sports", "--source-ports":
			if !takeValue(args, &index, &rule.SourcePort) {
				return opaque()
			}
		case "--dports", "--destination-ports":
			if !takeValue(args, &index, &rule.DestinationPort) {
				return opaque()
			}
		case "-i", "--in-interface":
			if !takeValue(args, &index, &rule.Interface) {
				return opaque()
			}
		case "-m", "--match":
			var module string
			if !takeValue(args, &index, &module) || (module != "tcp" && module != "udp" && module != "comment" && module != "conntrack" && module != "multiport") {
				return opaque()
			}
		case "--ctstate":
			var states string
			if !takeValue(args, &index, &states) {
				return opaque()
			}
			rule.ConnectionStates = strings.Split(states, ",")
		case "--comment":
			if !takeValue(args, &index, &comment) {
				return opaque()
			}
		case "-j", "--jump":
			var action string
			if !takeValue(args, &index, &action) {
				return opaque()
			}
			rule.Action = filter.Action(strings.ToLower(action))
		case "--reject-with":
			if !takeValue(args, &index, &rejectWith) {
				return opaque()
			}
		default:
			return opaque()
		}
	}
	if rule.Action != filter.ActionAccept && rule.Action != filter.ActionDrop && rule.Action != filter.ActionReject {
		return opaque()
	}
	if rejectWith != "" && !isDefaultRejectWith(scope.Family, rejectWith) {
		return opaque()
	}
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return opaque()
	}
	marker := ""
	if strings.HasPrefix(comment, "1panel-rule:") {
		marker = comment
	} else {
		normalized.Description = comment
	}
	return filter.ObservedRule{
		Rule: normalized, Locator: locator, Marker: marker, ParseStatus: filter.ParseStatusSupported, Raw: raw,
		Protected: filter.IsBuiltinProtectedRule(normalized),
	}
}

func isDefaultRejectWith(family filter.Family, value string) bool {
	if family == filter.FamilyIPv6 {
		return value == "icmp6-port-unreachable"
	}
	return value == "icmp-port-unreachable"
}

func takeValue(args []string, index *int, target *string) bool {
	if *index+1 >= len(args) {
		return false
	}
	*index = *index + 1
	*target = args[*index]
	return true
}

func (a *Adapter) SaveRules(ctx context.Context, scope filter.Scope) error {
	return a.writer.Save(ctx, scope)
}
