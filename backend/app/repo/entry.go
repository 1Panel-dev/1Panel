package repo

type RepoGroup struct {
	HostRepo
	BackupRepo
	GroupRepo
	CommandRepo
	OperationRepo
	CommonRepo
	SettingRepo
}

var RepoGroupApp = new(RepoGroup)
