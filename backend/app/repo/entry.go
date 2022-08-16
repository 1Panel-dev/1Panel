package repo

type RepoGroup struct {
	UserRepo
	OperationRepo
	CommonRepo
}

var RepoGroupApp = new(RepoGroup)
