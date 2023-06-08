package business

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	syncApp()
	syncInstalledApp()
}

func syncApp() {
	setting, err := service.NewISettingService().GetSettingInfo()
	if err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
		return
	}
	if setting.AppStoreLastModified != "0" {
		global.LOG.Info("no need to sync")
		return
	}
	global.LOG.Info("sync app start...")
	if err := service.NewIAppService().SyncAppListFromRemote(); err != nil {
		global.LOG.Errorf("sync app error")
		return
	}
	global.LOG.Info("sync app successful")
}

func syncInstalledApp() {
	if err := service.NewIAppInstalledService().SyncAll(true); err != nil {
		global.LOG.Errorf("sync instaled app error: %s", err.Error())
	}
}
