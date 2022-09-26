package repo

type RepoGroup struct {
	HostRepo
	BackupRepo
	GroupRepo
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
}

var RepoGroupApp = new(RepoGroup)
