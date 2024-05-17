package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type FtpRepo struct{}

type IFtpRepo interface {
	Get(opts ...DBOption) (model.Ftp, error)
	GetList(opts ...DBOption) ([]model.Ftp, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Ftp, error)
	Create(ftp *model.Ftp) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	WithByUser(user string) DBOption
}

func NewIFtpRepo() IFtpRepo {
	return &FtpRepo{}
}

func (u *FtpRepo) Get(opts ...DBOption) (model.Ftp, error) {
	var ftp model.Ftp
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&ftp).Error
	return ftp, err
}

func (h *FtpRepo) WithByUser(user string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(user) == 0 {
			return g
		}
		return g.Where("user like ?", "%"+user+"%")
	}
}

func (u *FtpRepo) GetList(opts ...DBOption) ([]model.Ftp, error) {
	var ftps []model.Ftp
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&ftps).Error
	return ftps, err
}

func (h *FtpRepo) Page(page, size int, opts ...DBOption) (int64, []model.Ftp, error) {
	var users []model.Ftp
	db := global.DB.Model(&model.Ftp{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (h *FtpRepo) Create(ftp *model.Ftp) error {
	return global.DB.Create(ftp).Error
}

func (h *FtpRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Ftp{}).Where("id = ?", id).Updates(vars).Error
}

func (h *FtpRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Ftp{}).Error
}
