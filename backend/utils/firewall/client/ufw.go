package client

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type Ufw struct{}

func NewUfw() (*Ufw, error) {
	return &Ufw{}, nil
}

func (f *Ufw) Status() (string, error) {
	stdout, err := cmd.Exec("sudo ufw status")
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, err: %s", stdout)
	}
	if stdout == "Status: inactive\n" {
		return "running", nil
	}
	return "not running", nil
}

func (f *Ufw) Start() error {
	stdout, err := cmd.Exec("sudo ufw enable")
	if err != nil {
		return fmt.Errorf("enable the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) Stop() error {
	stdout, err := cmd.Exec("sudo ufw disable")
	if err != nil {
		return fmt.Errorf("stop the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) Reload() error {
	stdout, err := cmd.Exec("sudo ufw reload")
	if err != nil {
		return fmt.Errorf("reload firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Ufw) ListPort() ([]FireInfo, error) {
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
		datas = append(datas, itemPort)
	}
	return datas, nil
}

func (f *Ufw) ListRichRules() ([]FireInfo, error) {
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
		var itemRule FireInfo
		ruleInfo := strings.Split(strings.ReplaceAll(rule, "\"", ""), " ")
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
		datas = append(datas, itemRule)
	}
	return datas, nil
}

func (f *Ufw) Port(port FireInfo, operation string) error {
	switch operation {
	case "add":
		operation = "allow"
	case "remove":
		operation = "deny"
	default:
		return fmt.Errorf("unsupport operation %s", operation)
	}

	command := fmt.Sprintf("ufw %s %s", operation, port.Port)
	if len(port.Protocol) != 0 {
		command += fmt.Sprintf("/%s", port.Protocol)
	}
	stdout, err := cmd.Exec(command)
	if err != nil {
		return fmt.Errorf("%s port failed, err: %s", operation, stdout)
	}
	return nil
}

func (f *Ufw) RichRules(rule FireInfo, operation string) error {
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
	if err := f.Reload(); err != nil {
		return err
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
