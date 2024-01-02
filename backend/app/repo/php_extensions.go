package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type PHPExtensionsRepo struct {
}

type IPHPExtensionsRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.PHPExtensions, error)
	Save(extension *model.PHPExtensions) error
	Create(extension *model.PHPExtensions) error
	GetFirst(opts ...DBOption) (model.PHPExtensions, error)
	DeleteBy(opts ...DBOption) error
	List() ([]model.PHPExtensions, error)
}

func NewIPHPExtensionsRepo() IPHPExtensionsRepo {
	return &PHPExtensionsRepo{}
}

func (p *PHPExtensionsRepo) Page(page, size int, opts ...DBOption) (int64, []model.PHPExtensions, error) {
	var (
		phpExtensions []model.PHPExtensions
	)
	db := getDb(opts...).Model(&model.PHPExtensions{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&phpExtensions).Error
	return count, phpExtensions, err
}

func (p *PHPExtensionsRepo) List() ([]model.PHPExtensions, error) {
	var (
		phpExtensions []model.PHPExtensions
	)
	err := getDb().Model(&model.PHPExtensions{}).Find(&phpExtensions).Error
	return phpExtensions, err
}

func (p *PHPExtensionsRepo) Save(extension *model.PHPExtensions) error {
	return getDb().Save(&extension).Error
}

func (p *PHPExtensionsRepo) Create(extension *model.PHPExtensions) error {
	return getDb().Create(&extension).Error
}

func (p *PHPExtensionsRepo) GetFirst(opts ...DBOption) (model.PHPExtensions, error) {
	var extension model.PHPExtensions
	db := getDb(opts...).Model(&model.PHPExtensions{})
	err := db.First(&extension).Error
	return extension, err
}

func (p *PHPExtensionsRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.PHPExtensions{}).Error
}
