package client

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

var ForwardListRegex = regexp.MustCompile(`^port=(\d{1,5}):proto=(.+?):toport=(\d{1,5}):toaddr=(.*)$`)

type Firewall struct{}

func NewFirewalld() (*Firewall, error) {
	return &Firewall{}, nil
}

func (f *Firewall) Name() string {
	return "firewalld"
}

func (f *Firewall) Status() (bool, error) {
	stdout, _ := cmd.RunDefaultWithStdoutBashC("LANGUAGE=en_US:en firewall-cmd --state")
	return stdout == "running\n", nil
}

func (f *Firewall) Version() (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashC("LANGUAGE=en_US:en firewall-cmd --version")
	if err != nil {
		return "", fmt.Errorf("load the firewall version failed, %v", err)
	}
	return strings.ReplaceAll(stdout, "\n ", ""), nil
}

func (f *Firewall) Start() error {
	if err := controller.HandleStart("firewalld"); err != nil {
		return fmt.Errorf("enable the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewall) Stop() error {
	if err := controller.HandleStop("firewalld"); err != nil {
		return fmt.Errorf("stop the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewall) Restart() error {
	if err := controller.HandleRestart("firewalld"); err != nil {
		return fmt.Errorf("restart the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewall) Reload() error {
	if err := cmd.RunDefaultBashC("firewall-cmd --reload"); err != nil {
		return fmt.Errorf("reload firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewall) ListPort() ([]FireInfo, error) {
	var wg sync.WaitGroup
	var datas []FireInfo
	wg.Add(2)
	go func() {
		defer wg.Done()
		stdout, err := cmd.RunDefaultWithStdoutBashC("firewall-cmd --zone=public --list-ports")
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
		stdout1, err := cmd.RunDefaultWithStdoutBashC("firewall-cmd --zone=public --list-rich-rules")
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
	stdout, err := cmd.RunDefaultWithStdoutBashC("firewall-cmd --zone=public --list-forward-ports")
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
	stdout, err := cmd.RunDefaultWithStdoutBashC("firewall-cmd --zone=public --list-rich-rules")
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
		return buserr.New("ErrCmdIllegal")
	}

	if err := cmd.RunDefaultBashCf("firewall-cmd --zone=public --%s-port=%s/%s --permanent", operation, port.Port, port.Protocol); err != nil {
		return fmt.Errorf("%s (port: %s/%s strategy: %s) failed, %v", operation, port.Port, port.Protocol, port.Strategy, err)
	}
	return nil
}

func (f *Firewall) RichRules(rule FireInfo, operation string) error {
	if cmd.CheckIllegal(operation, rule.Address, rule.Protocol, rule.Port, rule.Strategy) {
		return buserr.New("ErrCmdIllegal")
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
	if err := cmd.RunDefaultBashCf("firewall-cmd --zone=public --%s-rich-rule '%s' --permanent", operation, ruleStr); err != nil {
		return fmt.Errorf("%s rich rules (%s) failed, %v", operation, ruleStr, err)
	}
	if len(rule.Address) == 0 {
		if err := cmd.RunDefaultBashCf("firewall-cmd --zone=public --%s-rich-rule '%s' --permanent", operation, strings.ReplaceAll(ruleStr, "family=ipv4 ", "family=ipv6 ")); err != nil {
			return fmt.Errorf("%s rich rules (%s) failed, %v", operation, strings.ReplaceAll(ruleStr, "family=ipv4 ", "family=ipv6 "), err)
		}
	}
	return nil
}

func (f *Firewall) PortForward(info Forward, operation string) error {
	ruleStr := fmt.Sprintf("firewall-cmd --zone=public --%s-forward-port=port=%s:proto=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.TargetPort)
	if info.TargetIP != "" && info.TargetIP != "127.0.0.1" && info.TargetIP != "localhost" {
		ruleStr = fmt.Sprintf("firewall-cmd --zone=public --%s-forward-port=port=%s:proto=%s:toaddr=%s:toport=%s --permanent", operation, info.Port, info.Protocol, info.TargetIP, info.TargetPort)
	}

	if err := cmd.RunDefaultBashC(ruleStr); err != nil {
		return fmt.Errorf("%s port forward failed, %s", operation, err)
	}
	if err := f.Reload(); err != nil {
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
	stdout, err := cmd.RunDefaultWithStdoutBashC("firewall-cmd --zone=public --query-masquerade")
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(stdout), "no") {
			if err := cmd.RunDefaultBashC("firewall-cmd --zone=public --add-masquerade --permanent"); err != nil {
				return err
			}
			return f.Reload()
		}
		return err
	}

	return nil
}
