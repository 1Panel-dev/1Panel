package business

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	setting, err := service.NewISettingService().GetSettingInfo()
	if err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
	}
	if setting.AppStoreVersion != "" {
		return
	}
	if err := service.NewIAppService().SyncAppList(); err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
	}
}
