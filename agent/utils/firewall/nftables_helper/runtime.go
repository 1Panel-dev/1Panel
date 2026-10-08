package nftables_helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	firewallutil "github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const (
	TableName        = "nft_1panel_filter"
	InputChain       = "NFT_1PANEL_INPUT"
	BasicBeforeChain = "NFT_1PANEL_BASIC_BEFORE"
	BasicChain       = "NFT_1PANEL_BASIC"
	BasicAfterChain  = "NFT_1PANEL_BASIC_AFTER"
)

func TableFamily(family filter.Family) string {
	if family == filter.FamilyIPv6 {
		return "ip6"
	}
	return "ip"
}

func BasicChains() []string {
	return []string{BasicBeforeChain, BasicChain, BasicAfterChain}
}

func run(args ...string) (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudoAndStdout("nft", args...)
	if err != nil {
		return stdout, fmt.Errorf("command=nft %s failed: %w", strings.Join(args, " "), err)
	}
	return stdout, nil
}

func readNftObject(run func(...string) (string, error), args ...string) (string, bool, error) {
	output, err := run(args...)
	if err == nil {
		return output, true, nil
	}
	message := err.Error()
	if slices.Contains(args, "ip6") && (strings.Contains(message, "Address family not supported") || strings.Contains(message, "Protocol not supported")) {
		return "", false, fmt.Errorf("%w: %v", filter.ErrFamilyUnavailable, err)
	}
	if strings.Contains(message, "Error:") && strings.Contains(message, "No such file or directory") &&
		!strings.Contains(message, "Operation not permitted") && !strings.Contains(message, "Permission denied") {
		return "", false, nil
	}
	return "", false, err
}

func runCommand(args ...string) error {
	err := cmd.NewCommandMgr(cmd.WithTimeout(60*time.Second)).RunWithOptionalSudo("nft", args...)
	if err != nil {
		return fmt.Errorf("command=nft %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

func runBatch(commands ...[]string) error {
	script, err := buildBatchScript(commands...)
	if err != nil || script == "" {
		return err
	}
	return RunScript(script)
}

func RunScript(script string) error {
	return RunScriptContext(context.Background(), script)
}

func RunScriptContext(ctx context.Context, script string) error {
	return runScriptFile(script, func(file string) error {
		return cmd.NewCommandMgr(
			cmd.WithContext(ctx),
			cmd.WithEnv("LC_ALL=C"),
			cmd.WithTimeout(60*time.Second),
		).RunWithOptionalSudo("nft", "-f", file)
	})
}

func runScriptFile(script string, run func(string) error) error {
	file, err := os.CreateTemp("", "1panel-nft-*.nft")
	if err != nil {
		return err
	}
	name := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(name)
	}()
	if _, err := file.WriteString(script); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return firewallutil.WrapBatchCommandError("nft -f <generated-batch>", script, run(name))
}

func buildBatchScript(commands ...[]string) (string, error) {
	var script strings.Builder
	for _, command := range commands {
		if len(command) == 0 {
			continue
		}
		for _, token := range command {
			if strings.ContainsAny(token, "\r\n") {
				return "", fmt.Errorf("invalid newline in nftables batch command")
			}
		}
		script.WriteString(strings.Join(command, " "))
		script.WriteByte('\n')
	}
	return script.String(), nil
}

func LoadInitStatus(tab string) (bool, bool, error) {
	if tab != "base" {
		return false, false, nil
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		initialized, bound, err := loadFamilyInitStatus(family)
		if err != nil || !initialized {
			return false, false, err
		}
		if !bound {
			return true, false, nil
		}
	}
	return true, true, nil
}

func LoadFamilyInitStatus(family filter.Family, tab string) (bool, bool, error) {
	if tab != "base" {
		return false, false, nil
	}
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return false, false, nil
	}
	return loadFamilyInitStatus(family)
}

func LoadFamilyBindStatus(family filter.Family) (bool, error) {
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return false, nil
	}
	stdout, err := run("list", "chain", TableFamily(family), TableName, InputChain)
	if err != nil {
		return false, err
	}
	return hasBaseChainBinding(stdout), nil
}

func hasBaseChainBinding(output string) bool {
	for _, chain := range BasicChains() {
		if strings.Contains(output, "jump "+chain) {
			return true
		}
	}
	return false
}

func loadFamilyInitStatus(family filter.Family) (bool, bool, error) {
	initialized, bound, _, err := LoadFamilyState(family)
	return initialized, bound, err
}

func LoadFamilyState(family filter.Family) (bool, bool, bool, error) {
	output, exists, err := ReadTable(run, TableFamily(family), TableName)
	if err != nil || !exists {
		return false, false, false, err
	}
	chains := ParseTableChains(output)
	count := 0
	for _, chain := range append(BasicChains(), InputChain) {
		if _, ok := chains[chain]; ok {
			count++
		}
	}
	if count != len(BasicChains())+1 {
		return false, false, count > 0, nil
	}
	input, exists := chains[InputChain]
	if !exists {
		return false, false, false, nil
	}
	for _, chain := range BasicChains() {
		if !strings.Contains(input, "jump "+chain) {
			return true, false, false, nil
		}
	}
	return true, !TableDormant(output), false, nil
}

func ReadTable(run func(...string) (string, error), family, table string) (string, bool, error) {
	return readNftObject(run, "-a", "list", "table", family, table)
}

func ParseTableChains(output string) map[string]string {
	chains := make(map[string]string)
	lines := strings.Split(output, "\n")
	name, indent, start := "", "", 0
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if name == "" {
			fields := strings.Fields(trimmed)
			if len(fields) >= 3 && fields[0] == "chain" && fields[2] == "{" {
				name, indent, start = fields[1], line[:len(line)-len(strings.TrimLeft(line, " \t"))], index
			}
			continue
		}
		if strings.HasPrefix(line, indent+"}") {
			chains[name] = strings.Join(lines[start:index+1], "\n")
			name = ""
		}
	}
	return chains
}

var ErrChainNotFound = errors.New("nftables chain is not initialized")

func ReadChain(run func(...string) (string, error), family, table, chain string) (string, error) {
	output, exists, err := readNftObject(run, "-a", "list", "chain", family, table, chain)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrChainNotFound
	}
	return output, nil
}

func TableDormant(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSuffix(strings.TrimSpace(line), ";"))
		if len(fields) > 0 && fields[0] == "chain" {
			break
		}
		if len(fields) > 1 && fields[0] == "flags" {
			for _, flag := range strings.Split(strings.Join(fields[1:], ""), ",") {
				if flag == "dormant" {
					return true
				}
			}
		}
	}
	return false
}

func SetTableDormant(ctx context.Context, family, table string) error {
	manager := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second))
	output, exists, err := ReadTable(func(args ...string) (string, error) { return manager.RunWithOptionalSudoAndStdout("nft", args...) }, family, table)
	if err != nil || !exists || TableDormant(output) {
		return err
	}
	script, err := buildBatchScript([]string{"add", "table", family, table, "{", "flags", "dormant", ";", "}"})
	if err != nil {
		return err
	}
	return RunScriptContext(ctx, script)
}
