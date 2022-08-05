package v1

import "1Panel/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

var (
	userService = service.ServiceGroupApp.UserService
)
