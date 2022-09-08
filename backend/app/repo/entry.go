package repo

type RepoGroup struct {
	HostRepo
	GroupRepo
	CommandRepo
	OperationRepo
	CommonRepo
	SettingRepo
}

var RepoGroupApp = new(RepoGroup)
