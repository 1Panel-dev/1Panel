package iptables_helper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

const (
	InputChain       = "INPUT"
	BasicBeforeChain = constant.FirewallBasicBeforeChain
	BasicChain       = constant.FirewallBasicChain
	BasicAfterChain  = constant.FirewallBasicAfterChain
)

func BasicChains() []string {
	return []string{BasicBeforeChain, BasicChain, BasicAfterChain}
}

const (
	EstablishedRule = "-m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment \"ESTABLISHED Whitelist\""
	IoRuleIn        = "-i lo -j ACCEPT -m comment --comment \"Loopback Whitelist\""
	DropAllTcp      = "-p tcp -j DROP"
	DropAllUdp      = "-p udp -j DROP"
	AllowSSH        = "-p tcp --dport ssh -j ACCEPT"
)

const (
	ACCEPT   = "ACCEPT"
	DROP     = "DROP"
	REJECT   = "REJECT"
	ANYWHERE = "anywhere"
)

const (
	FilterTab = "filter"
	NatTab    = "nat"
)

func runTables(ctx context.Context, executable, tab string, ignoreExist1, withWait bool, ruleArgs ...string) (string, error) {
	options := []cmd.Option{cmd.WithContext(ctx), cmd.WithTimeout(60 * time.Second)}
	if ignoreExist1 {
		options = append(options, cmd.WithIgnoreExist1())
	}
	cmdMgr := cmd.NewCommandMgr(options...)
	args := []string{"-t", tab}
	if withWait {
		args = append(args, "-w")
	}
	args = append(args, ruleArgs...)
	return cmdMgr.RunWithOptionalSudoAndStdout(executable, args...)
}

func runIptables(ctx context.Context, tab string, ignoreExist1, withWait bool, ruleArgs ...string) (string, error) {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return "", err
	}
	return runTables(ctx, commands.IPv4, tab, ignoreExist1, withWait, ruleArgs...)
}

func RunWithStd(tab string, args ...string) (string, error) {
	return RunWithStdContext(context.Background(), tab, args...)
}

func RunWithStdContext(ctx context.Context, tab string, args ...string) (string, error) {
	stdout, err := runIptables(ctx, tab, true, true, args...)
	if err != nil {
		global.LOG.Errorf("iptables command failed [table=%s, args=%s]: %v", tab, strings.Join(args, " "), err)
		return stdout, err
	}
	return stdout, nil
}

func RunIPv6WithStd(tab string, args ...string) (string, error) {
	return RunIPv6WithStdContext(context.Background(), tab, args...)
}

func RunIPv6WithStdContext(ctx context.Context, tab string, args ...string) (string, error) {
	commands, executableErr := lifecycle.ResolveIptablesCommands()
	if executableErr != nil {
		return "", executableErr
	}
	if !commands.IPv6Available() {
		return "", fmt.Errorf("ip6tables command family is unavailable")
	}
	stdout, err := runTables(ctx, commands.IPv6, tab, true, true, args...)
	if err != nil {
		global.LOG.Errorf("ip6tables command failed [table=%s, args=%s]: %v", tab, strings.Join(args, " "), err)
		return stdout, err
	}
	return stdout, nil
}

func RunIPv6(tab string, args ...string) error {
	_, err := RunIPv6WithStd(tab, args...)
	return err
}

func CheckIPv6ChainExist(tab, chain string) (bool, error) {
	stdout, err := RunIPv6WithStd(tab, "-S")
	if err != nil {
		return false, fmt.Errorf("check IPv6 chain %s from tab %s exist failed: %w", chain, tab, err)
	}
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == "-N "+chain {
			return true, nil
		}
	}
	return false, nil
}

func RunWithoutIgnore(tab string, args ...string) (string, error) {
	stdout, err := runIptables(context.Background(), tab, false, false, args...)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}
func Run(tab string, args ...string) error {
	if _, err := RunWithStd(tab, args...); err != nil {
		return err
	}
	return nil
}

func CheckChainExist(tab, chain string) (bool, error) {
	stdout, err := RunWithStd(tab, "-S")
	if err != nil {
		global.LOG.Errorf("check chain %s from tab %s exist failed, err: %v", chain, tab, err)
		return false, fmt.Errorf("check chain %s from tab %s exist failed, err: %v", chain, tab, err)
	}
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == "-N "+chain {
			return true, nil
		}
	}
	return false, nil
}
