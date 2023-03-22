package repo

type RepoGroup struct {
	CommonRepo
	AppRepo
	AppTagRepo
	TagRepo
	AppDetailRepo
	AppInstallRepo
	AppInstallResourceRpo
	ImageRepoRepo
	ComposeTemplateRepo
	MysqlRepo
	CronjobRepo
	HostRepo
	CommandRepo
	GroupRepo
	SettingRepo
	BackupRepo
	WebsiteRepo
	WebsiteDomainRepo
	WebsiteDnsAccountRepo
	WebsiteSSLRepo
	WebsiteAcmeAccountRepo
	LogRepo
	SnapshotRepo
}

var RepoGroupApp = new(RepoGroup)
