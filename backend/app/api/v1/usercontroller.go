package v1

import "github.com/1Panel-dev/1Panel/app/service"

type UserController interface {
}

type userController struct {
	userService service.UserService
}

func NewUserController() UserController {
	return &userController{
		userService: service.NewUserService(),
	}
}

func (u userController) GetUser() {

}
