package service

import "github.com/1Panel-dev/1Panel/agent/app/repo"

var (
	appRepo                = repo.NewIAppRepo()
	appTagRepo             = repo.NewIAppTagRepo()
	appDetailRepo          = repo.NewIAppDetailRepo()
	tagRepo                = repo.NewITagRepo()
	appInstallRepo         = repo.NewIAppInstallRepo()
	launcherRepo           = repo.NewILauncherRepo()
	appInstallResourceRepo = repo.NewIAppInstallResourceRpo()
	appIgnoreUpgradeRepo   = repo.NewIAppIgnoreUpgradeRepo()

	aiRepo                = repo.NewIAiRepo()
	mcpServerRepo         = repo.NewIMcpServerRepo()
	tensorrtLLMRepo       = repo.NewITensorRTLLMRepo()
	agentRepo             = repo.NewIAgentRepo()
	agentAccountRepo      = repo.NewIAgentAccountRepo()
	agentAccountModelRepo = repo.NewIAgentAccountModelRepo()

	mysqlRepo             = repo.NewIMysqlRepo()
	postgresqlRepo        = repo.NewIPostgresqlRepo()
	mongodbRepo           = repo.NewIMongodbRepo()
	databaseRepo          = repo.NewIDatabaseRepo()
	databaseUserRepo      = repo.NewIDatabaseUserRepo()
	databaseUserGrantRepo = repo.NewIDatabaseUserGrantRepo()

	imageRepoRepo = repo.NewIImageRepoRepo()
	composeRepo   = repo.NewIComposeTemplateRepo()

	scriptRepo  = repo.NewIScriptRepo()
	cronjobRepo = repo.NewICronjobRepo()

	hostRepo    = repo.NewIHostRepo()
	ftpRepo     = repo.NewIFtpRepo()
	clamRepo    = repo.NewIClamRepo()
	monitorRepo = repo.NewIMonitorRepo()

	settingRepo = repo.NewISettingRepo()
	backupRepo  = repo.NewIBackupRepo()

	websiteRepo               = repo.NewIWebsiteRepo()
	websiteDomainRepo         = repo.NewIWebsiteDomainRepo()
	websiteDnsRepo            = repo.NewIWebsiteDnsAccountRepo()
	websiteSSLRepo            = repo.NewISSLRepo()
	websiteAcmeRepo           = repo.NewIAcmeAccountRepo()
	websiteCARepo             = repo.NewIWebsiteCARepo()
	websiteTemplateRepo       = repo.NewIWebsiteTemplateRepo()
	websiteTemplateOutputRepo = repo.NewIWebsiteTemplateOutputRepo()

	snapshotRepo = repo.NewISnapshotRepo()

	runtimeRepo       = repo.NewIRunTimeRepo()
	phpExtensionsRepo = repo.NewIPHPExtensionsRepo()

	favoriteRepo  = repo.NewIFavoriteRepo()
	fileShareRepo = repo.NewIFileShareRepo()

	taskRepo = repo.NewITaskRepo()

	groupRepo = repo.NewIGroupRepo()
	alertRepo = repo.NewIAlertRepo()
)
