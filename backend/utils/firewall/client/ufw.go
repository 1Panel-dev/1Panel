package client

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type Ufw struct {
	CmdStr string
}

func NewUfw() (*Ufw, error) {
	var ufw Ufw
	if cmd.HasNoPasswordSudo() {
		ufw.CmdStr = "sudo ufw"
	} else {
		ufw.CmdStr = "ufw"
	}
	return &ufw, nil
}

func (f *Ufw) Name() string {
	return "ufw"
}

func (f *Ufw) Status() (string, error) {
	stdout, _ := cmd.Execf("%s status | grep Status", f.CmdStr)
	if stdout == "Status: active\n" {
		return "running", nil
	}
	stdout1, _ := cmd.Execf("%s status | grep 状态", f.CmdStr)
	if stdout1 == "状态： 激活\n" {
		return "running", nil
	}
	return "not running", nil
}

func (f *Ufw) Version() (string, error) {
	stdout, err := cmd.Execf("%s version | grep ufw", f.CmdStr)
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, err: %s", stdout)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	return strings.ReplaceAll(info, "ufw ", ""), nil
}

func (f *Ufw) Start() error {
	stdout, err := cmd.Execf("echo y | %s enable", f.CmdStr)
	if err != nil {
		return fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) Stop() error {
	stdout, err := cmd.Execf("%s disable", f.CmdStr)
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
	stdout, err := cmd.Execf("%s status verbose", f.CmdStr)
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

func (f *Ufw) ListAddress() ([]FireInfo, error) {
	stdout, err := cmd.Execf("%s status verbose", f.CmdStr)
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
		return fmt.Errorf("unsupport strategy %s", port.Strategy)
	}
	if cmd.CheckIllegal(port.Protocol, port.Port) {
		return buserr.New(constant.ErrCmdIllegal)
	}

	command := fmt.Sprintf("%s %s %s", f.CmdStr, port.Strategy, port.Port)
	if operation == "remove" {
		command = fmt.Sprintf("%s delete %s %s", f.CmdStr, port.Strategy, port.Port)
	}
	if len(port.Protocol) != 0 {
		command += fmt.Sprintf("/%s", port.Protocol)
	}
	stdout, err := cmd.Exec(command)
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
		return fmt.Errorf("unsupport strategy %s", rule.Strategy)
	}

	if cmd.CheckIllegal(operation, rule.Protocol, rule.Address, rule.Port) {
		return buserr.New(constant.ErrCmdIllegal)
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

	stdout, err := cmd.Exec(ruleStr)
	if err != nil {
		if strings.Contains(stdout, "ERROR: Invalid position") {
			stdout, err := cmd.Exec(strings.ReplaceAll(ruleStr, "insert 1 ", ""))
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
	ruleStr := fmt.Sprintf("firewall-cmd --%s-forward-port=port=%s:proto=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.Target)
	if len(info.Address) != 0 {
		ruleStr = fmt.Sprintf("firewall-cmd --%s-forward-port=port=%s:proto=%s:toaddr=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.Address, info.Target)
	}

	stdout, err := cmd.Exec(ruleStr)
	if err != nil {
		return fmt.Errorf("%s port forward failed, err: %s", operation, stdout)
	}
	if err := f.Reload(); err != nil {
		return err
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
	if fields[1] == "(v6)" {
		return itemInfo
	}
	if fields[0] == "Anywhere" && fireType != "port" {
		itemInfo.Strategy = "drop"
		if fields[1] == "ALLOW" {
			itemInfo.Strategy = "accept"
		}
		itemInfo.Address = fields[3]
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
