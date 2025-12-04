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
	EstablishedRule = "-m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment 'ESTABLISHED Whitelist'"
	IoRuleIn        = "-i lo -j ACCEPT -m comment --comment 'Loopback Whitelist'"
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

func RunWithStd(tab, rule string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(60*time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -w -t %s %s", cmd.SudoHandleCmd(), tab, rule)
	if err != nil {
		global.LOG.Errorf("iptables command failed [table=%s, rule=%s]: %v", tab, rule, err)
		return stdout, err
	}
	return stdout, nil
}
func RunWithoutIgnore(tab, rule string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(60 * time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -t %s %s", cmd.SudoHandleCmd(), tab, rule)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}
func Run(tab, rule string) error {
	if _, err := RunWithStd(tab, rule); err != nil {
		return err
	}
	return nil
}

func NewChain(tab, chain string) error {
	return Run(tab, "-N "+chain)
}

func ClearChain(tab, chain string) error {
	return Run(tab, "-F "+chain)
}

func AddRule(tab, chain, rule string) error {
	if CheckRuleExist(tab, chain, rule) {
		return nil
	}
	return Run(tab, fmt.Sprintf("-A %s %s", chain, rule))
}
func DeleteRule(tab, chain, rule string) error {
	return Run(tab, fmt.Sprintf("-D %s %s", chain, rule))
}

func CheckChainExist(tab, chain string) (bool, error) {
	stdout, err := RunWithStd(tab, fmt.Sprintf("-S | grep -w 'N %s'", chain))
	if err != nil {
		global.LOG.Errorf("check chain %s from tab %s exist failed, err: %v", chain, tab, err)
		return false, fmt.Errorf("check chain %s from tab %s exist failed, err: %v", chain, tab, err)
	}
	if strings.TrimSpace(stdout) == "" {
		return false, nil
	}
	return true, nil
}
func CheckChainBind(tab, parentChain, chain string) (bool, error) {
	stdout, err := RunWithStd(tab, fmt.Sprintf("-S %s | grep -- '-j %s'", parentChain, chain))
	if err != nil {
		global.LOG.Errorf("check chain %s from tab %s is bind to %s failed, err: %v", chain, tab, parentChain, err)
		return false, fmt.Errorf("check chain %s from tab %s is bind to %s failed, err: %v", chain, tab, parentChain, err)
	}
	if strings.TrimSpace(stdout) == "" {
		return false, nil
	}
	return true, nil
}
func CheckRuleExist(tab, chain, rule string) bool {
	_, err := RunWithoutIgnore(tab, fmt.Sprintf("-C %s %s", chain, rule))
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
		if err := Run(tab, fmt.Sprintf("-I %s %d -j %s", targetChain, position, chain)); err != nil {
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
		return Run(tab, fmt.Sprintf("-D %s %v", targetChain, line))
	}
	return nil
}

func FindChainNum(tab, targetChain, chain string) (int, error) {
	stdout, err := RunWithStd(tab, fmt.Sprintf("-L %s --line-numbers -n | grep -w %s", targetChain, chain))
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
	return Run(tab, fmt.Sprintf("-A %s -j %s", parentChain, chain))
}
