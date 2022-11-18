package repo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type MysqlRepo struct{}

type IMysqlRepo interface {
	Get(opts ...DBOption) (model.DatabaseMysql, error)
	WithByMysqlName(mysqlName string) DBOption
	List(opts ...DBOption) ([]model.DatabaseMysql, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.DatabaseMysql, error)
	Create(ctx context.Context, mysql *model.DatabaseMysql) error
	Delete(ctx context.Context, opts ...DBOption) error
	Update(id uint, vars map[string]interface{}) error
	LoadRunningVersion(keys []string) ([]string, error)
	LoadBaseInfoByName(name string) (*RootInfo, error)
	LoadRedisBaseInfo() (*RootInfo, error)
	UpdateDatabaseInfo(id uint, vars map[string]interface{}) error
}

func NewIMysqlRepo() IMysqlRepo {
	return &MysqlRepo{}
}

func (u *MysqlRepo) Get(opts ...DBOption) (model.DatabaseMysql, error) {
	var mysql model.DatabaseMysql
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&mysql).Error
	return mysql, err
}

func (u *MysqlRepo) List(opts ...DBOption) ([]model.DatabaseMysql, error) {
	var users []model.DatabaseMysql
	db := global.DB.Model(&model.DatabaseMysql{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&users).Error
	return users, err
}

func (u *MysqlRepo) Page(page, size int, opts ...DBOption) (int64, []model.DatabaseMysql, error) {
	var users []model.DatabaseMysql
	db := global.DB.Model(&model.DatabaseMysql{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *MysqlRepo) LoadRunningVersion(keys []string) ([]string, error) {
	var (
		apps       []model.App
		appInstall []model.AppInstall
		results    []string
	)
	if err := global.DB.Where("name in (?)", keys).Find(&apps).Error; err != nil {
		return nil, err
	}
	for _, app := range apps {
		if err := global.DB.Where("app_id = ?", app.ID).Find(&appInstall).Error; err != nil {
			return nil, err
		}
		for _, item := range appInstall {
			results = append(results, item.Name)
		}
	}
	return results, nil
}

type RootInfo struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Port          int64  `json:"port"`
	Password      string `json:"password"`
	ContainerName string `json:"containerName"`
	Param         string `json:"param"`
	Env           string `json:"env"`
	Key           string `json:"key"`
}

func (u *MysqlRepo) LoadBaseInfoByName(name string) (*RootInfo, error) {
	var (
		app        model.App
		appInstall model.AppInstall
		info       RootInfo
	)
	if err := global.DB.Where("name = ?", name).First(&appInstall).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("id = ?", appInstall.AppId).First(&app).Error; err != nil {
		return nil, err
	}
	envMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(appInstall.Env), &envMap); err != nil {
		return nil, err
	}
	password, ok := envMap["PANEL_DB_ROOT_PASSWORD"].(string)
	if ok {
		info.Password = password
	} else {
		return nil, errors.New("error password in db")
	}
	port, ok := envMap["PANEL_APP_PORT_HTTP"].(float64)
	if ok {
		info.Port = int64(port)
	} else {
		return nil, errors.New("error port in db")
	}
	info.ID = appInstall.ID
	info.Key = app.Key
	info.ContainerName = appInstall.ContainerName
	info.Name = appInstall.Name
	info.Env = appInstall.Env
	info.Param = appInstall.Param
	return &info, nil
}

func (u *MysqlRepo) LoadRedisBaseInfo() (*RootInfo, error) {
	var (
		app        model.App
		appInstall model.AppInstall
		info       RootInfo
	)
	if err := global.DB.Where("key = ?", "redis").First(&app).Error; err != nil {
		return nil, err
	}
	if err := global.DB.Where("app_id = ?", app.ID).First(&appInstall).Error; err != nil {
		return nil, err
	}
	envMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(appInstall.Env), &envMap); err != nil {
		return nil, err
	}
	password, ok := envMap["PANEL_DB_ROOT_PASSWORD"].(string)
	if ok {
		info.Password = password
	} else {
		return nil, errors.New("error password in db")
	}
	port, ok := envMap["PANEL_APP_PORT_HTTP"].(float64)
	if ok {
		info.Port = int64(port)
	} else {
		return nil, errors.New("error port in db")
	}
	info.ID = appInstall.ID
	info.ContainerName = appInstall.ContainerName
	info.Name = appInstall.Name
	info.Env = appInstall.Env
	info.Param = appInstall.Param
	return &info, nil
}

func (u *MysqlRepo) Create(ctx context.Context, mysql *model.DatabaseMysql) error {
	return getTx(ctx).Create(mysql).Error
}

func (u *MysqlRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DatabaseMysql{}).Error
}

func (u *MysqlRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DatabaseMysql{}).Where("id = ?", id).Updates(vars).Error
}

func (u *MysqlRepo) UpdateDatabaseInfo(id uint, vars map[string]interface{}) error {
	if err := global.DB.Model(&model.AppInstall{}).Where("id = ?", id).Updates(vars).Error; err != nil {
		return err
	}
	return nil
}

func (u *MysqlRepo) WithByMysqlName(mysqlName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("mysql_name = ?", mysqlName)
	}
}
