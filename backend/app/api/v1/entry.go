package v1

import "github.com/1Panel-dev/1Panel/backend/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	authService            = service.ServiceGroupApp.AuthService
	hostService            = service.ServiceGroupApp.HostService
	backupService          = service.ServiceGroupApp.BackupService
	groupService           = service.ServiceGroupApp.GroupService
	containerService       = service.ServiceGroupApp.ContainerService
	composeTemplateService = service.ServiceGroupApp.ComposeTemplateService
	imageRepoService       = service.ServiceGroupApp.ImageRepoService
	imageService           = service.ServiceGroupApp.ImageService
	commandService         = service.ServiceGroupApp.CommandService
	operationService       = service.ServiceGroupApp.OperationService
	fileService            = service.ServiceGroupApp.FileService
	cronjobService         = service.ServiceGroupApp.CronjobService
	settingService         = service.ServiceGroupApp.SettingService
	appService             = service.ServiceGroupApp.AppService
)
