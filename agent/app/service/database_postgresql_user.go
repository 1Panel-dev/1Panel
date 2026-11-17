package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/postgresql/client"
	"gorm.io/gorm"
)

func (u *PostgresqlService) ListUsers(req dto.PostgresqlUserSearch) ([]dto.PostgresqlUser, error) {
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	users, err := client.ListUsers(cli)
	if err != nil {
		return nil, err
	}
	metadata, err := databaseUserRepo.List(repo.WithByType("postgresql"), databaseUserRepo.WithByDatabase(req.Database))
	if err != nil {
		return nil, err
	}
	legacy, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(req.Database))
	if err != nil {
		return nil, err
	}
	result := make([]dto.PostgresqlUser, 0, len(users))
	for _, user := range users {
		item := dto.PostgresqlUser{Username: user.Username, SuperUser: user.SuperUser, Protected: user.Protected, Databases: user.Databases}
		passwords := make(map[string]struct{})
		for _, db := range legacy {
			if db.Username == user.Username && db.Password != "" {
				passwords[db.Password] = struct{}{}
			}
		}
		if len(passwords) == 1 {
			for password := range passwords {
				item.Password = password
			}
		}
		for _, meta := range metadata {
			if meta.Username == user.Username {
				item.Password, item.Description = meta.Password, meta.Description
				break
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (u *PostgresqlService) CreateUser(req dto.PostgresqlUserCreate) error {
	if strings.HasPrefix(req.Username, "pg_") || req.Username == "postgres" {
		return buserr.New("ErrPgUserProtected")
	}
	if req.Username == "" || len(req.Username) > 63 || req.Password == "" || strings.ContainsAny(req.Username+req.Password, "\x00") {
		return buserr.New("ErrInvalidParams")
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.CreateUser(client.CreateInfo{Username: req.Username, Password: req.Password, SuperUser: req.SuperUser, Timeout: 300}, false); err != nil {
		return err
	}
	return saveDatabaseUserCredential("postgresql", req.Database, req.Username, "", req.Password, req.Description)
}

func (u *PostgresqlService) UpdateUser(req dto.PostgresqlUserUpdate) error {
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	if user.Protected {
		return buserr.New("ErrPgUserProtected")
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.ChangePrivileges(client.Privileges{Username: req.Username, SuperUser: req.SuperUser, Timeout: 300}); err != nil {
		return err
	}
	dbs, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(req.Database))
	if err != nil {
		return err
	}
	for _, db := range dbs {
		if db.Username == req.Username {
			if err := postgresqlRepo.Update(db.ID, map[string]interface{}{"super_user": req.SuperUser}); err != nil {
				return err
			}
		}
	}
	return saveDatabaseUserCredential("postgresql", req.Database, req.Username, "", user.Password, req.Description)
}

func (u *PostgresqlService) ChangeUserPassword(req dto.PostgresqlUserPassword) error {
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	if user.Protected {
		return buserr.New("ErrPgUserProtected")
	}
	if req.Password == "" || strings.ContainsRune(req.Password, 0) {
		return buserr.New("ErrInvalidParams")
	}
	targets, err := loadPostgresqlUserApps(req.Database, req.Username, "")
	if err != nil {
		return err
	}
	dbs, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(req.Database))
	if err != nil {
		return err
	}
	password, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.ChangePassword(client.PasswordChangeInfo{Username: req.Username, Password: req.Password, Timeout: 300}); err != nil {
		return err
	}
	if err := saveDatabaseUserCredential("postgresql", req.Database, req.Username, "", req.Password, user.Description); err != nil {
		return err
	}
	for _, db := range dbs {
		if db.Username == req.Username {
			if err := postgresqlRepo.Update(db.ID, map[string]interface{}{"password": password}); err != nil {
				return err
			}
		}
	}
	var failures []string
	for _, app := range targets {
		if err := updateInstallInfoInDB(app.App.Key, app.Name, "user-password", req.Password); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", app.Name, err))
		}
	}
	if len(failures) != 0 {
		return buserr.WithDetail("ErrPgUserPasswordSync", strings.Join(failures, "; "), nil)
	}
	return nil
}

func (u *PostgresqlService) SaveUserPassword(req dto.PostgresqlUserPassword) error {
	if req.Password == "" || strings.ContainsRune(req.Password, 0) {
		return buserr.New("ErrInvalidParams")
	}
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	if user.Protected {
		return buserr.New("ErrPgUserProtected")
	}
	metadata, err := databaseUserRepo.Get(repo.WithByType("postgresql"), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, ""))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	metadata.Type, metadata.Database, metadata.Username, metadata.Host = "postgresql", req.Database, req.Username, ""
	metadata.Password, metadata.Description, metadata.IsDelete = req.Password, user.Description, false
	return postgresqlRepo.SaveUserPassword(&metadata)
}

func (u *PostgresqlService) DeleteUser(req dto.PostgresqlUserDelete) error {
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	if user.Protected {
		return buserr.New("ErrPgUserProtected")
	}
	if err := checkPostgresqlUserUsage(req.Database, req.Username, ""); err != nil {
		return err
	}
	if len(user.Databases) != 0 {
		return buserr.New("ErrPgUserBound")
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Delete(client.DeleteInfo{Username: req.Username, Timeout: 300}); err != nil {
		return err
	}
	dbs, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(req.Database))
	if err != nil {
		return err
	}
	for _, db := range dbs {
		if db.Username == req.Username {
			if err := postgresqlRepo.Update(db.ID, map[string]interface{}{"username": "", "password": "", "super_user": false}); err != nil {
				return err
			}
		}
	}
	return databaseUserRepo.Delete(repo.WithByType("postgresql"), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, ""))
}

func (u *PostgresqlService) GrantUser(req dto.PostgresqlUserGrant) error {
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	if user.Protected {
		return buserr.New("ErrPgUserProtected")
	}
	if user.SuperUser {
		return buserr.New("ErrPgSuperUserGrant")
	}
	db, err := postgresqlRepo.Get(postgresqlRepo.WithByPostgresqlName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := client.GrantUser(cli, req.Name, req.Username, false); err != nil {
		return err
	}
	if db.Username == "" {
		password, err := encrypt.StringEncrypt(user.Password)
		if err != nil {
			return err
		}
		return postgresqlRepo.Update(db.ID, map[string]interface{}{"username": req.Username, "password": password, "super_user": false})
	}
	return nil
}

func (u *PostgresqlService) RevokeUser(req dto.PostgresqlUserGrant) error {
	user, err := u.loadPostgresqlUser(req.Database, req.Username)
	if err != nil {
		return err
	}
	if user.Protected {
		return buserr.New("ErrPgUserProtected")
	}
	if user.SuperUser {
		return buserr.New("ErrPgSuperUserGrant")
	}
	db, err := postgresqlRepo.Get(postgresqlRepo.WithByPostgresqlName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return err
	}
	if err := checkPostgresqlUserUsage(req.Database, req.Username, req.Name); err != nil {
		return err
	}
	cli, err := LoadPostgresqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := saveDatabaseUserCredential("postgresql", req.Database, req.Username, "", user.Password, user.Description); err != nil {
		return err
	}
	if err := client.GrantUser(cli, req.Name, req.Username, true); err != nil {
		return err
	}
	if db.Username == req.Username {
		return postgresqlRepo.Update(db.ID, map[string]interface{}{"username": "", "password": "", "super_user": false})
	}
	return nil
}

func (u *PostgresqlService) loadPostgresqlUser(database, username string) (dto.PostgresqlUser, error) {
	users, err := u.ListUsers(dto.PostgresqlUserSearch{Database: database})
	if err != nil {
		return dto.PostgresqlUser{}, err
	}
	for _, user := range users {
		if user.Username == username {
			return user, nil
		}
	}
	return dto.PostgresqlUser{}, buserr.New("ErrRecordNotFound")
}

func loadPostgresqlUserApps(database, username, dbName string) ([]model.AppInstall, error) {
	instance, err := databaseRepo.Get(repo.WithByName(database))
	if err != nil {
		return nil, err
	}
	dbs, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(database))
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(dbs))
	for _, db := range dbs {
		if dbName == "" || db.Name == dbName {
			ids = append(ids, db.ID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	opts := []repo.DBOption{appInstallResourceRepo.WithResourceIds(ids), appRepo.WithKey(instance.Type)}
	if instance.From == "local" {
		app, err := appInstallRepo.LoadBaseInfo(instance.Type, database)
		if err != nil {
			return nil, err
		}
		opts = append(opts, appInstallResourceRepo.WithLinkId(app.ID))
	}
	resources, err := appInstallResourceRepo.GetBy(opts...)
	if err != nil {
		return nil, err
	}
	appIDs := make([]uint, 0, len(resources))
	for _, resource := range resources {
		appIDs = append(appIDs, resource.AppInstallId)
	}
	if len(appIDs) == 0 {
		return nil, nil
	}
	apps, err := appInstallRepo.ListBy(context.Background(), repo.WithByIDs(appIDs))
	if err != nil {
		return nil, err
	}
	result := make([]model.AppInstall, 0, len(apps))
	for _, app := range apps {
		var env map[string]interface{}
		if err := json.Unmarshal([]byte(app.Env), &env); err != nil {
			return nil, err
		}
		if env["PANEL_DB_USER"] == username {
			result = append(result, app)
		}
	}
	return result, nil
}

func checkPostgresqlUserUsage(database, username, dbName string) error {
	apps, err := loadPostgresqlUserApps(database, username, dbName)
	if err != nil {
		return err
	}
	if len(apps) != 0 {
		return buserr.New("ErrPgUserUsed")
	}
	dbs, err := postgresqlRepo.List(postgresqlRepo.WithByPostgresqlName(database))
	if err != nil {
		return err
	}
	websites, err := websiteRepo.List(websiteRepo.WithDBTypes([]string{constant.AppPostgresql, constant.AppPostgresqlCluster}))
	if err != nil {
		return err
	}
	for _, db := range dbs {
		if db.Username != username || (dbName != "" && db.Name != dbName) {
			continue
		}
		for _, website := range websites {
			if website.DbID == db.ID {
				return buserr.New("ErrPgUserUsed")
			}
		}
	}
	return nil
}
