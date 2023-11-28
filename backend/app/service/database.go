package service

import (
	"context"
	"fmt"
	"os"
	"path"

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
}

func NewIDatabaseService() IDatabaseService {
	return &DatabaseService{}
}

func (u *DatabaseService) SearchWithPage(search dto.DatabaseSearch) (int64, interface{}, error) {
	total, dbs, err := databaseRepo.Page(search.Page, search.PageSize,
		databaseRepo.WithTypeList(search.Type),
		commonRepo.WithLikeName(search.Info),
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

func (u *DatabaseService) CheckDatabase(req dto.DatabaseCreate) bool {
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
		return false
	}
	return true
}

func (u *DatabaseService) Create(req dto.DatabaseCreate) error {
	db, _ := databaseRepo.Get(commonRepo.WithByName(req.Name))
	if db.ID != 0 {
		if db.From == "local" {
			return buserr.New(constant.ErrLocalExist)
		}
		return constant.ErrRecordExist
	}
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
		if err := mysqlRepo.Delete(context.Background(), mysqlRepo.WithByMysqlName(db.Name)); err != nil && !req.ForceDelete {
			return err
		}
	}
	return nil
}

func (u *DatabaseService) Update(req dto.DatabaseUpdate) error {
	if _, err := mysql.NewMysqlClient(client.DBInfo{
		From:     "remote",
		Address:  req.Address,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,

		SSL:        req.SSL,
		ClientKey:  req.ClientKey,
		ClientCert: req.ClientCert,
		RootCert:   req.RootCert,
		SkipVerify: req.SkipVerify,

		Timeout: 300,
	}); err != nil {
		return err
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
	upMap["ssl"] = req.SSL
	upMap["client_key"] = req.ClientKey
	upMap["client_cert"] = req.ClientCert
	upMap["root_cert"] = req.RootCert
	upMap["skip_verify"] = req.SkipVerify
	return databaseRepo.Update(req.ID, upMap)
}
