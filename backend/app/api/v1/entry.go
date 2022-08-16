package v1

import "github.com/1Panel-dev/1Panel/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	userService      = service.ServiceGroupApp.UserService
	operationService = service.ServiceGroupApp.OperationService
)
