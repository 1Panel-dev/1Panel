package service

import "github.com/1Panel-dev/1Panel/backend/app/repo"

type ServiceGroup struct {
	AuthService

	AppService
	AppInstallService

	ContainerService
	ImageService
	ImageRepoService
	ComposeTemplateService

	MysqlService
	RedisService

	CronjobService

	HostService
	GroupService
	CommandService
	FileService

	SettingService
	BackupService

	OperationService
	WebsiteGroupService
	WebsiteService
	WebSiteDnsAccountService
	WebSiteSSLService
	WebSiteAcmeAccountService
}

var ServiceGroupApp = new(ServiceGroup)

var (
	commonRepo = repo.RepoGroupApp.CommonRepo

	appInstallBackupRepo   = repo.RepoGroupApp.AppInstallBackupRepo
	appRepo                = repo.RepoGroupApp.AppRepo
	appTagRepo             = repo.RepoGroupApp.AppTagRepo
	appDetailRepo          = repo.RepoGroupApp.AppDetailRepo
	tagRepo                = repo.RepoGroupApp.TagRepo
	appInstallRepo         = repo.RepoGroupApp.AppInstallRepo
	appInstallResourceRepo = repo.RepoGroupApp.AppInstallResourceRpo
	dataBaseRepo           = repo.RepoGroupApp.DatabaseRepo

	mysqlRepo = repo.RepoGroupApp.MysqlRepo

	imageRepoRepo = repo.RepoGroupApp.ImageRepoRepo
	composeRepo   = repo.RepoGroupApp.ComposeTemplateRepo

	cronjobRepo = repo.RepoGroupApp.CronjobRepo

	hostRepo    = repo.RepoGroupApp.HostRepo
	groupRepo   = repo.RepoGroupApp.GroupRepo
	commandRepo = repo.RepoGroupApp.CommandRepo

	settingRepo = repo.RepoGroupApp.SettingRepo
	backupRepo  = repo.RepoGroupApp.BackupRepo

	operationRepo     = repo.RepoGroupApp.OperationRepo
	websiteRepo       = repo.RepoGroupApp.WebSiteRepo
	websiteGroupRepo  = repo.RepoGroupApp.WebSiteGroupRepo
	websiteDomainRepo = repo.RepoGroupApp.WebSiteDomainRepo
	websiteDnsRepo    = repo.RepoGroupApp.WebsiteDnsAccountRepo
	websiteSSLRepo    = repo.RepoGroupApp.WebsiteSSLRepo
	websiteAcmeRepo   = repo.RepoGroupApp.WebsiteAcmeAccountRepo
)
