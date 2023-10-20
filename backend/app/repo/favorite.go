package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type FavoriteRepo struct{}

type IFavoriteRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.Favorite, error)
	Create(group *model.Favorite) error
	Delete(opts ...DBOption) error
	GetFirst(opts ...DBOption) (model.Favorite, error)
	All() ([]model.Favorite, error)
	WithByPath(path string) DBOption
}

func NewIFavoriteRepo() IFavoriteRepo {
	return &FavoriteRepo{}
}

func (f *FavoriteRepo) WithByPath(path string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path = ?", path)
	}
}

func (f *FavoriteRepo) Page(page, size int, opts ...DBOption) (int64, []model.Favorite, error) {
	var (
		favorites []model.Favorite
		count     int64
	)
	count = int64(0)
	db := getDb(opts...).Model(&model.Favorite{})
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&favorites).Error
	return count, favorites, err
}

func (f *FavoriteRepo) Create(favorite *model.Favorite) error {
	return global.DB.Create(favorite).Error
}

func (f *FavoriteRepo) GetFirst(opts ...DBOption) (model.Favorite, error) {
	var favorite model.Favorite
	db := getDb(opts...).Model(&model.Favorite{})
	if err := db.First(&favorite).Error; err != nil {
		return favorite, err
	}
	return favorite, nil
}

func (f *FavoriteRepo) Delete(opts ...DBOption) error {
	db := getDb(opts...).Model(&model.Favorite{})
	return db.Delete(&model.Favorite{}).Error
}

func (f *FavoriteRepo) All() ([]model.Favorite, error) {
	var favorites []model.Favorite
	if err := getDb().Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}
