package client

import (
	"fmt"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

type Firewall struct{}

func NewFirewalld() (*Firewall, error) {
	return &Firewall{}, nil
}

func (f *Firewall) Name() string {
	return "firewalld"
}

func (f *Firewall) Status() (bool, error) {
	stdout, _ := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--state")
	return stdout == "running\n", nil
}

func (f *Firewall) Version() (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--version")
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
	if err := cmd.NewCommandMgr().Run("firewall-cmd", "--reload"); err != nil {
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
		stdout, err := cmd.NewCommandMgr().RunWithStdout("firewall-cmd", "--zone=public", "--list-ports")
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
		stdout1, err := cmd.NewCommandMgr().RunWithStdout("firewall-cmd", "--zone=public", "--list-rich-rules")
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

func (f *Firewall) ListAddress() ([]FireInfo, error) {
	stdout, err := cmd.NewCommandMgr().RunWithStdout("firewall-cmd", "--zone=public", "--list-rich-rules")
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

	if err := cmd.NewCommandMgr().Run("firewall-cmd", buildFirewalldPortArgs(port, operation)...); err != nil {
		return fmt.Errorf("%s (port: %s/%s strategy: %s) failed, %v", operation, port.Port, port.Protocol, port.Strategy, err)
	}
	return nil
}

func (f *Firewall) RichRules(rule FireInfo, operation string) error {
	if cmd.CheckIllegal(operation, rule.Address, rule.Protocol, rule.Port, rule.Strategy) {
		return buserr.New("ErrCmdIllegal")
	}
	for _, ruleStr := range buildFirewalldRichRuleStrings(rule) {
		if err := cmd.NewCommandMgr().Run("firewall-cmd", buildFirewalldRichRuleArgs(ruleStr, operation)...); err != nil {
			return fmt.Errorf("%s rich rules (%s) failed, %v", operation, ruleStr, err)
		}
	}
	return nil
}

func (f *Firewall) ExpandPortRule(rule FireInfo) []PortUnit {
	return expandPortRule(rule, rule.Chain)
}

func (f *Firewall) ApplyPortUnit(unit PortUnit, operation string) error {
	if needsRichRule(unit.Apply) {
		return f.RichRules(unit.Apply, operation)
	}
	return f.Port(unit.Apply, operation)
}

func (f *Firewall) ExpandAddressRule(rule FireInfo) []AddressUnit {
	return expandAddressRule(rule, "")
}

func (f *Firewall) ApplyAddressUnit(unit AddressUnit, operation string) error {
	return f.RichRules(unit.Apply, operation)
}

func (f *Firewall) AddPortWhiteList(list PortWhiteList) error {
	return addNativePortWhiteList(f, list)
}

func (f *Firewall) SyncPortWhiteList(list PortWhiteList) error {
	return syncNativePortWhiteList(f, list)
}

func buildFirewalldPortArgs(port FireInfo, operation string) []string {
	return []string{"--zone=public", "--" + operation + "-port=" + port.Port + "/" + port.Protocol, "--permanent"}
}

func buildFirewalldRichRuleString(rule FireInfo) string {
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
	return ruleStr + rule.Strategy
}

func buildFirewalldRichRuleStrings(rule FireInfo) []string {
	ruleStr := buildFirewalldRichRuleString(rule)
	rules := []string{ruleStr}
	if len(rule.Address) == 0 {
		rules = append(rules, strings.ReplaceAll(ruleStr, "family=ipv4 ", "family=ipv6 "))
	}
	return rules
}

func buildFirewalldRichRuleArgs(ruleStr, operation string) []string {
	return []string{"--zone=public", "--" + operation + "-rich-rule", ruleStr, "--permanent"}
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
