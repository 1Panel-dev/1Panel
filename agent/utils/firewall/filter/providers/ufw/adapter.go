package ufw

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

type CommandReader interface {
	Read(context.Context, ...string) (string, error)
}

type CommandWriter interface {
	Run(context.Context, filter.NativeCommand) error
}

type Adapter struct {
	reader CommandReader
	writer CommandWriter
}

func NewAdapter() *Adapter {
	backend := systemBackend{}
	return &Adapter{reader: backend, writer: backend}
}

func NewAdapterWithReader(reader CommandReader) *Adapter {
	return &Adapter{reader: reader}
}

func NewAdapterWithBackend(reader CommandReader, writer CommandWriter) *Adapter {
	return &Adapter{reader: reader, writer: writer}
}

func (a *Adapter) Provider() filter.Provider { return filter.ProviderUFW }

func (a *Adapter) Capabilities(context.Context) (filter.Capabilities, error) {
	return filter.Capabilities{
		Scopes: []filter.ScopePattern{{
			Provider: filter.ProviderUFW, Families: []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6},
			Chains: []string{filter.UFWInputChain}, Directions: []filter.Direction{filter.DirectionInput},
		}},
		Marker: true, ExplicitPosition: true,
	}, nil
}

func (a *Adapter) Observe(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	scope = scope.Normalize()
	if err := validateScope(scope); err != nil {
		return filter.Snapshot{}, err
	}
	if a.reader == nil {
		return filter.Snapshot{}, errors.New("ufw reader is required")
	}
	numbered, err := a.reader.Read(ctx, "status", "numbered")
	if err != nil {
		return filter.Snapshot{}, err
	}

	rules := parseNumberedRules(scope, numbered)
	snapshot, err := filter.NewSnapshot(scope, rules)
	if err != nil {
		return filter.Snapshot{}, err
	}
	snapshot.Notices = statusNotices(numbered)
	return snapshot, nil
}

// NativeDetail loads a UFW application profile only when its detail is opened.
func (a *Adapter) NativeDetail(ctx context.Context, profile string, _ bool) (string, error) {
	profile = strings.TrimSpace(profile)
	if !validApplicationProfileName(profile) {
		return "", fmt.Errorf("%w: invalid UFW application profile name", filter.ErrInvalidRule)
	}
	if a.reader == nil {
		return "", errors.New("ufw reader is required")
	}
	info, err := a.reader.Read(ctx, "app", "info", profile)
	if err != nil {
		return "", fmt.Errorf("read UFW application profile %q: %w", profile, err)
	}
	info = strings.TrimSpace(info)
	if info == "" {
		return "", fmt.Errorf("read UFW application profile %q: empty output", profile)
	}
	return info, nil
}

func (a *Adapter) Compile(snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, error) {
	if snapshot.Revision == "" {
		return filter.BackendPlan{}, filter.ErrRuleStale
	}
	if err := validateScope(snapshot.Scope); err != nil {
		return filter.BackendPlan{}, err
	}
	if hasScopeNotice(snapshot.Notices, filter.ScopeNoticeManagedScopeInactive) {
		return filter.BackendPlan{}, fmt.Errorf("%w: ufw is inactive", filter.ErrProviderUnavailable)
	}
	if len(changes) != 1 {
		return filter.BackendPlan{}, fmt.Errorf("%w: ufw plans currently require exactly one change", filter.ErrInvalidRule)
	}
	rulePlan, err := compileChange(snapshot, changes[0])
	if err != nil {
		return filter.BackendPlan{}, err
	}
	return filter.BackendPlan{
		Provider: filter.ProviderUFW, Scope: snapshot.Scope, SnapshotRevision: snapshot.Revision,
		Rules: []filter.NativeRulePlan{rulePlan},
	}, nil
}

func (a *Adapter) Apply(ctx context.Context, plan filter.BackendPlan) (filter.ApplyResult, error) {
	if err := validateBackendPlan(plan); err != nil {
		return filter.ApplyResult{}, err
	}
	if a.writer == nil {
		return filter.ApplyResult{}, errors.New("ufw writer is required")
	}
	for _, command := range append(append([]filter.NativeCommand(nil), plan.Rules[0].Commands...), plan.Rules[0].RollbackCommands...) {
		if err := validateCommand(command); err != nil {
			return filter.ApplyResult{}, err
		}
	}
	executed := 0
	for index, command := range plan.Rules[0].Commands {
		if err := a.writer.Run(ctx, command); err != nil {
			// UFW can update its native rules and still return a non-zero exit
			// status. Re-read the marker to distinguish that case from a command
			// that had no effect before choosing the rollback boundary.
			if a.failedCommandApplied(ctx, plan.Rules[0], index) {
				executed = index + 1
			}
			return filter.ApplyResult{}, a.compensate(ctx, plan.Rules[0], executed, fmt.Errorf("execute UFW rule: %w", err))
		}
		executed = index + 1
	}
	verification, err := a.verify(ctx, plan)
	if err != nil {
		return filter.ApplyResult{}, a.compensate(ctx, plan.Rules[0], executed, fmt.Errorf("verify UFW rule: %w", err))
	}
	if !verification.Matched {
		return filter.ApplyResult{}, a.compensate(ctx, plan.Rules[0], executed, errors.New("ufw write verification failed"))
	}
	return filter.ApplyResult{
		Applied:      []filter.ObservedRule{plan.Rules[0].Expected},
		Verification: &verification,
	}, nil
}

func (a *Adapter) failedCommandApplied(ctx context.Context, plan filter.NativeRulePlan, commandIndex int) bool {
	snapshot, err := a.Observe(ctx, plan.Expected.Rule.Scope)
	if err != nil {
		// If state cannot be read, prefer removing a possibly-applied rule over
		// leaving an untracked allow/deny rule behind.
		return true
	}
	markerCount := countMarker(snapshot, plan.Expected.Marker)
	switch plan.Operation {
	case filter.ChangeCreate, filter.ChangeAdopt:
		return markerCount > 0
	case filter.ChangeUpdate:
		if commandIndex == 0 {
			return markerCount == 0
		}
		return markerCount > 0
	case filter.ChangeDelete:
		return markerCount == 0
	default:
		return true
	}
}

func (a *Adapter) Verify(ctx context.Context, plan filter.BackendPlan) (filter.VerifyResult, error) {
	if err := validateBackendPlan(plan); err != nil {
		return filter.VerifyResult{}, err
	}
	return a.verify(ctx, plan)
}

func (a *Adapter) Rollback(ctx context.Context, plan filter.BackendPlan) error {
	if err := validateBackendPlan(plan); err != nil {
		return err
	}
	if a.writer == nil {
		return errors.New("ufw writer is required")
	}
	for ruleIndex := len(plan.Rules) - 1; ruleIndex >= 0; ruleIndex-- {
		rulePlan := plan.Rules[ruleIndex]
		if err := a.rollback(ctx, rulePlan, len(rulePlan.RollbackCommands)); err != nil {
			return err
		}
	}
	return nil
}

func validateScope(scope filter.Scope) error {
	if err := scope.ValidateMVP(); err != nil {
		return err
	}
	if scope.Provider != filter.ProviderUFW || scope.Chain != filter.UFWInputChain || scope.Direction != filter.DirectionInput {
		return fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, scope.Key())
	}
	return nil
}

func validateBackendPlan(plan filter.BackendPlan) error {
	if plan.Provider != filter.ProviderUFW || len(plan.Rules) != 1 || plan.SnapshotRevision == "" {
		return fmt.Errorf("%w: invalid ufw backend plan", filter.ErrInvalidRule)
	}
	if err := validateScope(plan.Scope); err != nil {
		return err
	}
	if len(plan.Rules[0].Commands) != len(plan.Rules[0].RollbackCommands) {
		return fmt.Errorf("%w: incomplete ufw rollback plan", filter.ErrInvalidRule)
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
	if err := validateWritableRule(normalized); err != nil {
		return filter.NativeRulePlan{}, err
	}
	if normalized.UUID == "" {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: rule UUID is required", filter.ErrInvalidRule)
	}
	if (change.Operation == filter.ChangeCreate || change.Operation == filter.ChangeUpdate) && isBroadDeny(normalized) {
		return filter.NativeRulePlan{}, filter.ErrLockoutRisk
	}
	marker := "1panel-rule:" + normalized.UUID
	position := insertionPosition(snapshot, normalized)
	expected := observedForRule(normalized, marker, position)
	plan := filter.NativeRulePlan{RuleUUID: normalized.UUID, Operation: change.Operation, Expected: expected}

	switch change.Operation {
	case filter.ChangeCreate:
		if normalized.OrderIndex != nil && *normalized.OrderIndex < 1 {
			return filter.NativeRulePlan{}, fmt.Errorf("%w: create target is out of range", filter.ErrInvalidRule)
		}
		command := insertCommand(position, normalized, marker)
		if change.Append {
			command = commentCommand(normalized, marker)
		}
		plan.Commands = []filter.NativeCommand{command}
		plan.RollbackCommands = []filter.NativeCommand{deleteRuleCommand(normalized, marker)}
	case filter.ChangeAdopt:
		target, targetErr := validateMutationTarget(snapshot, change, normalized, marker, false)
		if targetErr != nil {
			return filter.NativeRulePlan{}, targetErr
		}
		position = *target.Locator.Position
		plan.Previous = &target
		plan.Expected = observedForRule(normalized, marker, position)
		plan.Commands = []filter.NativeCommand{commentCommand(normalized, marker)}
		plan.RollbackCommands = []filter.NativeCommand{commentCommand(target.Rule, observedComment(target))}
	case filter.ChangeUpdate:
		target, targetErr := validateMutationTarget(snapshot, change, normalized, marker, true)
		if targetErr != nil {
			return filter.NativeRulePlan{}, targetErr
		}
		position = *target.Locator.Position
		targetPosition := position
		if normalized.OrderIndex != nil {
			if *normalized.OrderIndex < 1 {
				return filter.NativeRulePlan{}, fmt.Errorf("%w: update target is out of range", filter.ErrInvalidRule)
			}
			targetPosition = int(*normalized.OrderIndex)
		}
		plan.Previous = &target
		plan.Expected = observedForRule(normalized, marker, targetPosition)
		plan.Commands = []filter.NativeCommand{
			deletePositionCommand(position),
			positionedCommand(targetPosition, normalized, marker, change.Append),
		}
		plan.RollbackCommands = []filter.NativeCommand{
			positionedCommand(position, target.Rule, observedComment(target), change.RestoreAtEnd),
			deleteRuleCommand(normalized, marker),
		}
	case filter.ChangeDelete:
		target, targetErr := validateMutationTarget(snapshot, change, normalized, marker, true)
		if targetErr != nil {
			return filter.NativeRulePlan{}, targetErr
		}
		position = *target.Locator.Position
		plan.Previous = &target
		plan.Expected = target
		plan.Commands = []filter.NativeCommand{deletePositionCommand(position)}
		plan.RollbackCommands = []filter.NativeCommand{
			positionedCommand(position, target.Rule, observedComment(target), change.RestoreAtEnd),
		}
	case filter.ChangeReorder:
		return filter.NativeRulePlan{}, fmt.Errorf("%w: ufw reorder is not supported", filter.ErrUnsupportedScope)
	default:
		return filter.NativeRulePlan{}, fmt.Errorf("%w: unsupported operation %s", filter.ErrInvalidRule, change.Operation)
	}
	return plan, nil
}

func validateWritableRule(rule filter.FirewallRule) error {
	if rule.Scope.Provider != filter.ProviderUFW || (rule.Scope.Family != filter.FamilyIPv4 && rule.Scope.Family != filter.FamilyIPv6) ||
		rule.NativeKind != filter.NativeKindUFWRule {
		return fmt.Errorf("%w: unsupported ufw rule representation", filter.ErrInvalidRule)
	}
	if rule.SourcePort != "" || len(rule.ConnectionStates) != 0 || rule.Priority != nil || rule.OrderBucket != "" {
		return fmt.Errorf("%w: ufw source ports, states, priorities and backend options are not supported", filter.ErrInvalidRule)
	}
	if rule.Protocol != "all" && rule.Protocol != "tcp" && rule.Protocol != "udp" {
		return fmt.Errorf("%w: ufw protocol %q is not supported", filter.ErrInvalidRule, rule.Protocol)
	}
	return nil
}

func isBroadDeny(rule filter.FirewallRule) bool {
	return (rule.Action == filter.ActionDrop || rule.Action == filter.ActionReject) && rule.Protocol == "all" &&
		rule.SourceAddress == "" && rule.DestinationAddress == "" && rule.DestinationPort == ""
}

func validateMutationTarget(snapshot filter.Snapshot, change filter.DesiredChange, desired filter.FirewallRule, marker string, requireOwned bool) (filter.ObservedRule, error) {
	if change.Locator == nil || change.Locator.Position == nil {
		return filter.ObservedRule{}, fmt.Errorf("%w: ufw mutation requires a numbered locator", filter.ErrInvalidRule)
	}
	if change.Locator.Provider != "" && change.Locator.Provider != filter.ProviderUFW {
		return filter.ObservedRule{}, fmt.Errorf("%w: ufw locator provider mismatch", filter.ErrInvalidRule)
	}
	if change.Locator.ScopeKey != "" && change.Locator.ScopeKey != snapshot.Scope.Key() {
		return filter.ObservedRule{}, fmt.Errorf("%w: ufw locator scope mismatch", filter.ErrInvalidRule)
	}
	var matches []filter.ObservedRule
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position == nil || *observed.Locator.Position != *change.Locator.Position {
			continue
		}
		if change.Locator.NativeID != "" && observed.Locator.NativeID != change.Locator.NativeID {
			continue
		}
		if change.Locator.Canonical != "" && observed.Locator.Canonical != change.Locator.Canonical {
			continue
		}
		matches = append(matches, observed)
	}
	if len(matches) != 1 {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	target := matches[0]
	if target.Protected {
		return filter.ObservedRule{}, filter.ErrProtectedRule
	}
	if target.ParseStatus != filter.ParseStatusSupported {
		return filter.ObservedRule{}, fmt.Errorf("%w: opaque ufw rules cannot be modified", filter.ErrInvalidRule)
	}
	if requireOwned {
		if target.Marker != marker {
			return filter.ObservedRule{}, fmt.Errorf("%w: ufw rule is not owned by this rule UUID", filter.ErrInvalidRule)
		}
	} else if target.Marker != "" {
		return filter.ObservedRule{}, fmt.Errorf("%w: ufw adoption target is already marked", filter.ErrInvalidRule)
	}
	wantKey, err := filter.RuleKey(desired)
	if err != nil {
		return filter.ObservedRule{}, err
	}
	gotKey, err := filter.RuleKey(target.Rule)
	if err != nil {
		return filter.ObservedRule{}, err
	}
	if change.Operation == filter.ChangeAdopt && wantKey != gotKey {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	return target, nil
}

func insertionPosition(snapshot filter.Snapshot, rule filter.FirewallRule) int {
	if rule.OrderIndex != nil && *rule.OrderIndex > 0 {
		return int(*rule.OrderIndex)
	}
	maxPosition := 0
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position != nil && *observed.Locator.Position > maxPosition {
			maxPosition = *observed.Locator.Position
		}
	}
	position := maxPosition + 1
	return position
}

func observedForRule(rule filter.FirewallRule, marker string, position int) filter.ObservedRule {
	positionCopy := position
	order := int64(position)
	rule.OrderIndex = &order
	return filter.ObservedRule{
		Rule: rule, Marker: marker, ParseStatus: filter.ParseStatusSupported, Persistence: filter.PersistenceStatusConverged,
		Locator: filter.Locator{
			Provider: filter.ProviderUFW, ScopeKey: rule.Scope.Key(), NativeID: strconv.Itoa(position),
			Canonical: canonicalRule(rule), Position: &positionCopy,
		},
	}
}

func insertCommand(position int, rule filter.FirewallRule, comment string) filter.NativeCommand {
	args := []string{"insert", strconv.Itoa(position)}
	args = append(args, compileRuleArgs(rule, comment)...)
	return filter.NativeCommand{Executable: "ufw", Args: args}
}

func positionedCommand(position int, rule filter.FirewallRule, comment string, appendAtEnd bool) filter.NativeCommand {
	if appendAtEnd {
		return commentCommand(rule, comment)
	}
	return insertCommand(position, rule, comment)
}

func commentCommand(rule filter.FirewallRule, comment string) filter.NativeCommand {
	return filter.NativeCommand{Executable: "ufw", Args: compileRuleArgs(rule, comment)}
}

func deletePositionCommand(position int) filter.NativeCommand {
	return filter.NativeCommand{Executable: "ufw", Args: []string{"--force", "delete", strconv.Itoa(position)}}
}

func deleteRuleCommand(rule filter.FirewallRule, comment string) filter.NativeCommand {
	args := []string{"--force", "delete"}
	args = append(args, compileRuleArgs(rule, comment)...)
	return filter.NativeCommand{Executable: "ufw", Args: args}
}

func compileRuleArgs(rule filter.FirewallRule, comment string) []string {
	args := []string{nativeAction(rule.Action), "in"}
	if rule.Interface != "" {
		args = append(args, "on", rule.Interface)
	}
	if rule.Protocol != "all" {
		args = append(args, "proto", rule.Protocol)
	}
	args = append(args, "from", nativeAddress(rule.SourceAddress, rule.Scope.Family))
	args = append(args, "to", nativeAddress(rule.DestinationAddress, rule.Scope.Family))
	if rule.DestinationPort != "" {
		args = append(args, "port", strings.ReplaceAll(rule.DestinationPort, "-", ":"))
	}
	args = append(args, "comment", comment)
	return args
}

func nativeAction(action filter.Action) string {
	switch action {
	case filter.ActionAccept:
		return "allow"
	case filter.ActionDrop:
		return "deny"
	case filter.ActionReject:
		return "reject"
	default:
		return ""
	}
}

func nativeAddress(address string, family filter.Family) string {
	if address != "" {
		return address
	}
	if family == filter.FamilyIPv6 {
		return "::/0"
	}
	return "0.0.0.0/0"
}

func observedComment(observed filter.ObservedRule) string {
	if observed.Marker != "" {
		return observed.Marker
	}
	return observed.Rule.Description
}

func (a *Adapter) verify(ctx context.Context, plan filter.BackendPlan) (filter.VerifyResult, error) {
	snapshot, err := a.Observe(ctx, plan.Scope)
	if err != nil {
		return filter.VerifyResult{}, err
	}
	rulePlan := plan.Rules[0]
	marker := rulePlan.Expected.Marker
	count := countMarker(snapshot, marker)
	for _, scope := range relatedScopes(plan.Scope) {
		other, err := a.Observe(ctx, scope)
		if err != nil {
			return filter.VerifyResult{}, err
		}
		count += countMarker(other, marker)
	}
	if rulePlan.Operation == filter.ChangeDelete {
		return filter.VerifyResult{Snapshot: snapshot, Matched: count == 0}, nil
	}
	if count != 1 {
		return filter.VerifyResult{Snapshot: snapshot, Matched: false}, nil
	}
	wantKey, err := filter.RuleKey(rulePlan.Expected.Rule)
	if err != nil {
		return filter.VerifyResult{}, err
	}
	for _, observed := range snapshot.Rules {
		if observed.Marker != marker || observed.ParseStatus != filter.ParseStatusSupported {
			continue
		}
		gotKey, keyErr := filter.RuleKey(observed.Rule)
		if keyErr != nil || gotKey != wantKey || observed.Locator.Position == nil || rulePlan.Expected.Locator.Position == nil ||
			*observed.Locator.Position != *rulePlan.Expected.Locator.Position {
			return filter.VerifyResult{Snapshot: snapshot, Matched: false}, nil
		}
		return filter.VerifyResult{Snapshot: snapshot, Matched: true}, nil
	}
	return filter.VerifyResult{Snapshot: snapshot, Matched: false}, nil
}

func (a *Adapter) compensate(ctx context.Context, plan filter.NativeRulePlan, executed int, cause error) error {
	rollbackErr := a.rollback(ctx, plan, executed)
	if rollbackErr != nil {
		return fmt.Errorf("ufw apply failed: %w; compensation failed: %v", cause, rollbackErr)
	}
	return cause
}

func (a *Adapter) rollback(ctx context.Context, plan filter.NativeRulePlan, executed int) error {
	var rollbackErr error
	for index := executed - 1; index >= 0; index-- {
		if index >= len(plan.RollbackCommands) {
			if rollbackErr == nil {
				rollbackErr = errors.New("missing ufw rollback command")
			}
			continue
		}
		if err := a.writer.Run(ctx, plan.RollbackCommands[index]); err != nil && rollbackErr == nil {
			rollbackErr = err
		}
	}
	return rollbackErr
}

func relatedScopes(scope filter.Scope) []filter.Scope {
	scope = scope.Normalize()
	other := scope
	if other.Family == filter.FamilyIPv4 {
		other.Family = filter.FamilyIPv6
	} else {
		other.Family = filter.FamilyIPv4
	}
	return []filter.Scope{other}
}

func countMarker(snapshot filter.Snapshot, marker string) int {
	count := 0
	for _, observed := range snapshot.Rules {
		if observed.Marker == marker {
			count++
		}
	}
	return count
}

func hasScopeNotice(notices []filter.ScopeNotice, code filter.ScopeNoticeCode) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}
	return false
}

func validateCommand(command filter.NativeCommand) error {
	if command.Executable != "ufw" || len(command.Args) == 0 {
		return fmt.Errorf("%w: invalid ufw command", filter.ErrInvalidRule)
	}
	first := command.Args[0]
	if first == "--force" {
		if len(command.Args) < 3 || command.Args[1] != "delete" {
			return fmt.Errorf("%w: invalid forced ufw command", filter.ErrInvalidRule)
		}
		return nil
	}
	switch first {
	case "insert", "allow", "deny", "reject":
		return nil
	default:
		return fmt.Errorf("%w: unsupported ufw command %q", filter.ErrInvalidRule, first)
	}
}

func parseNumberedRules(scope filter.Scope, output string) []filter.ObservedRule {
	rules := make([]filter.ObservedRule, 0)
	for _, rawLine := range strings.Split(output, "\n") {
		raw := strings.TrimSpace(rawLine)
		matches := re.UFWNumberedRuleRegex.FindStringSubmatch(raw)
		if len(matches) == 0 {
			if opaque, family, inbound, ok := parseUnrecognizedNumberedRule(scope, raw); ok &&
				family == scope.Family && inbound {
				rules = append(rules, opaque)
			}
			continue
		}
		position, err := strconv.Atoi(matches[1])
		if err != nil || position < 1 {
			continue
		}
		if !isInboundNumberedRule(matches[4], matches[5]) {
			continue
		}
		family := familyForNumberedRule(matches[2], matches[5])
		if family != scope.Family {
			continue
		}
		rules = append(rules, parseNumberedRule(scope, position, matches[2], matches[3], matches[4], matches[5], raw))
	}
	return rules
}

func parseNumberedRule(scope filter.Scope, position int, destination, action, direction, source, raw string) filter.ObservedRule {
	comment, source := splitComment(source)
	positionCopy := position
	locator := filter.Locator{
		Provider: filter.ProviderUFW, ScopeKey: scope.Key(), NativeID: strconv.Itoa(position),
		Canonical: normalizedDisplay(raw), Position: &positionCopy,
	}
	ruleAction := map[string]filter.Action{
		"ALLOW":  filter.ActionAccept,
		"DENY":   filter.ActionDrop,
		"REJECT": filter.ActionReject,
		"LIMIT":  filter.ActionAccept,
	}[action]
	partialRule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule, Action: ruleAction,
	}
	opaque := func() filter.ObservedRule {
		return filter.ObservedRule{
			Rule:    partialRule,
			Locator: locator, ParseStatus: filter.ParseStatusOpaque, Raw: raw, Persistence: filter.PersistenceStatusConverged,
		}
	}
	partial := func() filter.ObservedRule {
		return filter.ObservedRule{
			Rule: partialRule, Locator: locator, ParseStatus: filter.ParseStatusPartial, Raw: raw,
			Persistence: filter.PersistenceStatusConverged,
		}
	}

	if direction == "FWD" || direction == "OUT" || strings.Contains(source, "(out)") {
		return opaque()
	}
	source = strings.ReplaceAll(source, "(out)", "")
	destination, destinationV6 := stripV6Marker(destination)
	source, sourceV6 := stripV6Marker(source)
	source, hasUnsupportedAnnotation := stripRuleAnnotations(source)
	observedV6 := destinationV6 || sourceV6 || containsIPv6Address(destination) || containsIPv6Address(source)
	if (scope.Family == filter.FamilyIPv6) != observedV6 {
		return opaque()
	}

	sourceAddress, ok := parseSource(source)
	if !ok {
		return opaque()
	}
	partialRule.SourceAddress = sourceAddress
	if ruleAction == "" {
		return opaque()
	}
	order := int64(position)
	destinationAddress, destinationPort, protocol, iface, destinationAnnotation, ok := parseDestination(destination)
	if !ok {
		profile, profileInterface, profileOK := applicationProfile(destination)
		if !profileOK {
			return opaque()
		}
		return filter.ObservedRule{
			Rule: filter.FirewallRule{
				Scope: scope, NativeKind: filter.NativeKindUFWApplication, Protocol: "",
				SourceAddress: sourceAddress, Interface: profileInterface, Action: ruleAction,
				OrderIndex: &order, Description: profile,
			},
			Locator: locator, ParseStatus: filter.ParseStatusOpaque, Raw: raw,
			Persistence: filter.PersistenceStatusConverged,
		}
	}
	partialRule.Protocol = protocol
	partialRule.DestinationAddress = destinationAddress
	partialRule.DestinationPort = destinationPort
	partialRule.Interface = iface
	partialRule.Description = comment
	if partialRule.Description == "" {
		partialRule.Description = destinationAnnotation
	}
	if destinationAnnotation != "" && validApplicationProfileName(destinationAnnotation) {
		partialRule.NativeKind = filter.NativeKindUFWApplication
		partialRule.Description = destinationAnnotation
		return partial()
	}
	if action == "LIMIT" || hasUnsupportedAnnotation {
		return partial()
	}
	rule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule, Protocol: protocol,
		SourceAddress: sourceAddress, DestinationAddress: destinationAddress, DestinationPort: destinationPort,
		Interface: iface, Action: ruleAction, OrderIndex: &order,
	}
	marker := ""
	if strings.HasPrefix(comment, "1panel-rule:") && strings.TrimSpace(strings.TrimPrefix(comment, "1panel-rule:")) != "" {
		marker = comment
	} else {
		rule.Description = comment
		if rule.Description == "" {
			rule.Description = destinationAnnotation
		}
	}
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return opaque()
	}
	locator.Canonical = canonicalRule(normalized)
	return filter.ObservedRule{
		Rule: normalized, Locator: locator, Marker: marker, ParseStatus: filter.ParseStatusSupported,
		Raw: raw, Persistence: filter.PersistenceStatusConverged,
	}
}

func stripRuleAnnotations(value string) (string, bool) {
	tokens := strings.Fields(value)
	filtered := tokens[:0]
	found := false
	for _, token := range tokens {
		switch strings.ToLower(token) {
		case "(log)", "(log-all)":
			found = true
		default:
			filtered = append(filtered, token)
		}
	}
	return strings.Join(filtered, " "), found
}

func applicationProfile(value string) (string, string, bool) {
	profile, iface, ok := splitInterface(value)
	profile = strings.TrimSpace(profile)
	if !ok || !validApplicationProfileName(profile) {
		return "", "", false
	}
	return profile, iface, true
}

func validApplicationProfileName(profile string) bool {
	profile = strings.TrimSpace(profile)
	if profile == "" || len(profile) > 255 || strings.ContainsAny(profile, "/\\\r\n\x00") {
		return false
	}
	hasLetter := false
	for _, char := range profile {
		if char < 0x20 || char == 0x7f {
			return false
		}
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
	}
	return hasLetter
}

func familyForNumberedRule(destination, source string) filter.Family {
	_, source = splitComment(source)
	if strings.Contains(destination, "(v6)") || strings.Contains(source, "(v6)") || containsIPv6Address(destination) || containsIPv6Address(source) {
		return filter.FamilyIPv6
	}
	return filter.FamilyIPv4
}

func containsIPv6Address(value string) bool {
	for _, token := range strings.Fields(value) {
		token = strings.Trim(token, "(),")
		if prefix, err := netip.ParsePrefix(token); err == nil && prefix.Addr().Is6() {
			return true
		}
		if address, err := netip.ParseAddr(token); err == nil && address.Is6() {
			return true
		}
	}
	return false
}

func parseUnrecognizedNumberedRule(scope filter.Scope, raw string) (filter.ObservedRule, filter.Family, bool, bool) {
	matches := re.UFWNumberedRulePrefixRegex.FindStringSubmatch(raw)
	if len(matches) == 0 {
		return filter.ObservedRule{}, "", false, false
	}
	position, err := strconv.Atoi(matches[1])
	if err != nil || position < 1 {
		return filter.ObservedRule{}, "", false, false
	}
	positionCopy := position
	_, familyInput := splitComment(matches[2])
	family := filter.FamilyIPv4
	if strings.Contains(familyInput, "(v6)") || containsIPv6Address(familyInput) {
		family = filter.FamilyIPv6
	}
	upper := " " + strings.ToUpper(strings.Join(strings.Fields(familyInput), " ")) + " "
	inbound := !strings.Contains(upper, " FWD ") && !strings.Contains(upper, " OUT ") &&
		!strings.Contains(strings.ToLower(familyInput), "(out)")
	return filter.ObservedRule{
		Rule: filter.FirewallRule{
			Scope: scope, NativeKind: filter.NativeKindUFWRule, Action: filter.ActionAccept,
		},
		Locator: filter.Locator{
			Provider: filter.ProviderUFW, ScopeKey: scope.Key(), NativeID: strconv.Itoa(position),
			Canonical: normalizedDisplay(raw), Position: &positionCopy,
		},
		ParseStatus: filter.ParseStatusOpaque, Raw: raw, Persistence: filter.PersistenceStatusConverged,
	}, family, inbound, true
}

func splitComment(source string) (string, string) {
	index := strings.Index(source, "#")
	if index < 0 {
		return "", strings.TrimSpace(source)
	}
	return strings.TrimSpace(source[index+1:]), strings.TrimSpace(source[:index])
}

func stripV6Marker(value string) (string, bool) {
	found := false
	for {
		index := strings.Index(strings.ToLower(value), "(v6)")
		if index < 0 {
			break
		}
		value = value[:index] + value[index+len("(v6)"):]
		found = true
	}
	return strings.Join(strings.Fields(value), " "), found
}

func parseDestination(value string) (address, port, protocol, iface, annotation string, ok bool) {
	value, iface, ok = splitInterface(value)
	if !ok {
		return "", "", "", "", "", false
	}
	value, annotation = splitDestinationAnnotation(value)
	tokens := strings.Fields(value)
	switch len(tokens) {
	case 1:
		if isAnywhere(tokens[0]) {
			return "", "", "all", iface, annotation, true
		}
		if parsedPort, parsedProtocol, portOK := parsePortProtocol(tokens[0]); portOK {
			return "", parsedPort, parsedProtocol, iface, annotation, true
		}
		if parsedAddress, addressOK := parseAddress(tokens[0]); addressOK {
			return parsedAddress, "", "all", iface, annotation, true
		}
	case 2:
		parsedAddress, addressOK := parseAddress(tokens[0])
		parsedPort, parsedProtocol, portOK := parsePortProtocol(tokens[1])
		if addressOK && portOK {
			return parsedAddress, parsedPort, parsedProtocol, iface, annotation, true
		}
	}
	return "", "", "", "", "", false
}

func splitDestinationAnnotation(value string) (string, string) {
	value = strings.TrimSpace(value)
	if !strings.HasSuffix(value, ")") {
		return value, ""
	}
	depth := 0
	for index := len(value) - 1; index >= 0; index-- {
		switch value[index] {
		case ')':
			depth++
		case '(':
			depth--
			if depth == 0 {
				endpoint := strings.TrimSpace(value[:index])
				annotation := strings.TrimSpace(value[index+1 : len(value)-1])
				if endpoint != "" && annotation != "" {
					return endpoint, annotation
				}
				return value, ""
			}
		}
	}
	return value, ""
}

func parseSource(value string) (string, bool) {
	if isAnywhere(value) {
		return "", true
	}
	if strings.Contains(value, " on ") || len(strings.Fields(value)) != 1 {
		return "", false
	}
	return parseAddress(value)
}

func isInboundNumberedRule(direction, source string) bool {
	return direction != "OUT" && direction != "FWD" && !strings.Contains(source, "(out)")
}

func splitInterface(value string) (endpoint, iface string, ok bool) {
	const delimiter = " on "
	index := strings.LastIndex(value, delimiter)
	if index < 0 {
		return strings.TrimSpace(value), "", true
	}
	endpoint = strings.TrimSpace(value[:index])
	iface = strings.TrimSpace(value[index+len(delimiter):])
	if endpoint == "" || iface == "" || strings.ContainsAny(iface, " \t") {
		return "", "", false
	}
	return endpoint, iface, true
}

func parsePortProtocol(value string) (string, string, bool) {
	parts := strings.Split(value, "/")
	protocol := "all"
	portValue := parts[0]
	if len(parts) == 2 {
		protocol = strings.ToLower(parts[1])
		if protocol != "tcp" && protocol != "udp" {
			return "", "", false
		}
	} else if len(parts) != 1 {
		return "", "", false
	}
	ports := strings.Split(portValue, ",")
	normalized := make([]string, 0, len(ports))
	for _, port := range ports {
		port = strings.TrimSpace(port)
		bounds := strings.Split(port, ":")
		if len(bounds) > 2 {
			return "", "", false
		}
		for _, bound := range bounds {
			value, err := strconv.Atoi(bound)
			if err != nil || value < 1 || value > 65535 {
				return "", "", false
			}
		}
		if len(bounds) == 2 {
			start, _ := strconv.Atoi(bounds[0])
			end, _ := strconv.Atoi(bounds[1])
			if start > end {
				return "", "", false
			}
			port = bounds[0] + "-" + bounds[1]
		}
		normalized = append(normalized, port)
	}
	return strings.Join(normalized, ","), protocol, len(normalized) > 0
}

func parseAddress(value string) (string, bool) {
	if isAnywhere(value) {
		return "", true
	}
	if prefix, err := netip.ParsePrefix(value); err == nil {
		return prefix.String(), true
	}
	if address, err := netip.ParseAddr(value); err == nil {
		return address.String(), true
	}
	return "", false
}

func isAnywhere(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "Anywhere")
}

func canonicalRule(rule filter.FirewallRule) string {
	return strings.Join([]string{
		string(rule.Action), rule.Protocol, rule.SourceAddress, rule.DestinationAddress,
		rule.DestinationPort, rule.Interface,
	}, "|")
}

func normalizedDisplay(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}

func statusNotices(numbered string) []filter.ScopeNotice {
	notices := make([]filter.ScopeNotice, 0, 1)
	if !statusActive(numbered) {
		notices = append(notices, filter.ScopeNotice{Code: filter.ScopeNoticeManagedScopeInactive})
	}
	return notices
}

func statusActive(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.EqualFold(strings.TrimSpace(line), "Status: active") {
			return true
		}
	}
	return false
}

type systemBackend struct{}

func (systemBackend) Read(_ context.Context, args ...string) (string, error) {
	return cmd.NewCommandMgr(
		cmd.WithTimeout(60*time.Second), cmd.WithEnv("LANGUAGE=en_US:en"),
	).RunWithOptionalSudoAndStdout("ufw", args...)
}

func (systemBackend) Run(_ context.Context, command filter.NativeCommand) error {
	if err := validateCommand(command); err != nil {
		return err
	}
	return cmd.NewCommandMgr(
		cmd.WithTimeout(60*time.Second), cmd.WithEnv("LANGUAGE=en_US:en"),
	).RunWithOptionalSudo(command.Executable, command.Args...)
}
