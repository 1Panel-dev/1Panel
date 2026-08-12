package iptables

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	native "github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/mattn/go-shellwords"
)

type RuleReader interface {
	ListChain(context.Context, filter.Scope) (string, error)
}

type RuleWriter interface {
	Run(context.Context, filter.NativeCommand) error
	Save(context.Context, filter.Scope) error
}

type MultiportChecker interface {
	CheckMultiport(context.Context, filter.Family) error
}

type Adapter struct {
	reader      RuleReader
	writer      RuleWriter
	checker     MultiportChecker
	multiportMu sync.Mutex
	multiportOK map[filter.Family]bool
}

func NewAdapter() *Adapter {
	backend := systemBackend{}
	return &Adapter{reader: backend, writer: backend, checker: backend}
}

func NewAdapterWithReader(reader RuleReader) *Adapter {
	adapter := &Adapter{reader: reader}
	adapter.checker, _ = reader.(MultiportChecker)
	return adapter
}

func NewAdapterWithBackend(reader RuleReader, writer RuleWriter) *Adapter {
	adapter := &Adapter{reader: reader, writer: writer}
	if checker, ok := reader.(MultiportChecker); ok {
		adapter.checker = checker
	} else if checker, ok := writer.(MultiportChecker); ok {
		adapter.checker = checker
	}
	return adapter
}

func (a *Adapter) Provider() filter.Provider { return filter.ProviderIptables }

func (a *Adapter) CheckRule(ctx context.Context, rule filter.FirewallRule) error {
	if !strings.Contains(rule.DestinationPort, ",") && !strings.Contains(rule.SourcePort, ",") {
		return nil
	}
	if rule.Protocol != "tcp" && rule.Protocol != "udp" {
		return fmt.Errorf("%w: iptables multiport requires tcp or udp", filter.ErrInvalidRule)
	}
	if a.checker == nil {
		return nil
	}
	a.multiportMu.Lock()
	defer a.multiportMu.Unlock()
	if a.multiportOK[rule.Scope.Family] {
		return nil
	}
	if err := a.checker.CheckMultiport(ctx, rule.Scope.Family); err != nil {
		return fmt.Errorf("%w: iptables multiport is unavailable for %s: %v", filter.ErrUnsupportedScope, rule.Scope.Family, err)
	}
	if a.multiportOK == nil {
		a.multiportOK = make(map[filter.Family]bool, 2)
	}
	a.multiportOK[rule.Scope.Family] = true
	return nil
}

func (a *Adapter) Capabilities(context.Context) (filter.Capabilities, error) {
	return filter.Capabilities{
		Scopes: []filter.ScopePattern{{
			Provider: filter.ProviderIptables, Families: []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}, Table: "filter",
			Chains:     []string{native.Chain1PanelBasicBefore, native.Chain1PanelBasic, native.Chain1PanelBasicAfter},
			Directions: []filter.Direction{filter.DirectionInput},
		}}, Marker: true, OwnedChains: true, ExplicitPosition: true,
		AtomicApply: false, TransactionalRollback: false,
	}, nil
}

func (a *Adapter) Observe(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	scope = scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return filter.Snapshot{}, err
	}
	if scope.Provider != filter.ProviderIptables {
		return filter.Snapshot{}, fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, scope.Key())
	}
	if a.reader == nil {
		return filter.Snapshot{}, fmt.Errorf("iptables reader is required")
	}
	output, err := a.reader.ListChain(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	rules := parseChainRules(scope, output)
	return filter.NewSnapshot(scope, rules)
}

func (a *Adapter) Compile(snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, error) {
	if snapshot.Revision == "" {
		return filter.BackendPlan{}, filter.ErrRuleStale
	}
	if err := validateAdapterScope(snapshot.Scope); err != nil {
		return filter.BackendPlan{}, err
	}
	if len(changes) != 1 {
		return filter.BackendPlan{}, fmt.Errorf("%w: iptables plans currently require exactly one change", filter.ErrInvalidRule)
	}
	plan := filter.BackendPlan{Provider: filter.ProviderIptables, Scope: snapshot.Scope, SnapshotRevision: snapshot.Revision}
	for _, change := range changes {
		rulePlan, err := compileChange(snapshot, change)
		if err != nil {
			return filter.BackendPlan{}, err
		}
		plan.Rules = append(plan.Rules, rulePlan)
	}
	return plan, nil
}

func (a *Adapter) Apply(ctx context.Context, plan filter.BackendPlan) (filter.ApplyResult, error) {
	if a.writer == nil {
		return filter.ApplyResult{}, fmt.Errorf("iptables writer is required")
	}
	if plan.Provider != filter.ProviderIptables {
		return filter.ApplyResult{}, fmt.Errorf("%w: backend plan provider %q", filter.ErrUnsupportedScope, plan.Provider)
	}
	if err := validateAdapterScope(plan.Scope); err != nil {
		return filter.ApplyResult{}, err
	}
	if len(plan.Rules) != 1 {
		return filter.ApplyResult{}, fmt.Errorf("%w: iptables plans currently require exactly one rule", filter.ErrInvalidRule)
	}
	for _, rulePlan := range plan.Rules {
		for _, command := range rulePlan.Commands {
			if err := validateNativeCommand(plan.Scope, command); err != nil {
				return filter.ApplyResult{}, err
			}
		}
	}
	for ruleIndex, rulePlan := range plan.Rules {
		executed := 0
		for _, command := range rulePlan.Commands {
			if err := a.writer.Run(ctx, command); err != nil {
				return filter.ApplyResult{}, a.compensate(ctx, plan, ruleIndex, executed, err)
			}
			executed++
		}
	}
	if err := a.writer.Save(ctx, plan.Scope); err != nil {
		return filter.ApplyResult{}, a.compensate(ctx, plan, len(plan.Rules)-1, -1, err)
	}
	applied := make([]filter.ObservedRule, 0, len(plan.Rules))
	for _, rulePlan := range plan.Rules {
		applied = append(applied, rulePlan.Expected)
	}
	return filter.ApplyResult{Applied: applied}, nil
}

func (a *Adapter) Verify(ctx context.Context, plan filter.BackendPlan) (filter.VerifyResult, error) {
	if plan.Provider != filter.ProviderIptables {
		return filter.VerifyResult{}, fmt.Errorf("%w: backend plan provider %q", filter.ErrUnsupportedScope, plan.Provider)
	}
	snapshot, err := a.Observe(ctx, plan.Scope)
	if err != nil {
		return filter.VerifyResult{}, err
	}
	for _, expected := range plan.Rules {
		markerMatches := 0
		semanticMatches := 0
		for _, observed := range snapshot.Rules {
			if observed.Marker != "" && observed.Marker == expected.Expected.Marker {
				markerMatches++
				want, wantErr := filter.RuleKey(expected.Expected.Rule)
				got, gotErr := filter.RuleKey(observed.Rule)
				if wantErr == nil && gotErr == nil && want == got {
					semanticMatches++
				}
			}
		}
		requiresPositionMatch := expected.Operation == filter.ChangeReorder ||
			(expected.Operation == filter.ChangeUpdate && expected.Expected.Rule.OrderIndex != nil)
		positionMatches := true
		if requiresPositionMatch {
			positionMatches = false
			for _, observed := range snapshot.Rules {
				if observed.Marker == expected.Expected.Marker && observed.Locator.Position != nil &&
					expected.Expected.Locator.Position != nil && *observed.Locator.Position == *expected.Expected.Locator.Position {
					positionMatches = true
					break
				}
			}
		}
		if (expected.Operation == filter.ChangeDelete && markerMatches != 0) ||
			(requiresPositionMatch && !positionMatches) ||
			(expected.Operation != filter.ChangeDelete && (markerMatches != 1 || semanticMatches != 1)) {
			return filter.VerifyResult{Snapshot: snapshot, Matched: false}, nil
		}
	}
	return filter.VerifyResult{Snapshot: snapshot, Matched: true}, nil
}

func (a *Adapter) Rollback(ctx context.Context, plan filter.BackendPlan) error {
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

func (a *Adapter) compensate(ctx context.Context, plan filter.BackendPlan, lastRule, lastCommandCount int, cause error) error {
	rollbackErr := a.rollback(ctx, plan, lastRule, lastCommandCount)
	if rollbackErr != nil {
		return fmt.Errorf("iptables apply failed: %w; compensation failed: %v", cause, rollbackErr)
	}
	return cause
}

func (a *Adapter) rollback(ctx context.Context, plan filter.BackendPlan, lastRule, lastCommandCount int) error {
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

func compileChange(snapshot filter.Snapshot, change filter.DesiredChange) (filter.NativeRulePlan, error) {
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	if rule == nil {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: %s rule is required", filter.ErrInvalidRule, change.Operation)
	}
	normalized, err := filter.NormalizeRule(*rule)
	if err != nil {
		return filter.NativeRulePlan{}, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Key() {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: change scope %s", filter.ErrUnsupportedScope, normalized.Scope.Key())
	}
	if normalized.UUID == "" {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: rule UUID is required", filter.ErrInvalidRule)
	}
	if (normalized.Scope.Family == filter.FamilyIPv4 && normalized.Protocol == "icmpv6") ||
		(normalized.Scope.Family == filter.FamilyIPv6 && normalized.Protocol == "icmp") {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: protocol %q does not match %s", filter.ErrInvalidRule, normalized.Protocol, normalized.Scope.Family)
	}
	if (change.Operation == filter.ChangeCreate || change.Operation == filter.ChangeUpdate) && isBroadDeny(normalized) {
		return filter.NativeRulePlan{}, filter.ErrLockoutRisk
	}
	marker := "1panel-rule:" + normalized.UUID
	position := len(snapshot.Rules) + 1
	verb := "-I"
	var target filter.ObservedRule
	switch change.Operation {
	case filter.ChangeCreate:
		if normalized.OrderIndex != nil && (*normalized.OrderIndex < 1 || *normalized.OrderIndex > int64(len(snapshot.Rules)+1)) {
			return filter.NativeRulePlan{}, fmt.Errorf("%w: create target is out of range", filter.ErrInvalidRule)
		}
		position = insertionPosition(snapshot, normalized)
	case filter.ChangeAdopt:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.NativeRulePlan{}, err
		}
		verb = "-R"
	case filter.ChangeUpdate:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.NativeRulePlan{}, err
		}
		targetPosition := position
		if normalized.OrderIndex != nil {
			if *normalized.OrderIndex < 1 || *normalized.OrderIndex > int64(len(snapshot.Rules)) {
				return filter.NativeRulePlan{}, fmt.Errorf("%w: update target is out of range", filter.ErrInvalidRule)
			}
			targetPosition = int(*normalized.OrderIndex)
		}
		if err := validateReorderPath(snapshot, position, targetPosition); err != nil {
			return filter.NativeRulePlan{}, err
		}
		if position != targetPosition {
			return positionalMutationPlan(snapshot, normalized, target, marker, position, targetPosition, change.Operation), nil
		}
		verb = "-R"
	case filter.ChangeDelete:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.NativeRulePlan{}, err
		}
		verb = "-D"
	case filter.ChangeReorder:
		position, target, err = validateMutationTarget(snapshot, change, normalized, marker)
		if err != nil {
			return filter.NativeRulePlan{}, err
		}
		if normalized.OrderIndex == nil || *normalized.OrderIndex < 1 || *normalized.OrderIndex > int64(len(snapshot.Rules)) {
			return filter.NativeRulePlan{}, fmt.Errorf("%w: reorder target is out of range", filter.ErrInvalidRule)
		}
		targetPosition := int(*normalized.OrderIndex)
		if err := validateReorderPath(snapshot, position, targetPosition); err != nil {
			return filter.NativeRulePlan{}, err
		}
		return positionalMutationPlan(snapshot, normalized, target, marker, position, targetPosition, change.Operation), nil
	default:
		return filter.NativeRulePlan{}, fmt.Errorf("%w: unsupported operation %s", filter.ErrInvalidRule, change.Operation)
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
	return filter.NativeRulePlan{
		RuleUUID: normalized.UUID, Operation: change.Operation,
		Commands:         []filter.NativeCommand{{Executable: executableForFamily(snapshot.Scope.Family), Args: args}},
		RollbackCommands: []filter.NativeCommand{{Executable: executableForFamily(snapshot.Scope.Family), Args: rollbackArgs}},
		Previous:         pointerToObserved(target, change.Operation != filter.ChangeCreate),
		Expected:         expected,
	}, nil
}

func positionalMutationPlan(
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
	previous filter.ObservedRule,
	marker string,
	position int,
	targetPosition int,
	operation filter.ChangeOperation,
) filter.NativeRulePlan {
	expected := filter.ObservedRule{
		Rule: rule, Marker: marker, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: snapshot.Scope.Key(), Position: &targetPosition},
	}
	plan := filter.NativeRulePlan{
		RuleUUID: rule.UUID, Operation: operation, Previous: &previous, Expected: expected,
	}
	if position == targetPosition {
		return plan
	}
	executable := executableForFamily(snapshot.Scope.Family)
	deleteArgs := []string{"-w", "-t", snapshot.Scope.Table, "-D", snapshot.Scope.Chain, strconv.Itoa(position)}
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

func pointerToObserved(rule filter.ObservedRule, include bool) *filter.ObservedRule {
	if !include {
		return nil
	}
	return &rule
}

func validateReorderPath(snapshot filter.Snapshot, from, to int) error {
	start, end := from, to
	if start > end {
		start, end = end, start
	}
	for position := start; position <= end; position++ {
		if position == from {
			continue
		}
		observed := snapshot.Rules[position-1]
		if observed.Protected {
			return filter.ErrProtectedRule
		}
		if observed.ParseStatus == filter.ParseStatusOpaque || observed.Marker == "" {
			return fmt.Errorf("%w: reorder cannot cross external or opaque rules", filter.ErrUnsupportedScope)
		}
	}
	return nil
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
			args = append(args, "-m", "multiport", "--sports", nativePort(rule.SourcePort))
		} else {
			args = append(args, "--sport", nativePort(rule.SourcePort))
		}
	}
	if rule.DestinationPort != "" {
		if strings.Contains(rule.DestinationPort, ",") {
			args = append(args, "-m", "multiport", "--dports", nativePort(rule.DestinationPort))
		} else {
			args = append(args, "--dport", nativePort(rule.DestinationPort))
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

func nativePort(port string) string {
	return strings.ReplaceAll(port, "-", ":")
}

func validateMutationTarget(snapshot filter.Snapshot, change filter.DesiredChange, after filter.FirewallRule, marker string) (int, filter.ObservedRule, error) {
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
	if change.Operation != filter.ChangeAdopt && observed.Marker != marker {
		return 0, filter.ObservedRule{}, filter.ErrRuleStale
	}
	return position, observed, nil
}

func insertionPosition(snapshot filter.Snapshot, rule filter.FirewallRule) int {
	if rule.OrderIndex != nil {
		return int(*rule.OrderIndex)
	}
	if snapshot.Scope.Chain == native.Chain1PanelBasicAfter && rule.Action == filter.ActionAccept {
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

func isBroadDeny(rule filter.FirewallRule) bool {
	return (rule.Action == filter.ActionDrop || rule.Action == filter.ActionReject) &&
		rule.SourceAddress == "" && rule.DestinationAddress == "" && rule.SourcePort == "" &&
		rule.DestinationPort == ""
}

type systemBackend struct{}

func (systemBackend) CheckMultiport(_ context.Context, family filter.Family) error {
	executable := executableForFamily(family)
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithOptionalSudo(executable, "-m", "multiport", "--help")
}

func (systemBackend) ListChain(_ context.Context, scope filter.Scope) (string, error) {
	if scope.Family == filter.FamilyIPv6 {
		return native.RunIPv6WithStd(scope.Table, "-S", scope.Chain)
	}
	return native.RunWithStd(scope.Table, "-S", scope.Chain)
}

func (systemBackend) Run(_ context.Context, command filter.NativeCommand) error {
	if command.Executable != "iptables" && command.Executable != "ip6tables" {
		return fmt.Errorf("unexpected iptables executable %q", command.Executable)
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudo(command.Executable, command.Args...)
}

func (systemBackend) Save(_ context.Context, scope filter.Scope) error {
	fileName := map[string]string{
		native.Chain1PanelBasicBefore: native.BasicBeforeFileName,
		native.Chain1PanelBasic:       native.BasicFileName,
		native.Chain1PanelBasicAfter:  native.BasicAfterFileName,
	}[scope.Chain]
	if fileName == "" {
		return fmt.Errorf("unsupported persistence chain %q", scope.Chain)
	}
	if scope.Family == filter.FamilyIPv6 {
		fileName = native.IPv6FileName(fileName)
		return native.SaveIPv6RulesToFile(scope.Table, scope.Chain, fileName)
	}
	return native.SaveRulesToFile(scope.Table, scope.Chain, fileName)
}

func executableForFamily(family filter.Family) string {
	if family == filter.FamilyIPv6 {
		return "ip6tables"
	}
	return "iptables"
}

func validateNativeCommand(scope filter.Scope, command filter.NativeCommand) error {
	expected := executableForFamily(scope.Family)
	if command.Executable != expected {
		return fmt.Errorf("%w: %s scope requires %s, got %s", filter.ErrInvalidRule, scope.Family, expected, command.Executable)
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
		Protected: isProtectedRule(scope, normalized, comment),
	}
}

func isDefaultRejectWith(family filter.Family, value string) bool {
	if family == filter.FamilyIPv6 {
		return value == "icmp6-port-unreachable"
	}
	return value == "icmp-port-unreachable"
}

func isProtectedRule(scope filter.Scope, rule filter.FirewallRule, comment string) bool {
	if scope.Chain == native.Chain1PanelBasicBefore || scope.Chain == native.Chain1PanelBasicAfter {
		return true
	}
	if rule.Action == filter.ActionAccept && rule.Interface == "lo" {
		return true
	}
	if rule.Action == filter.ActionAccept {
		states := make(map[string]struct{}, len(rule.ConnectionStates))
		for _, state := range rule.ConnectionStates {
			states[state] = struct{}{}
		}
		if _, established := states["established"]; established {
			return true
		}
	}
	if scope.Chain == native.Chain1PanelBasicAfter && rule.Action == filter.ActionDrop &&
		rule.SourceAddress == "" && rule.DestinationAddress == "" && rule.SourcePort == "" && rule.DestinationPort == "" {
		return true
	}
	return strings.Contains(strings.ToLower(comment), "whitelist")
}

func takeValue(args []string, index *int, target *string) bool {
	if *index+1 >= len(args) {
		return false
	}
	*index = *index + 1
	*target = args[*index]
	return true
}
