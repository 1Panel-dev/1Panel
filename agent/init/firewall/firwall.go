package firewall

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

func Init() {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return
	}
	clientName := client.Name()
	if clientName == "ufw" || clientName == "iptables" {
		if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelForward, iptables.ForwardFileName); err != nil {
			global.LOG.Errorf("load forward rules from file failed, err: %v", err)
			return
		}
		if err := iptables.LoadRulesFromFile(iptables.NatTab, iptables.Chain1PanelPreRouting, iptables.ForwardFileName1); err != nil {
			global.LOG.Errorf("load prerouting rules from file failed, err: %v", err)
			return
		}
		if err := iptables.LoadRulesFromFile(iptables.NatTab, iptables.Chain1PanelPostRouting, iptables.ForwardFileName2); err != nil {
			global.LOG.Errorf("load postrouting rules from file failed, err: %v", err)
			return
		}
		global.LOG.Infof("loaded iptables rules for forward from file successfully")
	}
	if clientName == "ufw" {
		_ = iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicAfter)
		_ = iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicBefore)
		_ = iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasic)
		_ = iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelInput)
		_ = iptables.UnbindChain(iptables.FilterTab, iptables.ChainOutput, iptables.Chain1PanelOutput)
	}
	if clientName == "iptables" {
		if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.BasicBeforeFileName); err != nil {
			global.LOG.Errorf("load basic before rules from file failed, err: %v", err)
			return
		}
		if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelBasic, iptables.BasicFileName); err != nil {
			global.LOG.Errorf("load basic rules from file failed, err: %v", err)
			return
		}
		if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.BasicAfterFileName); err != nil {
			global.LOG.Errorf("load basic after rules from file failed, err: %v", err)
			return
		}
		if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelInput, iptables.InputFileName); err != nil {
			global.LOG.Errorf("load input rules from file failed, err: %v", err)
			return
		}
		if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelOutput, iptables.OutputFileName); err != nil {
			global.LOG.Errorf("load output rules from file failed, err: %v", err)
			return
		}
		global.LOG.Infof("loaded iptables rules for basic, input and output from file successfully")

		panelPort := service.LoadPanelPort()
		if len(panelPort) == 0 {
			global.LOG.Errorf("find 1panel service port failed")
			return
		}
		if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicBefore, fmt.Sprintf("-p tcp -m tcp --dport %v -j ACCEPT", panelPort)); err != nil {
			global.LOG.Errorf("add port accept rule %v failed, err: %v", panelPort, err)
			return
		}

		iptablesService := service.IptablesService{}
		if err := iptablesService.Operate(dto.IptablesOp{Operate: "bind-base"}); err != nil {
			global.LOG.Errorf("bind base chains failed, err: %v", err)
			return
		}
		if err := iptablesService.Operate(dto.IptablesOp{Name: iptables.Chain1PanelOutput, Operate: "bind"}); err != nil {
			global.LOG.Errorf("bind output chains failed, err: %v", err)
			return
		}
		if err := iptablesService.Operate(dto.IptablesOp{Name: iptables.Chain1PanelInput, Operate: "bind"}); err != nil {
			global.LOG.Errorf("bind input chains failed, err: %v", err)
			return
		}
	}

}
