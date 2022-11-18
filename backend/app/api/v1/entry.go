package v1

import "github.com/1Panel-dev/1Panel/backend/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	authService = service.ServiceGroupApp.AuthService

	appService        = service.ServiceGroupApp.AppService
	appInstallService = service.ServiceGroupApp.AppInstallService

	containerService       = service.ServiceGroupApp.ContainerService
	composeTemplateService = service.ServiceGroupApp.ComposeTemplateService
	imageRepoService       = service.ServiceGroupApp.ImageRepoService
	imageService           = service.ServiceGroupApp.ImageService

	mysqlService = service.ServiceGroupApp.MysqlService
	redisService = service.ServiceGroupApp.RedisService

	cronjobService = service.ServiceGroupApp.CronjobService

	hostService  = service.ServiceGroupApp.HostService
	groupService = service.ServiceGroupApp.GroupService
	fileService  = service.ServiceGroupApp.FileService

	settingService = service.ServiceGroupApp.SettingService
	backupService  = service.ServiceGroupApp.BackupService

	operationService = service.ServiceGroupApp.OperationService
	commandService   = service.ServiceGroupApp.CommandService

	websiteGroupService       = service.ServiceGroupApp.WebsiteGroupService
	websiteService            = service.ServiceGroupApp.WebsiteService
	websiteDnsAccountService  = service.ServiceGroupApp.WebSiteDnsAccountService
	websiteSSLService         = service.ServiceGroupApp.WebSiteSSLService
	websiteAcmeAccountService = service.ServiceGroupApp.WebSiteAcmeAccountService
)
