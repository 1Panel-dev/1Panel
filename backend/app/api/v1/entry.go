package v1

import "github.com/1Panel-dev/1Panel/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	authService      = service.ServiceGroupApp.AuthService
	hostService      = service.ServiceGroupApp.HostService
	backupService    = service.ServiceGroupApp.BackupService
	groupService     = service.ServiceGroupApp.GroupService
	commandService   = service.ServiceGroupApp.CommandService
	operationService = service.ServiceGroupApp.OperationService
	fileService      = service.ServiceGroupApp.FileService
	cronjobService   = service.ServiceGroupApp.CronjobService
	settingService   = service.ServiceGroupApp.SettingService
)
