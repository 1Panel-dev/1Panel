package service

import "github.com/1Panel-dev/1Panel/backend/app/repo"

var (
	commonRepo = repo.NewCommonRepo()

	appRepo                = repo.NewIAppRepo()
	appTagRepo             = repo.NewIAppTagRepo()
	appDetailRepo          = repo.NewIAppDetailRepo()
	tagRepo                = repo.NewITagRepo()
	appInstallRepo         = repo.NewIAppInstallRepo()
	appInstallResourceRepo = repo.NewIAppInstallResourceRpo()

	mysqlRepo      = repo.NewIMysqlRepo()
	postgresqlRepo = repo.NewIPostgresqlRepo()
	databaseRepo   = repo.NewIDatabaseRepo()

	imageRepoRepo = repo.NewIImageRepoRepo()
	composeRepo   = repo.NewIComposeTemplateRepo()

	cronjobRepo = repo.NewICronjobRepo()

	hostRepo    = repo.NewIHostRepo()
	groupRepo   = repo.NewIGroupRepo()
	commandRepo = repo.NewICommandRepo()
	ftpRepo     = repo.NewIFtpRepo()
	clamRepo    = repo.NewIClamRepo()

	settingRepo = repo.NewISettingRepo()
	backupRepo  = repo.NewIBackupRepo()

	websiteRepo       = repo.NewIWebsiteRepo()
	websiteDomainRepo = repo.NewIWebsiteDomainRepo()
	websiteDnsRepo    = repo.NewIWebsiteDnsAccountRepo()
	websiteSSLRepo    = repo.NewISSLRepo()
	websiteAcmeRepo   = repo.NewIAcmeAccountRepo()
	websiteCARepo     = repo.NewIWebsiteCARepo()

	logRepo      = repo.NewILogRepo()
	snapshotRepo = repo.NewISnapshotRepo()

	runtimeRepo       = repo.NewIRunTimeRepo()
	phpExtensionsRepo = repo.NewIPHPExtensionsRepo()

	favoriteRepo = repo.NewIFavoriteRepo()
)
