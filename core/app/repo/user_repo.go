package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"gorm.io/gorm"
)

type IUserRepo interface {
	CreateUser(user *model.User) error
	UpdateUser(user *model.User) error
	GetUserByID(id uint) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserList(req *dto.UserPageRequest) (int64, []model.User, error)
	DeleteUser(id uint) error
	GetUsersByParentID(parentID uint) ([]model.User, error)

	AddPermission(permission *model.UserPermission) error
	DeletePermission(userID uint, permission string) error
	GetUserPermissions(userID uint) ([]string, error)
	UpdateUserPermissions(userID uint, permissions []string) error
	GetPermissionsByRole(role string) ([]string, error)

	AddLoginHistory(history *model.UserLoginHistory) error
	GetLoginHistory(userID uint, limit int) ([]model.UserLoginHistory, error)
}

type UserRepo struct{}

func NewIUserRepo() IUserRepo {
	return &UserRepo{}
}

func (u *UserRepo) CreateUser(user *model.User) error {
	return baseDb.Create(user).Error
}

func (u *UserRepo) UpdateUser(user *model.User) error {
	return baseDb.Save(user).Error
}

func (u *UserRepo) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := baseDb.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := baseDb.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) GetUserList(req *dto.UserPageRequest) (int64, []model.User, error) {
	var users []model.User
	var total int64

	query := baseDb
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Model(&model.User{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	offset := (req.PageNum - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		return 0, nil, err
	}

	return total, users, nil
}

func (u *UserRepo) DeleteUser(id uint) error {
	// Delete user permissions first
	if err := baseDb.Where("user_id = ?", id).Delete(&model.UserPermission{}).Error; err != nil {
		return err
	}
	// Delete user login history
	if err := baseDb.Where("user_id = ?", id).Delete(&model.UserLoginHistory{}).Error; err != nil {
		return err
	}
	// Delete the user
	return baseDb.Delete(&model.User{}, id).Error
}

func (u *UserRepo) GetUsersByParentID(parentID uint) ([]model.User, error) {
	var users []model.User
	if err := baseDb.Where("parent_id = ?", parentID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserRepo) AddPermission(permission *model.UserPermission) error {
	return baseDb.Create(permission).Error
}

func (u *UserRepo) DeletePermission(userID uint, permission string) error {
	return baseDb.Where("user_id = ? AND permission = ?", userID, permission).Delete(&model.UserPermission{}).Error
}

func (u *UserRepo) GetUserPermissions(userID uint) ([]string, error) {
	var permissions []string
	if err := baseDb.Model(&model.UserPermission{}).
		Where("user_id = ?", userID).
		Pluck("permission", &permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (u *UserRepo) UpdateUserPermissions(userID uint, permissions []string) error {
	return baseDb.Transaction(func(tx *gorm.DB) error {
		// Delete all existing permissions for this user
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserPermission{}).Error; err != nil {
			return err
		}

		// Add new permissions
		for _, permission := range permissions {
			if err := tx.Create(&model.UserPermission{
				UserID:     userID,
				Permission: permission,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *UserRepo) GetPermissionsByRole(role string) ([]string, error) {
	// This will be retrievedfrom constants in the service layer
	// but we keep this interface for future extensibility
	return nil, nil
}

func (u *UserRepo) AddLoginHistory(history *model.UserLoginHistory) error {
	return baseDb.Create(history).Error
}

func (u *UserRepo) GetLoginHistory(userID uint, limit int) ([]model.UserLoginHistory, error) {
	var histories []model.UserLoginHistory
	if err := baseDb.
		Where("user_id = ?", userID).
		Order("login_at DESC").
		Limit(limit).
		Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}
