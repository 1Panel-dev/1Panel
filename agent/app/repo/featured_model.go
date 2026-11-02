package repo

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type FeaturedModelRepo struct{}

type IFeaturedModelRepo interface {
	Page(page, size int, family, info string) (int64, []model.FeaturedModel, error)
	Families() ([]string, error)
	SyncAll(models []model.FeaturedModel, version string) error
}

func NewIFeaturedModelRepo() IFeaturedModelRepo {
	return &FeaturedModelRepo{}
}

func (r *FeaturedModelRepo) Page(page, size int, family, info string) (int64, []model.FeaturedModel, error) {
	var total int64
	var items []model.FeaturedModel
	db := global.DB.Model(&model.FeaturedModel{})
	if family != "" {
		db = db.Where("family = ?", family)
	}
	if info != "" {
		db = db.Where("LOWER(repo_id) LIKE ?", "%"+strings.ToLower(info)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if err := db.Order("sort asc").Order("id asc").Limit(size).Offset(size * (page - 1)).Find(&items).Error; err != nil {
		return 0, nil, err
	}
	return total, items, nil
}

func (r *FeaturedModelRepo) Families() ([]string, error) {
	families := make([]string, 0)
	err := global.DB.Model(&model.FeaturedModel{}).Distinct("family").Order("family asc").Pluck("family", &families).Error
	return families, err
}

func (r *FeaturedModelRepo) SyncAll(models []model.FeaturedModel, version string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.FeaturedModel{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models).Error; err != nil {
			return err
		}
		return tx.Where("key = ?", "FeaturedModelsVersion").
			Assign(model.Setting{Value: version}).
			FirstOrCreate(&model.Setting{Key: "FeaturedModelsVersion"}).Error
	})
}
