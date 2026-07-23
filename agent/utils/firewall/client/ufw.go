package client

import (
	"fmt"
	"strconv"
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
	stdout, _ := f.runWithStdout("status")
	if strings.Contains(stdout, "Status: active") {
		return true, nil
	}
	if strings.Contains(stdout, "状态： 激活") {
		return true, nil
	}
	return false, nil
}

func (f *Ufw) Version() (string, error) {
	stdout, err := f.runWithStdout("version")
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, %v", err)
	}
	info := strings.Split(strings.TrimSpace(stdout), "\n")[0]
	return strings.ReplaceAll(info, "ufw ", ""), nil
}

func (f *Ufw) Start() error {
	if err := f.run("--force", "enable"); err != nil {
		return fmt.Errorf("enable the firewall failed, %v", err)
	}
	return nil
}

func (f *Ufw) Stop() error {
	if err := f.run("disable"); err != nil {
		return fmt.Errorf("stop the firewall failed, %v", err)
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
	stdout, err := f.runWithStdout("status", "verbose")
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
	stdout, err := f.runWithStdout("status", "verbose")
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

	args := []string{port.Strategy, port.Port}
	if operation == "remove" {
		args = []string{"delete", port.Strategy, port.Port}
	}
	if len(port.Protocol) != 0 {
		args[len(args)-1] += "/" + port.Protocol
	}
	if err := f.run(args...); err != nil {
		return fmt.Errorf("%s (%s) failed, %v", operation, strings.Join(args, " "), err)
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

	insertNum := f.loadInsertNum(rule, operation)
	args := []string{"insert", strconv.Itoa(insertNum), rule.Strategy}
	if operation == "remove" {
		args = []string{"delete", rule.Strategy}
	}
	if len(rule.Protocol) != 0 {
		args = append(args, "proto", rule.Protocol)
	}
	if strings.Contains(rule.Address, "-") {
		parts := strings.Split(rule.Address, "-")
		args = append(args, "from", parts[0], "to", parts[1])
	} else {
		args = append(args, "from", rule.Address)
	}
	if len(rule.Port) != 0 {
		args = append(args, "to", "any", "port", rule.Port)
	}

	stdout, err := f.runWithStdout(args...)
	if err != nil {
		if strings.Contains(stdout, "ERROR: Invalid position") || strings.Contains(stdout, "ERROR: 无效位置") {
			fallbackArgs := args
			if len(args) >= 2 && args[0] == "insert" {
				fallbackArgs = append([]string{}, args[2:]...)
			}
			if err := f.run(fallbackArgs...); err != nil {
				return fmt.Errorf("%s rich rules (%s), failed, %v", operation, strings.Join(args, " "), err)
			}
			return nil
		}
		return fmt.Errorf("%s rich rules (%s), failed, %v", operation, strings.Join(args, " "), err)
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

func (f *Ufw) loadInsertNum(rule FireInfo, operation string) int {
	if !strings.Contains(rule.Address, ":") || operation == "remove" {
		return 1
	}
	rules, err := f.runWithStdout("status", "numbered")
	if err != nil {
		global.LOG.Errorf("load ufw rules failed, err: %v", err)
		return 1
	}
	lines := strings.Split(rules, "\n")
	i := 1
	for _, item := range lines {
		fields := strings.Fields(item)
		if len(fields) < 4 {
			continue
		}
		if !strings.Contains(item, "(v6)") {
			i++
		}
	}
	return i
}

func (f *Ufw) run(args ...string) error {
	_, err := f.runWithStdout(args...)
	return err
}

func (f *Ufw) runWithStdout(args ...string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en"))
	return cmdMgr.RunWithOptionalSudoAndStdout("ufw", args...)
}
