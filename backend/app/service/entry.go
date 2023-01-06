package service

import "github.com/1Panel-dev/1Panel/backend/app/repo"

type ServiceGroup struct {
	AuthService
	DashboardService

	AppService
	AppInstallService

	ContainerService
	ImageService
	ImageRepoService
	ComposeTemplateService
	DockerService

	MysqlService
	RedisService

	CronjobService

	HostService
	GroupService
	CommandService
	FileService

	SettingService
	BackupService

	WebsiteGroupService
	WebsiteService
	WebsiteDnsAccountService
	WebsiteSSLService
	WebsiteAcmeAccountService

	NginxService

	LogService
	SnapshotService
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

	mysqlRepo = repo.RepoGroupApp.MysqlRepo

	imageRepoRepo = repo.RepoGroupApp.ImageRepoRepo
	composeRepo   = repo.RepoGroupApp.ComposeTemplateRepo

	cronjobRepo = repo.RepoGroupApp.CronjobRepo

	hostRepo    = repo.RepoGroupApp.HostRepo
	groupRepo   = repo.RepoGroupApp.GroupRepo
	commandRepo = repo.RepoGroupApp.CommandRepo

	settingRepo = repo.RepoGroupApp.SettingRepo
	backupRepo  = repo.RepoGroupApp.BackupRepo

	websiteRepo       = repo.NewIWebsiteRepo()
	websiteGroupRepo  = repo.RepoGroupApp.WebsiteGroupRepo
	websiteDomainRepo = repo.RepoGroupApp.WebsiteDomainRepo
	websiteDnsRepo    = repo.RepoGroupApp.WebsiteDnsAccountRepo
	websiteSSLRepo    = repo.NewISSLRepo()
	websiteAcmeRepo   = repo.NewIAcmeAccountRepo()

	logRepo      = repo.RepoGroupApp.LogRepo
	snapshotRepo = repo.NewISnapshotRepo()
)
