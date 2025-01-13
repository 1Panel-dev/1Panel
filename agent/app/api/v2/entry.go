package v2

import "github.com/1Panel-dev/1Panel/agent/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

type BaseApi struct{}

var (
	dashboardService = service.NewIDashboardService()

	appService        = service.NewIAppService()
	appInstallService = service.NewIAppInstalledService()

	containerService       = service.NewIContainerService()
	composeTemplateService = service.NewIComposeTemplateService()
	imageRepoService       = service.NewIImageRepoService()
	imageService           = service.NewIImageService()
	dockerService          = service.NewIDockerService()

	dbCommonService   = service.NewIDBCommonService()
	mysqlService      = service.NewIMysqlService()
	postgresqlService = service.NewIPostgresqlService()
	databaseService   = service.NewIDatabaseService()
	redisService      = service.NewIRedisService()

	cronjobService = service.NewICronjobService()

	fileService     = service.NewIFileService()
	sshService      = service.NewISSHService()
	firewallService = service.NewIFirewallService()
	monitorService  = service.NewIMonitorService()

	deviceService   = service.NewIDeviceService()
	fail2banService = service.NewIFail2BanService()
	ftpService      = service.NewIFtpService()
	clamService     = service.NewIClamService()

	settingService      = service.NewISettingService()
	backupService       = service.NewIBackupService()
	backupRecordService = service.NewIBackupRecordService()

	websiteService            = service.NewIWebsiteService()
	websiteDnsAccountService  = service.NewIWebsiteDnsAccountService()
	websiteSSLService         = service.NewIWebsiteSSLService()
	websiteAcmeAccountService = service.NewIWebsiteAcmeAccountService()

	nginxService = service.NewINginxService()

	logService      = service.NewILogService()
	snapshotService = service.NewISnapshotService()

	runtimeService       = service.NewRuntimeService()
	processService       = service.NewIProcessService()
	phpExtensionsService = service.NewIPHPExtensionsService()

	hostToolService = service.NewIHostToolService()

	recycleBinService = service.NewIRecycleBinService()
	favoriteService   = service.NewIFavoriteService()

	websiteCAService = service.NewIWebsiteCAService()
	taskService      = service.NewITaskService()
)
