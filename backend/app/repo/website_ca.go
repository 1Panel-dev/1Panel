package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteCARepo struct {
}

func NewIWebsiteCARepo() IWebsiteCARepo {
	return &WebsiteCARepo{}
}

type IWebsiteCARepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteCA, error)
	GetFirst(opts ...DBOption) (model.WebsiteCA, error)
	List(opts ...DBOption) ([]model.WebsiteCA, error)
	Create(ctx context.Context, ca *model.WebsiteCA) error
	DeleteBy(opts ...DBOption) error
}

func (w WebsiteCARepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteCA, error) {
	var caList []model.WebsiteCA
	db := getDb(opts...).Model(&model.WebsiteCA{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&caList).Error
	return count, caList, err
}

func (w WebsiteCARepo) GetFirst(opts ...DBOption) (model.WebsiteCA, error) {
	var ca model.WebsiteCA
	db := getDb(opts...).Model(&model.WebsiteCA{})
	if err := db.First(&ca).Error; err != nil {
		return ca, err
	}
	return ca, nil
}

func (w WebsiteCARepo) List(opts ...DBOption) ([]model.WebsiteCA, error) {
	var caList []model.WebsiteCA
	db := getDb(opts...).Model(&model.WebsiteCA{})
	err := db.Find(&caList).Error
	return caList, err
}

func (w WebsiteCARepo) Create(ctx context.Context, ca *model.WebsiteCA) error {
	return getTx(ctx).Create(ca).Error
}

func (w WebsiteCARepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebsiteCA{}).Error
}
