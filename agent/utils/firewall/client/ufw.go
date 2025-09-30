package client

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type Ufw struct {
	CmdStr string
}

func NewUfw() (*Ufw, error) {
	var ufw Ufw
	ufw.CmdStr = fmt.Sprintf("LANGUAGE=en_US:en %s ufw", cmd.SudoHandleCmd())
	return &ufw, nil
}

func (f *Ufw) Name() string {
	return "ufw"
}

func (f *Ufw) Status() (bool, error) {
	stdout, _ := cmd.RunDefaultWithStdoutBashCf("%s status | grep Status", f.CmdStr)
	if stdout == "Status: active\n" {
		return true, nil
	}
	stdout1, _ := cmd.RunDefaultWithStdoutBashCf("%s status | grep 状态", f.CmdStr)
	if stdout1 == "状态： 激活\n" {
		return true, nil
	}
	return false, nil
}

func (f *Ufw) Version() (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s version | grep ufw", f.CmdStr)
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, err: %s", stdout)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	return strings.ReplaceAll(info, "ufw ", ""), nil
}

func (f *Ufw) Start() error {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("echo y | %s enable", f.CmdStr)
	if err != nil {
		return fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) Stop() error {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s disable", f.CmdStr)
	if err != nil {
		return fmt.Errorf("stop the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) Restart() error {
	if err := f.Stop(); err != nil {
		return err
	}
	if err := f.Start(); err != nil {
		return err
	}
	return nil
}

func (f *Ufw) Reload() error {
	return nil
}

func (f *Ufw) ListPort() ([]FireInfo, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s status verbose", f.CmdStr)
	if err != nil {
		return nil, err
	}
	portInfos := strings.Split(stdout, "\n")
	var datas []FireInfo
	isStart := false
	for _, line := range portInfos {
		if strings.HasPrefix(line, "-") {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}
		itemFire := f.loadInfo(line, "port")
		if len(itemFire.Port) != 0 && itemFire.Port != "Anywhere" && !strings.Contains(itemFire.Port, ".") {
			itemFire.Port = strings.ReplaceAll(itemFire.Port, ":", "-")
			datas = append(datas, itemFire)
		}
	}
	return datas, nil
}

func (f *Ufw) ListForward() ([]FireInfo, error) {
	if err := f.EnableForward(); err != nil {
		global.LOG.Errorf("init port forward failed, err: %v", err)
	}
	iptables, err := NewIptables()
	if err != nil {
		return nil, err
	}
	rules, err := iptables.NatList()
	if err != nil {
		return nil, err
	}

	var list []FireInfo
	for _, rule := range rules {
		dest := strings.Split(rule.DestPort, ":")
		if len(dest) < 2 {
			continue
		}
		if len(dest[0]) == 0 {
			dest[0] = "127.0.0.1"
		}
		list = append(list, FireInfo{
			Num:        rule.Num,
			Protocol:   rule.Protocol,
			Interface:  rule.InIface,
			Port:       rule.SrcPort,
			TargetIP:   dest[0],
			TargetPort: dest[1],
		})
	}
	return list, nil
}

func (f *Ufw) ListAddress() ([]FireInfo, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s status verbose", f.CmdStr)
	if err != nil {
		return nil, err
	}
	portInfos := strings.Split(stdout, "\n")
	var datas []FireInfo
	isStart := false
	for _, line := range portInfos {
		if strings.HasPrefix(line, "-") {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}
		if !strings.Contains(line, " IN") {
			continue
		}
		itemFire := f.loadInfo(line, "address")
		if strings.Contains(itemFire.Port, ".") {
			itemFire.Address += ("-" + itemFire.Port)
			itemFire.Port = ""
		}
		if len(itemFire.Port) == 0 && len(itemFire.Address) != 0 {
			datas = append(datas, itemFire)
		}
	}
	return datas, nil
}

func (f *Ufw) Port(port FireInfo, operation string) error {
	switch port.Strategy {
	case "accept":
		port.Strategy = "allow"
	case "drop":
		port.Strategy = "deny"
	default:
		return fmt.Errorf("unsupported strategy %s", port.Strategy)
	}
	if cmd.CheckIllegal(port.Protocol, port.Port) {
		return buserr.New("ErrCmdIllegal")
	}

	command := fmt.Sprintf("%s %s %s", f.CmdStr, port.Strategy, port.Port)
	if operation == "remove" {
		command = fmt.Sprintf("%s delete %s %s", f.CmdStr, port.Strategy, port.Port)
	}
	if len(port.Protocol) != 0 {
		command += fmt.Sprintf("/%s", port.Protocol)
	}
	stdout, err := cmd.RunDefaultWithStdoutBashC(command)
	if err != nil {
		return fmt.Errorf("%s (%s) failed, err: %s", operation, command, stdout)
	}
	return nil
}

func (f *Ufw) RichRules(rule FireInfo, operation string) error {
	switch rule.Strategy {
	case "accept":
		rule.Strategy = "allow"
	case "drop":
		rule.Strategy = "deny"
	default:
		return fmt.Errorf("unsupported strategy %s", rule.Strategy)
	}

	if cmd.CheckIllegal(operation, rule.Protocol, rule.Address, rule.Port) {
		return buserr.New("ErrCmdIllegal")
	}

	ruleStr := fmt.Sprintf("%s insert 1 %s ", f.CmdStr, rule.Strategy)
	if operation == "remove" {
		ruleStr = fmt.Sprintf("%s delete %s ", f.CmdStr, rule.Strategy)
	}
	if len(rule.Protocol) != 0 {
		ruleStr += fmt.Sprintf("proto %s ", rule.Protocol)
	}
	if strings.Contains(rule.Address, "-") {
		ruleStr += fmt.Sprintf("from %s to %s ", strings.Split(rule.Address, "-")[0], strings.Split(rule.Address, "-")[1])
	} else {
		ruleStr += fmt.Sprintf("from %s ", rule.Address)
	}
	if len(rule.Port) != 0 {
		ruleStr += fmt.Sprintf("to any port %s ", rule.Port)
	}

	stdout, err := cmd.RunDefaultWithStdoutBashC(ruleStr)
	if err != nil {
		if strings.Contains(stdout, "ERROR: Invalid position") || strings.Contains(stdout, "ERROR: 无效位置") {
			stdout, err := cmd.RunDefaultWithStdoutBashC(strings.ReplaceAll(ruleStr, "insert 1 ", ""))
			if err != nil {
				return fmt.Errorf("%s rich rules (%s), failed, err: %s", operation, ruleStr, stdout)
			}
			return nil
		}
		return fmt.Errorf("%s rich rules (%s), failed, err: %s", operation, ruleStr, stdout)
	}
	return nil
}

func (f *Ufw) PortForward(info Forward, operation string) error {
	iptables, err := NewIptables()
	if err != nil {
		return err
	}

	if operation == "add" {
		err = iptables.NatAdd(info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface, true)
	} else {
		err = iptables.NatRemove(info.Num, info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface)
	}
	if err != nil {
		return fmt.Errorf("%s port forward failed, err: %s", operation, err)
	}
	return nil
}

func (f *Ufw) loadInfo(line string, fireType string) FireInfo {
	fields := strings.Fields(line)
	var itemInfo FireInfo
	if strings.Contains(line, "LIMIT") || strings.Contains(line, "ALLOW FWD") {
		return itemInfo
	}
	if len(fields) < 4 {
		return itemInfo
	}
	if fields[1] == "(v6)" && fireType == "port" {
		return itemInfo
	}
	if fields[0] == "Anywhere" && fireType != "port" {
		itemInfo.Strategy = "drop"
		if fields[1] == "ALLOW" {
			itemInfo.Strategy = "accept"
		}
		if fields[1] == "(v6)" {
			if fields[2] == "ALLOW" {
				itemInfo.Strategy = "accept"
			}
			itemInfo.Address = fields[4]
		} else {
			itemInfo.Address = fields[3]
		}
		return itemInfo
	}
	if strings.Contains(fields[0], "/") {
		itemInfo.Port = strings.Split(fields[0], "/")[0]
		itemInfo.Protocol = strings.Split(fields[0], "/")[1]
	} else {
		itemInfo.Port = fields[0]
		itemInfo.Protocol = "tcp/udp"
	}
	itemInfo.Family = "ipv4"
	if fields[1] == "ALLOW" {
		itemInfo.Strategy = "accept"
	} else {
		itemInfo.Strategy = "drop"
	}
	itemInfo.Address = fields[3]

	return itemInfo
}

func (f *Ufw) EnableForward() error {
	iptables, err := NewIptables()
	if err != nil {
		return err
	}
	if err = iptables.Check(); err != nil {
		return err
	}

	_ = iptables.NewChain(NatTab, PreRoutingChain)
	_ = iptables.NewChain(NatTab, PostRoutingChain)
	_ = iptables.NewChain(FilterTab, ForwardChain)

	if err = f.enableChain(iptables); err != nil {
		return err
	}
	return iptables.Reload()
}

func (f *Ufw) enableChain(iptables *Iptables) error {
	rules, err := iptables.NatList("PREROUTING")
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.Target == PreRoutingChain {
			return nil
		}
	}

	_ = iptables.AppendChain(NatTab, "PREROUTING", PreRoutingChain)
	_ = iptables.AppendChain(NatTab, "POSTROUTING", PostRoutingChain)
	_ = iptables.AppendChain(FilterTab, "FORWARD", ForwardChain)
	return nil
}
