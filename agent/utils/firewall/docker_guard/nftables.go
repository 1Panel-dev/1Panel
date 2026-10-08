package docker_guard

import (
	"context"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"slices"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
)

const (
	NftTable     = "nft_1panel_docker"
	NftBaseChain = "NFT_1PANEL_DOCKER_FORWARD"
	NftChain     = "NFT_1PANEL_DOCKER"

	dockerNftTable = "docker-bridges"
)

type Nftables struct {
	runner Runner
}

func NewNftables(ctx context.Context) *Nftables { return &Nftables{runner: commandRunner{ctx: ctx}} }

func (m *Nftables) Initialize(policies []Policy, inventory PolicyInventory, families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if !m.runner.Exists("nft") {
		return errors.New("nft is not installed")
	}
	if err := CheckIPv4Forwarding(); err != nil {
		return err
	}
	if err := m.checkForwardPolicy(families...); err != nil {
		return err
	}
	if err := m.ensureFamily(FamilyIPv4, true); err != nil {
		return err
	}
	if slices.Contains(families, FamilyIPv6) {
		if err := m.ensureFamily(FamilyIPv6, false); err != nil {
			return &FamilyError{Family: FamilyIPv6, Err: err}
		}
	}
	return m.rebuildLocked(policies, inventory, families...)
}

func (m *Nftables) Bind(families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if err := m.bindExistingFamily(FamilyIPv4, true); err != nil {
		return err
	}
	if slices.Contains(families, FamilyIPv6) {
		if err := m.bindExistingFamily(FamilyIPv6, false); err != nil {
			return &FamilyError{Family: FamilyIPv6, Err: err}
		}
	}
	return nil
}

func (m *Nftables) ReplacePolicies(policies []Policy, inventory PolicyInventory) error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	return m.rebuildLocked(policies, inventory)
}

func (m *Nftables) ListPolicies() (PolicyInventory, error) {
	if !m.runner.Exists("nft") {
		return PolicyInventory{}, nil
	}
	inventory := PolicyInventory{Policies: make([]Policy, 0), RuleOrders: make(map[string][]int64)}
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		output, err := nftables_helper.ReadChain(m.run, tableFamily, NftTable, NftChain)
		if errors.Is(err, nftables_helper.ErrChainNotFound) || (family == FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable)) {
			continue
		}
		if err != nil {
			return PolicyInventory{}, &FamilyError{Family: family, Err: fmt.Errorf("list %s chain: %w", NftChain, err)}
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

func (m *Nftables) Unbind(families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	mutationMu.Lock()
	defer mutationMu.Unlock()
	for _, family := range families {
		if err := m.unbindFamily(family); err != nil {
			return &FamilyError{Family: family, Err: err}
		}
	}
	return nil
}

func (m *Nftables) Cleanup() error {
	mutationMu.Lock()
	defer mutationMu.Unlock()
	if !m.runner.Exists("nft") {
		return errors.New("nft is not installed")
	}
	commands := make([][]string, 0, 2)
	for _, family := range []string{FamilyIPv4, FamilyIPv6} {
		tableFamily := nftTableFamily(family)
		_, exists, err := nftables_helper.ReadTable(m.run, tableFamily, NftTable)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		commands = append(commands, []string{"delete", "table", tableFamily, NftTable})
	}
	return m.runBatch(commands)
}

func (m *Nftables) Initialized(family string) (bool, error) {
	if nftTableFamily(family) == "" || !m.runner.Exists("nft") {
		return false, nil
	}
	tableFamily := nftTableFamily(family)
	if !m.objectExists("chain", tableFamily, NftTable, NftBaseChain) {
		return false, nil
	}
	return m.objectExists("chain", tableFamily, NftTable, NftChain), nil
}

func (m *Nftables) Status(family string) FamilyStatus {
	tableFamily := nftTableFamily(family)
	if tableFamily == "" || !m.runner.Exists("nft") {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonCommandMissing}
	}
	if !m.objectExists("table", tableFamily, dockerNftTable) {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonDockerChainMissing}
	}
	output, err := m.run("-a", "list", "table", tableFamily, NftTable)
	if err != nil {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonGuardChainMissing}
	}
	baseExists, guardExists := false, false
	currentChain := ""
	var baseRules strings.Builder
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "chain" && fields[2] == "{" {
			currentChain = fields[1]
			baseExists = baseExists || currentChain == NftBaseChain
			guardExists = guardExists || currentChain == NftChain
		} else if strings.TrimSpace(line) == "}" {
			currentChain = ""
		}
		if currentChain == NftBaseChain {
			baseRules.WriteString(line)
			baseRules.WriteByte('\n')
		}
	}
	if !baseExists || !guardExists {
		return FamilyStatus{State: StatusDisabled, Reason: ReasonGuardChainMissing, Partial: baseExists || guardExists}
	}
	status := FamilyStatus{State: StatusNotEffective, Initialized: true}
	rules := baseRules.String()
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

func (m *Nftables) ensureFamily(family string, required bool) error {
	tableFamily := nftTableFamily(family)
	if tableFamily == "" {
		return fmt.Errorf("unsupported address family %q", family)
	}
	if !m.objectExists("table", tableFamily, dockerNftTable) {
		if required {
			return fmt.Errorf("%w %s", buserr.New("ErrDockerNftablesChainUnavailable"), family)
		}
		return nil
	}
	commands := make([][]string, 0, 6)
	tableExists := m.objectExists("table", tableFamily, NftTable)
	if !tableExists {
		commands = append(commands, []string{"add", "table", tableFamily, NftTable})
	}
	baseExists := tableExists && m.objectExists("chain", tableFamily, NftTable, NftBaseChain)
	if !baseExists {
		commands = append(commands, []string{
			"add", "chain", tableFamily, NftTable, NftBaseChain,
			"{", "type", "filter", "hook", "forward", "priority", "-1", ";", "policy", "accept", ";", "}",
		})
	}
	if !tableExists || !m.objectExists("chain", tableFamily, NftTable, NftChain) {
		commands = append(commands, []string{"add", "chain", tableFamily, NftTable, NftChain})
	}
	if baseExists {
		output, err := m.run("-a", "list", "chain", tableFamily, NftTable, NftBaseChain)
		if err != nil {
			return err
		}
		for _, handle := range nftJumpHandles(output) {
			commands = append(commands, []string{"delete", "rule", tableFamily, NftTable, NftBaseChain, "handle", handle})
		}
	}
	commands = append(commands, []string{"insert", "rule", tableFamily, NftTable, NftBaseChain, "jump", NftChain})
	return m.runBatch(commands)
}

func (m *Nftables) bindExistingFamily(family string, required bool) error {
	tableFamily := nftTableFamily(family)
	if !m.runner.Exists("nft") {
		if required {
			return errors.New("nft is not installed")
		}
		return nil
	}
	if !m.objectExists("table", tableFamily, dockerNftTable) {
		if required {
			return fmt.Errorf("%w %s", buserr.New("ErrDockerNftablesChainUnavailable"), family)
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

func (m *Nftables) ensureJump(family string) error {
	tableFamily := nftTableFamily(family)
	output, err := m.run("-a", "list", "chain", tableFamily, NftTable, NftBaseChain)
	if err != nil {
		return err
	}
	commands := make([][]string, 0, len(nftJumpHandles(output))+1)
	for _, handle := range nftJumpHandles(output) {
		commands = append(commands, []string{"delete", "rule", tableFamily, NftTable, NftBaseChain, "handle", handle})
	}
	commands = append(commands, []string{"insert", "rule", tableFamily, NftTable, NftBaseChain, "jump", NftChain})
	return m.runBatch(commands)
}

func (m *Nftables) rebuildLocked(policies []Policy, inventory PolicyInventory, families ...string) error {
	if len(families) == 0 {
		families = []string{FamilyIPv4, FamilyIPv6}
	}
	if !m.runner.Exists("nft") {
		return nil
	}
	for _, family := range families {
		tableFamily := nftTableFamily(family)
		if !m.objectExists("chain", tableFamily, NftTable, NftChain) {
			continue
		}
		commands := [][]string{{"flush", "chain", tableFamily, NftTable, NftChain}}
		commands = append(commands, []string{"add", "rule", tableFamily, NftTable, NftChain, "ct", "state", "{", "established,related", "}", "return"})
		commands = append(commands, orderedNftRules(family, policies, inventory)...)
		commands = append(commands, []string{"add", "rule", tableFamily, NftTable, NftChain, "return"})
		script, err := buildNftScript(commands)
		if err != nil {
			return &FamilyError{Family: family, Err: err}
		}
		if _, err := m.runner.RunInput("nft", script); err != nil {
			return &FamilyError{Family: family, Err: fmt.Errorf("restore rules: %w", err)}
		}
	}
	return nil
}

func orderedNftRules(family string, policies []Policy, inventory PolicyInventory) [][]string {
	tableFamily := nftTableFamily(family)
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
				if native.Family != family || len(native.Tokens) == 0 || native.Tokens[0] == "-A" {
					continue
				}
				command := []string{"add", "rule", tableFamily, NftTable, NftChain}
				command = append(command, quoteNftTokens(native.Tokens)...)
				segments = append(segments, orderedPolicyRule{order: native.Order, rule: command})
			}
			continue
		}
		compiled := compileNftPolicy(policy)
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

func quoteNftTokens(tokens []string) []string {
	quoted := make([]string, 0, len(tokens))
	for index, token := range tokens {
		if (index > 0 && tokens[index-1] == "comment" || strings.ContainsAny(token, " \t\\\"'")) && !strings.HasPrefix(token, `"`) {
			quoted = append(quoted, strconv.Quote(token))
			continue
		}
		quoted = append(quoted, token)
	}
	return quoted
}

func compileNftPolicy(policy Policy) [][]string {
	tableFamily := nftTableFamily(policy.Family)
	addressKeyword := tableFamily
	base := []string{"add", "rule", tableFamily, NftTable, NftChain, "meta", "l4proto", policy.Protocol}
	if !isWildcardHost(policy.Family, policy.HostIP) {
		base = append(base, "ct", "original", addressKeyword, "daddr", policy.HostIP)
	}
	base = append(base, "ct", "original", "proto-dst", strconv.Itoa(int(policy.HostPort)))
	marker := "1panel-docker:" + policy.UUID
	comment := strconv.Quote(marker)
	if policy.Mode == ModeAll || policy.Mode == ModeAcceptAll {
		target := "drop"
		if policy.Mode == ModeAcceptAll {
			target = "accept"
		}
		return [][]string{append(append([]string{}, base...), target, "comment", comment)}
	}
	target := "drop"
	capacity := len(policy.Sources)
	if policy.Mode == ModeAllow {
		target = "return"
		capacity++
	} else if policy.Mode == ModeAcceptSources {
		target = "accept"
	}
	rules := make([][]string, 0, capacity)
	for _, source := range policy.Sources {
		args := append([]string{}, base...)
		args = append(args, addressKeyword, "saddr", source, target, "comment", comment)
		rules = append(rules, args)
	}
	if policy.Mode == ModeAllow {
		rules = append(rules, append(append([]string{}, base...), "drop", "comment", comment))
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

func (m *Nftables) unbindFamily(family string) error {
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
	commands := make([][]string, 0, len(nftJumpHandles(output)))
	for _, handle := range nftJumpHandles(output) {
		commands = append(commands, []string{"delete", "rule", tableFamily, NftTable, NftBaseChain, "handle", handle})
	}
	return m.runBatch(commands)
}

func (m *Nftables) runBatch(commands [][]string) error {
	if len(commands) == 0 {
		return nil
	}
	script, err := buildNftScript(commands)
	if err != nil {
		return err
	}
	if _, err := m.runner.RunInput("nft", script); err != nil {
		return fmt.Errorf("batch update Docker guard lifecycle: %w", err)
	}
	return nil
}

func (m *Nftables) objectExists(kind string, args ...string) bool {
	command := append([]string{"list", kind}, args...)
	_, err := m.run(command...)
	return err == nil
}

func (m *Nftables) run(args ...string) (string, error) {
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
		if len(fields) == 0 || fields[0] == "table" || fields[0] == "chain" {
			continue
		}
		return len(fields) >= 2 && fields[0] == "jump" && fields[1] == NftChain
	}
	return false
}

func (m *Nftables) checkForwardPolicy(families ...string) error {
	for _, family := range []struct{ command, name string }{
		{"iptables", FamilyIPv4},
		{"ip6tables", FamilyIPv6},
	} {
		if len(families) > 0 && !slices.Contains(families, family.name) {
			continue
		}
		if !m.runner.Exists(family.command) {
			continue
		}
		output, err := m.runner.Run(family.command, "-t", "filter", "-w", "-S", "FORWARD")
		if err != nil {
			return &FamilyError{Family: family.name, Err: fmt.Errorf("inspect iptables FORWARD policy: %w", err)}
		}
		found := false
		for _, line := range strings.Split(output, "\n") {
			fields := strings.Fields(line)
			if len(fields) != 3 || fields[0] != "-P" || fields[1] != "FORWARD" {
				continue
			}
			found = true
			if fields[2] == "DROP" {
				return &FamilyError{Family: family.name, Err: buserr.New("ErrDockerForwardPolicyDrop")}
			}
			if fields[2] != "ACCEPT" {
				return &FamilyError{Family: family.name, Err: fmt.Errorf("unexpected iptables FORWARD policy: %s", fields[2])}
			}
		}
		if !found {
			return &FamilyError{Family: family.name, Err: errors.New("iptables FORWARD default policy was not found")}
		}
	}
	return nil
}

func (m *Nftables) OperateFamily(family string, initialize bool) error {
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
	if err := m.checkForwardPolicy(family); err != nil {
		return err
	}
	if initialize {
		return m.ensureFamily(family, true)
	}
	return m.bindExistingFamily(family, true)
}
