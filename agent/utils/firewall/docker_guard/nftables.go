package docker_guard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	NftTable     = "nft_1panel_docker"
	NftBaseChain = "NFT_1PANEL_DOCKER_FORWARD"
	NftChain     = "NFT_1PANEL_DOCKER"

	dockerNftTable = "docker-bridges"
)

// NftablesManager manages a separate nftables table that runs immediately
// before Docker's filter/forward base chain. Docker owns docker-bridges, so the
// guard only uses that table to detect whether a family is available.
type NftablesManager struct {
	runner Runner
}

func NewNftablesManager() *NftablesManager { return &NftablesManager{runner: commandRunner{}} }

func NewNftablesManagerWithRunner(runner Runner) *NftablesManager {
	return &NftablesManager{runner: runner}
}

func (m *NftablesManager) Initialize(policies []Policy) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if !m.runner.Exists("nft") {
		return errors.New("nft is not installed")
	}
	if err := m.ensureFamily(FamilyIPv4, true); err != nil {
		return err
	}
	if err := m.ensureFamily(FamilyIPv6, false); err != nil {
		return &FamilyError{Family: FamilyIPv6, Err: err}
	}
	return m.rebuildLocked(policies)
}

func (m *NftablesManager) Bind() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if err := m.bindExistingFamily(FamilyIPv4, true); err != nil {
		return err
	}
	if err := m.bindExistingFamily(FamilyIPv6, false); err != nil {
		return &FamilyError{Family: FamilyIPv6, Err: err}
	}
	return nil
}

func (m *NftablesManager) Reconcile(policies []Policy) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	return m.rebuildLocked(policies)
}

func (m *NftablesManager) Unbind() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		if err := m.unbindFamily(family); err != nil {
			return &FamilyError{Family: family, Err: err}
		}
	}
	return nil
}

func (m *NftablesManager) Cleanup() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if !m.runner.Exists("nft") {
		return nil
	}
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		if !m.objectExists("table", tableFamily, NftTable) {
			continue
		}
		if _, err := m.run("delete", "table", tableFamily, NftTable); err != nil {
			return &FamilyError{Family: family, Err: err}
		}
	}
	return nil
}

func (m *NftablesManager) Initialized(family string) (bool, error) {
	if nftTableFamily(family) == "" || !m.runner.Exists("nft") {
		return false, nil
	}
	tableFamily := nftTableFamily(family)
	if !m.objectExists("chain", tableFamily, NftTable, NftBaseChain) {
		return false, nil
	}
	return m.objectExists("chain", tableFamily, NftTable, NftChain), nil
}

func (m *NftablesManager) Status(family string) FamilyStatus {
	tableFamily := nftTableFamily(family)
	if tableFamily == "" || !m.runner.Exists("nft") {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonCommandMissing}
	}
	if !m.objectExists("table", tableFamily, dockerNftTable) {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonDockerChainMissing}
	}
	if !m.objectExists("chain", tableFamily, NftTable, NftBaseChain) ||
		!m.objectExists("chain", tableFamily, NftTable, NftChain) {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonGuardChainMissing}
	}
	status := FamilyStatus{State: StatusNotEffective, Initialized: true}
	rules, err := m.run("-a", "list", "chain", tableFamily, NftTable, NftBaseChain)
	if err != nil {
		status.Reason = ReasonInspectFailed
		return status
	}
	jumps := nftJumpHandles(rules)
	if len(jumps) == 0 {
		status.Reason = ReasonJumpMissing
		return status
	}
	if len(jumps) > 1 {
		status.Reason = ReasonJumpDuplicate
		return status
	}
	if !nftHasFirstUniqueJump(rules) {
		status.Reason = ReasonJumpNotFirst
		return status
	}
	status.State = StatusEffective
	status.Bound = true
	status.Effective = true
	return status
}

func (m *NftablesManager) ensureFamily(family string, required bool) error {
	tableFamily := nftTableFamily(family)
	if tableFamily == "" {
		return fmt.Errorf("unsupported address family %q", family)
	}
	if !m.objectExists("table", tableFamily, dockerNftTable) {
		if required {
			return fmt.Errorf("%w for nftables %s", ErrDockerChainUnavailable, family)
		}
		return nil
	}
	if !m.objectExists("table", tableFamily, NftTable) {
		if _, err := m.run("add", "table", tableFamily, NftTable); err != nil {
			return fmt.Errorf("create 1Panel Docker guard table: %w", err)
		}
	}
	if !m.objectExists("chain", tableFamily, NftTable, NftBaseChain) {
		if _, err := m.run(
			"add", "chain", tableFamily, NftTable, NftBaseChain,
			"{", "type", "filter", "hook", "forward", "priority", "filter", "-", "1", ";", "policy", "accept", ";", "}",
		); err != nil {
			return fmt.Errorf("create 1Panel Docker guard base chain: %w", err)
		}
	}
	if !m.objectExists("chain", tableFamily, NftTable, NftChain) {
		if _, err := m.run("add", "chain", tableFamily, NftTable, NftChain); err != nil {
			return fmt.Errorf("create 1Panel Docker guard chain: %w", err)
		}
	}
	return m.ensureJump(family)
}

func (m *NftablesManager) bindExistingFamily(family string, required bool) error {
	tableFamily := nftTableFamily(family)
	if !m.runner.Exists("nft") {
		if required {
			return errors.New("nft is not installed")
		}
		return nil
	}
	if !m.objectExists("table", tableFamily, dockerNftTable) {
		if required {
			return fmt.Errorf("%w for nftables %s", ErrDockerChainUnavailable, family)
		}
		return nil
	}
	if !m.objectExists("chain", tableFamily, NftTable, NftBaseChain) ||
		!m.objectExists("chain", tableFamily, NftTable, NftChain) {
		if required {
			return fmt.Errorf("%s chain is not initialized for nftables %s", NftChain, family)
		}
		return nil
	}
	return m.ensureJump(family)
}

func (m *NftablesManager) ensureJump(family string) error {
	tableFamily := nftTableFamily(family)
	output, err := m.run("-a", "list", "chain", tableFamily, NftTable, NftBaseChain)
	if err != nil {
		return err
	}
	for _, handle := range nftJumpHandles(output) {
		if _, err := m.run("delete", "rule", tableFamily, NftTable, NftBaseChain, "handle", handle); err != nil {
			return err
		}
	}
	_, err = m.run("insert", "rule", tableFamily, NftTable, NftBaseChain, "jump", NftChain)
	return err
}

func (m *NftablesManager) rebuildLocked(policies []Policy) error {
	if !m.runner.Exists("nft") {
		return nil
	}
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		if !m.objectExists("chain", tableFamily, NftTable, NftChain) {
			continue
		}
		commands := [][]string{{"flush", "chain", tableFamily, NftTable, NftChain}}
		commands = append(commands, []string{"add", "rule", tableFamily, NftTable, NftChain, "ct", "state", "{", "established,related", "}", "return"})
		for _, policy := range policies {
			if policy.Family == family {
				commands = append(commands, compileNftPolicy(policy)...)
			}
		}
		commands = append(commands, []string{"add", "rule", tableFamily, NftTable, NftChain, "return"})
		script, err := buildNftScript(commands)
		if err != nil {
			return &FamilyError{Family: family, Err: err}
		}
		if _, err := m.runner.RunInput("nft", script, "-f", "-"); err != nil {
			return &FamilyError{Family: family, Err: fmt.Errorf("restore rules: %w", err)}
		}
	}
	return nil
}

func compileNftPolicy(policy Policy) [][]string {
	tableFamily := nftTableFamily(policy.Family)
	addressKeyword := tableFamily
	base := []string{"add", "rule", tableFamily, NftTable, NftChain, "meta", "l4proto", policy.Protocol}
	if !isWildcardHost(policy.Family, policy.HostIP) {
		base = append(base, "ct", "original", addressKeyword, "daddr", policy.HostIP)
	}
	base = append(base, "ct", "original", "proto-dst", strconv.Itoa(int(policy.HostPort)))
	comment := strconv.Quote("1panel-docker:" + policy.UUID)
	if policy.Mode == ModeAll {
		return [][]string{append(append([]string{}, base...), "comment", comment, "drop")}
	}
	target := "drop"
	capacity := len(policy.Sources)
	if policy.Mode == ModeAllow {
		target = "return"
		capacity++
	}
	rules := make([][]string, 0, capacity)
	for _, source := range policy.Sources {
		args := append([]string{}, base...)
		args = append(args, addressKeyword, "saddr", source, "comment", comment, target)
		rules = append(rules, args)
	}
	if policy.Mode == ModeAllow {
		rules = append(rules, append(append([]string{}, base...), "comment", comment, "drop"))
	}
	return rules
}

func buildNftScript(commands [][]string) (string, error) {
	var script strings.Builder
	for _, command := range commands {
		for index, token := range command {
			if !validNftToken(token) {
				return "", fmt.Errorf("invalid nftables token %q", token)
			}
			if index > 0 {
				script.WriteByte(' ')
			}
			script.WriteString(token)
		}
		script.WriteByte('\n')
	}
	return script.String(), nil
}

func validNftToken(token string) bool {
	if token == "" || strings.ContainsAny(token, "\r\n") {
		return false
	}
	if strings.HasPrefix(token, `"`) {
		_, err := strconv.Unquote(token)
		return err == nil
	}
	return !strings.ContainsAny(token, " \t\\\"'")
}

func (m *NftablesManager) unbindFamily(family string) error {
	if !m.runner.Exists("nft") {
		return nil
	}
	tableFamily := nftTableFamily(family)
	if !m.objectExists("chain", tableFamily, NftTable, NftBaseChain) {
		return nil
	}
	output, err := m.run("-a", "list", "chain", tableFamily, NftTable, NftBaseChain)
	if err != nil {
		return err
	}
	for _, handle := range nftJumpHandles(output) {
		if _, err := m.run("delete", "rule", tableFamily, NftTable, NftBaseChain, "handle", handle); err != nil {
			return err
		}
	}
	return nil
}

func (m *NftablesManager) objectExists(kind string, args ...string) bool {
	command := append([]string{"list", kind}, args...)
	_, err := m.run(command...)
	return err == nil
}

func (m *NftablesManager) run(args ...string) (string, error) {
	return m.runner.Run("nft", args...)
}

func nftTableFamily(family string) string {
	switch family {
	case FamilyIPv4:
		return "ip"
	case FamilyIPv6:
		return "ip6"
	default:
		return ""
	}
}

func nftJumpHandles(output string) []string {
	handles := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 4 || fields[0] != "jump" || fields[1] != NftChain {
			continue
		}
		for index := 2; index+1 < len(fields); index++ {
			if fields[index] == "handle" {
				handles = append(handles, fields[index+1])
				break
			}
		}
	}
	return handles
}

func nftHasFirstUniqueJump(output string) bool {
	if len(nftJumpHandles(output)) != 1 {
		return false
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "# handle ") {
			continue
		}
		fields := strings.Fields(line)
		return len(fields) >= 2 && fields[0] == "jump" && fields[1] == NftChain
	}
	return false
}
