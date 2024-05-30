package repo

import (
	"context"
	"encoding/json"

	"github.com/1Panel-dev/1Panel/backend/constant"

	"gorm.io/gorm/clause"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type AppInstallRepo struct{}

type IAppInstallRepo interface {
	WithDetailIdsIn(detailIds []uint) DBOption
	WithDetailIdNotIn(detailIds []uint) DBOption
	WithAppId(appId uint) DBOption
	WithAppIdsIn(appIds []uint) DBOption
	WithStatus(status string) DBOption
	WithServiceName(serviceName string) DBOption
	WithContainerName(containerName string) DBOption
	WithPort(port int) DBOption
	WithIdNotInWebsite() DBOption
	WithIDNotIs(id uint) DBOption
	ListBy(opts ...DBOption) ([]model.AppInstall, error)
	GetFirst(opts ...DBOption) (model.AppInstall, error)
	Create(ctx context.Context, install *model.AppInstall) error
	Save(ctx context.Context, install *model.AppInstall) error
	DeleteBy(opts ...DBOption) error
	Delete(ctx context.Context, install model.AppInstall) error
	Page(page, size int, opts ...DBOption) (int64, []model.AppInstall, error)
	BatchUpdateBy(maps map[string]interface{}, opts ...DBOption) error
	LoadBaseInfo(key string, name string) (*RootInfo, error)
	GetFirstByCtx(ctx context.Context, opts ...DBOption) (model.AppInstall, error)
}

func NewIAppInstallRepo() IAppInstallRepo {
	return &AppInstallRepo{}
}

func (a *AppInstallRepo) WithDetailIdsIn(detailIds []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id in (?)", detailIds)
	}
}

func (a *AppInstallRepo) WithDetailIdNotIn(detailIds []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id not in (?)", detailIds)
	}
}

func (a *AppInstallRepo) WithAppId(appId uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_id = ?", appId)
	}
}

func (a *AppInstallRepo) WithIDNotIs(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id != ?", id)
	}
}

func (a *AppInstallRepo) WithAppIdsIn(appIds []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_id in (?)", appIds)
	}
}

func (a *AppInstallRepo) WithStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ?", status)
	}
}

func (a *AppInstallRepo) WithServiceName(serviceName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("service_name = ?", serviceName)
	}
}

func (a *AppInstallRepo) WithContainerName(containerName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("container_name = ?", containerName)
	}
}

func (a *AppInstallRepo) WithPort(port int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("https_port = ? or  http_port = ?", port, port)
	}
}

func (a *AppInstallRepo) WithIdNotInWebsite() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id not in (select app_install_id from websites)")
	}
}

func (a *AppInstallRepo) ListBy(opts ...DBOption) ([]model.AppInstall, error) {
	var install []model.AppInstall
	db := getDb(opts...).Model(&model.AppInstall{})
	err := db.Preload("App").Find(&install).Error
	return install, err
}

func (a *AppInstallRepo) GetFirst(opts ...DBOption) (model.AppInstall, error) {
	var install model.AppInstall
	db := getDb(opts...).Model(&model.AppInstall{})
	err := db.Preload("App").First(&install).Error
	return install, err
}

func (a *AppInstallRepo) GetFirstByCtx(ctx context.Context, opts ...DBOption) (model.AppInstall, error) {
	var install model.AppInstall
	db := getTx(ctx, opts...).Model(&model.AppInstall{})
	err := db.Preload("App").First(&install).Error
	return install, err
}

func (a *AppInstallRepo) Create(ctx context.Context, install *model.AppInstall) error {
	db := getTx(ctx).Model(&model.AppInstall{})
	return db.Omit(clause.Associations).Create(&install).Error
}

func (a *AppInstallRepo) Save(ctx context.Context, install *model.AppInstall) error {
	return getTx(ctx).Save(&install).Error
}

func (a *AppInstallRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.AppInstall{}).Error
}

func (a *AppInstallRepo) Delete(ctx context.Context, install model.AppInstall) error {
	db := getTx(ctx).Model(&model.AppInstall{})
	return db.Delete(&install).Error
}

func (a *AppInstallRepo) Page(page, size int, opts ...DBOption) (int64, []model.AppInstall, error) {
	var apps []model.AppInstall
	db := getDb(opts...).Model(&model.AppInstall{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Preload("App").Find(&apps).Error
	return count, apps, err
}

func (a *AppInstallRepo) BatchUpdateBy(maps map[string]interface{}, opts ...DBOption) error {
	db := getDb(opts...).Model(&model.AppInstall{})
	if len(opts) == 0 {
		db = db.Where("1=1")
	}
	return db.Updates(&maps).Error
}

type RootInfo struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Port          int64  `json:"port"`
	HttpsPort     int64  `json:"httpsPort"`
	UserName      string `json:"userName"`
	Password      string `json:"password"`
	UserPassword  string `json:"userPassword"`
	ContainerName string `json:"containerName"`
	ServiceName   string `json:"serviceName"`
	Param         string `json:"param"`
	Env           string `json:"env"`
	Key           string `json:"key"`
	Version       string `json:"version"`
	AppPath       string `json:"app_path"`
}

func (a *AppInstallRepo) LoadBaseInfo(key string, name string) (*RootInfo, error) {
	var (
		app        model.App
		appInstall model.AppInstall
		info       RootInfo
	)
	if err := global.DB.Where("key = ?", key).First(&app).Error; err != nil {
		return nil, err
	}
	if len(name) == 0 {
		if err := global.DB.Where("app_id = ?", app.ID).First(&appInstall).Error; err != nil {
			return nil, err
		}
	} else {
		if err := global.DB.Where("app_id = ? AND name = ?", app.ID, name).First(&appInstall).Error; err != nil {
			return nil, err
		}
	}
	envMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(appInstall.Env), &envMap); err != nil {
		return nil, err
	}
	switch app.Key {
	case "mysql", "mariadb":
		password, ok := envMap["PANEL_DB_ROOT_PASSWORD"].(string)
		if ok {
			info.Password = password
		}
	case "redis":
		password, ok := envMap["PANEL_REDIS_ROOT_PASSWORD"].(string)
		if ok {
			info.Password = password
		}
	case "mongodb", constant.AppPostgresql:
		user, ok := envMap["PANEL_DB_ROOT_USER"].(string)
		if ok {
			info.UserName = user
		}
		password, ok := envMap["PANEL_DB_ROOT_PASSWORD"].(string)
		if ok {
			info.Password = password
		}
	}

	userPassword, ok := envMap["PANEL_DB_USER_PASSWORD"].(string)
	if ok {
		info.UserPassword = userPassword
	}
	info.Port = int64(appInstall.HttpPort)
	info.HttpsPort = int64(appInstall.HttpsPort)
	info.ID = appInstall.ID
	info.ServiceName = appInstall.ServiceName
	info.ContainerName = appInstall.ContainerName
	info.Name = appInstall.Name
	info.Env = appInstall.Env
	info.Param = appInstall.Param
	info.Version = appInstall.Version
	info.Key = app.Key
	appInstall.App = app
	info.AppPath = appInstall.GetAppPath()
	info.Status = appInstall.Status
	return &info, nil
}
