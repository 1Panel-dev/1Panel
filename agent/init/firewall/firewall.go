package firewall

import (
	"context"
	"fmt"
	"os"
	"time"

	"errors"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/init/migration/migrations"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"gorm.io/gorm"
)

func Init() {
	ctx := context.Background()
	defer initDockerPortGuard(ctx)
	if err := initForwardingRules(ctx); err != nil {
		global.LOG.Warnf("restore forwarding rules from file failed, err: %v", err)
	}
	client, err := service.NewSelectedSystemFirewallClient()
	if err != nil {
		global.LOG.Errorf("select system firewall provider failed, err: %v", err)
		return
	}
	clientName := client.Name()
	initialize := false
	defer func() {
		if err := migrations.TransferFirewalldSSHService(ctx, client, service.NewIFirewallService().SyncPortWhitelist); err != nil {
			global.LOG.Warnf("synchronize firewall whitelist on startup failed, err: %v", err)
		}
	}()
	families := []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6}
	nftFamilies := []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
	ipv6Status, err := repo.NewISettingRepo().GetValueByKey(constant.FirewallIPv6SupportKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.LOG.Errorf("load firewall IPv6 setting failed: %v", err)
		return
	}
	if ipv6Status == constant.StatusDisable {
		global.LOG.Info("IPv6 firewall support is disabled; skipping IPv6 host chains and saved rules during startup")
		families = families[:1]
		nftFamilies = nftFamilies[:1]
	}
	initialize = needInit()
	if !initialize {
		repairIptablesBaseChains(clientName, families...)
		return
	}
	InitPingStatus()
	global.LOG.Info("initializing firewall settings...")
	if clientName == "nftables" {
		if err := nftables_helper.Restore(nftFamilies...); err != nil {
			global.LOG.Errorf("restore nftables rules failed, err: %v", err)
		}
		status, _ := repo.NewISettingRepo().GetValueByKey("IptablesStatus")
		if status == constant.StatusEnable {
			if err := nftables_helper.Bind(nftFamilies...); err != nil {
				global.LOG.Errorf("bind nftables base chains failed, err: %v", err)
			}
		}
		return
	}

	if clientName != "iptables" {
		return
	}
	settingRepo := repo.NewISettingRepo()
	requiredPorts, err := service.LoadRequiredFirewallPortWhiteList()
	if err != nil {
		global.LOG.Errorf("load required firewall ports failed, err: %v", err)
		return
	}
	if err := iptables_helper.RestoreBaseChains(requiredPorts, families...); err != nil {
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

func repairIptablesBaseChains(clientName string, families ...string) {
	if clientName != constant.FirewallProviderIptables {
		return
	}
	settingRepo := repo.NewISettingRepo()
	status, _ := settingRepo.GetValueByKey("IptablesStatus")
	if status != constant.StatusEnable {
		return
	}
	ports, err := service.LoadRequiredFirewallPortWhiteList()
	if err != nil {
		global.LOG.Warnf("load required firewall ports for base chain repair failed, err: %v", err)
		return
	}
	if err := iptables_helper.RepairBaseChains(ports, families...); err != nil {
		global.LOG.Warnf("repair iptables base chains failed, err: %v", err)
	}
}

func initDockerPortGuard(ctx context.Context) {
	const (
		attempts = 12
		delay    = 5 * time.Second
	)
	var restoreErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		restoreErr = service.RestoreDockerPortGuard(ctx)
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

func initForwardingRules(ctx context.Context) error {
	return service.NewIForwardingService().Restore(ctx)
}
