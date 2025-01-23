package business

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	go syncApp()
	go syncInstalledApp()
	go syncRuntime()
	go syncSSL()
	generateKey()
}

func syncApp() {
	_ = service.NewISettingService().Update("AppStoreSyncStatus", constant.SyncSuccess)
	if err := service.NewIAppService().SyncAppListFromRemote(); err != nil {
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

func generateKey() {
	if err := service.NewISettingService().GenerateRSAKey(); err != nil {
		global.LOG.Errorf("generate rsa key error : %s", err.Error())
	}
}
