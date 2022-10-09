package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
)

type AppContainerRepo struct {
}

func (a AppContainerRepo) WithAppId(appId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_id = ?", appId)
	}
}

func (a AppContainerRepo) WithServiceName(serviceName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("service_name = ?", serviceName)
	}
}

func (a AppContainerRepo) GetBy(opts ...DBOption) ([]model.AppContainer, error) {
	db := global.DB.Model(&model.AppContainer{})
	var appContainers []model.AppContainer
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&appContainers).Error
	return appContainers, err
}

func (a AppContainerRepo) GetFirst(opts ...DBOption) (model.AppContainer, error) {
	db := global.DB.Model(&model.AppContainer{})
	var appContainer model.AppContainer
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&appContainer).Error
	return appContainer, err
}

func (a AppContainerRepo) Create(container *model.AppContainer) error {
	db := global.DB.Model(&model.AppContainer{})
	return db.Create(&container).Error
}

func (a AppContainerRepo) BatchCreate(ctx context.Context, containers []*model.AppContainer) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Model(&model.AppContainer{}).Create(&containers).Error
}
