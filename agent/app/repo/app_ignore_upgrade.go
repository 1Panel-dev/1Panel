package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type AppIgnoreUpgradeRepo struct {
}

type IAppIgnoreUpgradeRepo interface {
	WithScope(scope string) DBOption
	List(opts ...DBOption) ([]model.AppIgnoreUpgrade, error)
	Create(appIgnoreUpgrade *model.AppIgnoreUpgrade) error
	Delete(opts ...DBOption) error
}

func NewIAppIgnoreUpgradeRepo() IAppIgnoreUpgradeRepo {
	return &AppIgnoreUpgradeRepo{}
}

func (a AppIgnoreUpgradeRepo) WithScope(scope string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("scope = ?", scope)
	}
}

func (a AppIgnoreUpgradeRepo) List(opts ...DBOption) ([]model.AppIgnoreUpgrade, error) {
	var appIgnoreUpgradeList []model.AppIgnoreUpgrade
	err := getDb(opts...).Find(&appIgnoreUpgradeList).Error
	return appIgnoreUpgradeList, err
}

func (a AppIgnoreUpgradeRepo) Create(appIgnoreUpgrade *model.AppIgnoreUpgrade) error {
	return global.DB.Create(appIgnoreUpgrade).Error
}

func (a AppIgnoreUpgradeRepo) Delete(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.AppIgnoreUpgrade{}).Error
}
