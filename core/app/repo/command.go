package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

type CommandRepo struct{}

type ICommandRepo interface {
	List(opts ...DBOption) ([]model.Command, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Command, error)
	Create(command *model.Command) error
	Update(id uint, vars map[string]interface{}) error
	UpdateGroup(group, newGroup uint) error
	Delete(opts ...DBOption) error
	Get(opts ...DBOption) (model.Command, error)
}

func NewICommandRepo() ICommandRepo {
	return &CommandRepo{}
}

func (u *CommandRepo) Get(opts ...DBOption) (model.Command, error) {
	var command model.Command
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&command).Error
	return command, err
}

func (u *CommandRepo) Page(page, size int, opts ...DBOption) (int64, []model.Command, error) {
	var users []model.Command
	db := global.DB.Model(&model.Command{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *CommandRepo) List(opts ...DBOption) ([]model.Command, error) {
	var commands []model.Command
	db := global.DB.Model(&model.Command{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&commands).Error
	return commands, err
}

func (u *CommandRepo) Create(command *model.Command) error {
	return global.DB.Create(command).Error
}

func (u *CommandRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Command{}).Where("id = ?", id).Updates(vars).Error
}
func (h *CommandRepo) UpdateGroup(group, newGroup uint) error {
	return global.DB.Model(&model.Command{}).Where("group_id = ?", group).Updates(map[string]interface{}{"group_id": newGroup}).Error
}

func (u *CommandRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Command{}).Error
}
