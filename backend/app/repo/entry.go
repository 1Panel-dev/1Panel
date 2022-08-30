package repo

type RepoGroup struct {
	UserRepo
	HostRepo
	GroupRepo
	CommandRepo
	OperationRepo
	CommonRepo
}

var RepoGroupApp = new(RepoGroup)
