package service

import "github.com/1Panel-dev/1Panel/app/repo"

type ServiceGroup struct {
	UserService
	HostService
	OperationService
}

var ServiceGroupApp = new(ServiceGroup)

var (
	userRepo      = repo.RepoGroupApp.UserRepo
	hostRepo      = repo.RepoGroupApp.HostRepo
	operationRepo = repo.RepoGroupApp.OperationRepo
	commonRepo    = repo.RepoGroupApp.CommonRepo
)
