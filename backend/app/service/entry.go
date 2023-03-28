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

	mysqlRepo = repo.NewIMysqlRepo()

	imageRepoRepo = repo.NewIImageRepoRepo()
	composeRepo   = repo.NewIComposeTemplateRepo()

	cronjobRepo = repo.NewICronjobRepo()

	hostRepo    = repo.NewIHostRepo()
	groupRepo   = repo.NewIGroupRepo()
	commandRepo = repo.NewICommandRepo()

	settingRepo = repo.NewISettingRepo()
	backupRepo  = repo.NewIBackupRepo()

	websiteRepo       = repo.NewIWebsiteRepo()
	websiteDomainRepo = repo.NewIWebsiteDomainRepo()
	websiteDnsRepo    = repo.NewIWebsiteDnsAccountRepo()
	websiteSSLRepo    = repo.NewISSLRepo()
	websiteAcmeRepo   = repo.NewIAcmeAccountRepo()

	logRepo      = repo.NewILogRepo()
	snapshotRepo = repo.NewISnapshotRepo()

	runtimeRepo = repo.NewIRunTimeRepo()
)
