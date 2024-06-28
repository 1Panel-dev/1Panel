package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/postgresql"
	"github.com/1Panel-dev/1Panel/backend/utils/postgresql/client"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type PostgresqlService struct{}

type IPostgresqlService interface {
	SearchWithPage(search dto.PostgresqlDBSearch) (int64, interface{}, error)
	ListDBOption() ([]dto.PostgresqlOption, error)
	BindUser(req dto.PostgresqlBindUser) error
	Create(ctx context.Context, req dto.PostgresqlDBCreate) (*model.DatabasePostgresql, error)
	LoadFromRemote(database string) error
	ChangePrivileges(req dto.PostgresqlPrivileges) error
	ChangePassword(info dto.ChangeDBInfo) error
	UpdateDescription(req dto.UpdateDescription) error
	DeleteCheck(req dto.PostgresqlDBDeleteCheck) ([]string, error)
	Delete(ctx context.Context, req dto.PostgresqlDBDelete) error
}

func NewIPostgresqlService() IPostgresqlService {
	return &PostgresqlService{}
}

func (u *PostgresqlService) SearchWithPage(search dto.PostgresqlDBSearch) (int64, interface{}, error) {
	total, postgresqls, err := postgresqlRepo.Page(search.Page, search.PageSize,
		postgresqlRepo.WithByPostgresqlName(search.Database),
		commonRepo.WithLikeName(search.Info),
		commonRepo.WithOrderRuleBy(search.OrderBy, search.Order),
	)
	var dtoPostgresqls []dto.PostgresqlDBInfo
	for _, pg := range postgresqls {
		var item dto.PostgresqlDBInfo
		if err := copier.Copy(&item, &pg); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoPostgresqls = append(dtoPostgresqls, item)
	}
	return total, dtoPostgresqls, err
}

func (u *PostgresqlService) ListDBOption() ([]dto.PostgresqlOption, error) {
	postgresqls, err := postgresqlRepo.List()
	if err != nil {
		return nil, err
	}

	databases, err := databaseRepo.GetList(databaseRepo.WithTypeList("postgresql,mariadb"))
	if err != nil {
		return nil, err
	}
	var dbs []dto.PostgresqlOption
	for _, pg := range postgresqls {
		var item dto.PostgresqlOption
		if err := copier.Copy(&item, &pg); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		item.Database = pg.PostgresqlName
		for _, database := range databases {
			if database.Name == item.Database {
				item.Type = database.Type
			}
		}
		dbs = append(dbs, item)
	}
	return dbs, err
}

func (u *PostgresqlService) BindUser(req dto.PostgresqlBindUser) error {
	if cmd.CheckIllegal(req.Name, req.Username, req.Password) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	dbItem, err := postgresqlRepo.Get(postgresqlRepo.WithByPostgresqlName(req.Database), commonRepo.WithByName(req.Name))
	if err != nil {
		return err
	}

	pgClient, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	if err := pgClient.CreateUser(client.CreateInfo{
		Name:      req.Name,
		Username:  req.Username,
		Password:  req.Password,
		SuperUser: req.SuperUser,
		Timeout:   300,
	}, false); err != nil {
		return err
	}
	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return fmt.Errorf("decrypt database db password failed, err: %v", err)
	}
	if err := postgresqlRepo.Update(dbItem.ID, map[string]interface{}{
		"username":   req.Username,
		"password":   pass,
		"super_user": req.SuperUser,
	}); err != nil {
		return err
	}
	return nil
}

func (u *PostgresqlService) Create(ctx context.Context, req dto.PostgresqlDBCreate) (*model.DatabasePostgresql, error) {
	if cmd.CheckIllegal(req.Name, req.Username, req.Password, req.Format) {
		return nil, buserr.New(constant.ErrCmdIllegal)
	}

	pgsql, _ := postgresqlRepo.Get(commonRepo.WithByName(req.Name), postgresqlRepo.WithByPostgresqlName(req.Database), databaseRepo.WithByFrom(req.From))
	if pgsql.ID != 0 {
		return nil, constant.ErrRecordExist
	}

	var createItem model.DatabasePostgresql
	if err := copier.Copy(&createItem, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	if req.From == "local" && req.Username == "root" {
		return nil, errors.New("Cannot set root as user name")
	}

	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	createItem.PostgresqlName = req.Database
	defer cli.Close()
	if err := cli.Create(client.CreateInfo{
		Name:      req.Name,
		Username:  req.Username,
		Password:  req.Password,
		SuperUser: req.SuperUser,
		Timeout:   300,
	}); err != nil {
		return nil, err
	}

	global.LOG.Infof("create database %s successful!", req.Name)
	if err := postgresqlRepo.Create(ctx, &createItem); err != nil {
		return nil, err
	}
	return &createItem, nil
}

func LoadPostgresqlClientByFrom(database string) (postgresql.PostgresqlClient, error) {
	var (
		dbInfo client.DBInfo
		err    error
	)

	dbInfo.Timeout = 300
	databaseItem, err := databaseRepo.Get(commonRepo.WithByName(database))
	if err != nil {
		return nil, err
	}
	dbInfo.From = databaseItem.From
	dbInfo.Database = database
	if dbInfo.From != "local" {
		dbInfo.Address = databaseItem.Address
		dbInfo.Port = databaseItem.Port
		dbInfo.Username = databaseItem.Username
		dbInfo.Password = databaseItem.Password
	} else {
		app, err := appInstallRepo.LoadBaseInfo(databaseItem.Type, database)
		if err != nil {
			return nil, err
		}
		dbInfo.From = "local"
		dbInfo.Address = app.ContainerName
		dbInfo.Username = app.UserName
		dbInfo.Password = app.Password
		dbInfo.Port = uint(app.Port)
	}

	cli, err := postgresql.NewPostgresqlClient(dbInfo)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (u *PostgresqlService) LoadFromRemote(database string) error {
	client, err := LoadPostgresqlClientByFrom(database)
	if err != nil {
		return err
	}
	defer client.Close()

	databases, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(database))
	if err != nil {
		return err
	}
	datas, err := client.SyncDB()
	if err != nil {
		return err
	}
	deleteList := databases
	for _, data := range datas {
		hasOld := false
		for i := 0; i < len(databases); i++ {
			if strings.EqualFold(databases[i].Name, data.Name) && strings.EqualFold(databases[i].PostgresqlName, data.PostgresqlName) {
				hasOld = true
				if databases[i].IsDelete {
					_ = postgresqlRepo.Update(databases[i].ID, map[string]interface{}{"is_delete": false})
				}
				deleteList = append(deleteList[:i], deleteList[i+1:]...)
				break
			}
		}
		if !hasOld {
			var createItem model.DatabasePostgresql
			if err := copier.Copy(&createItem, &data); err != nil {
				return errors.WithMessage(constant.ErrStructTransform, err.Error())
			}
			if err := postgresqlRepo.Create(context.Background(), &createItem); err != nil {
				return err
			}
		}
	}
	for _, delItem := range deleteList {
		_ = postgresqlRepo.Update(delItem.ID, map[string]interface{}{"is_delete": true})
	}
	return nil
}

func (u *PostgresqlService) UpdateDescription(req dto.UpdateDescription) error {
	return postgresqlRepo.Update(req.ID, map[string]interface{}{"description": req.Description})
}

func (u *PostgresqlService) DeleteCheck(req dto.PostgresqlDBDeleteCheck) ([]string, error) {
	var appInUsed []string
	db, err := postgresqlRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return appInUsed, err
	}

	if db.From == "local" {
		app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
		if err != nil {
			return appInUsed, err
		}
		apps, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(app.ID), appInstallResourceRepo.WithResourceId(db.ID))
		for _, app := range apps {
			appInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(app.AppInstallId))
			if appInstall.ID != 0 {
				appInUsed = append(appInUsed, appInstall.Name)
			}
		}
	} else {
		apps, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithResourceId(db.ID), appRepo.WithKey(req.Type))
		for _, app := range apps {
			appInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(app.AppInstallId))
			if appInstall.ID != 0 {
				appInUsed = append(appInUsed, appInstall.Name)
			}
		}
	}

	return appInUsed, nil
}

func (u *PostgresqlService) Delete(ctx context.Context, req dto.PostgresqlDBDelete) error {
	db, err := postgresqlRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil && !req.ForceDelete {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Delete(client.DeleteInfo{
		Name:        db.Name,
		Username:    db.Username,
		ForceDelete: req.ForceDelete,
		Timeout:     300,
	}); err != nil && !req.ForceDelete {
		return err
	}

	if req.DeleteBackup {
		uploadDir := path.Join(global.CONF.System.BaseDir, fmt.Sprintf("1panel/uploads/database/%s/%s/%s", req.Type, req.Database, db.Name))
		if _, err := os.Stat(uploadDir); err == nil {
			_ = os.RemoveAll(uploadDir)
		}
		localDir, err := loadLocalDir()
		if err != nil && !req.ForceDelete {
			return err
		}
		backupDir := path.Join(localDir, fmt.Sprintf("database/%s/%s/%s", req.Type, db.PostgresqlName, db.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType(req.Type), commonRepo.WithByName(req.Database), backupRepo.WithByDetailName(db.Name))
		global.LOG.Infof("delete database %s-%s backups successful", req.Database, db.Name)
	}

	_ = postgresqlRepo.Delete(ctx, commonRepo.WithByID(db.ID))
	return nil
}

func (u *PostgresqlService) ChangePrivileges(req dto.PostgresqlPrivileges) error {
	if cmd.CheckIllegal(req.Database, req.Username) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	dbItem, err := postgresqlRepo.Get(postgresqlRepo.WithByPostgresqlName(req.Database), commonRepo.WithByName(req.Name))
	if err != nil {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ChangePrivileges(client.Privileges{Username: req.Username, SuperUser: req.SuperUser, Timeout: 300}); err != nil {
		return err
	}
	if err := postgresqlRepo.Update(dbItem.ID, map[string]interface{}{
		"super_user": req.SuperUser,
	}); err != nil {
		return err
	}
	return nil
}

func (u *PostgresqlService) ChangePassword(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	var (
		postgresqlData model.DatabasePostgresql
		passwordInfo   client.PasswordChangeInfo
	)
	passwordInfo.Password = req.Value
	passwordInfo.Timeout = 300

	if req.ID != 0 {
		postgresqlData, err = postgresqlRepo.Get(commonRepo.WithByID(req.ID))
		if err != nil {
			return err
		}
		passwordInfo.Username = postgresqlData.Username
	} else {
		dbItem, err := databaseRepo.Get(commonRepo.WithByType(req.Type), commonRepo.WithByFrom(req.From))
		if err != nil {
			return err
		}
		passwordInfo.Username = dbItem.Username
	}
	if err := cli.ChangePassword(passwordInfo); err != nil {
		return err
	}

	if req.ID != 0 {
		var appRess []model.AppInstallResource
		if req.From == "local" {
			app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
			if err != nil {
				return err
			}
			appRess, _ = appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(app.ID), appInstallResourceRepo.WithResourceId(postgresqlData.ID))
		} else {
			appRess, _ = appInstallResourceRepo.GetBy(appInstallResourceRepo.WithResourceId(postgresqlData.ID))
		}
		for _, appRes := range appRess {
			appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(appRes.AppInstallId))
			if err != nil {
				return err
			}
			appModel, err := appRepo.GetFirst(commonRepo.WithByID(appInstall.AppId))
			if err != nil {
				return err
			}

			global.LOG.Infof("start to update postgresql password used by app %s-%s", appModel.Key, appInstall.Name)
			if err := updateInstallInfoInDB(appModel.Key, appInstall.Name, "user-password", req.Value); err != nil {
				return err
			}
		}
		global.LOG.Info("execute password change sql successful")
		pass, err := encrypt.StringEncrypt(req.Value)
		if err != nil {
			return fmt.Errorf("decrypt database db password failed, err: %v", err)
		}
		_ = postgresqlRepo.Update(postgresqlData.ID, map[string]interface{}{"password": pass})
		return nil
	}

	if err := updateInstallInfoInDB(req.Type, req.Database, "password", req.Value); err != nil {
		return err
	}
	if req.From == "local" {
		remote, err := databaseRepo.Get(commonRepo.WithByName(req.Database))
		if err != nil {
			return err
		}
		pass, err := encrypt.StringEncrypt(req.Value)
		if err != nil {
			return fmt.Errorf("decrypt database password failed, err: %v", err)
		}
		_ = databaseRepo.Update(remote.ID, map[string]interface{}{"password": pass})
	}
	return nil
}
