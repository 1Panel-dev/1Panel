package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type CommandRepo struct{}

type ICommandRepo interface {
	GetList(opts ...DBOption) ([]model.Command, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Command, error)
	WithByInfo(info string) DBOption
	Create(command *model.Command) error
	Update(id uint, vars map[string]interface{}) error
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

func (u *CommandRepo) GetList(opts ...DBOption) ([]model.Command, error) {
	var commands []model.Command
	db := global.DB.Model(&model.Command{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&commands).Error
	return commands, err
}

func (c *CommandRepo) WithByInfo(info string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		infoStr := "%" + info + "%"
		return g.Where("name LIKE ? OR addr LIKE ?", infoStr, infoStr)
	}
}

func (u *CommandRepo) Create(command *model.Command) error {
	return global.DB.Create(command).Error
}

func (u *CommandRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Command{}).Where("id = ?", id).Updates(vars).Error
}

func (u *CommandRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Command{}).Error
}
