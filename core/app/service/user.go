package service

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
)

type UserService struct {
	userRepo repo.IUserRepo
}

type IUserService interface {
	CreateUser(req *dto.CreateUserRequest) (*model.User, error)
	UpdateUser(req *dto.UpdateUserRequest) error
	GetUserByID(id uint) (*dto.UserDetailResponse, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserList(req *dto.UserPageRequest) (*dto.UserListResponse, error)
	DeleteUser(id uint) error
	ChangePassword(req *dto.ChangePasswordRequest) error
	ResetPassword(req *dto.ResetPasswordRequest) error
	AssignPermissions(req *dto.AssignPermissionRequest) error
	GetUserPermissions(userID uint) ([]string, error)
	HasPermission(userID uint, permission string) (bool, error)
	CheckUserRole(userID uint, allowedRoles ...string) (bool, error)
	CreateDefaultAdmin(username, password string) error
	RecordLoginHistory(userID uint, req *dto.UserLoginResponse, ip, agent, address string, status string) error
	GetLoginHistory(userID uint) ([]model.UserLoginHistory, error)
}

func NewIUserService() IUserService {
	return &UserService{
		userRepo: repo.NewIUserRepo(),
	}
}

func (u *UserService) CreateUser(req *dto.CreateUserRequest) (*model.User, error) {
	// Check if user already exists
	existingUser, _ := u.userRepo.GetUserByUsername(req.Username)
	if existingUser != nil {
		return nil, buserr.New("ErrUserAlreadyExists")
	}

	// Encrypt password
	encryptedPassword, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: encryptedPassword,
		Role:     req.Role,
		Status:   constant.UserStatusActive,
		RealName: req.RealName,
		Phone:    req.Phone,
		Remark:   req.Remark,
	}

	if err := u.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	// Assign default permissions for the role
	if defaultPermissions, ok := constant.RolePermissions[req.Role]; ok {
		for _, perm := range defaultPermissions {
			u.userRepo.AddPermission(&model.UserPermission{
				UserID:     user.ID,
				Permission: perm,
			})
		}
	}

	return user, nil
}

func (u *UserService) UpdateUser(req *dto.UpdateUserRequest) error {
	user, err := u.userRepo.GetUserByID(req.ID)
	if err != nil {
		return buserr.New("ErrRecordNotFound")
	}

	user.Email = req.Email
	user.Role = req.Role
	user.Status = req.Status
	user.RealName = req.RealName
	user.Phone = req.Phone
	user.Remark = req.Remark

	return u.userRepo.UpdateUser(user)
}

func (u *UserService) GetUserByID(id uint) (*dto.UserDetailResponse, error) {
	user, err := u.userRepo.GetUserByID(id)
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}

	permissions, _ := u.userRepo.GetUserPermissions(id)

	return &dto.UserDetailResponse{
		User: userToDTO(user),
		Permissions: permissions,
	}, nil
}

func (u *UserService) GetUserByUsername(username string) (*model.User, error) {
	return u.userRepo.GetUserByUsername(username)
}

func (u *UserService) GetUserList(req *dto.UserPageRequest) (*dto.UserListResponse, error) {
	total, users, err := u.userRepo.GetUserList(req)
	if err != nil {
		return nil, err
	}

	var items []dto.UserResponse
	for _, user := range users {
		items = append(items, userToDTO(&user))
	}

	return &dto.UserListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (u *UserService) DeleteUser(id uint) error {
	user, err := u.userRepo.GetUserByID(id)
	if err != nil {
		return buserr.New("ErrRecordNotFound")
	}

	// Prevent deletion of the main admin
	if user.Role == constant.RoleAdminMain && user.ParentID == 0 {
		return buserr.New("ErrCannotDeleteMainAdmin")
	}

	return u.userRepo.DeleteUser(id)
}

func (u *UserService) ChangePassword(req *dto.ChangePasswordRequest) error {
	user, err := u.userRepo.GetUserByID(req.UserID)
	if err != nil {
		return buserr.New("ErrRecordNotFound")
	}

	// Verify old password
	if err := encrypt.StringDecrypt(req.OldPassword, user.Password); err != nil {
		return buserr.New("ErrOldPasswordIncorrect")
	}

	// Encrypt new password
	encryptedPassword, err := encrypt.StringEncrypt(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = encryptedPassword
	return u.userRepo.UpdateUser(user)
}

func (u *UserService) ResetPassword(req *dto.ResetPasswordRequest) error {
	user, err := u.userRepo.GetUserByID(req.UserID)
	if err != nil {
		return buserr.New("ErrRecordNotFound")
	}

	encryptedPassword, err := encrypt.StringEncrypt(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = encryptedPassword
	return u.userRepo.UpdateUser(user)
}

func (u *UserService) AssignPermissions(req *dto.AssignPermissionRequest) error {
	user, err := u.userRepo.GetUserByID(req.UserID)
	if err != nil {
		return buserr.New("ErrRecordNotFound")
	}

	// Only super admins can assign permissions
	if user.Role != constant.RoleAdminMain {
		return buserr.New("ErrUnauthorized")
	}

	return u.userRepo.UpdateUserPermissions(req.UserID, req.Permissions)
}

func (u *UserService) GetUserPermissions(userID uint) ([]string, error) {
	user, err := u.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}

	// Super admin has all permissions
	if user.Role == constant.RoleAdminMain {
		var allPerms []string
		for _, perms := range constant.RolePermissions {
			allPerms = append(allPerms, perms...)
		}
		return allPerms, nil
	}

	// Get custom permissions or use role default
	permissions, err := u.userRepo.GetUserPermissions(userID)
	if err != nil || len(permissions) == 0 {
		// Fall back to role-based permissions
		if defaultPerms, ok := constant.RolePermissions[user.Role]; ok {
			return defaultPerms, nil
		}
	}

	return permissions, nil
}

func (u *UserService) HasPermission(userID uint, permission string) (bool, error) {
	user, err := u.userRepo.GetUserByID(userID)
	if err != nil {
		return false, buserr.New("ErrRecordNotFound")
	}

	// Super admin has all permissions
	if user.Role == constant.RoleAdminMain {
		return true, nil
	}

	permissions, err := u.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permission || perm == constant.PermissionAdminAll {
			return true, nil
		}
	}

	return false, nil
}

func (u *UserService) CheckUserRole(userID uint, allowedRoles ...string) (bool, error) {
	user, err := u.userRepo.GetUserByID(userID)
	if err != nil {
		return false, buserr.New("ErrRecordNotFound")
	}

	for _, role := range allowedRoles {
		if user.Role == role {
			return true, nil
		}
	}

	return false, nil
}

func (u *UserService) CreateDefaultAdmin(username, password string) error {
	encryptedPassword, err := encrypt.StringEncrypt(password)
	if err != nil {
		return err
	}

	adminUser := &model.User{
		Username: username,
		Password: encryptedPassword,
		Email:    "",
		Role:     constant.RoleAdminMain,
		Status:   constant.UserStatusActive,
		RealName: "Admin",
	}

	if err := u.userRepo.CreateUser(adminUser); err != nil {
		return err
	}

	// Assign all admin permissions
	for _, perm := range constant.RolePermissions[constant.RoleAdminMain] {
		u.userRepo.AddPermission(&model.UserPermission{
			UserID:     adminUser.ID,
			Permission: perm,
		})
	}

	return nil
}

func (u *UserService) RecordLoginHistory(userID uint, loginInfo *dto.UserLoginResponse, ip, agent, address string, status string) error {
	history := &model.UserLoginHistory{
		UserID:  userID,
		IP:      ip,
		Address: address,
		Agent:   agent,
		Status:  status,
		LoginAt: time.Now(),
	}

	if err := u.userRepo.AddLoginHistory(history); err != nil {
		return err
	}

	// Update user's last login time
	user, err := u.userRepo.GetUserByID(userID)
	if err == nil {
		now := time.Now()
		user.LastLogin = &now
		u.userRepo.UpdateUser(user)
	}

	return nil
}

func (u *UserService) GetLoginHistory(userID uint) ([]model.UserLoginHistory, error) {
	return u.userRepo.GetLoginHistory(userID, 100)
}

func userToDTO(user *model.User) dto.UserResponse {
	lastLogin := int64(0)
	if user.LastLogin != nil {
		lastLogin = user.LastLogin.Unix()
	}

	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		RealName:  user.RealName,
		Phone:     user.Phone,
		LastLogin: lastLogin,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}
