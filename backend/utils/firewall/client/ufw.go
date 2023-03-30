package client

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
)

type Ufw struct {
	Client ssh.ConnInfo
}

func NewUfw() (*Ufw, error) {
	ConnInfo := ssh.ConnInfo{
		Addr:     "172.16.10.234",
		User:     "ubuntu",
		AuthMode: "password",
		Port:     22,
	}
	return &Ufw{Client: ConnInfo}, nil
}

func (f *Ufw) Name() string {
	return "ufw"
}

func (f *Ufw) Status() (string, error) {
	stdout, err := f.Client.Run("sudo ufw status | grep Status")
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, err: %s", stdout)
	}
	if stdout == "Status: active\n" {
		return "running", nil
	}
	return "not running", nil
}

func (f *Ufw) Version() (string, error) {
	stdout, err := f.Client.Run("sudo ufw version | grep ufw")
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, err: %s", stdout)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	return strings.ReplaceAll(info, "ufw ", ""), nil
}

func (f *Ufw) Start() error {
	stdout, err := f.Client.Run("echo y | sudo ufw enable")
	if err != nil {
		return fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) PingStatus() (string, error) {
	stdout, err := f.Client.Run("cat /etc/ufw/sysctl.conf | grep net/ipv4/icmp_echo_ignore_all= ")
	if err != nil {
		return constant.StatusDisable, fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	if stdout == "net/ipv4/icmp_echo_ignore_all=1\n" {
		return constant.StatusEnable, nil
	}
	return constant.StatusDisable, nil
}

func (f *Ufw) Stop() error {
	stdout, err := f.Client.Run("sudo ufw disable")
	if err != nil {
		return fmt.Errorf("stop the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) Reload() error {
	return nil
}

func (f *Ufw) ListPort() ([]FireInfo, error) {
	stdout, err := f.Client.Run("sudo ufw  status verbose")
	if err != nil {
		return nil, err
	}
	portInfos := strings.Split(stdout, "\n")
	var datas []FireInfo
	isStart := false
	for _, line := range portInfos {
		if strings.HasPrefix(line, "--") {
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
	stdout, err := f.Client.Run("sudo ufw  status verbose")
	if err != nil {
		return nil, err
	}
	portInfos := strings.Split(stdout, "\n")
	var datas []FireInfo
	isStart := false
	for _, line := range portInfos {
		if strings.HasPrefix(line, "--") {
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

	command := fmt.Sprintf("sudo ufw %s %s", port.Strategy, port.Port)
	if operation == "remove" {
		command = fmt.Sprintf("sudo ufw delete %s %s", port.Strategy, port.Port)
	}
	if len(port.Protocol) != 0 {
		command += fmt.Sprintf("/%s", port.Protocol)
	}
	stdout, err := f.Client.Run(command)
	if err != nil {
		return fmt.Errorf("%s port failed, err: %s", operation, stdout)
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

	ruleStr := fmt.Sprintf("sudo ufw %s ", rule.Strategy)
	if operation == "remove" {
		ruleStr = fmt.Sprintf("sudo ufw delete %s ", rule.Strategy)
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

	stdout, err := f.Client.Run(ruleStr)
	if err != nil {
		return fmt.Errorf("%s rich rules failed, err: %s", operation, stdout)
	}
	return nil
}

func (f *Ufw) PortForward(info Forward, operation string) error {
	ruleStr := fmt.Sprintf("firewall-cmd --%s-forward-port=port=%s:proto=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.Target)
	if len(info.Address) != 0 {
		ruleStr = fmt.Sprintf("firewall-cmd --%s-forward-port=port=%s:proto=%s:toaddr=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.Address, info.Target)
	}

	stdout, err := f.Client.Run(ruleStr)
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
