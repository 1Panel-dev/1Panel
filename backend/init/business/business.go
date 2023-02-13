package business

import (
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
)

func Init() {
	setting, err := service.NewISettingService().GetSettingInfo()
	if err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
	}
	if common.CompareVersion(setting.AppStoreVersion, "0.0") {
		return
	}
	appService := service.AppService{}
	if err := appService.SyncAppList(); err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
	}
}
