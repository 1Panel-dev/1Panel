package service

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/backend/utils/postgresql"
	pgclient "github.com/1Panel-dev/1Panel/backend/utils/postgresql/client"
	redisclient "github.com/1Panel-dev/1Panel/backend/utils/redis"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql/client"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type DatabaseService struct{}

type IDatabaseService interface {
	Get(name string) (dto.DatabaseInfo, error)
	SearchWithPage(search dto.DatabaseSearch) (int64, interface{}, error)
	CheckDatabase(req dto.DatabaseCreate) bool
	Create(req dto.DatabaseCreate) error
	Update(req dto.DatabaseUpdate) error
	DeleteCheck(id uint) ([]string, error)
	Delete(req dto.DatabaseDelete) error
	List(dbType string) ([]dto.DatabaseOption, error)
	LoadItems(dbType string) ([]dto.DatabaseItem, error)
}

func NewIDatabaseService() IDatabaseService {
	return &DatabaseService{}
}

func (u *DatabaseService) SearchWithPage(search dto.DatabaseSearch) (int64, interface{}, error) {
	total, dbs, err := databaseRepo.Page(search.Page, search.PageSize,
		databaseRepo.WithTypeList(search.Type),
		commonRepo.WithLikeName(search.Info),
		commonRepo.WithOrderRuleBy(search.OrderBy, search.Order),
		databaseRepo.WithoutByFrom("local"),
	)
	var datas []dto.DatabaseInfo
	for _, db := range dbs {
		var item dto.DatabaseInfo
		if err := copier.Copy(&item, &db); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		datas = append(datas, item)
	}
	return total, datas, err
}

func (u *DatabaseService) Get(name string) (dto.DatabaseInfo, error) {
	var data dto.DatabaseInfo
	remote, err := databaseRepo.Get(commonRepo.WithByName(name))
	if err != nil {
		return data, err
	}
	if err := copier.Copy(&data, &remote); err != nil {
		return data, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return data, nil
}

func (u *DatabaseService) List(dbType string) ([]dto.DatabaseOption, error) {
	dbs, err := databaseRepo.GetList(databaseRepo.WithTypeList(dbType))
	if err != nil {
		return nil, err
	}
	var datas []dto.DatabaseOption
	for _, db := range dbs {
		var item dto.DatabaseOption
		if err := copier.Copy(&item, &db); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		item.Database = db.Name
		datas = append(datas, item)
	}
	return datas, err
}

func (u *DatabaseService) LoadItems(dbType string) ([]dto.DatabaseItem, error) {
	dbs, err := databaseRepo.GetList(databaseRepo.WithTypeList(dbType))
	var datas []dto.DatabaseItem
	for _, db := range dbs {
		if dbType == "postgresql" {
			items, _ := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(db.Name))
			for _, item := range items {
				var dItem dto.DatabaseItem
				if err := copier.Copy(&dItem, &item); err != nil {
					continue
				}
				dItem.Database = db.Name
				datas = append(datas, dItem)
			}
		} else {
			items, _ := mysqlRepo.List(mysqlRepo.WithByMysqlName(db.Name))
			for _, item := range items {
				var dItem dto.DatabaseItem
				if err := copier.Copy(&dItem, &item); err != nil {
					continue
				}
				dItem.Database = db.Name
				datas = append(datas, dItem)
			}
		}
	}
	return datas, err
}

func (u *DatabaseService) CheckDatabase(req dto.DatabaseCreate) bool {
	switch req.Type {
	case constant.AppPostgresql:
		_, err := postgresql.NewPostgresqlClient(pgclient.DBInfo{
			From:     "remote",
			Address:  req.Address,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,
			Timeout:  6,
		})
		return err == nil
	case constant.AppRedis:
		_, err := redisclient.NewRedisClient(redisclient.DBInfo{
			Address:  req.Address,
			Port:     req.Port,
			Password: req.Password,
		})
		return err == nil
	case "mysql", "mariadb":
		_, err := mysql.NewMysqlClient(client.DBInfo{
			From:     "remote",
			Address:  req.Address,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,

			SSL:        req.SSL,
			RootCert:   req.RootCert,
			ClientKey:  req.ClientKey,
			ClientCert: req.ClientCert,
			SkipVerify: req.SkipVerify,
			Timeout:    6,
		})
		return err == nil
	}

	return false
}

func (u *DatabaseService) Create(req dto.DatabaseCreate) error {
	db, _ := databaseRepo.Get(commonRepo.WithByName(req.Name))
	if db.ID != 0 {
		if db.From == "local" {
			return buserr.New(constant.ErrLocalExist)
		}
		return constant.ErrRecordExist
	}
	switch req.Type {
	case constant.AppPostgresql:
		if _, err := postgresql.NewPostgresqlClient(pgclient.DBInfo{
			From:     "remote",
			Address:  req.Address,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,
			Timeout:  6,
		}); err != nil {
			return err
		}
	case constant.AppRedis:
		if _, err := redisclient.NewRedisClient(redisclient.DBInfo{
			Address:  req.Address,
			Port:     req.Port,
			Password: req.Password,
		}); err != nil {
			return err
		}
	case "mysql", "mariadb":
		if _, err := mysql.NewMysqlClient(client.DBInfo{
			From:     "remote",
			Address:  req.Address,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,

			SSL:        req.SSL,
			RootCert:   req.RootCert,
			ClientKey:  req.ClientKey,
			ClientCert: req.ClientCert,
			SkipVerify: req.SkipVerify,
			Timeout:    6,
		}); err != nil {
			return err
		}
	default:
		return errors.New("database type not supported")
	}

	if err := copier.Copy(&db, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := databaseRepo.Create(context.Background(), &db); err != nil {
		return err
	}
	return nil
}

func (u *DatabaseService) DeleteCheck(id uint) ([]string, error) {
	var appInUsed []string
	apps, _ := appInstallResourceRepo.GetBy(databaseRepo.WithByFrom("remote"), appInstallResourceRepo.WithLinkId(id))
	for _, app := range apps {
		appInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(app.AppInstallId))
		if appInstall.ID != 0 {
			appInUsed = append(appInUsed, appInstall.Name)
		}
	}

	return appInUsed, nil
}

func (u *DatabaseService) Delete(req dto.DatabaseDelete) error {
	db, _ := databaseRepo.Get(commonRepo.WithByID(req.ID))
	if db.ID == 0 {
		return constant.ErrRecordNotFound
	}

	if req.DeleteBackup {
		uploadDir := path.Join(global.CONF.System.BaseDir, fmt.Sprintf("1panel/uploads/database/%s/%s", db.Type, db.Name))
		if _, err := os.Stat(uploadDir); err == nil {
			_ = os.RemoveAll(uploadDir)
		}
		localDir, err := loadLocalDir()
		if err != nil && !req.ForceDelete {
			return err
		}
		backupDir := path.Join(localDir, fmt.Sprintf("database/%s/%s", db.Type, db.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		_ = backupRepo.DeleteRecord(context.Background(), commonRepo.WithByType(db.Type), commonRepo.WithByName(db.Name))
		global.LOG.Infof("delete database %s-%s backups successful", db.Type, db.Name)
	}

	if err := databaseRepo.Delete(context.Background(), commonRepo.WithByID(req.ID)); err != nil && !req.ForceDelete {
		return err
	}
	if db.From != "local" {
		if db.Type == "mysql" || db.Type == "mariadb" {
			if err := mysqlRepo.Delete(context.Background(), mysqlRepo.WithByMysqlName(db.Name)); err != nil && !req.ForceDelete {
				return err
			}
		} else {
			if err := postgresqlRepo.Delete(context.Background(), postgresqlRepo.WithByPostgresqlName(db.Name)); err != nil && !req.ForceDelete {
				return err
			}
		}
	}
	return nil
}

func (u *DatabaseService) Update(req dto.DatabaseUpdate) error {
	switch req.Type {
	case constant.AppPostgresql:
		if _, err := postgresql.NewPostgresqlClient(pgclient.DBInfo{
			From:     "remote",
			Address:  req.Address,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,
			Timeout:  300,
		}); err != nil {
			return err
		}
	case constant.AppRedis:
		if _, err := redisclient.NewRedisClient(redisclient.DBInfo{
			Address:  req.Address,
			Port:     req.Port,
			Password: req.Password,
		}); err != nil {
			return err
		}
	case "mysql", "mariadb":
		if _, err := mysql.NewMysqlClient(client.DBInfo{
			From:     "remote",
			Address:  req.Address,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,

			SSL:        req.SSL,
			RootCert:   req.RootCert,
			ClientKey:  req.ClientKey,
			ClientCert: req.ClientCert,
			SkipVerify: req.SkipVerify,
			Timeout:    300,
		}); err != nil {
			return err
		}
	default:
		return errors.New("database type not supported")
	}

	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return fmt.Errorf("decrypt database password failed, err: %v", err)
	}

	upMap := make(map[string]interface{})
	upMap["type"] = req.Type
	upMap["version"] = req.Version
	upMap["address"] = req.Address
	upMap["port"] = req.Port
	upMap["username"] = req.Username
	upMap["password"] = pass
	upMap["description"] = req.Description
	upMap["ssl"] = req.SSL
	upMap["client_key"] = req.ClientKey
	upMap["client_cert"] = req.ClientCert
	upMap["root_cert"] = req.RootCert
	upMap["skip_verify"] = req.SkipVerify
	return databaseRepo.Update(req.ID, upMap)
}
