package firewalld

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/mattn/go-shellwords"
)

type CommandReader interface {
	Read(context.Context, ...string) (string, error)
}

type CommandWriter interface {
	Run(context.Context, filter.NativeCommand) error
}

type Adapter struct {
	reader                    CommandReader
	writer                    CommandWriter
	prioritySupportMu         sync.Mutex
	richRulePrioritySupported *bool
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

func (a *Adapter) Provider() filter.Provider { return filter.ProviderFirewalld }

func (a *Adapter) Capabilities(ctx context.Context) (filter.Capabilities, error) {
	explicitPriority, err := a.supportsRichRulePriority(ctx)
	if err != nil {
		return filter.Capabilities{}, err
	}
	return filter.Capabilities{
		Scopes: []filter.ScopePattern{{
			Provider: filter.ProviderFirewalld, Families: []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6, filter.FamilyInet},
			Zone: filter.FirewalldInputZone, Directions: []filter.Direction{filter.DirectionInput},
		}},
		ExplicitPriority: explicitPriority,
		NativePort:       true,
	}, nil
}

func (a *Adapter) CheckRule(ctx context.Context, rule filter.FirewallRule) error {
	if rule.Priority == nil || *rule.Priority == 0 {
		return nil
	}
	supported, err := a.supportsRichRulePriority(ctx)
	if err != nil {
		return err
	}
	if !supported {
		return fmt.Errorf("%w: firewalld versions before 0.7.0 do not support rich rule priority", filter.ErrUnsupportedScope)
	}
	return nil
}

func (a *Adapter) supportsRichRulePriority(ctx context.Context) (bool, error) {
	a.prioritySupportMu.Lock()
	defer a.prioritySupportMu.Unlock()
	if a.richRulePrioritySupported != nil {
		return *a.richRulePrioritySupported, nil
	}
	if a.reader == nil {
		return false, errors.New("firewalld reader is required")
	}
	output, err := a.reader.Read(ctx, "--version")
	if err != nil {
		return false, fmt.Errorf("load firewalld version: %w", err)
	}
	major, minor, err := parseFirewalldVersion(output)
	if err != nil {
		return false, err
	}
	supported := major > 0 || minor >= 7
	a.richRulePrioritySupported = &supported
	return supported, nil
}

func parseFirewalldVersion(output string) (int, int, error) {
	version := strings.TrimSpace(output)
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid firewalld version %q", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid firewalld version %q", version)
	}
	minorDigits := strings.TrimLeftFunc(parts[1], func(r rune) bool { return r < '0' || r > '9' })
	minorEnd := strings.IndexFunc(minorDigits, func(r rune) bool { return r < '0' || r > '9' })
	if minorEnd >= 0 {
		minorDigits = minorDigits[:minorEnd]
	}
	minor, err := strconv.Atoi(minorDigits)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid firewalld version %q", version)
	}
	return major, minor, nil
}

func (a *Adapter) Observe(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	scope = scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return filter.Snapshot{}, err
	}
	if scope.Provider != filter.ProviderFirewalld {
		return filter.Snapshot{}, fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, scope.Key())
	}
	scope.Family = filter.FamilyInet
	if a.reader == nil {
		return filter.Snapshot{}, errors.New("firewalld reader is required")
	}

	var runtime, permanent zoneOutput
	var runtimeErr, permanentErr error
	var reads sync.WaitGroup
	reads.Add(2)
	go func() {
		defer reads.Done()
		runtime, runtimeErr = a.readScope(ctx, scope, false)
	}()
	go func() {
		defer reads.Done()
		permanent, permanentErr = a.readScope(ctx, scope, true)
	}()
	reads.Wait()
	if err := errors.Join(runtimeErr, permanentErr); err != nil {
		return filter.Snapshot{}, err
	}
	rules, err := mergeZoneObjects(scope, runtime, permanent)
	if err != nil {
		return filter.Snapshot{}, err
	}
	snapshot, err := filter.NewSnapshot(scope, rules)
	if err != nil {
		return filter.Snapshot{}, err
	}
	snapshot.Notices = publicZoneNotices(runtime, permanent)
	return snapshot, nil
}

func (a *Adapter) PrepareRule(rule filter.FirewallRule) (filter.FirewallRule, error) {
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return filter.FirewallRule{}, err
	}
	if normalized.Scope.Provider != filter.ProviderFirewalld {
		return filter.FirewallRule{}, fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, normalized.Scope.Key())
	}
	if normalized.NativeKind == filter.NativeKindRule {
		if isNativeZonePort(normalized) {
			normalized.NativeKind = filter.NativeKindZonePort
		} else {
			normalized.NativeKind = filter.NativeKindRichRule
		}
		normalized, err = filter.NormalizeRule(normalized)
		if err != nil {
			return filter.FirewallRule{}, err
		}
	}
	if err := validateWritableRule(normalized); err != nil {
		return filter.FirewallRule{}, err
	}
	return normalized, nil
}

func (a *Adapter) Compile(snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, error) {
	if snapshot.Revision == "" {
		return filter.BackendPlan{}, filter.ErrRuleStale
	}
	if err := validateFirewalldScope(snapshot.Scope); err != nil {
		return filter.BackendPlan{}, err
	}
	if len(changes) != 1 {
		return filter.BackendPlan{}, fmt.Errorf("%w: firewalld plans currently require exactly one change", filter.ErrInvalidRule)
	}
	for _, observed := range snapshot.Rules {
		if observed.Persistence != "" && observed.Persistence != filter.PersistenceStatusConverged {
			return filter.BackendPlan{}, fmt.Errorf("%w: firewalld runtime and permanent state differ", filter.ErrRuleStale)
		}
	}
	rulePlan, err := a.compileChange(snapshot, changes[0])
	if err != nil {
		return filter.BackendPlan{}, err
	}
	return filter.BackendPlan{
		Provider: filter.ProviderFirewalld, Scope: snapshot.Scope, SnapshotRevision: snapshot.Revision,
		Rules: []filter.NativeRulePlan{rulePlan},
	}, nil
}

func (a *Adapter) Apply(ctx context.Context, plan filter.BackendPlan) (filter.ApplyResult, error) {
	if plan.Provider != filter.ProviderFirewalld || len(plan.Rules) != 1 {
		return filter.ApplyResult{}, fmt.Errorf("%w: invalid firewalld backend plan", filter.ErrInvalidRule)
	}
	if err := validateFirewalldScope(plan.Scope); err != nil {
		return filter.ApplyResult{}, err
	}
	rulePlan := plan.Rules[0]
	if len(rulePlan.Commands) != 0 && a.writer == nil {
		return filter.ApplyResult{}, errors.New("firewalld writer is required")
	}
	if len(rulePlan.Commands) != len(rulePlan.RollbackCommands) {
		return filter.ApplyResult{}, fmt.Errorf("%w: incomplete firewalld rollback plan", filter.ErrInvalidRule)
	}
	for _, command := range append(append([]filter.NativeCommand(nil), rulePlan.Commands...), rulePlan.RollbackCommands...) {
		if err := validateScopeCommand(plan.Scope, command); err != nil {
			return filter.ApplyResult{}, err
		}
	}
	executed := 0
	for index, command := range rulePlan.Commands {
		if err := a.writer.Run(ctx, command); err != nil {
			return filter.ApplyResult{}, a.compensate(ctx, rulePlan, executed, err)
		}
		executed = index + 1
	}
	return filter.ApplyResult{Applied: []filter.ObservedRule{rulePlan.Expected}}, nil
}

func (a *Adapter) Verify(ctx context.Context, plan filter.BackendPlan) (filter.VerifyResult, error) {
	if plan.Provider != filter.ProviderFirewalld || len(plan.Rules) != 1 {
		return filter.VerifyResult{}, fmt.Errorf("%w: invalid firewalld backend plan", filter.ErrInvalidRule)
	}
	snapshot, err := a.Observe(ctx, plan.Scope)
	if err != nil {
		return filter.VerifyResult{}, err
	}
	rulePlan := plan.Rules[0]
	if rulePlan.Operation == filter.ChangeDelete {
		return filter.VerifyResult{Snapshot: snapshot, Matched: countCanonical(snapshot, rulePlan.Previous) == 0}, nil
	}
	if rulePlan.Operation == filter.ChangeUpdate && rulePlan.Previous != nil &&
		rulePlan.Previous.Locator.Canonical != rulePlan.Expected.Locator.Canonical && countCanonical(snapshot, rulePlan.Previous) != 0 {
		return filter.VerifyResult{Snapshot: snapshot, Matched: false}, nil
	}
	matches := 0
	for _, observed := range snapshot.Rules {
		if observed.Locator.Canonical != rulePlan.Expected.Locator.Canonical || observed.Persistence != filter.PersistenceStatusConverged {
			continue
		}
		want, wantErr := filter.RuleKey(rulePlan.Expected.Rule)
		got, gotErr := filter.RuleKey(observed.Rule)
		if wantErr == nil && gotErr == nil && want == got {
			matches++
		}
	}
	return filter.VerifyResult{Snapshot: snapshot, Matched: matches == 1}, nil
}

func (a *Adapter) Rollback(ctx context.Context, plan filter.BackendPlan) error {
	if plan.Provider != filter.ProviderFirewalld {
		return fmt.Errorf("%w: invalid firewalld backend plan", filter.ErrInvalidRule)
	}
	if err := validateFirewalldScope(plan.Scope); err != nil {
		return err
	}
	if a.writer == nil {
		return errors.New("firewalld writer is required")
	}
	for ruleIndex := len(plan.Rules) - 1; ruleIndex >= 0; ruleIndex-- {
		rulePlan := plan.Rules[ruleIndex]
		if err := a.rollback(ctx, plan.Scope, rulePlan, len(rulePlan.RollbackCommands)); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) compileChange(snapshot filter.Snapshot, change filter.DesiredChange) (filter.NativeRulePlan, error) {
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	if rule == nil {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: %s rule is required", filter.ErrInvalidRule, change.Operation)
	}
	normalized, err := a.PrepareRule(*rule)
	if err != nil {
		return filter.NativeRulePlan{}, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Key() {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: change scope %s", filter.ErrUnsupportedScope, normalized.Scope.Key())
	}
	if normalized.UUID == "" {
		return filter.NativeRulePlan{}, fmt.Errorf("%w: rule UUID is required", filter.ErrInvalidRule)
	}
	if (change.Operation == filter.ChangeCreate || change.Operation == filter.ChangeUpdate) && isBroadDeny(normalized) {
		return filter.NativeRulePlan{}, filter.ErrLockoutRisk
	}

	expected := observedForRule(normalized)
	plan := filter.NativeRulePlan{RuleUUID: normalized.UUID, Operation: change.Operation, Expected: expected}
	switch change.Operation {
	case filter.ChangeCreate:
		plan.Commands, plan.RollbackCommands = pairedCommands(normalized, "add", "remove")
	case filter.ChangeAdopt:
		target, targetErr := validateMutationTarget(snapshot, change, normalized, false)
		if targetErr != nil {
			return filter.NativeRulePlan{}, targetErr
		}
		plan.Previous = &target
		plan.Expected = target
		plan.Expected.Rule.UUID = normalized.UUID
	case filter.ChangeUpdate:
		target, targetErr := validateMutationTarget(snapshot, change, normalized, true)
		if targetErr != nil {
			return filter.NativeRulePlan{}, targetErr
		}
		plan.Previous = &target
		removeCommands, restoreCommands := pairedCommands(target.Rule, "remove", "add")
		addCommands, removeNewCommands := pairedCommands(normalized, "add", "remove")
		plan.Commands = append(removeCommands, addCommands...)
		plan.RollbackCommands = append(restoreCommands, removeNewCommands...)
	case filter.ChangeDelete:
		target, targetErr := validateMutationTarget(snapshot, change, normalized, true)
		if targetErr != nil {
			return filter.NativeRulePlan{}, targetErr
		}
		plan.Previous = &target
		plan.Expected = target
		plan.Expected.Rule.UUID = normalized.UUID
		plan.Commands, plan.RollbackCommands = pairedCommands(target.Rule, "remove", "add")
	default:
		return filter.NativeRulePlan{}, fmt.Errorf("%w: unsupported operation %s", filter.ErrInvalidRule, change.Operation)
	}
	return plan, nil
}

func validateFirewalldScope(scope filter.Scope) error {
	scope = scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return err
	}
	if scope.Provider != filter.ProviderFirewalld {
		return fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, scope.Key())
	}
	return nil
}

func isNativeZonePort(rule filter.FirewallRule) bool {
	return rule.Scope.Family == filter.FamilyInet && rule.Action == filter.ActionAccept &&
		(rule.Protocol == "tcp" || rule.Protocol == "udp") && rule.DestinationPort != "" &&
		rule.SourceAddress == "" && rule.SourcePort == "" && rule.DestinationAddress == "" && rule.Interface == "" &&
		len(rule.ConnectionStates) == 0 && rule.Priority == nil
}

func validateWritableRule(rule filter.FirewallRule) error {
	if rule.SourcePort != "" || rule.Interface != "" || len(rule.ConnectionStates) != 0 {
		return fmt.Errorf("%w: firewalld rule uses unsupported options", filter.ErrInvalidRule)
	}
	switch rule.NativeKind {
	case filter.NativeKindZonePort:
		if !isNativeZonePort(rule) {
			return fmt.Errorf("%w: invalid firewalld native port", filter.ErrInvalidRule)
		}
	case filter.NativeKindRichRule:
		if rule.Scope.Family == filter.FamilyInet && (rule.SourceAddress != "" || rule.DestinationAddress != "") {
			return fmt.Errorf("%w: address rich rules require an explicit family", filter.ErrInvalidRule)
		}
	default:
		return fmt.Errorf("%w: firewalld native kind %q is not writable", filter.ErrInvalidRule, rule.NativeKind)
	}
	return nil
}

func isBroadDeny(rule filter.FirewallRule) bool {
	return (rule.Action == filter.ActionDrop || rule.Action == filter.ActionReject) &&
		rule.SourceAddress == "" && rule.DestinationAddress == "" && rule.SourcePort == "" && rule.DestinationPort == ""
}

func observedForRule(rule filter.FirewallRule) filter.ObservedRule {
	canonical := nativeCanonical(rule)
	return filter.ObservedRule{
		Rule: rule, Locator: firewalldLocator(rule.Scope, canonical), ParseStatus: filter.ParseStatusSupported,
		Persistence: filter.PersistenceStatusConverged,
	}
}

func nativeCanonical(rule filter.FirewallRule) string {
	if rule.NativeKind == filter.NativeKindZonePort {
		return "port:" + rule.DestinationPort + "/" + rule.Protocol
	}
	return "rich:" + canonicalRichRule(rule)
}

func pairedCommands(rule filter.FirewallRule, operation, inverse string) ([]filter.NativeCommand, []filter.NativeCommand) {
	option := nativeOption(rule, operation)
	rollback := nativeOption(rule, inverse)
	selector := scopeSelector(rule.Scope)
	commands := []filter.NativeCommand{
		{Executable: "firewall-cmd", Args: []string{selector, option}},
		{Executable: "firewall-cmd", Args: []string{"--permanent", selector, option}},
	}
	rollbackCommands := []filter.NativeCommand{
		{Executable: "firewall-cmd", Args: []string{selector, rollback}},
		{Executable: "firewall-cmd", Args: []string{"--permanent", selector, rollback}},
	}
	return commands, rollbackCommands
}

func nativeOption(rule filter.FirewallRule, operation string) string {
	if rule.NativeKind == filter.NativeKindZonePort {
		return "--" + operation + "-port=" + rule.DestinationPort + "/" + rule.Protocol
	}
	return "--" + operation + "-rich-rule=" + canonicalRichRule(rule)
}

func validateMutationTarget(snapshot filter.Snapshot, change filter.DesiredChange, normalized filter.FirewallRule, requireOwned bool) (filter.ObservedRule, error) {
	if change.Locator == nil || change.Locator.Canonical == "" {
		return filter.ObservedRule{}, fmt.Errorf("%w: firewalld mutation requires canonical locator", filter.ErrInvalidRule)
	}
	if change.Locator.Provider != "" && change.Locator.Provider != filter.ProviderFirewalld {
		return filter.ObservedRule{}, fmt.Errorf("%w: locator provider mismatch", filter.ErrInvalidRule)
	}
	if change.Locator.ScopeKey != "" && change.Locator.ScopeKey != snapshot.Scope.Key() {
		return filter.ObservedRule{}, fmt.Errorf("%w: locator scope mismatch", filter.ErrInvalidRule)
	}
	matches := make([]filter.ObservedRule, 0, 1)
	for _, observed := range snapshot.Rules {
		if observed.Locator.Canonical == change.Locator.Canonical {
			matches = append(matches, observed)
		}
	}
	if len(matches) != 1 {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	target := matches[0]
	if target.Protected {
		return filter.ObservedRule{}, filter.ErrProtectedRule
	}
	if target.ParseStatus != filter.ParseStatusSupported || target.Persistence != filter.PersistenceStatusConverged {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	want := normalized
	if requireOwned {
		if change.Before == nil || change.Before.UUID == "" {
			return filter.ObservedRule{}, fmt.Errorf("%w: managed previous rule is required", filter.ErrInvalidRule)
		}
		prepared, err := (&Adapter{}).PrepareRule(*change.Before)
		if err != nil {
			return filter.ObservedRule{}, err
		}
		want = prepared
	}
	wantKey, wantErr := filter.RuleKey(want)
	targetKey, targetErr := filter.RuleKey(target.Rule)
	if wantErr != nil || targetErr != nil || wantKey != targetKey {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	return target, nil
}

func validateScopeCommand(scope filter.Scope, command filter.NativeCommand) error {
	if command.Executable != "firewall-cmd" {
		return fmt.Errorf("%w: unexpected firewalld executable %q", filter.ErrInvalidRule, command.Executable)
	}
	expected := scopeSelector(scope)
	foundSelector := false
	for _, arg := range command.Args {
		if arg == expected {
			foundSelector = true
		}
		if (strings.HasPrefix(arg, "--zone=") || strings.HasPrefix(arg, "--policy=")) && arg != expected {
			return fmt.Errorf("%w: firewalld command targets another scope", filter.ErrUnsupportedScope)
		}
	}
	if !foundSelector {
		return fmt.Errorf("%w: firewalld command must explicitly target %s", filter.ErrUnsupportedScope, expected)
	}
	return nil
}

func scopeSelector(scope filter.Scope) string {
	return "--zone=" + filter.FirewalldInputZone
}

func (a *Adapter) compensate(ctx context.Context, plan filter.NativeRulePlan, executed int, cause error) error {
	rollbackErr := a.rollback(ctx, plan.Expected.Rule.Scope, plan, executed)
	if rollbackErr != nil {
		return fmt.Errorf("firewalld apply failed: %w; compensation failed: %v", cause, rollbackErr)
	}
	return cause
}

func (a *Adapter) rollback(ctx context.Context, scope filter.Scope, plan filter.NativeRulePlan, executed int) error {
	var rollbackErr error
	for index := executed - 1; index >= 0; index-- {
		if index >= len(plan.RollbackCommands) {
			if rollbackErr == nil {
				rollbackErr = errors.New("firewalld rollback plan is incomplete")
			}
			continue
		}
		command := plan.RollbackCommands[index]
		if err := validateScopeCommand(scope, command); err != nil {
			if rollbackErr == nil {
				rollbackErr = err
			}
			continue
		}
		if err := a.writer.Run(ctx, command); err != nil && rollbackErr == nil {
			rollbackErr = err
		}
	}
	return rollbackErr
}

func countCanonical(snapshot filter.Snapshot, expected *filter.ObservedRule) int {
	if expected == nil {
		return 0
	}
	count := 0
	for _, observed := range snapshot.Rules {
		if observed.Locator.Canonical == expected.Locator.Canonical {
			count++
		}
	}
	return count
}

type zoneOutput struct {
	ports    string
	rich     string
	services string
	active   bool
}

func (a *Adapter) readScope(ctx context.Context, scope filter.Scope, permanent bool) (zoneOutput, error) {
	args := make([]string, 0, 3)
	if permanent {
		args = append(args, "--permanent")
	}
	args = append(args, scopeSelector(scope), "--list-all")
	output, err := a.reader.Read(ctx, args...)
	if err != nil {
		return zoneOutput{}, err
	}
	return parseZoneOutput(output), nil
}

func parseZoneOutput(output string) zoneOutput {
	var parsed zoneOutput
	richRules := make([]string, 0)
	inRichRules := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == filter.FirewalldInputZone+" (active)":
			parsed.active = true
			inRichRules = false
		case strings.HasPrefix(line, "ports:"):
			parsed.ports = strings.TrimSpace(strings.TrimPrefix(line, "ports:"))
			inRichRules = false
		case strings.HasPrefix(line, "services:"):
			parsed.services = strings.TrimSpace(strings.TrimPrefix(line, "services:"))
			inRichRules = false
		case strings.HasPrefix(line, "rich rules:"):
			inRichRules = true
			if value := strings.TrimSpace(strings.TrimPrefix(line, "rich rules:")); value != "" {
				richRules = append(richRules, value)
			}
		case inRichRules && line != "":
			richRules = append(richRules, line)
		}
	}
	parsed.rich = strings.Join(richRules, "\n")
	return parsed
}

func publicZoneNotices(runtime, permanent zoneOutput) []filter.ScopeNotice {
	notices := make([]filter.ScopeNotice, 0, 2)
	if !runtime.active {
		notices = append(notices, filter.ScopeNotice{Code: filter.ScopeNoticeManagedScopeInactive})
	}
	mismatched := make([]string, 0, 3)
	if !sameFields(runtime.ports, permanent.ports) {
		mismatched = append(mismatched, "ports")
	}
	if !sameLines(runtime.rich, permanent.rich) {
		mismatched = append(mismatched, "rich_rules")
	}
	if !sameFields(runtime.services, permanent.services) {
		mismatched = append(mismatched, "services")
	}
	if len(mismatched) != 0 {
		notices = append(notices, filter.ScopeNotice{Code: filter.ScopeNoticeRuntimePermanentMismatch, Values: mismatched})
	}
	return notices
}

func (a *Adapter) NativeDetail(ctx context.Context, service string, permanent bool) (string, error) {
	service = strings.TrimSpace(service)
	if !validServiceName(service) {
		return "", fmt.Errorf("%w: invalid firewalld service name", filter.ErrInvalidRule)
	}
	if a.reader == nil {
		return "", errors.New("firewalld reader is required")
	}
	args := make([]string, 0, 3)
	if permanent {
		args = append(args, "--permanent")
	}
	args = append(args, "--info-service="+service)
	info, err := a.reader.Read(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("read firewalld service %q: %w", service, err)
	}
	info = strings.TrimSpace(info)
	if info == "" {
		return "", fmt.Errorf("read firewalld service %q: empty output", service)
	}
	return info, nil
}

func validServiceName(service string) bool {
	if service == "" {
		return false
	}
	for _, char := range service {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.':
		default:
			return false
		}
	}
	return true
}

type mergedObject struct {
	rule      filter.ObservedRule
	runtime   bool
	permanent bool
}

func mergeZoneObjects(scope filter.Scope, runtime, permanent zoneOutput) ([]filter.ObservedRule, error) {
	objects := make(map[string]*mergedObject)
	add := func(rule filter.ObservedRule, runtimeState bool) {
		key := string(rule.Rule.NativeKind) + "\x00" + rule.Locator.Canonical
		object, exists := objects[key]
		if !exists {
			copy := rule
			object = &mergedObject{rule: copy}
			objects[key] = object
		}
		if runtimeState {
			object.runtime = true
		} else {
			object.permanent = true
		}
	}
	parse := func(output zoneOutput, runtimeState bool) error {
		for _, rule := range parseZonePorts(scope, output.ports) {
			add(rule, runtimeState)
		}
		for _, raw := range nonEmptyLines(output.rich) {
			rule, family, supported := parseRichRule(scope, raw)
			if !supported {
				opaqueScope := scope
				if family == filter.FamilyIPv4 || family == filter.FamilyIPv6 {
					opaqueScope.Family = family
				}
				rule = opaqueRichRule(opaqueScope, raw)
			}
			add(rule, runtimeState)
		}
		for _, service := range strings.Fields(output.services) {
			if scope.Family == filter.FamilyInet {
				add(opaqueZoneService(scope, service), runtimeState)
			}
		}
		return nil
	}
	if err := parse(runtime, true); err != nil {
		return nil, err
	}
	if err := parse(permanent, false); err != nil {
		return nil, err
	}

	rules := make([]filter.ObservedRule, 0, len(objects))
	for _, object := range objects {
		switch {
		case object.runtime && object.permanent:
			object.rule.Persistence = filter.PersistenceStatusConverged
		case object.runtime:
			object.rule.Persistence = filter.PersistenceStatusRuntimeOnly
		default:
			object.rule.Persistence = filter.PersistenceStatusPermanentOnly
		}
		rules = append(rules, object.rule)
	}
	sort.SliceStable(rules, func(i, j int) bool {
		left, right := rules[i], rules[j]
		leftRank, rightRank := bucketRank(left.Rule.OrderBucket), bucketRank(right.Rule.OrderBucket)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if left.Rule.Priority != nil && right.Rule.Priority != nil && *left.Rule.Priority != *right.Rule.Priority {
			return *left.Rule.Priority < *right.Rule.Priority
		}
		return left.Locator.Canonical < right.Locator.Canonical
	})
	return rules, nil
}

func parseZonePorts(scope filter.Scope, output string) []filter.ObservedRule {
	rules := make([]filter.ObservedRule, 0)
	for _, value := range strings.Fields(output) {
		port, protocol, ok := strings.Cut(value, "/")
		canonical := "port:" + strings.TrimSpace(value)
		if !ok {
			rules = append(rules, opaqueZoneObject(scope, filter.NativeKindZonePort, canonical, value))
			continue
		}
		rule, err := filter.NormalizeRule(filter.FirewallRule{
			Scope: scope, NativeKind: filter.NativeKindZonePort, Protocol: protocol,
			DestinationPort: port, Action: filter.ActionAccept,
		})
		if err != nil {
			rules = append(rules, opaqueZoneObject(scope, filter.NativeKindZonePort, canonical, value))
			continue
		}
		canonical = "port:" + rule.DestinationPort + "/" + rule.Protocol
		rules = append(rules, filter.ObservedRule{
			Rule: rule, Locator: firewalldLocator(scope, canonical), ParseStatus: filter.ParseStatusSupported, Raw: value,
		})
	}
	return rules
}

func parseRichRule(baseScope filter.Scope, raw string) (filter.ObservedRule, filter.Family, bool) {
	tokens, err := shellwords.Parse(raw)
	if err != nil || len(tokens) < 2 || tokens[0] != "rule" {
		return filter.ObservedRule{}, richRuleFamily(raw), false
	}
	family := filter.FamilyInet
	priority := 0
	rule := filter.FirewallRule{NativeKind: filter.NativeKindRichRule, Protocol: "all"}
	section := ""
	seenAction := false
	seenSource, seenDestination, seenPort, seenProtocol := false, false, false, false
	seenFamily, seenPriority := false, false
	for _, token := range tokens[1:] {
		if seenAction {
			return filter.ObservedRule{}, family, false
		}
		switch {
		case token == "source":
			if seenSource {
				return filter.ObservedRule{}, family, false
			}
			seenSource = true
			section = token
		case token == "destination":
			if seenDestination {
				return filter.ObservedRule{}, family, false
			}
			seenDestination = true
			section = token
		case token == "port":
			if seenPort || seenProtocol {
				return filter.ObservedRule{}, family, false
			}
			seenPort = true
			section = token
		case token == "protocol":
			if seenProtocol || seenPort {
				return filter.ObservedRule{}, family, false
			}
			seenProtocol = true
			section = token
		case strings.HasPrefix(token, "family="):
			if seenFamily {
				return filter.ObservedRule{}, family, false
			}
			seenFamily = true
			value := strings.TrimPrefix(token, "family=")
			family = filter.Family(strings.ToLower(value))
			if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
				return filter.ObservedRule{}, family, false
			}
			section = ""
		case strings.HasPrefix(token, "priority="):
			if seenPriority {
				return filter.ObservedRule{}, family, false
			}
			seenPriority = true
			value, parseErr := strconv.Atoi(strings.TrimPrefix(token, "priority="))
			if parseErr != nil || value < -32768 || value > 32767 {
				return filter.ObservedRule{}, family, false
			}
			priority = value
			section = ""
		case strings.HasPrefix(token, "address=") && section == "source":
			if rule.SourceAddress != "" {
				return filter.ObservedRule{}, family, false
			}
			rule.SourceAddress = strings.TrimPrefix(token, "address=")
		case strings.HasPrefix(token, "address=") && section == "destination":
			if rule.DestinationAddress != "" {
				return filter.ObservedRule{}, family, false
			}
			rule.DestinationAddress = strings.TrimPrefix(token, "address=")
		case strings.HasPrefix(token, "port=") && section == "port":
			if rule.DestinationPort != "" {
				return filter.ObservedRule{}, family, false
			}
			rule.DestinationPort = strings.TrimPrefix(token, "port=")
		case strings.HasPrefix(token, "protocol=") && section == "port":
			if rule.Protocol != "all" {
				return filter.ObservedRule{}, family, false
			}
			rule.Protocol = strings.TrimPrefix(token, "protocol=")
		case strings.HasPrefix(token, "value=") && section == "protocol":
			if rule.Protocol != "all" {
				return filter.ObservedRule{}, family, false
			}
			rule.Protocol = strings.TrimPrefix(token, "value=")
		case token == "accept" || token == "drop" || token == "reject":
			if seenAction {
				return filter.ObservedRule{}, family, false
			}
			rule.Action = filter.Action(token)
			seenAction = true
			section = ""
		default:
			return filter.ObservedRule{}, family, false
		}
	}
	if !seenAction || (seenSource && rule.SourceAddress == "") ||
		(seenDestination && rule.DestinationAddress == "") ||
		(seenPort && (rule.DestinationPort == "" || rule.Protocol == "all")) ||
		(seenProtocol && rule.Protocol == "all") {
		return filter.ObservedRule{}, family, false
	}
	if family == filter.FamilyInet && (rule.SourceAddress != "" || rule.DestinationAddress != "") {
		return filter.ObservedRule{}, family, false
	}
	scope := baseScope
	scope.Family = family
	rule.Scope = scope
	rule.Priority = &priority
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return filter.ObservedRule{}, family, false
	}
	canonical := canonicalRichRule(normalized)
	return filter.ObservedRule{
		Rule: normalized, Locator: firewalldLocator(scope, "rich:"+canonical), ParseStatus: filter.ParseStatusSupported, Raw: strings.TrimSpace(raw),
	}, family, true
}

func richRuleFamily(raw string) filter.Family {
	tokens, err := shellwords.Parse(raw)
	if err != nil {
		return filter.FamilyInet
	}
	for _, token := range tokens {
		if strings.HasPrefix(token, "family=") {
			return filter.Family(strings.ToLower(strings.TrimPrefix(token, "family=")))
		}
	}
	return filter.FamilyInet
}

func canonicalRichRule(rule filter.FirewallRule) string {
	parts := []string{"rule"}
	if rule.Scope.Family != filter.FamilyInet {
		parts = append(parts, `family=`+strconv.Quote(string(rule.Scope.Family)))
	}
	if rule.Priority != nil && *rule.Priority != 0 {
		parts = append(parts, `priority=`+strconv.Quote(strconv.Itoa(*rule.Priority)))
	}
	if rule.SourceAddress != "" {
		parts = append(parts, "source", `address=`+strconv.Quote(rule.SourceAddress))
	}
	if rule.DestinationAddress != "" {
		parts = append(parts, "destination", `address=`+strconv.Quote(rule.DestinationAddress))
	}
	if rule.DestinationPort != "" {
		parts = append(parts, "port", `port=`+strconv.Quote(rule.DestinationPort), `protocol=`+strconv.Quote(rule.Protocol))
	} else if rule.Protocol != "all" {
		parts = append(parts, "protocol", `value=`+strconv.Quote(rule.Protocol))
	}
	parts = append(parts, string(rule.Action))
	return strings.Join(parts, " ")
}

func opaqueRichRule(scope filter.Scope, raw string) filter.ObservedRule {
	canonical := "rich:" + strings.TrimSpace(raw)
	rule := opaqueZoneObject(scope, filter.NativeKindRichRule, canonical, raw)
	priority := 0
	action := filter.ActionAccept
	if tokens, err := shellwords.Parse(raw); err == nil {
		for _, token := range tokens {
			if strings.HasPrefix(token, "priority=") {
				if parsed, parseErr := strconv.Atoi(strings.TrimPrefix(token, "priority=")); parseErr == nil && parsed >= -32768 && parsed <= 32767 {
					priority = parsed
				}
			}
			if token == "accept" || token == "drop" || token == "reject" {
				action = filter.Action(token)
			}
		}
	}
	rule.Rule.Priority = &priority
	rule.Rule.Action = action
	rule.Rule.OrderBucket = richOrderBucket(priority, action)
	return rule
}

func opaqueZoneService(scope filter.Scope, service string) filter.ObservedRule {
	rule := opaqueZoneObject(scope, filter.NativeKindZoneService, "service:"+service, service)
	rule.Rule.Protocol = ""
	rule.Rule.Description = service
	return rule
}

func opaqueZoneObject(scope filter.Scope, kind filter.NativeKind, canonical, raw string) filter.ObservedRule {
	return filter.ObservedRule{
		Rule: filter.FirewallRule{
			Scope: scope, NativeKind: kind, Protocol: "all", Action: filter.ActionAccept,
			OrderBucket: filter.OrderBucketZonePrimitiveAllow,
		},
		Locator: firewalldLocator(scope, canonical), ParseStatus: filter.ParseStatusOpaque, Raw: strings.TrimSpace(raw),
	}
}

func firewalldLocator(scope filter.Scope, canonical string) filter.Locator {
	return filter.Locator{Provider: filter.ProviderFirewalld, ScopeKey: scope.Key(), NativeID: canonical, Canonical: canonical}
}

func nonEmptyLines(output string) []string {
	lines := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func sameFields(left, right string) bool {
	return sameStringSet(strings.Fields(left), strings.Fields(right))
}

func sameLines(left, right string) bool {
	return sameStringSet(nonEmptyLines(left), nonEmptyLines(right))
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func bucketRank(bucket string) int {
	switch bucket {
	case filter.OrderBucketRichPre:
		return 0
	case filter.OrderBucketRichZeroDeny:
		return 1
	case filter.OrderBucketZonePrimitiveAllow, filter.OrderBucketRichZeroAllow:
		return 2
	case filter.OrderBucketRichPost:
		return 3
	default:
		return 4
	}
}

func richOrderBucket(priority int, action filter.Action) string {
	if priority < 0 {
		return filter.OrderBucketRichPre
	}
	if priority > 0 {
		return filter.OrderBucketRichPost
	}
	if action == filter.ActionDrop || action == filter.ActionReject {
		return filter.OrderBucketRichZeroDeny
	}
	return filter.OrderBucketRichZeroAllow
}

type systemBackend struct{}

func (systemBackend) Read(ctx context.Context, args ...string) (string, error) {
	return cmd.NewCommandMgr(
		cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second), cmd.WithEnv("LANGUAGE=en_US:en"),
	).RunWithStdout("firewall-cmd", args...)
}

func (systemBackend) Run(ctx context.Context, command filter.NativeCommand) error {
	if err := validateSystemCommand(command); err != nil {
		return err
	}
	return cmd.NewCommandMgr(
		cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second), cmd.WithEnv("LANGUAGE=en_US:en"),
	).RunWithOptionalSudo(command.Executable, command.Args...)
}

func validateSystemCommand(command filter.NativeCommand) error {
	for _, arg := range command.Args {
		switch {
		case arg == "--zone="+filter.FirewalldInputZone:
			return validateScopeCommand(filter.Scope{Provider: filter.ProviderFirewalld, Family: filter.FamilyInet, Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput}, command)
		}
	}
	return fmt.Errorf("%w: firewalld command must target the managed input zone", filter.ErrUnsupportedScope)
}
