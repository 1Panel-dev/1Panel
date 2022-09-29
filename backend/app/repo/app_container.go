package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
)

type AppContainerRepo struct {
}

func (a AppContainerRepo) Create(container *model.AppContainer) error {
	db := global.DB.Model(&model.AppContainer{})
	return db.Create(&container).Error
}

func (a AppContainerRepo) BatchCreate(ctx context.Context, containers []*model.AppContainer) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Model(&model.AppContainer{}).Create(&containers).Error
}
