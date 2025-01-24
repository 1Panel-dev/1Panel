package model

import (
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type App struct {
	BaseModel
	Name                 string  `json:"name" gorm:"not null"`
	Key                  string  `json:"key" gorm:"not null;"`
	ShortDescZh          string  `json:"shortDescZh" yaml:"shortDescZh"`
	ShortDescEn          string  `json:"shortDescEn" yaml:"shortDescEn"`
	Description          string  `json:"description"`
	Icon                 string  `json:"icon"`
	Type                 string  `json:"type" gorm:"not null"`
	Status               string  `json:"status" gorm:"not null"`
	Required             string  `json:"required"`
	CrossVersionUpdate   bool    `json:"crossVersionUpdate" yaml:"crossVersionUpdate"`
	Limit                int     `json:"limit" gorm:"not null"`
	Website              string  `json:"website" gorm:"not null"`
	Github               string  `json:"github" gorm:"not null"`
	Document             string  `json:"document" gorm:"not null"`
	Recommend            int     `json:"recommend" gorm:"not null"`
	Resource             string  `json:"resource" gorm:"not null;default:remote"`
	ReadMe               string  `json:"readMe"`
	LastModified         int     `json:"lastModified"`
	Architectures        string  `json:"architectures"`
	MemoryRequired       int     `json:"memoryRequired"`
	GpuSupport           bool    `json:"gpuSupport"`
	RequiredPanelVersion float64 `json:"requiredPanelVersion"`

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
		return filepath.Join(global.Dir.LocalAppResourceDir, strings.TrimPrefix(i.Key, "local"))
	}
	return filepath.Join(global.Dir.RemoteAppResourceDir, i.Key)
}
