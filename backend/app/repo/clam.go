package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type ClamRepo struct{}

type IClamRepo interface {
	Page(limit, offset int, opts ...DBOption) (int64, []model.Clam, error)
	Create(clam *model.Clam) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	Get(opts ...DBOption) (model.Clam, error)
	List(opts ...DBOption) ([]model.Clam, error)
}

func NewIClamRepo() IClamRepo {
	return &ClamRepo{}
}

func (u *ClamRepo) Get(opts ...DBOption) (model.Clam, error) {
	var clam model.Clam
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&clam).Error
	return clam, err
}

func (u *ClamRepo) List(opts ...DBOption) ([]model.Clam, error) {
	var clam []model.Clam
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&clam).Error
	return clam, err
}

func (u *ClamRepo) Page(page, size int, opts ...DBOption) (int64, []model.Clam, error) {
	var users []model.Clam
	db := global.DB.Model(&model.Clam{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *ClamRepo) Create(clam *model.Clam) error {
	return global.DB.Create(clam).Error
}

func (u *ClamRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Clam{}).Where("id = ?", id).Updates(vars).Error
}

func (u *ClamRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Clam{}).Error
}
