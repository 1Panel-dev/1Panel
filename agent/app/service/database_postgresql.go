package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/postgresql"
	"github.com/1Panel-dev/1Panel/agent/utils/postgresql/client"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type PostgresqlService struct{}

type IPostgresqlService interface {
	ListUsers(req dto.PostgresqlUserSearch) ([]dto.PostgresqlUser, error)
	CreateUser(req dto.PostgresqlUserCreate) error
	UpdateUser(req dto.PostgresqlUserUpdate) error
	ChangeUserPassword(req dto.PostgresqlUserPassword) error
	SaveUserPassword(req dto.PostgresqlUserPassword) error
	DeleteUser(req dto.PostgresqlUserDelete) error
	GrantUser(req dto.PostgresqlUserGrant) error
	RevokeUser(req dto.PostgresqlUserGrant) error
	SearchWithPage(search dto.PostgresqlDBSearch) (int64, interface{}, error)
	ListDBOption() ([]dto.PostgresqlOption, error)
	BindUser(req dto.PostgresqlBindUser) error
	Create(ctx context.Context, req dto.PostgresqlDBCreate) (*model.DatabasePostgresql, error)
	LoadFromRemote(database string) error
	UpdateDescription(req dto.UpdateDescription) error
	DeleteCheck(req dto.PostgresqlDBDeleteCheck) ([]dto.DBResource, error)
	Delete(ctx context.Context, req dto.PostgresqlDBDelete) error
	ChangePrivileges(req dto.PostgresqlPrivileges) error
	ChangePassword(info dto.ChangeDBInfo) error
}

func NewIPostgresqlService() IPostgresqlService {
	return &PostgresqlService{}
}

func (u *PostgresqlService) SearchWithPage(search dto.PostgresqlDBSearch) (int64, interface{}, error) {
	total, postgresqls, err := postgresqlRepo.Page(search.Page, search.PageSize,
		postgresqlRepo.WithByPostgresqlName(search.Database),
		repo.WithByLikeName(search.Info),
		repo.WithOrderRuleBy(search.OrderBy, search.Order),
	)
	var dtoPostgresqls []dto.PostgresqlDBInfo
	for _, pg := range postgresqls {
		var item dto.PostgresqlDBInfo
		if err := copier.Copy(&item, &pg); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
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
			return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
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
	dbItem, err := postgresqlRepo.Get(postgresqlRepo.WithByPostgresqlName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return err
	}
	if err := u.CreateUser(dto.PostgresqlUserCreate{Database: req.Database, Username: req.Username, Password: req.Password, SuperUser: req.SuperUser}); err != nil {
		return err
	}
	if !req.SuperUser {
		if err := u.GrantUser(dto.PostgresqlUserGrant{Database: req.Database, Name: req.Name, Username: req.Username}); err != nil {
			return err
		}
	}
	if dbItem.Username != "" {
		return nil
	}
	password, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return err
	}
	return postgresqlRepo.Update(dbItem.ID, map[string]interface{}{"username": req.Username, "password": password, "super_user": req.SuperUser})
}

func (u *PostgresqlService) Create(ctx context.Context, req dto.PostgresqlDBCreate) (*model.DatabasePostgresql, error) {
	if cmd.CheckIllegal(req.Name, req.Format) || (!req.ExistingUser && (cmd.CheckIllegal(req.Username, req.Password) || strings.Contains(req.Password, " "))) {
		return nil, buserr.New("ErrCmdIllegal")
	}

	pgsql, _ := postgresqlRepo.Get(repo.WithByName(req.Name), postgresqlRepo.WithByPostgresqlName(req.Database), repo.WithByFrom(req.From))
	if pgsql.ID != 0 {
		return nil, buserr.New("ErrRecordExist")
	}

	if req.ExistingUser {
		users, err := u.ListUsers(dto.PostgresqlUserSearch{Database: req.Database})
		if err != nil {
			return nil, err
		}
		found := false
		for _, user := range users {
			if user.Username == req.Username {
				req.Password, req.SuperUser = user.Password, user.SuperUser
				found = true
				break
			}
		}
		if !found {
			return nil, buserr.New("ErrRecordNotFound")
		}
	} else if req.Password == "" {
		return nil, buserr.New("ErrInvalidParams")
	}

	var createItem model.DatabasePostgresql
	if err := copier.Copy(&createItem, &req); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}

	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	createItem.PostgresqlName = req.Database
	if err := cli.Create(client.CreateInfo{
		ExistingUser: req.ExistingUser,
		Name:         req.Name,
		Username:     req.Username,
		Password:     req.Password,
		SuperUser:    req.SuperUser,
		Timeout:      300,
	}); err != nil {
		return nil, err
	}

	global.LOG.Infof("create database %s successful!", req.Name)
	if err := postgresqlRepo.Create(ctx, &createItem); err != nil {
		return nil, err
	}
	if !req.ExistingUser {
		user, err := databaseUserRepo.GetEncryptedWithContext(ctx, repo.WithByType("postgresql"), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, ""))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return &createItem, err
		}
		user.Type, user.Database, user.Username, user.Host = "postgresql", req.Database, req.Username, ""
		user.Password, user.Description, user.IsDelete = createItem.Password, "", false
		if err := databaseUserRepo.SaveEncryptedWithContext(ctx, &user); err != nil {
			return &createItem, err
		}
	}
	return &createItem, nil
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
				return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
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

func (u *PostgresqlService) DeleteCheck(req dto.PostgresqlDBDeleteCheck) ([]dto.DBResource, error) {
	var res []dto.DBResource
	db, err := postgresqlRepo.Get(repo.WithByID(req.ID))
	if err != nil {
		return res, err
	}

	website, _ := websiteRepo.GetFirst(websiteRepo.WithDBType(constant.AppPostgresql), websiteRepo.WithDBID(req.ID))
	if website.ID != 0 {
		res = append(res, dto.DBResource{
			Type: constant.TypeWebsite,
			Name: website.PrimaryDomain,
		})
	}
	website, _ = websiteRepo.GetFirst(websiteRepo.WithDBType(constant.AppPostgresqlCluster), websiteRepo.WithDBID(req.ID))
	if website.ID != 0 {
		res = append(res, dto.DBResource{
			Type: constant.TypeWebsite,
			Name: website.PrimaryDomain,
		})
	}

	if db.From == "local" {
		app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
		if err != nil {
			return res, err
		}
		apps, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(app.ID), appInstallResourceRepo.WithResourceId(db.ID))
		for _, app := range apps {
			appInstall, _ := appInstallRepo.GetFirst(repo.WithByID(app.AppInstallId))
			if appInstall.ID != 0 {
				res = append(res, dto.DBResource{
					Type: constant.TypeApp,
					Name: appInstall.Name,
				})
			}
		}
	} else {
		apps, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithResourceId(db.ID), appRepo.WithKey(req.Type))
		for _, app := range apps {
			appInstall, _ := appInstallRepo.GetFirst(repo.WithByID(app.AppInstallId))
			if appInstall.ID != 0 {
				res = append(res, dto.DBResource{
					Type: constant.TypeApp,
					Name: appInstall.Name,
				})
			}
		}
	}

	return res, nil
}

func (u *PostgresqlService) Delete(ctx context.Context, req dto.PostgresqlDBDelete) error {
	db, err := postgresqlRepo.Get(repo.WithByID(req.ID), postgresqlRepo.WithByPostgresqlName(req.Database))
	if err != nil {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if db.Username != "" {
		user, err := databaseUserRepo.GetEncryptedWithContext(ctx, repo.WithByType("postgresql"), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(db.Username, ""))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			password, encryptErr := encrypt.StringEncrypt(db.Password)
			if encryptErr != nil {
				return encryptErr
			}
			user = model.DatabaseUser{Type: "postgresql", Database: req.Database, Username: db.Username, Password: password}
			if err := databaseUserRepo.SaveEncryptedWithContext(ctx, &user); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	if err := cli.Delete(client.DeleteInfo{
		Name:        db.Name,
		ForceDelete: req.ForceDelete,
		Timeout:     300,
	}); err != nil && !req.ForceDelete {
		return err
	}

	if req.DeleteBackup {
		uploadDir := path.Join(global.Dir.DataDir, fmt.Sprintf("uploads/database/%s/%s/%s", req.Type, req.Database, db.Name))
		if _, err := os.Stat(uploadDir); err == nil {
			_ = os.RemoveAll(uploadDir)
		}
		backupDir := path.Join(global.Dir.LocalBackupDir, fmt.Sprintf("database/%s/%s/%s", req.Type, db.PostgresqlName, db.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		_ = backupRepo.DeleteRecord(ctx, repo.WithByType(req.Type), repo.WithByName(req.Database), repo.WithByDetailName(db.Name))
		global.LOG.Infof("delete database %s-%s backups successful", req.Database, db.Name)
	}

	return postgresqlRepo.Delete(ctx, repo.WithByID(db.ID))
}

func (u *PostgresqlService) ChangePrivileges(req dto.PostgresqlPrivileges) error {
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	return u.UpdateUser(dto.PostgresqlUserUpdate{Database: req.Database, Username: req.Username, SuperUser: req.SuperUser, Description: user.Description})
}

func (u *PostgresqlService) ChangePassword(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) || strings.Contains(req.Value, " ") {
		return buserr.New("ErrCmdIllegal")
	}
	if req.ID != 0 {
		db, err := postgresqlRepo.Get(repo.WithByID(req.ID), postgresqlRepo.WithByPostgresqlName(req.Database))
		if err != nil {
			return err
		}
		return u.ChangeUserPassword(dto.PostgresqlUserPassword{Database: req.Database, Username: db.Username, Password: req.Value})
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	dbItem, err := databaseRepo.Get(repo.WithByName(req.Database))
	if err != nil {
		return err
	}
	passwordInfo := client.PasswordChangeInfo{Username: dbItem.Username, Password: req.Value, Timeout: 300}
	if err := cli.ChangePassword(passwordInfo); err != nil {
		return err
	}

	if err := updateInstallInfoInDB(req.Type, req.Database, "password", req.Value); err != nil {
		return err
	}
	if req.From == "local" {
		remote, err := databaseRepo.Get(repo.WithByName(req.Database))
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

func LoadPostgresqlClientByFrom(database string) (postgresql.PostgresqlClient, error) {
	var (
		dbInfo client.DBInfo
		err    error
	)

	dbInfo.Timeout = 300
	databaseItem, err := databaseRepo.Get(repo.WithByName(database))
	if err != nil {
		return nil, err
	}
	if databaseItem.Type != constant.AppPostgresql && databaseItem.Type != constant.AppPostgresqlCluster && databaseItem.Type != constant.AppPostgres {
		return nil, buserr.New("ErrInvalidParams")
	}
	dbInfo.From = databaseItem.From
	dbInfo.Database = database
	if dbInfo.From != "local" {
		dbInfo.Address = databaseItem.Address
		dbInfo.Port = databaseItem.Port
		dbInfo.Username = databaseItem.Username
		dbInfo.Password = databaseItem.Password
		dbInfo.InitialDB = databaseItem.InitialDB
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
