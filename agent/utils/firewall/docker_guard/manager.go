package docker_guard

import (
	"context"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	firewallutil "github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
)

type Runner interface {
	Run(executable string, args ...string) (string, error)
	RunInput(executable, input string, args ...string) (string, error)
	Exists(executable string) bool
}

type commandRunner struct{ ctx context.Context }

func (r commandRunner) Run(executable string, args ...string) (string, error) {
	executable = dockerGuardExecutable(executable)
	manager := cmd.NewCommandMgr(cmd.WithContext(r.ctx), cmd.WithTimeout(60*time.Second))
	stdout, err := manager.RunWithOptionalSudoAndStdout(executable, args...)
	err = errors.Join(err, r.ctx.Err())
	if err != nil {
		return stdout, fmt.Errorf("command=%s %s failed: %w", executable, strings.Join(args, " "), err)
	}
	return stdout, nil
}

func (r commandRunner) RunInput(executable, input string, args ...string) (string, error) {
	executable = dockerGuardExecutable(executable)
	if executable == "nft" {
		return "", errors.Join(nftables_helper.RunScriptContext(r.ctx, input), r.ctx.Err())
	}
	manager := cmd.NewCommandMgr(cmd.WithContext(r.ctx), cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(input)))
	stdout, err := manager.RunWithOptionalSudoAndStdout(executable, args...)
	return stdout, errors.Join(firewallutil.WrapBatchCommandError(executable+" "+strings.Join(args, " "), input, err), r.ctx.Err())
}

func (commandRunner) Exists(executable string) bool {
	resolved := dockerGuardExecutable(executable)
	return resolved != "" && cmd.Which(resolved)
}

func dockerGuardExecutable(logical string) string {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return logical
	}
	switch logical {
	case "iptables":
		return commands.IPv4
	case "iptables-restore":
		return commands.Restore4
	case "ip6tables":
		return commands.IPv6
	case "ip6tables-restore":
		return commands.Restore6
	default:
		return logical
	}
}

type Iptables struct {
	runner Runner
}

var mutationMu sync.Mutex

func NewIptables(ctx context.Context) *Iptables { return &Iptables{runner: commandRunner{ctx: ctx}} }

func (m *Iptables) Initialize(policies []Policy, inventory PolicyInventory, families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if err := CheckIPv4Forwarding(); err != nil {
		return err
	}
	if !m.runner.Exists("iptables-restore") {
		return errors.New("iptables-restore is not installed")
	}
	if err := m.ensureFamily("iptables", true); err != nil {
		return err
	}
	if slices.Contains(families, FamilyIPv6) && m.runner.Exists("ip6tables") {
		available, err := m.chainExists("ip6tables", DockerChain)
		if err != nil {
			return &FamilyError{Family: FamilyIPv6, Err: fmt.Errorf("inspect %s chain: %w", DockerChain, err)}
		}
		if available {
			if !m.runner.Exists("ip6tables-restore") {
				return &FamilyError{Family: FamilyIPv6, Err: errors.New("ip6tables-restore is not installed")}
			}
			if err := m.ensureFamily("ip6tables", false); err != nil {
				return &FamilyError{Family: FamilyIPv6, Err: err}
			}
		}
	}
	return m.rebuildLocked(policies, inventory, families...)
}

func (m *Iptables) Bind(families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if err := m.bindExistingFamily("iptables", true); err != nil {
		return err
	}
	if slices.Contains(families, FamilyIPv6) && m.runner.Exists("ip6tables") {
		if err := m.bindExistingFamily("ip6tables", false); err != nil {
			return &FamilyError{Family: FamilyIPv6, Err: err}
		}
	}
	return nil
}

func (m *Iptables) ReplacePolicies(policies []Policy, inventory PolicyInventory) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	return m.rebuildLocked(policies, inventory)
}

func (m *Iptables) ListPolicies() (PolicyInventory, error) {
	inventory := PolicyInventory{Policies: make([]Policy, 0), RuleOrders: make(map[string][]int64)}
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		executable := executableForFamily(family)
		if executable == "" || !m.runner.Exists(executable) {
			continue
		}
		exists, err := m.chainExists(executable, Chain)
		if family == FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return PolicyInventory{}, &FamilyError{Family: family, Err: fmt.Errorf("inspect %s chain: %w", Chain, err)}
		}
		if !exists {
			continue
		}
		output, err := m.run(executable, "-S", Chain)
		if err != nil {
			return PolicyInventory{}, &FamilyError{Family: family, Err: fmt.Errorf("list %s chain: %w", Chain, err)}
		}
		parsed, err := parseDockerGuardPolicies(output, family)
		if err != nil {
			return PolicyInventory{}, &FamilyError{Family: family, Err: err}
		}
		inventory.Policies = append(inventory.Policies, parsed.Policies...)
		for key, orders := range parsed.RuleOrders {
			inventory.RuleOrders[key] = append([]int64(nil), orders...)
		}
	}
	return inventory, nil
}

func (m *Iptables) Unbind(families ...string) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	for _, family := range families {
		if err := m.unbindFamily(executableForFamily(family)); err != nil {
			return err
		}
	}
	return nil
}

func (m *Iptables) Cleanup() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	for _, executable := range []string{"iptables", "ip6tables"} {
		if !m.runner.Exists(executable) {
			continue
		}
		output, err := m.run(executable, "-S")
		if err != nil {
			return err
		}
		rules := dockerGuardLifecycleRules(output, false, false)
		if len(rules) == 0 {
			continue
		}
		if err := m.restoreLifecycle(executable, rules); err != nil {
			return err
		}
	}
	return nil
}

func (m *Iptables) Initialized(family string) (bool, error) {
	executable := executableForFamily(family)
	if executable == "" || !m.runner.Exists(executable) {
		return false, nil
	}
	return m.chainExists(executable, Chain)
}

func (m *Iptables) Status(family string) FamilyStatus {
	executable := executableForFamily(family)
	if executable == "" || !m.runner.Exists(executable) {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonCommandMissing}
	}
	chains, err := m.run(executable, "-S")
	if err != nil {
		return FamilyStatus{State: StatusNotEffective, Reason: ReasonInspectFailed}
	}
	if !chainDeclared(chains, DockerChain) {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonDockerChainMissing}
	}
	if !chainDeclared(chains, Chain) {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonGuardChainMissing}
	}
	status := FamilyStatus{State: StatusNotEffective, Initialized: true}
	jumps := countJumps(chains)
	if jumps == 0 {
		status.Reason = ReasonJumpMissing
		return status
	}
	if jumps > 1 {
		status.Reason = ReasonJumpDuplicate
		return status
	}
	if !hasFirstUniqueJump(chains) {
		status.Reason = ReasonJumpNotFirst
		return status
	}
	status.State = StatusEffective
	status.Bound = true
	status.Effective = true
	return status
}

func (m *Iptables) bindExistingFamily(executable string, required bool) error {
	if !m.runner.Exists(executable) {
		if required {
			return fmt.Errorf("%s is not installed", executable)
		}
		return nil
	}
	output, err := m.run(executable, "-S")
	if err != nil {
		return err
	}
	if !chainDeclared(output, DockerChain) {
		if required {
			return buserr.New("ErrDockerIptablesChainUnavailable")
		}
		return nil
	}
	if !chainDeclared(output, Chain) {
		if required {
			return fmt.Errorf("%s chain is not initialized for %s", Chain, executable)
		}
		return nil
	}
	return m.restoreLifecycle(executable, dockerGuardLifecycleRules(output, true, false))
}

func (m *Iptables) ensureFamily(executable string, required bool) error {
	if !m.runner.Exists(executable) {
		if required {
			return fmt.Errorf("%s is not installed", executable)
		}
		return nil
	}
	output, err := m.run(executable, "-S")
	if err != nil {
		return err
	}
	if !chainDeclared(output, DockerChain) {
		if required {
			return buserr.New("ErrDockerIptablesChainUnavailable")
		}
		return nil
	}
	return m.restoreLifecycle(executable, dockerGuardLifecycleRules(output, true, !chainDeclared(output, Chain)))
}

func dockerGuardLifecycleRules(output string, bind, createOwned bool) [][]string {
	rules := make([][]string, 0, countJumps(output)+3)
	if createOwned {
		rules = append(rules, []string{"-N", Chain})
	}
	for i := 0; i < countJumps(output); i++ {
		rules = append(rules, []string{"-D", DockerChain, "-j", Chain})
	}
	if bind {
		rules = append(rules, []string{"-I", DockerChain, "1", "-j", Chain})
	} else if chainDeclared(output, Chain) {
		rules = append(rules, []string{"-F", Chain}, []string{"-X", Chain})
	}
	return rules
}

func (m *Iptables) restoreLifecycle(executable string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}
	restoreExecutable := executable + "-restore"
	if !m.runner.Exists(restoreExecutable) {
		return fmt.Errorf("%s is not installed", restoreExecutable)
	}
	script, err := buildRestoreScript(rules)
	if err != nil {
		return err
	}
	if _, err := m.runner.RunInput(restoreExecutable, script, "--noflush", "--wait"); err != nil {
		return fmt.Errorf("batch update Docker guard lifecycle: %w", err)
	}
	return nil
}

func (m *Iptables) rebuildLocked(policies []Policy, inventory PolicyInventory, families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	for _, family := range families {
		executable := executableForFamily(family)
		if executable == "" || !m.runner.Exists(executable) {
			continue
		}
		exists, err := m.chainExists(executable, Chain)
		if family == FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) && !slices.ContainsFunc(policies, func(policy Policy) bool { return policy.Family == family }) {
			continue
		}
		if err != nil {
			return &FamilyError{Family: family, Err: fmt.Errorf("inspect %s chain: %w", Chain, err)}
		}
		if !exists {
			continue
		}
		rules := [][]string{{"-F", Chain}, {"-A", Chain, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "RETURN"}}
		rules = append(rules, orderedIPTablesRules(family, policies, inventory)...)
		rules = append(rules, []string{"-A", Chain, "-j", "RETURN"})
		script, err := buildRestoreScript(rules)
		if err != nil {
			return err
		}
		restoreExecutable := executable + "-restore"
		if !m.runner.Exists(restoreExecutable) {
			return &FamilyError{Family: family, Err: fmt.Errorf("%s is not installed", restoreExecutable)}
		}
		if _, err := m.runner.RunInput(restoreExecutable, script, "--noflush", "--wait"); err != nil {
			return &FamilyError{Family: family, Err: fmt.Errorf("restore rules: %w", err)}
		}
	}
	return nil
}

func orderedIPTablesRules(family string, policies []Policy, inventory PolicyInventory) [][]string {
	segments := make([]orderedPolicyRule, 0, len(policies))
	maxOrder := int64(0)
	for _, orders := range inventory.RuleOrders {
		for _, order := range orders {
			maxOrder = max(maxOrder, order)
		}
	}
	for _, policy := range policies {
		for _, rule := range policy.NativeRules {
			maxOrder = max(maxOrder, rule.Order)
		}
	}
	for _, policy := range policies {
		if policy.Family != family {
			continue
		}
		if len(policy.NativeRules) > 0 {
			for _, native := range policy.NativeRules {
				if native.Family != family || len(native.Tokens) < 2 || native.Tokens[0] != "-A" || native.Tokens[1] != Chain {
					continue
				}
				segments = append(segments, orderedPolicyRule{order: native.Order, rule: append([]string(nil), native.Tokens...)})
			}
			continue
		}
		compiled := compilePolicy(policy)
		orders := inventory.RuleOrders[policy.Family+"\x00"+policy.UUID]
		for ruleIndex, rule := range compiled {
			order := int64(0)
			if ruleIndex < len(orders) {
				order = orders[ruleIndex]
			} else {
				maxOrder++
				order = maxOrder
			}
			segments = append(segments, orderedPolicyRule{order: order, rule: rule})
		}
	}
	return sortPolicyRules(segments)
}

func buildRestoreScript(rules [][]string) (string, error) {
	var script strings.Builder
	script.WriteString("*filter\n")
	for _, rule := range rules {
		for i, token := range rule {
			if token == "" || strings.ContainsAny(token, "\r\n\x00") {
				return "", fmt.Errorf("invalid iptables-restore token %q", token)
			}
			if i > 0 {
				script.WriteByte(' ')
			}
			if strings.ContainsAny(token, " \t\\\"'") {
				script.WriteString(strconv.Quote(token))
			} else {
				script.WriteString(token)
			}
		}
		script.WriteByte('\n')
	}
	script.WriteString("COMMIT\n")
	return script.String(), nil
}

func compilePolicy(policy Policy) [][]string {
	base := []string{"-A", Chain, "-p", policy.Protocol, "-m", "conntrack"}
	if !isWildcardHost(policy.Family, policy.HostIP) {
		base = append(base, "--ctorigdst", policy.HostIP)
	}
	base = append(base, "--ctorigdstport", strconv.Itoa(int(policy.HostPort)))
	comment := "1panel-docker:" + policy.UUID
	if policy.Mode == ModeAll || policy.Mode == ModeAcceptAll {
		target := "DROP"
		if policy.Mode == ModeAcceptAll {
			target = "ACCEPT"
		}
		return [][]string{append(append([]string{}, base...), "-m", "comment", "--comment", comment, "-j", target)}
	}
	target := "DROP"
	capacity := len(policy.Sources)
	if policy.Mode == ModeAllow {
		target = "RETURN"
		capacity++
	} else if policy.Mode == ModeAcceptSources {
		target = "ACCEPT"
	}
	rules := make([][]string, 0, capacity)
	for _, source := range policy.Sources {
		args := append([]string{}, base...)
		args = append(args, "-s", source, "-m", "comment", "--comment", comment, "-j", target)
		rules = append(rules, args)
	}
	if policy.Mode == ModeAllow {
		rules = append(rules, append(append([]string{}, base...), "-m", "comment", "--comment", comment, "-j", "DROP"))
	}
	return rules
}

func (m *Iptables) unbindFamily(executable string) error {
	if !m.runner.Exists(executable) {
		return nil
	}
	if output, err := m.run(executable, "-S", DockerChain); err == nil {
		rules := make([][]string, 0, countJumps(output))
		for i := 0; i < countJumps(output); i++ {
			rules = append(rules, []string{"-D", DockerChain, "-j", Chain})
		}
		return m.restoreLifecycle(executable, rules)
	}
	return nil
}

func (m *Iptables) chainExists(executable, chain string) (bool, error) {
	output, err := m.run(executable, "-S")
	if err != nil {
		return false, err
	}
	return chainDeclared(output, chain), nil
}

func chainDeclared(output, chain string) bool {
	needle := "-N " + chain
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == needle {
			return true
		}
	}
	return false
}

func (m *Iptables) run(executable string, args ...string) (string, error) {
	commandArgs := append([]string{"-w", "-t", "filter"}, args...)
	output, err := m.runner.Run(executable, commandArgs...)
	if executable == "ip6tables" && err != nil && (strings.Contains(err.Error(), "Address family not supported") || strings.Contains(err.Error(), "Protocol not supported")) {
		return output, fmt.Errorf("%w: %v", filter.ErrFamilyUnavailable, err)
	}
	return output, err
}

func executableForFamily(family string) string {
	switch family {
	case FamilyIPv4:
		return "iptables"
	case FamilyIPv6:
		return "ip6tables"
	default:
		return ""
	}
}

func countJumps(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == "-A" && fields[1] == DockerChain && fields[2] == "-j" && fields[3] == Chain {
			count++
		}
	}
	return count
}

func hasFirstUniqueJump(output string) bool {
	firstRule := ""
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-A "+DockerChain+" ") {
			if firstRule == "" {
				firstRule = line
			}
		}
	}
	return firstRule == "-A "+DockerChain+" -j "+Chain && countJumps(output) == 1
}

func isWildcardHost(family, hostIP string) bool {
	return (family == FamilyIPv4 && (hostIP == "" || hostIP == "0.0.0.0")) ||
		(family == FamilyIPv6 && (hostIP == "" || hostIP == "::"))
}

func (m *Iptables) OperateFamily(family string, initialize bool) error {
	if family != FamilyIPv4 && family != FamilyIPv6 {
		return fmt.Errorf("unsupported Docker firewall family %q", family)
	}
	if family == FamilyIPv4 {
		if err := CheckIPv4Forwarding(); err != nil {
			return err
		}
	}
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if initialize {
		return m.ensureFamily(executableForFamily(family), true)
	}
	return m.bindExistingFamily(executableForFamily(family), true)
}
