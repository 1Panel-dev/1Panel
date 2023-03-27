package client

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
)

type Firewall struct {
	Client ssh.ConnInfo
}

func NewFirewalld() (*Firewall, error) {
	ConnInfo := ssh.ConnInfo{
		Addr:     "172.16.10.143",
		User:     "root",
		AuthMode: "password",
		Port:     22,
	}
	return &Firewall{Client: ConnInfo}, nil
}

func (f *Firewall) Status() (string, error) {
	stdout, err := f.Client.Run("firewall-cmd --state")
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, err: %s", stdout)
	}
	return strings.ReplaceAll(stdout, "\n", ""), nil
}

func (f *Firewall) Start() error {
	stdout, err := f.Client.Run("systemctl start firewalld")
	if err != nil {
		return fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) Stop() error {
	stdout, err := f.Client.Run("systemctl stop firewalld")
	if err != nil {
		return fmt.Errorf("stop the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) Reload() error {
	stdout, err := f.Client.Run("firewall-cmd --reload")
	if err != nil {
		return fmt.Errorf("reload firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) ListPort() ([]FireInfo, error) {
	stdout, err := f.Client.Run("firewall-cmd --zone=public --list-ports")
	if err != nil {
		return nil, err
	}
	ports := strings.Split(strings.ReplaceAll(stdout, "\n", ""), " ")
	var datas []FireInfo
	for _, port := range ports {
		var itemPort FireInfo
		if strings.Contains(port, "/") {
			itemPort.Port = strings.Split(port, "/")[0]
			itemPort.Protocol = strings.Split(port, "/")[1]
		}
		itemPort.Strategy = "accept"
		datas = append(datas, itemPort)
	}

	stdout1, err := f.Client.Run("firewall-cmd --zone=public --list-rich-rules")
	if err != nil {
		return nil, err
	}
	rules := strings.Split(stdout1, "\n")
	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		itemRule := f.loadInfo(rule)
		if len(itemRule.Port) != 0 {
			datas = append(datas, itemRule)
		}
	}
	return datas, nil
}

func (f *Firewall) ListAddress() ([]FireInfo, error) {
	stdout, err := f.Client.Run("firewall-cmd --zone=public --list-rich-rules")
	if err != nil {
		return nil, err
	}
	var datas []FireInfo
	rules := strings.Split(stdout, "\n")
	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		itemRule := f.loadInfo(rule)
		if len(itemRule.Port) == 0 {
			datas = append(datas, itemRule)
		}
	}
	return datas, nil
}

func (f *Firewall) Port(port FireInfo, operation string) error {
	stdout, err := f.Client.Run(fmt.Sprintf("firewall-cmd --zone=public --%s-port=%s/%s --permanent", operation, port.Port, port.Protocol))
	if err != nil {
		return fmt.Errorf("%s port failed, err: %s", operation, stdout)
	}
	return nil
}

func (f *Firewall) RichRules(rule FireInfo, operation string) error {
	ruleStr := "rule family=ipv4 "
	if len(rule.Address) != 0 {
		ruleStr += fmt.Sprintf("source address=%s ", rule.Address)
	}
	if len(rule.Port) != 0 {
		ruleStr += fmt.Sprintf("port port=%s ", rule.Port)
	}
	if len(rule.Protocol) != 0 {
		ruleStr += fmt.Sprintf("protocol=%s ", rule.Protocol)
	}
	ruleStr += rule.Strategy

	stdout, err := f.Client.Run(fmt.Sprintf("firewall-cmd --zone=public --%s-rich-rule '%s' --permanent", operation, ruleStr))
	if err != nil {
		return fmt.Errorf("%s rich rules failed, err: %s", operation, stdout)
	}
	return nil
}

func (f *Firewall) PortForward(info Forward, operation string) error {
	ruleStr := fmt.Sprintf("firewall-cmd --%s-forward-port=port=%s:proto=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.Target)
	if len(info.Address) != 0 {
		ruleStr = fmt.Sprintf("firewall-cmd --%s-forward-port=port=%s:proto=%s:toaddr=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.Address, info.Target)
	}

	stdout, err := f.Client.Run(ruleStr)
	if err != nil {
		return fmt.Errorf("%s port forward failed, err: %s", operation, stdout)
	}
	return nil
}

func (f *Firewall) loadInfo(line string) FireInfo {
	var itemRule FireInfo
	ruleInfo := strings.Split(strings.ReplaceAll(line, "\"", ""), " ")
	for _, item := range ruleInfo {
		switch {
		case strings.Contains(item, "family="):
			itemRule.Family = strings.ReplaceAll(item, "family=", "")
		case strings.Contains(item, "address="):
			itemRule.Address = strings.ReplaceAll(item, "address=", "")
		case strings.Contains(item, "port="):
			itemRule.Port = strings.ReplaceAll(item, "port=", "")
		case strings.Contains(item, "protocol="):
			itemRule.Protocol = strings.ReplaceAll(item, "protocol=", "")
		case item == "accept" || item == "drop":
			itemRule.Strategy = item
		}
	}
	return itemRule
}
