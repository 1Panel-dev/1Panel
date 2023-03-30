package client

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type Firewall struct{}

func NewFirewalld() (*Firewall, error) {
	return &Firewall{}, nil
}

func (f *Firewall) Name() string {
	return "firewalld"
}

func (f *Firewall) Status() (string, error) {
	stdout, _ := cmd.Exec("firewall-cmd --state")
	return strings.ReplaceAll(stdout, "\n", ""), nil
}

func (f *Firewall) Version() (string, error) {
	stdout, err := cmd.Exec("firewall-cmd --version")
	if err != nil {
		return "", fmt.Errorf("load the firewall version failed, err: %s", stdout)
	}
	return strings.ReplaceAll(stdout, "\n ", ""), nil
}

func (f *Firewall) Start() error {
	stdout, err := cmd.Exec("systemctl start firewalld")
	if err != nil {
		return fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) PingStatus() (string, error) {
	stdout, _ := cmd.Exec("firewall-cmd --query-rich-rule='rule protocol value=icmp drop'")
	if stdout == "yes\n" {
		return constant.StatusEnable, nil
	}
	return constant.StatusDisable, nil
}

func (f *Firewall) UpdatePingStatus(enabel string) error {
	operation := "add"
	if enabel == "0" {
		operation = "remove"
	}
	stdout, err := cmd.Execf("firewall-cmd --permanent --%s-rich-rule='rule protocol value=icmp drop'", operation)
	if err != nil {
		return fmt.Errorf("update firewall ping status failed, err: %s", stdout)
	}
	return f.Reload()
}

func (f *Firewall) Stop() error {
	stdout, err := cmd.Exec("systemctl stop firewalld")
	if err != nil {
		return fmt.Errorf("stop the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) Reload() error {
	stdout, err := cmd.Exec("firewall-cmd --reload")
	if err != nil {
		return fmt.Errorf("reload firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) ListPort() ([]FireInfo, error) {
	stdout, err := cmd.Exec("firewall-cmd --zone=public --list-ports")
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

	stdout1, err := cmd.Exec("firewall-cmd --zone=public --list-rich-rules")
	if err != nil {
		return nil, err
	}
	rules := strings.Split(stdout1, "\n")
	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		itemRule := f.loadInfo(rule)
		if len(itemRule.Port) != 0 && itemRule.Family == "ipv4" {
			datas = append(datas, itemRule)
		}
	}
	return datas, nil
}

func (f *Firewall) ListAddress() ([]FireInfo, error) {
	stdout, err := cmd.Exec("firewall-cmd --zone=public --list-rich-rules")
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
	stdout, err := cmd.Execf("firewall-cmd --zone=public --%s-port=%s/%s --permanent", operation, port.Port, port.Protocol)
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

	stdout, err := cmd.Execf("firewall-cmd --zone=public --%s-rich-rule '%s' --permanent", operation, ruleStr)
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

	stdout, err := cmd.Exec(ruleStr)
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
