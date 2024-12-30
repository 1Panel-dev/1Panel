package client

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

var ForwardListRegex = regexp.MustCompile(`^port=(\d{1,5}):proto=(.+?):toport=(\d{1,5}):toaddr=(.*)$`)

type Firewall struct{}

func NewFirewalld() (*Firewall, error) {
	return &Firewall{}, nil
}

func (f *Firewall) Name() string {
	return "firewalld"
}

func (f *Firewall) Status() (string, error) {
	stdout, _ := cmd.Exec("LANGUAGE=en_US:en firewall-cmd --state")
	if stdout == "running\n" {
		return "running", nil
	}
	return "not running", nil
}

func (f *Firewall) Version() (string, error) {
	stdout, err := cmd.Exec("LANGUAGE=en_US:en firewall-cmd --version")
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

func (f *Firewall) Stop() error {
	stdout, err := cmd.Exec("systemctl stop firewalld")
	if err != nil {
		return fmt.Errorf("stop the firewall failed, err: %s", stdout)
	}
	return nil
}

func (f *Firewall) Restart() error {
	stdout, err := cmd.Exec("systemctl restart firewalld")
	if err != nil {
		return fmt.Errorf("restart the firewall failed, err: %s", stdout)
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
	var wg sync.WaitGroup
	var datas []FireInfo
	wg.Add(2)
	go func() {
		defer wg.Done()
		stdout, err := cmd.Exec("firewall-cmd --zone=public --list-ports")
		if err != nil {
			return
		}
		ports := strings.Split(strings.ReplaceAll(stdout, "\n", ""), " ")
		for _, port := range ports {
			if len(port) == 0 {
				continue
			}
			var itemPort FireInfo
			if strings.Contains(port, "/") {
				itemPort.Port = strings.Split(port, "/")[0]
				itemPort.Protocol = strings.Split(port, "/")[1]
			}
			itemPort.Strategy = "accept"
			datas = append(datas, itemPort)
		}
	}()

	go func() {
		defer wg.Done()
		stdout1, err := cmd.Exec("firewall-cmd --zone=public --list-rich-rules")
		if err != nil {
			return
		}
		rules := strings.Split(stdout1, "\n")
		for _, rule := range rules {
			if len(rule) == 0 {
				continue
			}
			itemRule := f.loadInfo(rule)
			if len(itemRule.Port) != 0 && (itemRule.Family == "ipv4" || (itemRule.Family == "ipv6" && len(itemRule.Address) != 0)) {
				datas = append(datas, itemRule)
			}
		}
	}()
	wg.Wait()
	return datas, nil
}

func (f *Firewall) ListForward() ([]FireInfo, error) {
	if err := f.EnableForward(); err != nil {
		global.LOG.Errorf("init port forward failed, err: %v", err)
	}
	stdout, err := cmd.Exec("firewall-cmd --zone=public --list-forward-ports")
	if err != nil {
		return nil, err
	}
	var datas []FireInfo
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimFunc(line, func(r rune) bool {
			return r <= 32
		})
		if ForwardListRegex.MatchString(line) {
			match := ForwardListRegex.FindStringSubmatch(line)
			if len(match) < 4 {
				continue
			}
			if len(match[4]) == 0 {
				match[4] = "127.0.0.1"
			}
			datas = append(datas, FireInfo{
				Port:       match[1],
				Protocol:   match[2],
				TargetIP:   match[4],
				TargetPort: match[3],
			})
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
		if len(itemRule.Port) == 0 && len(itemRule.Address) != 0 {
			datas = append(datas, itemRule)
		}
	}
	return datas, nil
}

func (f *Firewall) Port(port FireInfo, operation string) error {
	if cmd.CheckIllegal(operation, port.Protocol, port.Port) {
		return buserr.New(constant.ErrCmdIllegal)
	}

	stdout, err := cmd.Execf("firewall-cmd --zone=public --%s-port=%s/%s --permanent", operation, port.Port, port.Protocol)
	if err != nil {
		return fmt.Errorf("%s (port: %s/%s strategy: %s) failed, err: %s", operation, port.Port, port.Protocol, port.Strategy, stdout)
	}
	return nil
}

func (f *Firewall) RichRules(rule FireInfo, operation string) error {
	if cmd.CheckIllegal(operation, rule.Address, rule.Protocol, rule.Port, rule.Strategy) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	ruleStr := "rule family=ipv4 "
	if strings.Contains(rule.Address, ":") {
		ruleStr = "rule family=ipv6 "
	}
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
		return fmt.Errorf("%s rich rules (%s) failed, err: %s", operation, ruleStr, stdout)
	}
	if len(rule.Address) == 0 {
		stdout1, err := cmd.Execf("firewall-cmd --zone=public --%s-rich-rule '%s' --permanent", operation, strings.ReplaceAll(ruleStr, "family=ipv4 ", "family=ipv6 "))
		if err != nil {
			return fmt.Errorf("%s rich rules (%s) failed, err: %s", operation, strings.ReplaceAll(ruleStr, "family=ipv4 ", "family=ipv6 "), stdout1)
		}
	}
	return nil
}

func (f *Firewall) PortForward(info Forward, operation string) error {
	ruleStr := fmt.Sprintf("firewall-cmd --zone=public --%s-forward-port=port=%s:proto=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.TargetPort)
	if info.TargetIP != "" && info.TargetIP != "127.0.0.1" && info.TargetIP != "localhost" {
		ruleStr = fmt.Sprintf("firewall-cmd --zone=public --%s-forward-port=port=%s:proto=%s:toaddr=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.TargetIP, info.TargetPort)
	}

	stdout, err := cmd.Exec(ruleStr)
	if err != nil {
		return fmt.Errorf("%s port forward failed, err: %s", operation, stdout)
	}
	if err = f.Reload(); err != nil {
		return err
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
		case strings.Contains(item, "ipset="):
			itemRule.Address = strings.ReplaceAll(item, "ipset=", "")
		case strings.Contains(item, "address="):
			itemRule.Address = strings.ReplaceAll(item, "address=", "")
		case strings.Contains(item, "port="):
			itemRule.Port = strings.ReplaceAll(item, "port=", "")
		case strings.Contains(item, "protocol="):
			itemRule.Protocol = strings.ReplaceAll(item, "protocol=", "")
		case item == "accept" || item == "drop" || item == "reject":
			itemRule.Strategy = item
		}
	}
	return itemRule
}

func (f *Firewall) EnableForward() error {
	stdout, err := cmd.Exec("firewall-cmd --zone=public --query-masquerade")
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(stdout), "no") {
			stdout, err = cmd.Exec("firewall-cmd --zone=public --add-masquerade --permanent")
			if err != nil {
				return fmt.Errorf("%s: %s", err, stdout)
			}
			return f.Reload()
		}
		return fmt.Errorf("%s: %s", err, stdout)
	}

	return nil
}
