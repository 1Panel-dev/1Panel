package repo

type RepoGroup struct {
	CommonRepo
	AppRepo
	AppTagRepo
	TagRepo
	AppDetailRepo
	AppInstallRepo
	AppInstallResourceRpo
	DatabaseRepo
	AppInstallBackupRepo
	ImageRepoRepo
	ComposeTemplateRepo
	MysqlRepo
	CronjobRepo
	HostRepo
	CommandRepo
	GroupRepo
	SettingRepo
	BackupRepo
	WebSiteRepo
	WebSiteDomainRepo
	WebSiteGroupRepo
	WebsiteDnsAccountRepo
	WebsiteSSLRepo
	WebsiteAcmeAccountRepo
	LogRepo
}

var RepoGroupApp = new(RepoGroup)
