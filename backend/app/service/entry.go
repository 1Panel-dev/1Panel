package service

import "github.com/1Panel-dev/1Panel/app/repo"

type ServiceGroup struct {
	UserService
}

var ServiceGroupApp = new(ServiceGroup)

var (
	userRepo   = repo.RepoGroupApp.UserRepo
	commonRepo = repo.RepoGroupApp.CommonRepo
)
