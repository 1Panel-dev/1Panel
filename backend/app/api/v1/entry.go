package v1

import "github.com/1Panel-dev/1Panel/backend/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	authService      = service.ServiceGroupApp.AuthService
	dashboardService = service.ServiceGroupApp.DashboardService

	appService        = service.NewIAppService()
	appInstallService = service.ServiceGroupApp.AppInstallService

	containerService       = service.ServiceGroupApp.ContainerService
	composeTemplateService = service.ServiceGroupApp.ComposeTemplateService
	imageRepoService       = service.ServiceGroupApp.ImageRepoService
	imageService           = service.ServiceGroupApp.ImageService
	dockerService          = service.ServiceGroupApp.DockerService

	mysqlService = service.ServiceGroupApp.MysqlService
	redisService = service.ServiceGroupApp.RedisService

	cronjobService = service.ServiceGroupApp.CronjobService

	hostService     = service.ServiceGroupApp.HostService
	groupService    = service.ServiceGroupApp.GroupService
	fileService     = service.ServiceGroupApp.FileService
	firewallService = service.NewIFirewallService()

	settingService = service.ServiceGroupApp.SettingService
	backupService  = service.ServiceGroupApp.BackupService

	commandService = service.ServiceGroupApp.CommandService

	websiteService            = service.ServiceGroupApp.WebsiteService
	websiteDnsAccountService  = service.ServiceGroupApp.WebsiteDnsAccountService
	websiteSSLService         = service.ServiceGroupApp.WebsiteSSLService
	websiteAcmeAccountService = service.ServiceGroupApp.WebsiteAcmeAccountService

	nginxService = service.ServiceGroupApp.NginxService

	logService      = service.ServiceGroupApp.LogService
	snapshotService = service.ServiceGroupApp.SnapshotService
	upgradeService  = service.ServiceGroupApp.UpgradeService
)
