package docker_guard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

const (
	Chain       = "1PANEL_DOCKER"
	DockerChain = "DOCKER-USER"
	FamilyIPv4  = constant.FirewallFamilyIPv4
	FamilyIPv6  = constant.FirewallFamilyIPv6
	ModeSources = "deny_sources"
	ModeAllow   = "allow_sources"
	ModeAll     = "deny_all"

	StatusEffective    = "effective"
	StatusDisabled     = "disabled"
	StatusNotEffective = "not_effective"

	ReasonCommandMissing     = "command_missing"
	ReasonDockerChainMissing = "docker_chain_missing"
	ReasonGuardChainMissing  = "guard_chain_missing"
	ReasonJumpMissing        = "jump_missing"
	ReasonJumpNotFirst       = "jump_not_first"
	ReasonJumpDuplicate      = "jump_duplicate"
	ReasonInspectFailed      = "inspect_failed"
)

var ErrDockerChainUnavailable = errors.New("Docker DOCKER-USER chain is unavailable")

type FamilyError struct {
	Family string
	Err    error
}

func (e *FamilyError) Error() string { return fmt.Sprintf("%s Docker port guard: %v", e.Family, e.Err) }
func (e *FamilyError) Unwrap() error { return e.Err }

type Policy struct {
	UUID     string
	Family   string
	HostIP   string
	HostPort uint16
	Protocol string
	Mode     string
	Sources  []string
}

type FamilyStatus struct {
	State       string
	Reason      string
	Initialized bool
	Bound       bool
	Effective   bool
}

type Runner interface {
	Run(executable string, args ...string) (string, error)
	RunInput(executable, input string, args ...string) (string, error)
	Exists(executable string) bool
}

type commandRunner struct{}

func (commandRunner) Run(executable string, args ...string) (string, error) {
	executable = dockerGuardExecutable(executable)
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60 * time.Second))
	return manager.RunWithOptionalSudoAndStdout(executable, args...)
}

func (commandRunner) RunInput(executable, input string, args ...string) (string, error) {
	executable = dockerGuardExecutable(executable)
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(input)))
	return manager.RunWithOptionalSudoAndStdout(executable, args...)
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

type Manager struct {
	runner Runner
}

var mutationMu sync.Mutex

func NewManager() *Manager { return &Manager{runner: commandRunner{}} }

func NewManagerWithRunner(runner Runner) *Manager { return &Manager{runner: runner} }

func (m *Manager) Initialize(policies []Policy) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if !m.runner.Exists("iptables-restore") {
		return errors.New("iptables-restore is not installed")
	}
	if err := m.ensureFamily("iptables", true); err != nil {
		return err
	}
	if m.runner.Exists("ip6tables") {
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
	return m.rebuildLocked(policies)
}

func (m *Manager) Bind() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if err := m.bindExistingFamily("iptables", true); err != nil {
		return err
	}
	if m.runner.Exists("ip6tables") {
		if err := m.bindExistingFamily("ip6tables", false); err != nil {
			return &FamilyError{Family: FamilyIPv6, Err: err}
		}
	}
	return nil
}

func (m *Manager) Reconcile(policies []Policy) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	return m.rebuildLocked(policies)
}

func (m *Manager) ListPolicies() ([]Policy, error) {
	policies := make([]Policy, 0)
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		executable := executableForFamily(family)
		if executable == "" || !m.runner.Exists(executable) {
			continue
		}
		exists, err := m.chainExists(executable, Chain)
		if err != nil {
			return nil, &FamilyError{Family: family, Err: fmt.Errorf("inspect %s chain: %w", Chain, err)}
		}
		if !exists {
			continue
		}
		output, err := m.run(executable, "-S", Chain)
		if err != nil {
			return nil, &FamilyError{Family: family, Err: fmt.Errorf("list %s chain: %w", Chain, err)}
		}
		parsed, err := parseDockerGuardPolicies(output, family)
		if err != nil {
			return nil, &FamilyError{Family: family, Err: err}
		}
		policies = append(policies, parsed...)
	}
	return policies, nil
}

func (m *Manager) Unbind() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if err := m.unbindFamily("iptables"); err != nil {
		return err
	}
	if m.runner.Exists("ip6tables") {
		if err := m.unbindFamily("ip6tables"); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Cleanup() error {
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
		rules := dockerGuardLifecycleRules(output, false, chainDeclared(output, Chain))
		if len(rules) == 0 {
			continue
		}
		if err := m.restoreLifecycle(executable, rules); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Initialized(family string) (bool, error) {
	executable := executableForFamily(family)
	if executable == "" || !m.runner.Exists(executable) {
		return false, nil
	}
	return m.chainExists(executable, Chain)
}

func (m *Manager) Status(family string) FamilyStatus {
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
	rules, err := m.run(executable, "-S", DockerChain)
	if err != nil {
		status.Reason = ReasonInspectFailed
		return status
	}
	jumps := countJumps(rules)
	if jumps == 0 {
		status.Reason = ReasonJumpMissing
		return status
	}
	if jumps > 1 {
		status.Reason = ReasonJumpDuplicate
		return status
	}
	if !hasFirstUniqueJump(rules) {
		status.Reason = ReasonJumpNotFirst
		return status
	}
	status.State = StatusEffective
	status.Bound = true
	status.Effective = true
	return status
}

func (m *Manager) bindExistingFamily(executable string, required bool) error {
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
			return fmt.Errorf("%w for %s", ErrDockerChainUnavailable, executable)
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

func (m *Manager) ensureFamily(executable string, required bool) error {
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
			return fmt.Errorf("%w for %s", ErrDockerChainUnavailable, executable)
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

func (m *Manager) restoreLifecycle(executable string, rules [][]string) error {
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

func (m *Manager) rebuildLocked(policies []Policy) error {
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		executable := executableForFamily(family)
		if executable == "" || !m.runner.Exists(executable) {
			continue
		}
		exists, err := m.chainExists(executable, Chain)
		if err != nil {
			return &FamilyError{Family: family, Err: fmt.Errorf("inspect %s chain: %w", Chain, err)}
		}
		if !exists {
			continue
		}
		rules := [][]string{{"-F", Chain}, {"-A", Chain, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "RETURN"}}
		for _, policy := range policies {
			if policy.Family != family {
				continue
			}
			rules = append(rules, compilePolicy(policy)...)
		}
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

func buildRestoreScript(rules [][]string) (string, error) {
	var script strings.Builder
	script.WriteString("*filter\n")
	for _, rule := range rules {
		for i, token := range rule {
			if token == "" || strings.ContainsAny(token, " \t\r\n\\\"'") {
				return "", fmt.Errorf("invalid iptables-restore token %q", token)
			}
			if i > 0 {
				script.WriteByte(' ')
			}
			script.WriteString(token)
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
	if policy.Mode == ModeAll {
		return [][]string{append(append([]string{}, base...), "-m", "comment", "--comment", comment, "-j", "DROP")}
	}
	target := "DROP"
	capacity := len(policy.Sources)
	if policy.Mode == ModeAllow {
		target = "RETURN"
		capacity++
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

func (m *Manager) unbindFamily(executable string) error {
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

func (m *Manager) chainExists(executable, chain string) (bool, error) {
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

func (m *Manager) run(executable string, args ...string) (string, error) {
	commandArgs := append([]string{"-w", "-t", "filter"}, args...)
	return m.runner.Run(executable, commandArgs...)
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
