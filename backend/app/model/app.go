package model

import (
	"github.com/1Panel-dev/1Panel/backend/constant"
	"path/filepath"
	"strings"
)

type App struct {
	BaseModel
	Name               string `json:"name" gorm:"type:varchar(64);not null"`
	Key                string `json:"key" gorm:"type:varchar(64);not null;"`
	ShortDescZh        string `json:"shortDescZh" yaml:"shortDescZh" gorm:"type:longtext;"`
	ShortDescEn        string `json:"shortDescEn" yaml:"shortDescEn" gorm:"type:longtext;"`
	Icon               string `json:"icon" gorm:"type:longtext;"`
	Type               string `json:"type" gorm:"type:varchar(64);not null"`
	Status             string `json:"status" gorm:"type:varchar(64);not null"`
	Required           string `json:"required" gorm:"type:varchar(64);"`
	CrossVersionUpdate bool   `json:"crossVersionUpdate"`
	Limit              int    `json:"limit" gorm:"type:Integer;not null"`
	Website            string `json:"website" gorm:"type:varchar(64);not null"`
	Github             string `json:"github" gorm:"type:varchar(64);not null"`
	Document           string `json:"document" gorm:"type:varchar(64);not null"`
	Recommend          int    `json:"recommend" gorm:"type:Integer;not null"`
	Resource           string `json:"resource" gorm:"type:varchar;not null;default:remote"`
	ReadMe             string `json:"readMe" gorm:"type:varchar;"`
	LastModified       int    `json:"lastModified" gorm:"type:Integer;"`

	Details []AppDetail `json:"-" gorm:"-:migration"`
	TagsKey []string    `json:"tags" yaml:"tags" gorm:"-"`
	AppTags []AppTag    `json:"-" gorm:"-:migration"`
}

func (i *App) IsLocalApp() bool {
	return i.Resource == constant.ResourceLocal
}
func (i *App) GetAppResourcePath() string {
	if i.IsLocalApp() {
		//这里要去掉本地应用的local前缀
		return filepath.Join(constant.LocalAppResourceDir, strings.TrimPrefix(i.Key, "local"))
	}
	return filepath.Join(constant.RemoteAppResourceDir, i.Key)
}
