package model

import (
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
	"path"
)

type AppInstall struct {
	BaseModel
	Name        string         `json:"name" gorm:"type:varchar(64);not null"`
	Version     string         `json:"version" gorm:"type:varchar(256);not null"`
	AppId       uint           `json:"appId" gorm:"type:integer;not null"`
	AppDetailId uint           `json:"appDetailId" gorm:"type:integer;not null"`
	Params      string         `json:"params"  gorm:"type:longtext;"`
	Status      string         `json:"status" gorm:"type:varchar(256);not null"`
	Description string         `json:"description" gorm:"type:varchar(256);"`
	Message     string         `json:"message"  gorm:"type:longtext;"`
	CanUpdate   bool           `json:"canUpdate"`
	App         App            `json:"-"`
	Containers  []AppContainer `json:"containers"`
}

func (i *AppInstall) GetPath() string {
	return path.Join(global.CONF.System.AppDir, i.App.Key, i.Name)
}

func (i *AppInstall) GetComposePath() string {
	return path.Join(global.CONF.System.AppDir, i.App.Key, i.Name, "docker-compose.yml")
}

func (i *AppInstall) BeforeDelete(tx *gorm.DB) (err error) {

	if err = tx.Model(AppContainer{}).Debug().Where("app_install_id = ?", i.ID).Delete(AppContainer{}).Error; err != nil {
		return err
	}

	return
}
