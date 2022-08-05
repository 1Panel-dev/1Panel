package repo

type RepoGroup struct {
	UserRepo
	CommonRepo
}

var RepoGroupApp = new(RepoGroup)
