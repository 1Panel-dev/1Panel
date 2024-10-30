package v2

import "github.com/1Panel-dev/1Panel/core/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	hostService        = service.NewIHostService()
	authService        = service.NewIAuthService()
	backupService      = service.NewIBackupService()
	settingService     = service.NewISettingService()
	logService         = service.NewILogService()
	upgradeService     = service.NewIUpgradeService()
	groupService       = service.NewIGroupService()
	commandService     = service.NewICommandService()
	appLauncherService = service.NewIAppLauncher()
)
