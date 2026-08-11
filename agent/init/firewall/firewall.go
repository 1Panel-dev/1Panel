package firewall

import (
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
)

func Init() {
	if !needInit() {
		if err := initIPv6BaseIfEnabled(); err != nil {
			global.LOG.Errorf("initialize IPv6 iptables rules failed, err: %v", err)
		}
		return
	}
	InitPingStatus()
	global.LOG.Info("initializing firewall settings...")
	if err := service.NewIForwardingService().Replay(); err != nil {
		global.LOG.Errorf("replay forwarding rules failed, err: %v", err)
		return
	}
	client, err := lifecycle.NewClient()
	if err != nil {
		return
	}
	clientName := client.Name()

	if clientName != "iptables" {
		return
	}
	settingRepo := repo.NewISettingRepo()
	panelPort := service.LoadPanelPort()
	if len(panelPort) == 0 {
		global.LOG.Errorf("find 1panel service port failed")
		return
	}
	if err := iptables_helper.RestoreIPv4BaseChains(panelPort); err != nil {
		global.LOG.Errorf("restore iptables base chains failed, err: %v", err)
		return
	}
	global.LOG.Infof("loaded iptables rules for basic from file successfully")
	firewallService := service.NewIFirewallService()
	iptablesStatus, _ := settingRepo.GetValueByKey("IptablesStatus")
	if iptablesStatus == constant.StatusEnable {
		if err := firewallService.OperateFilterChain(dto.IptablesOp{Operate: "bind-base-without-init"}); err != nil {
			global.LOG.Errorf("bind base chains failed, err: %v", err)
			return
		}
	}

}

func initIPv6BaseIfEnabled() error {
	if !cmd.Which("ip6tables") || !cmd.Which("ip6tables-restore") {
		return nil
	}
	client, err := lifecycle.NewClient()
	if err != nil || client.Name() != "iptables" {
		return err
	}
	status, err := repo.NewISettingRepo().GetValueByKey("IptablesStatus")
	if err != nil || status != constant.StatusEnable {
		return err
	}
	panelPort := service.LoadPanelPort()
	if panelPort == "" {
		return fmt.Errorf("find 1panel service port failed")
	}
	return iptables_helper.EnsureIPv6BaseChains(panelPort)
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
	status := ping.LoadStatus()
	statusInDB, _ := repo.NewISettingRepo().GetValueByKey("BanPing")
	if statusInDB == status {
		return
	}

	enable := "1"
	if statusInDB == constant.StatusDisable {
		enable = "0"
	}
	if err := ping.UpdateStatus(enable); err != nil {
		global.LOG.Errorf("initialize ping status failed: %v", err)
	}
}
