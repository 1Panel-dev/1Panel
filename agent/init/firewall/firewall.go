package firewall

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	migrationutils "github.com/1Panel-dev/1Panel/agent/init/migration/migrations/utils"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
)

func Init() {
	ctx := context.Background()
	client, err := service.NewSelectedSystemFirewallClient()
	if err != nil {
		global.LOG.Errorf("select system firewall provider failed, err: %v", err)
		return
	}
	clientName := client.Name()
	if err := migrationutils.TransferHostFirewall(ctx, clientName); err != nil {
		global.LOG.Errorf("transfer legacy host firewall records failed, err: %v", err)
		return
	}
	if err := migrationutils.TransferLegacyHostFirewallRuleOwnership(ctx, clientName, service.AdoptLegacyHostFirewallRuleOwnership); err != nil {
		global.LOG.Warnf("transfer legacy host firewall rule ownership failed, err: %v", err)
	}
	if err := migrationutils.TransferFirewallForwarding(ctx); err != nil {
		global.LOG.Errorf("transfer legacy forwarding rules failed, err: %v", err)
		return
	}
	if err := initForwardingRules(ctx); err != nil {
		global.LOG.Warnf("restore forwarding rules failed, manual synchronization is available, err: %v", err)
	}
	if !needInit() {
		repairIptablesIPv6BaseChains(clientName)
		return
	}
	defer initDockerPortGuard(ctx)
	InitPingStatus()
	global.LOG.Info("initializing firewall settings...")
	if clientName == "nftables" {
		if err := nftables_helper.Restore(); err != nil {
			global.LOG.Errorf("restore nftables rules failed, err: %v", err)
		}
		status, _ := repo.NewISettingRepo().GetValueByKey("IptablesStatus")
		if status == constant.StatusEnable {
			if err := nftables_helper.Bind(); err != nil {
				global.LOG.Errorf("bind nftables base chains failed, err: %v", err)
			}
		}
		return
	}

	if clientName != "iptables" {
		return
	}
	settingRepo := repo.NewISettingRepo()
	panelPort := service.LoadPanelPort()
	if len(panelPort) == 0 {
		global.LOG.Errorf("find 1panel service port failed")
		return
	}
	if err := iptables_helper.RestoreBaseChains(panelPort); err != nil {
		global.LOG.Errorf("restore iptables base chains failed, err: %v", err)
		return
	}
	global.LOG.Infof("loaded iptables rules for basic from file successfully")
	firewallService := service.NewIFirewallService()
	iptablesStatus, _ := settingRepo.GetValueByKey("IptablesStatus")
	if iptablesStatus == constant.StatusEnable {
		if err := firewallService.OperateFilterChain(dto.FilterChainOperation{Operate: string(firewall.BaseOperationBindWithoutInit)}); err != nil {
			global.LOG.Errorf("bind base chains failed, err: %v", err)
			return
		}
	}

}

func repairIptablesIPv6BaseChains(clientName string) {
	if clientName != constant.FirewallProviderIptables {
		return
	}
	settingRepo := repo.NewISettingRepo()
	status, _ := settingRepo.GetValueByKey("IptablesStatus")
	if status != constant.StatusEnable {
		return
	}
	initialized, bound, err := iptables_helper.LoadFamilyInitStatus(constant.FirewallFamilyIPv6, "base")
	if err == nil && initialized && bound {
		return
	}
	panelPort := service.LoadPanelPort()
	if err := iptables_helper.RepairIPv6BaseChains(panelPort); err != nil {
		global.LOG.Warnf("repair IPv6 iptables base chains failed, err: %v", err)
		return
	}
	global.LOG.Info("repaired IPv6 iptables base chains successfully")
}

func initDockerPortGuard(ctx context.Context) {
	const (
		attempts = 12
		delay    = 5 * time.Second
	)
	var restoreErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		restoreErr = service.ReconcileDockerPortGuard(ctx)
		if restoreErr == nil {
			return
		}
		if attempt == attempts {
			break
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			global.LOG.Warnf("restore Docker port guard on startup canceled, err: %v", ctx.Err())
			return
		case <-timer.C:
		}
	}
	global.LOG.Warnf("restore Docker port guard on startup failed after %d attempts, err: %v", attempts, restoreErr)
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

func initForwardingRules(ctx context.Context) error {
	return service.NewIForwardingService().Restore(ctx)
}
