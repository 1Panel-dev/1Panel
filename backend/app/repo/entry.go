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
}

var RepoGroupApp = new(RepoGroup)
