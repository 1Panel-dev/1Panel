package job

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type app struct{}

func NewAppStoreJob() *app {
	return &app{}
}

func (a *app) Run() {
	global.LOG.Info("AppStore scheduled task in progress ...")
	appService := service.NewIAppService()
	if err := appService.SyncAppListFromRemote(); err != nil {
		global.LOG.Errorf("AppStore sync failed %s", err.Error())
	}
	appService.SyncAppListFromLocal()
	global.LOG.Info("AppStore scheduled task has completed")
}
