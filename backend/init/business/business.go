package business

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	go syncApp()
	go syncInstalledApp()
}

func syncApp() {
	if err := service.NewIAppService().SyncAppListFromRemote(); err != nil {
		global.LOG.Errorf("App Store synchronization failed")
		return
	}
}

func syncInstalledApp() {
	if err := service.NewIAppInstalledService().SyncAll(true); err != nil {
		global.LOG.Errorf("sync instaled app error: %s", err.Error())
	}
}
