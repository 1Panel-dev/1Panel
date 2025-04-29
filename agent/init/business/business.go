package business

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func Init() {
	go syncApp()
	go syncInstalledApp()
	go syncRuntime()
	go syncSSL()
	go syncTask()
	go initAcmeAccount()
}

func syncApp() {
	_ = service.NewISettingService().Update("AppStoreSyncStatus", constant.StatusSyncSuccess)
	if err := service.NewIAppService().SyncAppListFromRemote(""); err != nil {
		global.LOG.Errorf("App Store synchronization failed")
		return
	}
}

func syncInstalledApp() {
	if err := service.NewIAppInstalledService().SyncAll(true); err != nil {
		global.LOG.Errorf("sync installed app error: %s", err.Error())
	}
}

func syncRuntime() {
	if err := service.NewRuntimeService().SyncForRestart(); err != nil {
		global.LOG.Errorf("sync runtime status error : %s", err.Error())
	}
}

func syncSSL() {
	if err := service.NewIWebsiteSSLService().SyncForRestart(); err != nil {
		global.LOG.Errorf("sync ssl status error : %s", err.Error())
	}
}

func syncTask() {
	if err := service.NewITaskService().SyncForRestart(); err != nil {
		global.LOG.Errorf("sync task status error : %s", err.Error())
	}
}

func initAcmeAccount() {
	acmeAccountService := service.NewIWebsiteAcmeAccountService()
	search := dto.PageInfo{
		Page:     1,
		PageSize: 10,
	}
	count, _, _ := acmeAccountService.Page(search)
	if count == 0 {
		createAcmeAccount := request.WebsiteAcmeAccountCreate{
			Email:    "acme@1paneldev.com",
			Type:     "letsencrypt",
			KeyType:  "2048",
			UseProxy: true,
		}
		if _, err := acmeAccountService.Create(createAcmeAccount); err != nil {
			global.LOG.Errorf("create acme account error: %s", err.Error())
		}
	}
}
