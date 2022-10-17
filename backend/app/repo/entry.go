package repo

type RepoGroup struct {
	HostRepo
	BackupRepo
	GroupRepo
	ImageRepoRepo
	ComposeTemplateRepo
	CommandRepo
	OperationRepo
	CommonRepo
	CronjobRepo
	SettingRepo
	AppRepo
	AppTagRepo
	TagRepo
	AppDetailRepo
	AppInstallRepo
	AppInstallResourceRpo
	DatabaseRepo
	AppInstallBackupRepo
}

var RepoGroupApp = new(RepoGroup)
