package iptables

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	Chain1PanelPreRouting  = "1PANEL_PREROUTING"
	Chain1PanelPostRouting = "1PANEL_POSTROUTING"
	Chain1PanelForward     = "1PANEL_FORWARD"
	ChainInput             = "INPUT"
	ChainOutput            = "OUTPUT"
	Chain1PanelInput       = "1PANEL_INPUT"
	Chain1PanelOutput      = "1PANEL_OUTPUT"
	Chain1PanelBasicBefore = "1PANEL_BASIC_BEFORE"
	Chain1PanelBasic       = "1PANEL_BASIC"
	Chain1PanelBasicAfter  = "1PANEL_BASIC_AFTER"
)

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

func runIptables(tab string, ignoreExist1, withWait bool, ruleArgs ...string) (string, error) {
	options := []cmd.Option{cmd.WithTimeout(60 * time.Second)}
	if ignoreExist1 {
		options = append(options, cmd.WithIgnoreExist1())
	}
	cmdMgr := cmd.NewCommandMgr(options...)
	args := []string{"-t", tab}
	if withWait {
		args = append(args, "-w")
	}
	args = append(args, ruleArgs...)
	return cmdMgr.RunWithOptionalSudoAndStdout("iptables", args...)
}

func RunWithStd(tab string, args ...string) (string, error) {
	stdout, err := runIptables(tab, true, true, args...)
	if err != nil {
		global.LOG.Errorf("iptables command failed [table=%s, args=%s]: %v", tab, strings.Join(args, " "), err)
		return stdout, err
	}
	return stdout, nil
}
func RunWithoutIgnore(tab string, args ...string) (string, error) {
	stdout, err := runIptables(tab, false, false, args...)
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

func NewChain(tab, chain string) error {
	return Run(tab, "-N", chain)
}

func ClearChain(tab, chain string) error {
	return Run(tab, "-F", chain)
}

func AddRule(tab, chain string, ruleArgs ...string) error {
	if CheckRuleExist(tab, chain, ruleArgs...) {
		return nil
	}
	args := append([]string{"-A", chain}, ruleArgs...)
	return Run(tab, args...)
}
func DeleteRule(tab, chain string, ruleArgs ...string) error {
	args := append([]string{"-D", chain}, ruleArgs...)
	return Run(tab, args...)
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
func CheckChainBind(tab, parentChain, chain string) (bool, error) {
	stdout, err := RunWithStd(tab, "-S", parentChain)
	if err != nil {
		global.LOG.Errorf("check chain %s from tab %s is bind to %s failed, err: %v", chain, tab, parentChain, err)
		return false, fmt.Errorf("check chain %s from tab %s is bind to %s failed, err: %v", chain, tab, parentChain, err)
	}
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "-j "+chain) {
			return true, nil
		}
	}
	return false, nil
}
func CheckRuleExist(tab, chain string, ruleArgs ...string) bool {
	args := append([]string{"-C", chain}, ruleArgs...)
	_, err := RunWithoutIgnore(tab, args...)
	return err == nil
}

func AddChain(tab, chain string) error {
	exists, err := CheckChainExist(tab, chain)
	if err != nil {
		return fmt.Errorf("check chain %s exist from tab %s failed, err: %w", chain, tab, err)
	}
	if !exists {
		if err := NewChain(tab, chain); err != nil {
			return fmt.Errorf("add chain %s for tab %s failed, err: %w", tab, chain, err)
		}
	}
	return nil
}
func BindChain(tab, targetChain, chain string, position int) error {
	line, err := FindChainNum(tab, targetChain, chain)
	if err != nil {
		return fmt.Errorf("find chain %s number from %s failed, err: %w", chain, targetChain, err)
	}
	if line == 0 {
		if err := Run(tab, "-I", targetChain, strconv.Itoa(position), "-j", chain); err != nil {
			return fmt.Errorf("bind chain %s to %s failed, err: %w", chain, targetChain, err)
		}
	}
	return nil
}
func UnbindChain(tab, targetChain, chain string) error {
	line, err := FindChainNum(tab, targetChain, chain)
	if err != nil {
		return fmt.Errorf("find chain %s number from %s failed, err: %w", chain, targetChain, err)
	}
	if line != 0 {
		return Run(tab, "-D", targetChain, strconv.Itoa(line))
	}
	return nil
}

func FindChainNum(tab, targetChain, chain string) (int, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(60*time.Second))
	commandName, commandArgs := cmd.WrapWithOptionalSudo("iptables", "-w", "-t", tab, "-L", targetChain, "--line-numbers", "-n")
	stdout, err := cmdMgr.RunPipe(
		cmd.PipeCommand{Name: commandName, Args: commandArgs},
		cmd.PipeCommand{Name: "grep", Args: []string{"-w", chain}},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to list rules in chain %s: %w", targetChain, err)
	}

	lineItem := strings.TrimSpace(stdout)
	lines := strings.Split(lineItem, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[1] == chain {
			itemNum, err := strconv.Atoi(fields[0])
			return itemNum, err
		}
	}
	return 0, nil
}

func AddChainWithAppend(tab, parentChain, chain string) error {
	exists, err := CheckChainExist(tab, chain)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", chain, err)
	}
	if !exists {
		if err := NewChain(tab, chain); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", chain, err)
		}
	}
	isBind, err := CheckChainBind(tab, parentChain, chain)
	if err != nil {
		return fmt.Errorf("check chain %s bind to %s failed, err: %w", parentChain, chain, err)
	}
	if !isBind {
		if err := AppendChain(tab, parentChain, chain); err != nil {
			return fmt.Errorf("failed to append %s to %s: %w", chain, parentChain, err)
		}
	}
	return nil
}
func AppendChain(tab string, parentChain, chain string) error {
	return Run(tab, "-A", parentChain, "-j", chain)
}
