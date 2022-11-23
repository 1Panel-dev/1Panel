package business

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	appService := service.AppService{}
	if err := appService.SyncAppList(); err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
	}
}
