package v1

import "github.com/1Panel-dev/1Panel/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	userService      = service.ServiceGroupApp.UserService
	hostService      = service.ServiceGroupApp.HostService
	groupService     = service.ServiceGroupApp.GroupService
	commandService   = service.ServiceGroupApp.CommandService
	operationService = service.ServiceGroupApp.OperationService
	fileService      = service.ServiceGroupApp.FileService
)
