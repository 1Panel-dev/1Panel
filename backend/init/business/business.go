package business

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	setting, err := service.NewISettingService().GetSettingInfo()
	if err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
		return
	}
	if setting.AppStoreVersion != "" {
		fmt.Println(setting.AppStoreVersion)
		global.LOG.Info("do not sync")
		return
	}
	global.LOG.Info("sync app start...")
	if err := service.NewIAppService().SyncAppList(); err != nil {
		global.LOG.Errorf("sync app error: %s", err.Error())
	}
	global.LOG.Info("sync app success")
}
