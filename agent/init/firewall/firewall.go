package firewall

import (
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

func Init() {
	if !needInit() {
		return
	}
	InitPingStatus()
	global.LOG.Info("initializing firewall settings...")
	provider, err := firewall.DetectProvider()
	if err != nil {
		return
	}
	if err := service.NewIForwardingService().Replay(); err != nil {
		global.LOG.Errorf("replay forwarding rules failed, err: %v", err)
		return
	}

	if provider != "iptables" {
		return
	}
	settingRepo := repo.NewISettingRepo()
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
	panelPort := service.LoadPanelPort()
	if len(panelPort) == 0 {
		global.LOG.Errorf("find 1panel service port failed")
		return
	}
	if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicBefore, "-p", "tcp", "-m", "tcp", "--dport", panelPort, "-j", "ACCEPT"); err != nil {
		global.LOG.Errorf("add port accept rule %v failed, err: %v", panelPort, err)
		return
	}
	global.LOG.Infof("loaded iptables rules for basic from file successfully")
	iptablesService := service.IptablesService{}
	iptablesStatus, _ := settingRepo.GetValueByKey("IptablesStatus")
	if iptablesStatus == constant.StatusEnable {
		if err := iptablesService.Operate(dto.IptablesOp{Operate: "bind-base-without-init"}); err != nil {
			global.LOG.Errorf("bind base chains failed, err: %v", err)
			return
		}
	}

	if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelInput, iptables.InputFileName); err != nil {
		global.LOG.Errorf("load input rules from file failed, err: %v", err)
		return
	}
	if err := iptables.LoadRulesFromFile(iptables.FilterTab, iptables.Chain1PanelOutput, iptables.OutputFileName); err != nil {
		global.LOG.Errorf("load output rules from file failed, err: %v", err)
		return
	}
	global.LOG.Infof("loaded iptables rules for input and output from file successfully")
	iptablesInputStatus, _ := settingRepo.GetValueByKey("IptablesInputStatus")
	if iptablesInputStatus == constant.StatusEnable {
		if err := iptablesService.Operate(dto.IptablesOp{Name: iptables.Chain1PanelInput, Operate: "bind"}); err != nil {
			global.LOG.Errorf("bind input chains failed, err: %v", err)
			return
		}
	}
	iptablesOutputStatus, _ := settingRepo.GetValueByKey("IptablesOutputStatus")
	if iptablesOutputStatus == constant.StatusEnable {
		if err := iptablesService.Operate(dto.IptablesOp{Name: iptables.Chain1PanelOutput, Operate: "bind"}); err != nil {
			global.LOG.Errorf("bind output chains failed, err: %v", err)
			return
		}
	}
}

func needInit() bool {
	file, err := os.OpenFile("/run/1panel_boot_mark", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return false
		}
		global.LOG.Errorf("check boot mark file failed: %v", err)
		return true
	}
	defer file.Close()
	fmt.Fprintf(file, "Boot Mark for 1panel\n")
	return true
}

func InitPingStatus() {
	global.LOG.Info("initializing ban ping status from settings...")
	status := firewall.LoadPingStatus()
	statusInDB, _ := repo.NewISettingRepo().GetValueByKey("BanPing")
	if statusInDB == status {
		return
	}

	enable := "1"
	if statusInDB == constant.StatusDisable {
		enable = "0"
	}
	if err := firewall.UpdatePingStatus(enable); err != nil {
		global.LOG.Errorf("initialize ping status failed: %v", err)
	}
}
