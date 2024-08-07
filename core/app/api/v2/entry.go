package v2

import "github.com/1Panel-dev/1Panel/core/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	authService    = service.NewIAuthService()
	settingService = service.NewISettingService()
	logService     = service.NewILogService()
	upgradeService = service.NewIUpgradeService()
)
