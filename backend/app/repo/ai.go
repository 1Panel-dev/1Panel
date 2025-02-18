package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type AiRepo struct{}

type IAiRepo interface {
	Get(opts ...DBOption) (model.OllamaModel, error)
	List(opts ...DBOption) ([]model.OllamaModel, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.OllamaModel, error)
	Create(cronjob *model.OllamaModel) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewIAiRepo() IAiRepo {
	return &AiRepo{}
}

func (u *AiRepo) Get(opts ...DBOption) (model.OllamaModel, error) {
	var item model.OllamaModel
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&item).Error
	return item, err
}

func (u *AiRepo) List(opts ...DBOption) ([]model.OllamaModel, error) {
	var list []model.OllamaModel
	db := global.DB.Model(&model.OllamaModel{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&list).Error
	return list, err
}

func (u *AiRepo) Page(page, size int, opts ...DBOption) (int64, []model.OllamaModel, error) {
	var list []model.OllamaModel
	db := global.DB.Model(&model.OllamaModel{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&list).Error
	return count, list, err
}

func (u *AiRepo) Create(item *model.OllamaModel) error {
	return global.DB.Create(item).Error
}

func (u *AiRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.OllamaModel{}).Where("id = ?", id).Updates(vars).Error
}

func (u *AiRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.OllamaModel{}).Error
}
