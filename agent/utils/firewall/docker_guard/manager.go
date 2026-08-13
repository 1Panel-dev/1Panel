package docker_guard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	Chain       = "1PANEL_DOCKER"
	DockerChain = "DOCKER-USER"
	FamilyIPv4  = "ipv4"
	FamilyIPv6  = "ipv6"
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
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60 * time.Second))
	return manager.RunWithOptionalSudoAndStdout(executable, args...)
}

func (commandRunner) RunInput(executable, input string, args ...string) (string, error) {
	manager := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second), cmd.WithStdin(strings.NewReader(input)))
	return manager.RunWithOptionalSudoAndStdout(executable, args...)
}

func (commandRunner) Exists(executable string) bool { return cmd.Which(executable) }

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
	dockerChain, err := m.chainExists(executable, DockerChain)
	if err != nil {
		return err
	}
	if !dockerChain {
		if required {
			return fmt.Errorf("%w for %s", ErrDockerChainUnavailable, executable)
		}
		return nil
	}
	owned, err := m.chainExists(executable, Chain)
	if err != nil {
		return err
	}
	if !owned {
		if required {
			return fmt.Errorf("%s chain is not initialized for %s", Chain, executable)
		}
		return nil
	}
	return m.ensureJump(executable)
}

func (m *Manager) ensureFamily(executable string, required bool) error {
	if !m.runner.Exists(executable) {
		if required {
			return fmt.Errorf("%s is not installed", executable)
		}
		return nil
	}
	dockerChain, err := m.chainExists(executable, DockerChain)
	if err != nil {
		return err
	}
	if !dockerChain {
		if required {
			return fmt.Errorf("%w for %s", ErrDockerChainUnavailable, executable)
		}
		return nil
	}
	owned, err := m.chainExists(executable, Chain)
	if err != nil {
		return err
	}
	if !owned {
		if _, err := m.run(executable, "-N", Chain); err != nil {
			return err
		}
	}
	return m.ensureJump(executable)
}

func (m *Manager) ensureJump(executable string) error {
	output, err := m.run(executable, "-S", DockerChain)
	if err != nil {
		return err
	}
	for i := 0; i < countJumps(output); i++ {
		if _, err := m.run(executable, "-D", DockerChain, "-j", Chain); err != nil {
			return err
		}
	}
	_, err = m.run(executable, "-I", DockerChain, "1", "-j", Chain)
	return err
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
		for i := 0; i < countJumps(output); i++ {
			if _, err := m.run(executable, "-D", DockerChain, "-j", Chain); err != nil {
				return err
			}
		}
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
