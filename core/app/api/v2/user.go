package v2

import (
	"net/http"
	"strconv"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

type UserApi struct{}

var userService = service.NewIUserService()

// @Tags User
// @Summary Create a new user
// @Accept json
// @Param request body dto.CreateUserRequest true "create user request"
// @Success 200 {object} dto.UserResponse
// @Router /core/users [post]
func (u *UserApi) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// Only admin can create users
	if err := helper.CheckUserRole(c, "admin"); err != nil {
		helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
		return
	}

	user, err := userService.CreateUser(&req)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrCreateUserFailed", err)
		return
	}

	helper.SuccessWithData(c, user)
}

// @Tags User
// @Summary Update user information
// @Accept json
// @Param request body dto.UpdateUserRequest true "update user request"
// @Success 200
// @Router /core/users [put]
func (u *UserApi) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// Only admin or the user themselves can update
	currentUserID, err := helper.GetCurrentUserID(c)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusUnauthorized, "ErrUnauthorized", err)
		return
	}

	if currentUserID != req.ID {
		if err := helper.CheckUserRole(c, "admin"); err != nil {
			helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
			return
		}
	}

	if err := userService.UpdateUser(&req); err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrUpdateUserFailed", err)
		return
	}

	helper.SuccessWithMsg(c, "UserUpdatedSuccessfully")
}

// @Tags User
// @Summary Get user by ID
// @Param id path int true "user id"
// @Success 200 {object} dto.UserDetailResponse
// @Router /core/users/:id [get]
func (u *UserApi) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidUserID", err)
		return
	}

	userDetail, err := userService.GetUserByID(uint(id))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusNotFound, "ErrUserNotFound", err)
		return
	}

	helper.SuccessWithData(c, userDetail)
}

// @Tags User
// @Summary Get user list with pagination
// @Accept json
// @Param request body dto.UserPageRequest true "user page request"
// @Success 200 {object} dto.UserListResponse
// @Router /core/users [get]
func (u *UserApi) ListUsers(c *gin.Context) {
	pageNum := c.DefaultQuery("pageNum", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	username := c.Query("username")
	role := c.Query("role")
	status := c.Query("status")

	num, _ := strconv.Atoi(pageNum)
	size, _ := strconv.Atoi(pageSize)

	if num < 1 {
		num = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	req := &dto.UserPageRequest{
		PageNum:  num,
		PageSize: size,
		Username: username,
		Role:     role,
		Status:   status,
	}

	userList, err := userService.GetUserList(req)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrGetUserListFailed", err)
		return
	}

	helper.SuccessWithData(c, userList)
}

// @Tags User
// @Summary Delete user
// @Param id path int true "user id"
// @Success 200
// @Router /core/users/:id [delete]
func (u *UserApi) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidUserID", err)
		return
	}

	// Only admin can delete users
	if err := helper.CheckUserRole(c, "admin"); err != nil {
		helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
		return
	}

	if err := userService.DeleteUser(uint(id)); err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrDeleteUserFailed", err)
		return
	}

	helper.SuccessWithMsg(c, "UserDeletedSuccessfully")
}

// @Tags User
// @Summary Change user password
// @Accept json
// @Param request body dto.ChangePasswordRequest true "change password request"
// @Success 200
// @Router /core/users/password/change [post]
func (u *UserApi) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// User can only change their own password
	currentUserID, err := helper.GetCurrentUserID(c)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusUnauthorized, "ErrUnauthorized", err)
		return
	}

	if currentUserID != req.UserID {
		// Check if user is admin
		if err := helper.CheckUserRole(c, "admin"); err != nil {
			helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
			return
		}
	}

	if err := userService.ChangePassword(&req); err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrChangePasswordFailed", err)
		return
	}

	helper.SuccessWithMsg(c, "PasswordChangedSuccessfully")
}

// @Tags User
// @Summary Reset user password (admin only)
// @Accept json
// @Param request body dto.ResetPasswordRequest true "reset password request"
// @Success 200
// @Router /core/users/password/reset [post]
func (u *UserApi) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// Only admin can reset passwords
	if err := helper.CheckUserRole(c, "admin"); err != nil {
		helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
		return
	}

	if err := userService.ResetPassword(&req); err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrResetPasswordFailed", err)
		return
	}

	helper.SuccessWithMsg(c, "PasswordResetSuccessfully")
}

// @Tags User
// @Summary Assign permissions to user
// @Accept json
// @Param request body dto.AssignPermissionRequest true "assign permission request"
// @Success 200
// @Router /core/users/permissions [post]
func (u *UserApi) AssignPermissions(c *gin.Context) {
	var req dto.AssignPermissionRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	// Only admin can assign permissions
	if err := helper.CheckUserRole(c, "admin"); err != nil {
		helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
		return
	}

	if err := userService.AssignPermissions(&req); err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrAssignPermissionFailed", err)
		return
	}

	helper.SuccessWithMsg(c, "PermissionsAssignedSuccessfully")
}

// @Tags User
// @Summary Get user permissions
// @Param id path int true "user id"
// @Success 200 {object} []string
// @Router /core/users/:id/permissions [get]
func (u *UserApi) GetUserPermissions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidUserID", err)
		return
	}

	permissions, err := userService.GetUserPermissions(uint(id))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrGetPermissionsFailed", err)
		return
	}

	helper.SuccessWithData(c, permissions)
}

// @Tags User
// @Summary Get current user profile
// @Success 200 {object} dto.UserDetailResponse
// @Router /core/users/profile [get]
func (u *UserApi) GetProfile(c *gin.Context) {
	currentUserID, err := helper.GetCurrentUserID(c)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusUnauthorized, "ErrUnauthorized", err)
		return
	}

	userDetail, err := userService.GetUserByID(currentUserID)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusNotFound, "ErrUserNotFound", err)
		return
	}

	helper.SuccessWithData(c, userDetail)
}

// @Tags User
// @Summary Get login history
// @Param id path int true "user id"
// @Success 200 {object} []model.UserLoginHistory
// @Router /core/users/:id/login-history [get]
func (u *UserApi) GetLoginHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidUserID", err)
		return
	}

	// Only admin or the user themselves can view login history
	currentUserID, err := helper.GetCurrentUserID(c)
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusUnauthorized, "ErrUnauthorized", err)
		return
	}

	if currentUserID != uint(id) {
		if err := helper.CheckUserRole(c, "admin"); err != nil {
			helper.ErrorWithDetail(c, http.StatusForbidden, "ErrUnauthorized", err)
			return
		}
	}

	histories, err := userService.GetLoginHistory(uint(id))
	if err != nil {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrGetLoginHistoryFailed", err)
		return
	}

	helper.SuccessWithData(c, histories)
}
