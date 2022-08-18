package repo

type RepoGroup struct {
	UserRepo
	HostRepo
	OperationRepo
	CommonRepo
}

var RepoGroupApp = new(RepoGroup)
