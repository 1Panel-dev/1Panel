package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/mysql"
	"github.com/1Panel-dev/1Panel/agent/utils/mysql/client"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.MysqlDBSearch) (int64, interface{}, error)
	ListDBOption() ([]dto.MysqlOption, error)
	Create(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error)
	LoadFromRemote(req dto.MysqlLoadDB) error
	ChangeAccess(info dto.ChangeDBInfo) error
	ChangePassword(info dto.ChangeDBInfo) error
	UpdateVariables(req dto.MysqlVariablesUpdate) error
	UpdateDescription(req dto.UpdateDescription) error
	DeleteCheck(req dto.MysqlDBDeleteCheck) ([]dto.DBResource, error)
	Delete(ctx context.Context, req dto.MysqlDBDelete) error

	ListUsers(req dto.MysqlUserSearch) ([]dto.MysqlUser, error)
	ListGrants(req dto.MysqlUserSearch) ([]dto.MysqlGrant, error)
	ListGrantSummary(req dto.MysqlGrantSummarySearch) (map[string][]dto.MysqlUser, error)
	CreateUser(req dto.MysqlUserCreate) error
	UpdateUser(req dto.MysqlUserUpdate) error
	ChangeUserPassword(req dto.MysqlUserPassword) error
	SaveUserPassword(req dto.MysqlUserPassword) error
	DeleteUser(req dto.MysqlUserDelete) error
	GrantUser(req dto.MysqlGrantCreate) error
	RevokeGrant(req dto.MysqlGrantDelete) error

	LoadFormatOption(req dto.OperationWithName) []dto.MysqlFormatCollationOption
	LoadStatus(req dto.OperationWithNameAndType) (*dto.MysqlStatus, error)
	LoadVariables(req dto.OperationWithNameAndType) (*dto.MysqlVariables, error)
	LoadRemoteAccess(req dto.OperationWithNameAndType) (bool, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func normalizeDatabaseUserType(dbType string) string {
	if len(dbType) == 0 {
		return "mysql"
	}
	return dbType
}

func resolveDatabaseUserType(database string) (string, error) {
	databaseItem, err := databaseRepo.Get(repo.WithByName(database))
	if err != nil {
		return "", err
	}
	return normalizeDatabaseUserType(databaseItem.Type), nil
}

func databaseUserKey(username, host string) string {
	return username + "@" + host
}

func splitMysqlHosts(permission string) []string {
	hosts := strings.Split(permission, ",")
	res := make([]string, 0, len(hosts))
	seen := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		res = append(res, host)
	}
	return res
}

func parseMysqlHosts(permission string, allowMultiple bool) ([]string, error) {
	hosts := splitMysqlHosts(permission)
	if len(hosts) == 0 {
		return nil, errors.New("mysql user host is required")
	}
	if !allowMultiple && len(hosts) > 1 {
		return nil, errors.New("multiple mysql user hosts are not supported for this operation")
	}
	return hosts, nil
}

func checkMysqlNormalUser(username string) error {
	if isMysqlSystemUser(username) {
		return errors.New("mysql system user does not support this operation")
	}
	return nil
}

// isMysqlSystemUser identifies internal accounts which must only be managed by
// MySQL or its container runtime. The root account is managed through the
// dedicated root API.
func isMysqlSystemUser(username string) bool {
	switch strings.ToLower(username) {
	case "root",
		"mysql.session", "mysql.sys", "mysql.infoschema", "mysqlxsys",
		"mariadb.sys", "mariadb-sys",
		"debian-sys-maint",
		"healthcheck":
		return true
	default:
		return false
	}
}

func saveDatabaseUserCredential(dbType, database, username, host, password, description string) error {
	dbType = normalizeDatabaseUserType(dbType)
	user, err := databaseUserRepo.Get(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(database), databaseUserRepo.WithByUser(username, host))
	if err != nil {
		user = model.DatabaseUser{
			Type:     dbType,
			Database: database,
			Username: username,
			Host:     host,
		}
	}
	user.Password = password
	user.IsDelete = false
	if len(description) != 0 {
		user.Description = description
	}
	return databaseUserRepo.Save(&user)
}

func saveDatabaseUserCredentials(dbType, database, username, permission, password, description string) error {
	for _, host := range splitMysqlHosts(permission) {
		if err := saveDatabaseUserCredential(dbType, database, username, host, password, description); err != nil {
			return err
		}
	}
	return nil
}

type mysqlUserAppTarget struct {
	Key  string
	Name string
}

func loadMysqlUserAppTargets(dbType, database, username, host string, dbNames ...string) ([]mysqlUserAppTarget, error) {
	if host != "%" {
		return nil, nil
	}
	dbNameSet := make(map[string]struct{}, len(dbNames))
	for _, dbName := range dbNames {
		dbNameSet[dbName] = struct{}{}
	}
	dbItems, err := mysqlRepo.List(mysqlRepo.WithByMysqlName(database))
	if err != nil {
		return nil, err
	}
	localResourceIDs := make([]uint, 0, len(dbItems))
	remoteResourceIDs := make([]uint, 0, len(dbItems))
	for _, dbItem := range dbItems {
		if len(dbNameSet) != 0 {
			if _, ok := dbNameSet[dbItem.Name]; !ok {
				continue
			}
		}
		if dbItem.ID == 0 {
			continue
		}
		if dbItem.From == "local" {
			localResourceIDs = append(localResourceIDs, dbItem.ID)
		} else {
			remoteResourceIDs = append(remoteResourceIDs, dbItem.ID)
		}
	}

	appResources := make([]model.AppInstallResource, 0)
	if len(localResourceIDs) != 0 {
		app, err := appInstallRepo.LoadBaseInfo(dbType, database)
		if err != nil {
			return nil, err
		}
		resources, err := appInstallResourceRepo.GetBy(
			appInstallResourceRepo.WithLinkId(app.ID),
			appInstallResourceRepo.WithResourceIds(localResourceIDs),
		)
		if err != nil {
			return nil, err
		}
		appResources = append(appResources, resources...)
	}
	if len(remoteResourceIDs) != 0 {
		resources, err := appInstallResourceRepo.GetBy(
			appInstallResourceRepo.WithResourceIds(remoteResourceIDs),
			appRepo.WithKey(dbType),
		)
		if err != nil {
			return nil, err
		}
		appResources = append(appResources, resources...)
	}

	appInstallIDs := make([]uint, 0, len(appResources))
	appInstallIDSet := make(map[uint]struct{}, len(appResources))
	for _, appResource := range appResources {
		if appResource.AppInstallId == 0 {
			continue
		}
		if _, ok := appInstallIDSet[appResource.AppInstallId]; ok {
			continue
		}
		appInstallIDSet[appResource.AppInstallId] = struct{}{}
		appInstallIDs = append(appInstallIDs, appResource.AppInstallId)
	}
	if len(appInstallIDs) == 0 {
		return nil, nil
	}
	appInstalls, err := appInstallRepo.ListBy(context.Background(), repo.WithByIDs(appInstallIDs))
	if err != nil {
		return nil, err
	}
	targets := make([]mysqlUserAppTarget, 0, len(appInstalls))
	for _, appInstall := range appInstalls {
		var envMap map[string]interface{}
		if err := json.Unmarshal([]byte(appInstall.Env), &envMap); err != nil {
			return nil, err
		}
		appUsername, ok := envMap["PANEL_DB_USER"].(string)
		if !ok || appUsername != username {
			continue
		}
		targets = append(targets, mysqlUserAppTarget{Key: appInstall.App.Key, Name: appInstall.Name})
	}
	return targets, nil
}

func checkMysqlUserAppUsage(dbType, database, username, host string, dbNames ...string) error {
	targets, err := loadMysqlUserAppTargets(dbType, database, username, host, dbNames...)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	appNames := make([]string, 0, len(targets))
	for _, target := range targets {
		appNames = append(appNames, target.Name)
	}
	sort.Strings(appNames)
	return buserr.WithDetail("ErrMysqlUserUsedByApps", strings.Join(appNames, ", "), nil)
}

func updateMysqlPasswordAppTargets(targets []mysqlUserAppTarget, password string) error {
	for _, target := range targets {
		global.LOG.Infof("start to update mysql password used by app %s-%s", target.Key, target.Name)
		if err := updateInstallInfoInDB(target.Key, target.Name, "user-password", password); err != nil {
			return err
		}
	}
	return nil
}

func syncDatabaseUserMetadata(dbType, database string, users []client.UserInfo) error {
	dbType = normalizeDatabaseUserType(dbType)
	metas, err := databaseUserRepo.List(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(database))
	if err != nil {
		return err
	}
	metaMap := make(map[string]model.DatabaseUser, len(metas))
	for _, item := range metas {
		metaMap[databaseUserKey(item.Username, item.Host)] = item
	}
	userMap := make(map[string]struct{}, len(users))
	for _, item := range users {
		if isMysqlSystemUser(item.Username) {
			continue
		}
		key := databaseUserKey(item.Username, item.Host)
		userMap[key] = struct{}{}
		if meta, ok := metaMap[key]; ok {
			if meta.IsDelete {
				if err := databaseUserRepo.Update(map[string]interface{}{"is_delete": false, "password": ""}, repo.WithByType(dbType), databaseUserRepo.WithByDatabase(database), databaseUserRepo.WithByUser(item.Username, item.Host)); err != nil {
					return err
				}
			}
			continue
		}
		if err := databaseUserRepo.Save(&model.DatabaseUser{
			Type:     dbType,
			Database: database,
			Username: item.Username,
			Host:     item.Host,
		}); err != nil {
			return err
		}
	}
	for _, item := range metas {
		if isMysqlSystemUser(item.Username) {
			continue
		}
		if _, ok := userMap[databaseUserKey(item.Username, item.Host)]; ok {
			continue
		}
		if item.IsDelete {
			continue
		}
		if err := databaseUserRepo.Update(map[string]interface{}{"is_delete": true, "password": ""}, repo.WithByType(dbType), databaseUserRepo.WithByDatabase(database), databaseUserRepo.WithByUser(item.Username, item.Host)); err != nil {
			return err
		}
	}
	return nil
}

func saveDatabaseUserGrant(dbType, database, dbName, username, host string) error {
	dbType = normalizeDatabaseUserType(dbType)
	grant, err := databaseUserGrantRepo.Get(
		repo.WithByType(dbType),
		databaseUserGrantRepo.WithByDatabase(database),
		databaseUserGrantRepo.WithByDBName(dbName),
		databaseUserGrantRepo.WithByUser(username, host),
	)
	if err == nil && grant.ID != 0 {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return databaseUserGrantRepo.Save(&model.DatabaseUserGrant{
		Type:     dbType,
		Database: database,
		DBName:   dbName,
		Username: username,
		Host:     host,
	})
}

func syncDatabaseUserGrants(dbType, database string, grants []client.GrantInfo) error {
	dbType = normalizeDatabaseUserType(dbType)
	items := make([]model.DatabaseUserGrant, 0, len(grants))
	grantMap := make(map[string]struct{}, len(grants))
	for _, item := range grants {
		if isMysqlSystemUser(item.Username) || item.Database == "*" {
			continue
		}
		key := item.Database + "\x00" + item.Username + "\x00" + item.Host
		if _, ok := grantMap[key]; ok {
			continue
		}
		grantMap[key] = struct{}{}
		items = append(items, model.DatabaseUserGrant{
			Type:     dbType,
			Database: database,
			DBName:   item.Database,
			Username: item.Username,
			Host:     item.Host,
		})
	}
	return databaseUserGrantRepo.Replace(dbType, database, items)
}

func (u *MysqlService) SearchWithPage(search dto.MysqlDBSearch) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize,
		mysqlRepo.WithByMysqlName(search.Database),
		repo.WithByLikeName(search.Info),
		repo.WithOrderRuleBy(search.OrderBy, search.Order),
	)
	var dtoMysqls []dto.MysqlDBInfo
	for _, mysql := range mysqls {
		var item dto.MysqlDBInfo
		if err := copier.Copy(&item, &mysql); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		item.Username = ""
		item.Password = ""
		item.Permission = ""
		dtoMysqls = append(dtoMysqls, item)
	}
	return total, dtoMysqls, err
}

func (u *MysqlService) ListDBOption() ([]dto.MysqlOption, error) {
	mysqls, err := mysqlRepo.List()
	if err != nil {
		return nil, err
	}

	databases, err := databaseRepo.GetList(databaseRepo.WithTypeList("mysql,mariadb"))
	if err != nil {
		return nil, err
	}
	var dbs []dto.MysqlOption
	for _, mysql := range mysqls {
		var item dto.MysqlOption
		if err := copier.Copy(&item, &mysql); err != nil {
			return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		item.Database = mysql.MysqlName
		for _, database := range databases {
			if database.Name == item.Database {
				item.Type = database.Type
			}
		}
		dbs = append(dbs, item)
	}
	return dbs, err
}

func (u *MysqlService) Create(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error) {
	if cmd.CheckIllegal(req.Name, req.Username, req.Password, req.Format, req.Collation, req.Permission) {
		return nil, buserr.New("ErrCmdIllegal")
	}
	if len(req.Username) != 0 && len(req.Password) == 0 {
		return nil, errors.New("password is required when creating mysql user")
	}
	permissionHosts := make([]string, 0)
	if len(req.Username) != 0 {
		var err error
		permissionHosts, err = parseMysqlHosts(req.Permission, true)
		if err != nil {
			return nil, err
		}
		req.Permission = strings.Join(permissionHosts, ",")
	}
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return nil, err
	}

	mysql, _ := mysqlRepo.Get(repo.WithByName(req.Name), mysqlRepo.WithByMysqlName(req.Database), repo.WithByFrom(req.From))
	if mysql.ID != 0 {
		return nil, buserr.New("ErrRecordExist")
	}

	var createItem model.DatabaseMysql
	if err := copier.Copy(&createItem, &req); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	createItem.Username = ""
	createItem.Password = ""
	createItem.Permission = ""

	if req.From == "local" && req.Username == "root" {
		return nil, errors.New("cannot set root as user name")
	}

	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return nil, err
	}
	createItem.MysqlName = req.Database
	defer cli.Close()
	if err := cli.Create(client.CreateInfo{
		Name:       req.Name,
		Format:     req.Format,
		Collation:  req.Collation,
		Username:   req.Username,
		Password:   req.Password,
		Permission: req.Permission,
		Version:    version,
		Timeout:    300,
	}); err != nil {
		return nil, err
	}
	if len(req.Username) != 0 {
		if err := saveDatabaseUserCredentials(dbType, req.Database, req.Username, req.Permission, req.Password, req.Description); err != nil {
			return nil, err
		}
		for _, host := range permissionHosts {
			if err := saveDatabaseUserGrant(dbType, req.Database, req.Name, req.Username, host); err != nil {
				return nil, err
			}
		}
	}

	global.LOG.Infof("create database %s successful!", req.Name)
	if err := mysqlRepo.Create(ctx, &createItem); err != nil {
		return nil, err
	}
	return &createItem, nil
}

func (u *MysqlService) ListUsers(req dto.MysqlUserSearch) ([]dto.MysqlUser, error) {
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return nil, err
	}
	users, err := databaseUserRepo.List(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database))
	if err != nil {
		return nil, err
	}
	res := make([]dto.MysqlUser, 0, len(users))
	for _, user := range users {
		if isMysqlSystemUser(user.Username) {
			continue
		}
		res = append(res, dto.MysqlUser{
			Username:    user.Username,
			Host:        user.Host,
			Password:    user.Password,
			Description: user.Description,
			IsDelete:    user.IsDelete,
		})
	}
	return res, nil
}

func (u *MysqlService) ListGrants(req dto.MysqlUserSearch) ([]dto.MysqlGrant, error) {
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return nil, err
	}
	grants, err := databaseUserGrantRepo.List(repo.WithByType(dbType), databaseUserGrantRepo.WithByDatabase(req.Database))
	if err != nil {
		return nil, err
	}
	res := make([]dto.MysqlGrant, 0, len(grants))
	for _, grant := range grants {
		res = append(res, dto.MysqlGrant{Database: grant.DBName, Username: grant.Username, Host: grant.Host})
	}
	return res, nil
}

func (u *MysqlService) ListGrantSummary(req dto.MysqlGrantSummarySearch) (map[string][]dto.MysqlUser, error) {
	res := make(map[string][]dto.MysqlUser, len(req.DBs))
	dbMap := make(map[string]struct{}, len(req.DBs))
	dbNames := make([]string, 0, len(req.DBs))
	for _, item := range req.DBs {
		if item == "" {
			continue
		}
		if _, ok := dbMap[item]; ok {
			continue
		}
		dbMap[item] = struct{}{}
		dbNames = append(dbNames, item)
		res[item] = []dto.MysqlUser{}
	}
	if len(dbMap) == 0 {
		return res, nil
	}
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return nil, err
	}
	grants, err := databaseUserGrantRepo.List(
		repo.WithByType(dbType),
		databaseUserGrantRepo.WithByDatabase(req.Database),
		databaseUserGrantRepo.WithByDBNames(dbNames),
	)
	if err != nil {
		return nil, err
	}
	userList := make([][2]string, 0, len(grants))
	userSet := make(map[string]struct{}, len(grants))
	for _, grant := range grants {
		key := databaseUserKey(grant.Username, grant.Host)
		if _, ok := userSet[key]; ok {
			continue
		}
		userSet[key] = struct{}{}
		userList = append(userList, [2]string{grant.Username, grant.Host})
	}
	userMetas, err := databaseUserRepo.List(
		repo.WithByType(dbType),
		databaseUserRepo.WithByDatabase(req.Database),
		databaseUserRepo.WithByUserList(userList),
	)
	if err != nil {
		return nil, err
	}
	metaMap := make(map[string]model.DatabaseUser, len(userMetas))
	for _, item := range userMetas {
		metaMap[databaseUserKey(item.Username, item.Host)] = item
	}

	for _, grant := range grants {
		if _, ok := dbMap[grant.DBName]; !ok {
			continue
		}
		meta, ok := metaMap[databaseUserKey(grant.Username, grant.Host)]
		if !ok || meta.IsDelete {
			continue
		}
		item := dto.MysqlUser{Username: grant.Username, Host: grant.Host, Description: meta.Description}
		res[grant.DBName] = append(res[grant.DBName], item)
	}
	return res, nil
}

func (u *MysqlService) CreateUser(req dto.MysqlUserCreate) error {
	if cmd.CheckIllegal(req.Username, req.Password, req.Host) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	hosts, err := parseMysqlHosts(req.Host, true)
	if err != nil {
		return err
	}
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	createdHosts := make([]string, 0, len(hosts))
	rollbackCreatedHosts := func() {
		for i := len(createdHosts) - 1; i >= 0; i-- {
			host := createdHosts[i]
			if rollbackErr := cli.DeleteUser(client.UserInfo{Username: req.Username, Host: host}, version, 300); rollbackErr != nil {
				global.LOG.Errorf("rollback mysql user %s@%s failed, err: %v", req.Username, host, rollbackErr)
			}
		}
	}
	for _, host := range hosts {
		if err := cli.CreateUserOnly(client.UserInfo{Username: req.Username, Host: host}, req.Password, 300); err != nil {
			rollbackCreatedHosts()
			return err
		}
		createdHosts = append(createdHosts, host)
	}
	savedHosts := make([]string, 0, len(hosts))
	for _, host := range hosts {
		if err := saveDatabaseUserCredential(dbType, req.Database, req.Username, host, req.Password, req.Description); err != nil {
			for _, savedHost := range savedHosts {
				if rollbackErr := databaseUserRepo.Delete(
					repo.WithByType(dbType),
					databaseUserRepo.WithByDatabase(req.Database),
					databaseUserRepo.WithByUser(req.Username, savedHost),
				); rollbackErr != nil {
					global.LOG.Errorf("rollback mysql user record %s@%s failed, err: %v", req.Username, savedHost, rollbackErr)
				}
			}
			rollbackCreatedHosts()
			return err
		}
		savedHosts = append(savedHosts, host)
	}
	return nil
}

func (u *MysqlService) UpdateUser(req dto.MysqlUserUpdate) error {
	if cmd.CheckIllegal(req.Username, req.Host, req.NewHost) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	oldHosts, err := parseMysqlHosts(req.Host, false)
	if err != nil {
		return err
	}
	newHosts, err := parseMysqlHosts(req.NewHost, false)
	if err != nil {
		return err
	}
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	req.Host = oldHosts[0]
	req.NewHost = newHosts[0]
	if req.Host != req.NewHost {
		targetUser, _ := databaseUserRepo.Get(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, req.NewHost))
		if targetUser.ID != 0 {
			return buserr.New("ErrRecordExist")
		}
		if err := checkMysqlUserAppUsage(dbType, req.Database, req.Username, req.Host); err != nil {
			return err
		}
		cli, _, err := LoadMysqlClientByFrom(req.Database)
		if err != nil {
			return err
		}
		defer cli.Close()
		if err := cli.UpdateUser(client.UserUpdateInfo{Username: req.Username, Host: req.Host, NewHost: req.NewHost}, 300); err != nil {
			return err
		}
		user, err := databaseUserRepo.Get(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, req.Host))
		if err != nil {
			user = model.DatabaseUser{
				Type:     dbType,
				Database: req.Database,
				Username: req.Username,
			}
		}
		user.Host = req.NewHost
		user.IsDelete = false
		user.Description = req.Description
		if err := databaseUserRepo.Save(&user); err != nil {
			if rollbackErr := cli.UpdateUser(client.UserUpdateInfo{Username: req.Username, Host: req.NewHost, NewHost: req.Host}, 300); rollbackErr != nil {
				global.LOG.Errorf("rollback mysql user %s host from %s to %s failed, err: %v", req.Username, req.NewHost, req.Host, rollbackErr)
			}
			return err
		}
		if err := databaseUserGrantRepo.Update(
			map[string]interface{}{"host": req.NewHost},
			repo.WithByType(dbType),
			databaseUserGrantRepo.WithByDatabase(req.Database),
			databaseUserGrantRepo.WithByUser(req.Username, req.Host),
		); err != nil {
			return err
		}
		return nil
	}
	user, err := databaseUserRepo.Get(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, req.Host))
	if err != nil {
		user = model.DatabaseUser{
			Type:     dbType,
			Database: req.Database,
			Username: req.Username,
			Host:     req.Host,
		}
	}
	user.IsDelete = false
	user.Description = req.Description
	return databaseUserRepo.Save(&user)
}

func (u *MysqlService) ChangeUserPassword(req dto.MysqlUserPassword) error {
	if cmd.CheckIllegal(req.Username, req.Host, req.Password) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	hosts, err := parseMysqlHosts(req.Host, false)
	if err != nil {
		return err
	}
	req.Host = hosts[0]
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	grants, err := cli.ListGrants(300)
	if err != nil {
		return err
	}
	grantDBs := make([]string, 0)
	for _, grant := range grants {
		if grant.Username == req.Username && grant.Host == req.Host {
			grantDBs = append(grantDBs, grant.Database)
		}
	}
	var appTargets []mysqlUserAppTarget
	if len(grantDBs) != 0 {
		appTargets, err = loadMysqlUserAppTargets(dbType, req.Database, req.Username, req.Host, grantDBs...)
		if err != nil {
			return err
		}
	}
	if err := cli.ChangePassword(client.PasswordChangeInfo{
		Username:   req.Username,
		Permission: req.Host,
		Password:   req.Password,
		Version:    version,
		Timeout:    300,
	}); err != nil {
		return err
	}
	if err := saveDatabaseUserCredential(dbType, req.Database, req.Username, req.Host, req.Password, ""); err != nil {
		return err
	}
	return updateMysqlPasswordAppTargets(appTargets, req.Password)
}

func (u *MysqlService) SaveUserPassword(req dto.MysqlUserPassword) error {
	if cmd.CheckIllegal(req.Username, req.Host, req.Password) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	hosts, err := parseMysqlHosts(req.Host, false)
	if err != nil {
		return err
	}
	req.Host = hosts[0]
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	user, err := databaseUserRepo.Get(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, req.Host))
	if err != nil {
		return err
	}
	if user.IsDelete {
		return errors.New("cannot save password for a deleted mysql user")
	}
	user.Password = req.Password
	return databaseUserRepo.Save(&user)
}

func (u *MysqlService) DeleteUser(req dto.MysqlUserDelete) error {
	if cmd.CheckIllegal(req.Username, req.Host) {
		return buserr.New("ErrCmdIllegal")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	hosts, err := parseMysqlHosts(req.Host, false)
	if err != nil {
		return err
	}
	req.Host = hosts[0]
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	if err := checkMysqlUserAppUsage(dbType, req.Database, req.Username, req.Host); err != nil {
		return err
	}
	user, err := databaseUserRepo.Get(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, req.Host))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err != nil || !user.IsDelete {
		cli, version, err := LoadMysqlClientByFrom(req.Database)
		if err != nil {
			return err
		}
		defer cli.Close()
		if err := cli.DeleteUser(client.UserInfo{Username: req.Username, Host: req.Host}, version, 300); err != nil {
			return err
		}
	}
	if err := databaseUserGrantRepo.Delete(repo.WithByType(dbType), databaseUserGrantRepo.WithByDatabase(req.Database), databaseUserGrantRepo.WithByUser(req.Username, req.Host)); err != nil {
		return err
	}
	return databaseUserRepo.Delete(repo.WithByType(dbType), databaseUserRepo.WithByDatabase(req.Database), databaseUserRepo.WithByUser(req.Username, req.Host))
}

func (u *MysqlService) GrantUser(req dto.MysqlGrantCreate) error {
	if cmd.CheckIllegal(req.DB, req.Username, req.Host) {
		return buserr.New("ErrCmdIllegal")
	}
	if req.DB == "*" {
		return errors.New("global mysql privileges must be managed outside 1Panel")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	hosts, err := parseMysqlHosts(req.Host, true)
	if err != nil {
		return err
	}
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	cli, _, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	for _, host := range hosts {
		if err := cli.GrantUser(client.GrantInfo{Database: req.DB, Username: req.Username, Host: host}, 300); err != nil {
			return err
		}
		if err := saveDatabaseUserGrant(dbType, req.Database, req.DB, req.Username, host); err != nil {
			return err
		}
	}
	return nil
}

func (u *MysqlService) RevokeGrant(req dto.MysqlGrantDelete) error {
	if cmd.CheckIllegal(req.DB, req.Username, req.Host) {
		return buserr.New("ErrCmdIllegal")
	}
	if req.DB == "*" {
		return errors.New("global mysql privileges must be managed outside 1Panel")
	}
	if err := checkMysqlNormalUser(req.Username); err != nil {
		return err
	}
	hosts, err := parseMysqlHosts(req.Host, false)
	if err != nil {
		return err
	}
	req.Host = hosts[0]
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	if err := checkMysqlUserAppUsage(dbType, req.Database, req.Username, req.Host, req.DB); err != nil {
		return err
	}
	cli, _, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.RevokeGrant(client.GrantInfo{Database: req.DB, Username: req.Username, Host: req.Host}, 300); err != nil {
		return err
	}
	return databaseUserGrantRepo.Delete(
		repo.WithByType(dbType),
		databaseUserGrantRepo.WithByDatabase(req.Database),
		databaseUserGrantRepo.WithByDBName(req.DB),
		databaseUserGrantRepo.WithByUser(req.Username, req.Host),
	)
}

func (u *MysqlService) LoadFromRemote(req dto.MysqlLoadDB) error {
	dbType, err := resolveDatabaseUserType(req.Database)
	if err != nil {
		return err
	}
	client, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer client.Close()

	databases, err := mysqlRepo.List(mysqlRepo.WithByMysqlName(req.Database))
	if err != nil {
		return err
	}
	datas, err := client.SyncDB(version)
	if err != nil {
		return err
	}
	users, err := client.ListUsers(300)
	if err != nil {
		return err
	}
	grants, err := client.ListGrants(300)
	if err != nil {
		return err
	}
	if err := syncDatabaseUserMetadata(dbType, req.Database, users); err != nil {
		return err
	}
	if err := syncDatabaseUserGrants(dbType, req.Database, grants); err != nil {
		return err
	}
	deleteList := databases
	for _, data := range datas {
		hasOld := false
		for i := 0; i < len(databases); i++ {
			if strings.EqualFold(databases[i].Name, data.Name) && strings.EqualFold(databases[i].MysqlName, data.MysqlName) {
				hasOld = true
				if databases[i].IsDelete {
					_ = mysqlRepo.Update(databases[i].ID, map[string]interface{}{"is_delete": false})
				}
				deleteList = append(deleteList[:i], deleteList[i+1:]...)
				break
			}
		}
		if !hasOld {
			var createItem model.DatabaseMysql
			if err := copier.Copy(&createItem, &data); err != nil {
				return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
			}
			createItem.Username = ""
			createItem.Password = ""
			createItem.Permission = ""
			if err := mysqlRepo.Create(context.Background(), &createItem); err != nil {
				return err
			}
		}
	}
	for _, delItem := range deleteList {
		_ = mysqlRepo.Update(delItem.ID, map[string]interface{}{"is_delete": true})
	}
	return nil
}

func (u *MysqlService) UpdateDescription(req dto.UpdateDescription) error {
	return mysqlRepo.Update(req.ID, map[string]interface{}{"description": req.Description})
}

func loadMysqlDeleteTarget(id uint) (model.DatabaseMysql, string, error) {
	db, err := mysqlRepo.Get(repo.WithByID(id))
	if err != nil {
		return db, "", err
	}
	dbType, err := resolveDatabaseUserType(db.MysqlName)
	if err != nil {
		return db, "", err
	}
	return db, dbType, nil
}

func (u *MysqlService) deleteCheck(db model.DatabaseMysql, dbType string) ([]dto.DBResource, error) {
	var res []dto.DBResource
	websites, err := websiteRepo.GetBy(
		websiteRepo.WithDBTypes([]string{constant.AppMysql, constant.AppMariaDB, constant.AppMysqlCluster}),
		websiteRepo.WithDBID(db.ID),
	)
	if err != nil {
		return res, err
	}
	for _, website := range websites {
		res = append(res, dto.DBResource{
			Type: constant.TypeWebsite,
			Name: website.PrimaryDomain,
		})
	}

	if db.From == "local" {
		app, err := appInstallRepo.LoadBaseInfo(dbType, db.MysqlName)
		if err != nil {
			return res, err
		}
		apps, err := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(app.ID), appInstallResourceRepo.WithResourceId(db.ID))
		if err != nil {
			return res, err
		}
		for _, app := range apps {
			appInstall, err := appInstallRepo.GetFirst(repo.WithByID(app.AppInstallId))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return res, err
			}
			if appInstall.ID != 0 {
				res = append(res, dto.DBResource{
					Type: constant.TypeApp,
					Name: appInstall.Name,
				})
			}
		}
	} else {
		apps, err := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithResourceId(db.ID), appRepo.WithKey(dbType))
		if err != nil {
			return res, err
		}
		for _, app := range apps {
			appInstall, err := appInstallRepo.GetFirst(repo.WithByID(app.AppInstallId))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return res, err
			}
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

func (u *MysqlService) DeleteCheck(req dto.MysqlDBDeleteCheck) ([]dto.DBResource, error) {
	db, dbType, err := loadMysqlDeleteTarget(req.ID)
	if err != nil {
		return nil, err
	}
	return u.deleteCheck(db, dbType)
}

func (u *MysqlService) Delete(ctx context.Context, req dto.MysqlDBDelete) error {
	return u.delete(ctx, req, nil)
}

func (u *MysqlService) delete(ctx context.Context, req dto.MysqlDBDelete, exclusions []dto.DBResource) error {
	db, dbType, err := loadMysqlDeleteTarget(req.ID)
	if err != nil {
		return err
	}
	resources, err := u.deleteCheck(db, dbType)
	if err != nil {
		return err
	}
	if len(exclusions) != 0 {
		exclusionMap := make(map[string]struct{}, len(exclusions))
		for _, exclusion := range exclusions {
			exclusionMap[exclusion.Type+"\x00"+exclusion.Name] = struct{}{}
		}
		filtered := resources[:0]
		for _, resource := range resources {
			if _, ok := exclusionMap[resource.Type+"\x00"+resource.Name]; ok {
				continue
			}
			filtered = append(filtered, resource)
		}
		resources = filtered
	}
	if len(resources) != 0 {
		names := make([]string, 0, len(resources))
		for _, resource := range resources {
			names = append(names, resource.Name)
		}
		sort.Strings(names)
		return buserr.WithDetail("ErrInUsed", strings.Join(names, ", "), nil)
	}
	cli, version, err := LoadMysqlClientByFrom(db.MysqlName)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.DeleteDatabase(client.DeleteInfo{
		Name:    db.Name,
		Version: version,
		Timeout: 300,
	}); err != nil && !req.ForceDelete {
		return err
	}

	if req.DeleteBackup {
		uploadDir := filepath.Join(global.Dir.DataDir, fmt.Sprintf("uploads/database/%s/%s/%s", dbType, db.MysqlName, db.Name))
		if _, err := os.Stat(uploadDir); err == nil {
			_ = os.RemoveAll(uploadDir)
		}
		backupDir := filepath.Join(global.Dir.LocalBackupDir, fmt.Sprintf("database/%s/%s/%s", dbType, db.MysqlName, db.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		_ = backupRepo.DeleteRecord(ctx, repo.WithByType(dbType), repo.WithByName(db.MysqlName), repo.WithByDetailName(db.Name))
		global.LOG.Infof("delete database %s-%s backups successful", db.MysqlName, db.Name)
	}

	_ = mysqlRepo.Delete(ctx, repo.WithByID(db.ID))
	return nil
}

func deleteMysqlDatabaseForResourceOwner(ctx context.Context, req dto.MysqlDBDelete, owner dto.DBResource) error {
	return (&MysqlService{}).delete(ctx, req, []dto.DBResource{owner})
}

func isMysqlDatabaseResourceInUseError(err error) bool {
	businessErr, ok := err.(buserr.BusinessError)
	return ok && businessErr.Msg == "ErrInUsed"
}

func (u *MysqlService) ChangePassword(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) {
		return buserr.New("ErrCmdIllegal")
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if req.ID != 0 {
		return errors.New("mysql user password should be changed by user api")
	}
	passwordInfo := client.PasswordChangeInfo{
		Username: "root",
		Password: req.Value,
		Timeout:  300,
		Version:  version,
	}
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

func (u *MysqlService) ChangeAccess(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) {
		return buserr.New("ErrCmdIllegal")
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if req.ID != 0 {
		return errors.New("mysql user access should be changed by user api")
	}
	accessInfo := client.AccessChangeInfo{
		Username:   "root",
		Permission: req.Value,
		Timeout:    300,
		Version:    version,
	}
	if err := cli.ChangeAccess(accessInfo); err != nil {
		return err
	}

	return nil
}

func (u *MysqlService) UpdateVariables(req dto.MysqlVariablesUpdate) error {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
	if err != nil {
		return err
	}
	var files []string

	path := fmt.Sprintf("%s/%s/%s/conf/my.cnf", global.Dir.AppInstallDir, req.Type, app.Name)
	lineBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	files = strings.Split(string(lineBytes), "\n")

	group := "[mysqld]"
	for _, info := range req.Variables {
		if info.Param == "slow_query_log" && info.Value == "ON" {
			logFilePath := filepath.Join(global.Dir.DataDir, fmt.Sprintf("apps/%s/%s/data/1Panel-slow.log", app.Key, app.Name))
			if req.Type == "mariadb" {
				logFilePath = filepath.Join(global.Dir.DataDir, fmt.Sprintf("apps/%s/%s/db/data/1Panel-slow.log", app.Key, app.Name))
			}
			file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
		}
		if !strings.HasPrefix(app.Version, "5.7") && !strings.HasPrefix(app.Version, "5.6") {
			if info.Param == "query_cache_size" {
				continue
			}
		}

		if _, ok := info.Value.(float64); ok {
			files = updateMyCnf(files, group, info.Param, common.LoadSizeUnit(info.Value.(float64)))
		} else {
			files = updateMyCnf(files, group, info.Param, info.Value)
		}
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(files, "\n"))
	if err != nil {
		return err
	}

	if _, err := compose.Restart(fmt.Sprintf("%s/%s/%s/docker-compose.yml", global.Dir.AppInstallDir, req.Type, app.Name)); err != nil {
		return err
	}

	return nil
}

func (u *MysqlService) LoadRemoteAccess(req dto.OperationWithNameAndType) (bool, error) {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Name)
	if err != nil {
		return false, err
	}
	hosts, err := executeSqlForRows(app.ContainerName, app.Key, app.Password, "select host from mysql.user where user='root';")
	if err != nil {
		return false, err
	}
	for _, host := range hosts {
		if host == "%" {
			return true, nil
		}
	}

	return false, nil
}

func (u *MysqlService) LoadVariables(req dto.OperationWithNameAndType) (*dto.MysqlVariables, error) {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Name)
	if err != nil {
		return nil, err
	}
	variableMap, err := executeSqlForMaps(app.ContainerName, app.Key, app.Password, "show global variables;")
	if err != nil {
		return nil, err
	}
	var info dto.MysqlVariables
	arr, err := json.Marshal(variableMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)
	return &info, nil
}

func (u *MysqlService) LoadStatus(req dto.OperationWithNameAndType) (*dto.MysqlStatus, error) {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Name)
	if err != nil {
		return nil, err
	}

	statusMap, err := executeSqlForMaps(app.ContainerName, app.Key, app.Password, "show global status;")
	if err != nil {
		return nil, err
	}

	var info dto.MysqlStatus
	arr, err := json.Marshal(statusMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)

	if value, ok := statusMap["Run"]; ok {
		uptime, _ := strconv.Atoi(value)
		info.Run = time.Unix(time.Now().Unix()-int64(uptime), 0).Format(constant.DateTimeLayout)
	} else {
		if value, ok := statusMap["Uptime"]; ok {
			uptime, _ := strconv.Atoi(value)
			info.Run = time.Unix(time.Now().Unix()-int64(uptime), 0).Format(constant.DateTimeLayout)
		}
	}

	info.File = "OFF"
	info.Position = "OFF"
	masterStatus := "show master status;"
	if common.CompareAppVersion(app.Version, "8.4.0") && (req.Type == constant.AppMysql || req.Type == constant.AppMysqlCluster) {
		masterStatus = "show binary log status;"
	}
	rows, err := executeSqlForRows(app.ContainerName, app.Key, app.Password, masterStatus)
	if err != nil {
		return nil, err
	}
	if len(rows) > 2 {
		itemValue := strings.Split(rows[1], "\t")
		if len(itemValue) > 2 {
			info.File = itemValue[0]
			info.Position = itemValue[1]
		}
	}

	return &info, nil
}

func (u *MysqlService) LoadFormatOption(req dto.OperationWithName) []dto.MysqlFormatCollationOption {
	defaultList := []dto.MysqlFormatCollationOption{{Format: "utf8mb4"}, {Format: "utf8mb3"}, {Format: "gbk"}, {Format: "big5"}}
	client, _, err := LoadMysqlClientByFrom(req.Name)
	if err != nil {
		return defaultList
	}
	options, err := client.LoadFormatCollation(3)
	if err != nil {
		return defaultList
	}
	return options
}

func executeSqlForMaps(containerName, dbType, password, command string) (map[string]string, error) {
	if dbType == "mysql-cluster" {
		dbType = "mysql"
	}
	cmd := exec.Command("docker", "exec", containerName, dbType, "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(stdStr, "ERROR ") {
		return nil, errors.New(stdStr)
	}

	rows := strings.Split(stdStr, "\n")
	rowMap := make(map[string]string)
	for _, v := range rows {
		itemRow := strings.Split(v, "\t")
		if len(itemRow) == 2 {
			rowMap[itemRow[0]] = itemRow[1]
		}
	}
	return rowMap, nil
}

func executeSqlForRows(containerName, dbType, password, command string) ([]string, error) {
	if dbType == "mysql-cluster" {
		dbType = "mysql"
	}
	cmd := exec.Command("docker", "exec", containerName, dbType, "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(stdStr, "ERROR ") {
		return nil, errors.New(stdStr)
	}
	return strings.Split(stdStr, "\n"), nil
}

func updateMyCnf(oldFiles []string, group string, param string, value interface{}) []string {
	isOn := false
	hasGroup := false
	hasKey := false
	regItem := re.GetRegex(re.MysqlGroupPattern)
	var newFiles []string
	i := 0
	for _, line := range oldFiles {
		i++
		if strings.HasPrefix(line, group) {
			isOn = true
			hasGroup = true
			newFiles = append(newFiles, line)
			continue
		}
		if !isOn {
			newFiles = append(newFiles, line)
			continue
		}
		if strings.HasPrefix(line, param+"=") || strings.HasPrefix(line, "# "+param+"=") {
			newFiles = append(newFiles, fmt.Sprintf("%s=%v", param, value))
			hasKey = true
			continue
		}
		if regItem.Match([]byte(line)) || i == len(oldFiles) {
			isOn = false
			if !hasKey {
				newFiles = append(newFiles, fmt.Sprintf("%s=%v", param, value))
			}
			newFiles = append(newFiles, line)
			continue
		}
		newFiles = append(newFiles, line)
	}
	if !hasGroup {
		newFiles = append(newFiles, group+"\n")
		newFiles = append(newFiles, fmt.Sprintf("%s=%v\n", param, value))
	}
	return newFiles
}

func LoadMysqlClientByFrom(database string) (mysql.MysqlClient, string, error) {
	var (
		dbInfo  client.DBInfo
		version string
		err     error
	)

	dbInfo.Timeout = 300
	databaseItem, err := databaseRepo.Get(repo.WithByName(database))
	if err != nil {
		return nil, "", err
	}
	dbInfo.Type = databaseItem.Type
	dbInfo.From = databaseItem.From
	dbInfo.Database = database
	if dbInfo.From != "local" {
		dbInfo.Address = databaseItem.Address
		dbInfo.Port = databaseItem.Port
		dbInfo.Username = databaseItem.Username
		dbInfo.Password = databaseItem.Password
		dbInfo.SSL = databaseItem.SSL
		dbInfo.ClientKey = databaseItem.ClientKey
		dbInfo.ClientCert = databaseItem.ClientCert
		dbInfo.RootCert = databaseItem.RootCert
		dbInfo.SkipVerify = databaseItem.SkipVerify
		version = databaseItem.Version

	} else {
		app, err := appInstallRepo.LoadBaseInfo(databaseItem.Type, database)
		if err != nil {
			return nil, "", err
		}
		dbInfo.Address = app.ContainerName
		dbInfo.Username = "root"
		dbInfo.Password = app.Password
		version = app.Version
	}

	cli, err := mysql.NewMysqlClient(dbInfo)
	if err != nil {
		return nil, "", err
	}
	return cli, version, nil
}
