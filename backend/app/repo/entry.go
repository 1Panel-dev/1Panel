package repo

type RepoGroup struct {
	UserRepo
	HostRepo
	GroupRepo
	CommandRepo
	OperationRepo
	CommonRepo
	SettingRepo
}

var RepoGroupApp = new(RepoGroup)
