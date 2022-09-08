package service

import "github.com/1Panel-dev/1Panel/app/repo"

type ServiceGroup struct {
	AuthService
	HostService
	GroupService
	CommandService
	OperationService
	FileService
	SettingService
}

var ServiceGroupApp = new(ServiceGroup)

var (
	hostRepo      = repo.RepoGroupApp.HostRepo
	groupRepo     = repo.RepoGroupApp.GroupRepo
	commandRepo   = repo.RepoGroupApp.CommandRepo
	operationRepo = repo.RepoGroupApp.OperationRepo
	commonRepo    = repo.RepoGroupApp.CommonRepo
	settingRepo   = repo.RepoGroupApp.SettingRepo
)
