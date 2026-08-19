package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MongodbService struct{}

type IMongodbService interface {
	SearchWithPage(search dto.MongodbDBSearch, readOnly ...bool) (int64, interface{}, error)
	Create(ctx context.Context, req dto.MongodbDBCreate) (*model.DatabaseMongodb, error)
	LoadFromRemote(req dto.MongodbLoadDB) error
	UpdateDescription(req dto.UpdateDescription) error
	BindUser(req dto.MongodbBind) error
	ChangePassword(req dto.MongodbPassword) error
	ChangeRootPassword(req dto.ChangeDBInfo) error
	LoadPrivileges(req dto.MongodbPrivilegesLoad) (string, error)
	ChangePrivileges(req dto.MongodbPrivileges) error
	DeleteCheck(req dto.MongodbDBDeleteCheck) ([]dto.DBResource, error)
	Delete(ctx context.Context, req dto.MongodbDBDelete) error
}

func NewIMongodbService() IMongodbService {
	return &MongodbService{}
}

func (u *MongodbService) SearchWithPage(search dto.MongodbDBSearch, readOnly ...bool) (int64, interface{}, error) {
	total, mongodbs, err := mongodbRepo.Page(
		search.Page,
		search.PageSize,
		mongodbRepo.WithByMongodbName(search.Database),
		repo.WithByLikeName(search.Info),
		repo.WithOrderRuleBy(search.OrderBy, search.Order),
	)
	var dtoMongodbs []dto.MongodbDBInfo
	for _, mongodb := range mongodbs {
		var item dto.MongodbDBInfo
		if err := copier.Copy(&item, &mongodb); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		if isDemoReadOnly(readOnly...) {
			item.Password = ""
		}
		dtoMongodbs = append(dtoMongodbs, item)
	}
	return total, dtoMongodbs, err
}

func (u *MongodbService) UpdateDescription(req dto.UpdateDescription) error {
	return mongodbRepo.Update(req.ID, map[string]interface{}{"description": req.Description})
}

func (u *MongodbService) Create(ctx context.Context, req dto.MongodbDBCreate) (*model.DatabaseMongodb, error) {
	if cmd.CheckIllegal(req.Name, req.Username, req.Password) {
		return nil, buserr.New("ErrCmdIllegal")
	}
	if !isSupportedMongodbPrivilege(req.Permission) {
		return nil, buserr.New("ErrCmdIllegal")
	}

	mongodb, _ := mongodbRepo.Get(repo.WithByName(req.Name), mongodbRepo.WithByMongodbName(req.Database), repo.WithByFrom(req.From))
	if mongodb.ID != 0 {
		return nil, buserr.New("ErrRecordExist")
	}

	var createItem model.DatabaseMongodb
	if err := copier.Copy(&createItem, &req); err != nil {
		return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	createItem.MongodbName = req.Database

	if err := runMongodbCreate(req.From, req.Database, req.Name, req.Username, req.Password, req.Permission); err != nil {
		return nil, err
	}

	global.LOG.Infof("create mongodb database %s successful", req.Name)
	if err := mongodbRepo.Create(ctx, &createItem); err != nil {
		return nil, err
	}
	return &createItem, nil
}

func (u *MongodbService) LoadFromRemote(req dto.MongodbLoadDB) error {
	databases, err := mongodbRepo.List(mongodbRepo.WithByMongodbName(req.Database))
	if err != nil {
		return err
	}

	datas, err := loadMongodbDatabases(req)
	if err != nil {
		return err
	}
	global.LOG.Infof("sync mongodb databases from %s:%s, found %d items", req.From, req.Database, len(datas))

	deleteList := append([]model.DatabaseMongodb(nil), databases...)
	for _, data := range datas {
		hasOld := false
		for i := 0; i < len(databases); i++ {
			if strings.EqualFold(databases[i].Name, data.Name) && strings.EqualFold(databases[i].MongodbName, req.Database) {
				hasOld = true
				updateMap := map[string]interface{}{"is_delete": false}
				if len(data.Username) != 0 {
					updateMap["username"] = data.Username
				}
				_ = mongodbRepo.Update(databases[i].ID, updateMap)
				for j := 0; j < len(deleteList); j++ {
					if deleteList[j].ID == databases[i].ID {
						deleteList = append(deleteList[:j], deleteList[j+1:]...)
						break
					}
				}
				break
			}
		}
		if !hasOld {
			createItem := model.DatabaseMongodb{
				Name:        data.Name,
				From:        req.From,
				MongodbName: req.Database,
				Username:    data.Username,
				Password:    "",
				Description: "",
			}
			if err := mongodbRepo.Create(context.Background(), &createItem); err != nil {
				return err
			}
		}
	}
	for _, delItem := range deleteList {
		_ = mongodbRepo.Update(delItem.ID, map[string]interface{}{"is_delete": true})
	}
	return nil
}

func (u *MongodbService) BindUser(req dto.MongodbBind) error {
	if cmd.CheckIllegal(req.Database, req.Name, req.Username, req.Password) {
		return buserr.New("ErrCmdIllegal")
	}

	dbItem, err := mongodbRepo.Get(mongodbRepo.WithByMongodbName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return err
	}
	if err := bindMongodbUser(dbItem.From, dbItem.MongodbName, dbItem.Name, req.Username, req.Password); err != nil {
		return err
	}
	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return fmt.Errorf("encrypt mongodb db %s password failed, err: %v", req.Name, err)
	}
	return mongodbRepo.Update(dbItem.ID, map[string]interface{}{
		"username": req.Username,
		"password": pass,
	})
}

func (u *MongodbService) ChangePassword(req dto.MongodbPassword) error {
	if cmd.CheckIllegal(req.Database, req.Name, req.Password) {
		return buserr.New("ErrCmdIllegal")
	}

	dbItem, err := mongodbRepo.Get(mongodbRepo.WithByMongodbName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return err
	}
	if dbItem.Username == "" {
		return buserr.New("ErrRecordNotFound")
	}
	if err := updateMongodbPassword(dbItem.From, dbItem.MongodbName, dbItem.Name, dbItem.Username, req.Password); err != nil {
		return err
	}
	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return fmt.Errorf("encrypt mongodb db %s password failed, err: %v", req.Name, err)
	}
	return mongodbRepo.Update(dbItem.ID, map[string]interface{}{"password": pass})
}

func (u *MongodbService) ChangeRootPassword(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) {
		return buserr.New("ErrCmdIllegal")
	}
	if req.From != constant.AppResourceLocal {
		return buserr.New("ErrRecordNotFound")
	}

	appInfo, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
	if err != nil {
		return err
	}
	if appInfo.UserName == "" {
		return buserr.New("ErrRecordNotFound")
	}
	if err := updateMongodbPassword(req.From, req.Database, "admin", appInfo.UserName, req.Value); err != nil {
		return err
	}

	if err := updateInstallInfoInDB(req.Type, req.Database, "password", req.Value); err != nil {
		return err
	}
	remote, err := databaseRepo.Get(repo.WithByName(req.Database))
	if err != nil {
		return err
	}
	pass, err := encrypt.StringEncrypt(req.Value)
	if err != nil {
		return fmt.Errorf("encrypt mongodb root password failed, err: %v", err)
	}
	_ = databaseRepo.Update(remote.ID, map[string]interface{}{"password": pass})
	return nil
}

func (u *MongodbService) DeleteCheck(req dto.MongodbDBDeleteCheck) ([]dto.DBResource, error) {
	var res []dto.DBResource
	db, err := mongodbRepo.Get(repo.WithByID(req.ID))
	if err != nil {
		return res, err
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

func (u *MongodbService) LoadPrivileges(req dto.MongodbPrivilegesLoad) (string, error) {
	dbItem, err := mongodbRepo.Get(mongodbRepo.WithByMongodbName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return "", err
	}
	return loadMongodbPrivilege(dbItem.From, dbItem.MongodbName, dbItem.Name, req.Username)
}

func (u *MongodbService) ChangePrivileges(req dto.MongodbPrivileges) error {
	if cmd.CheckIllegal(req.Database, req.Username) {
		return buserr.New("ErrCmdIllegal")
	}
	if !isSupportedMongodbPrivilege(req.Permission) {
		return buserr.New("ErrCmdIllegal")
	}
	dbItem, err := mongodbRepo.Get(mongodbRepo.WithByMongodbName(req.Database), repo.WithByName(req.Name))
	if err != nil {
		return err
	}
	return updateMongodbPrivilege(dbItem.From, dbItem.MongodbName, dbItem.Name, req.Username, req.Permission)
}

func (u *MongodbService) Delete(ctx context.Context, req dto.MongodbDBDelete) error {
	db, err := mongodbRepo.Get(repo.WithByID(req.ID))
	if err != nil && !req.ForceDelete {
		return err
	}

	if err := runMongodbDelete(db.From, req.Database, db.Name); err != nil && !req.ForceDelete {
		return err
	}

	if req.DeleteBackup {
		uploadDir := filepath.Join(global.Dir.DataDir, fmt.Sprintf("uploads/database/%s/%s/%s", req.Type, req.Database, db.Name))
		if _, err := os.Stat(uploadDir); err == nil {
			_ = os.RemoveAll(uploadDir)
		}
		backupDir := filepath.Join(global.Dir.LocalBackupDir, fmt.Sprintf("database/%s/%s/%s", req.Type, req.Database, db.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		_ = backupRepo.DeleteRecord(ctx, repo.WithByType(req.Type), repo.WithByName(req.Database), repo.WithByDetailName(db.Name))
		global.LOG.Infof("delete mongodb database %s-%s backups successful", req.Database, db.Name)
	}

	_ = mongodbRepo.Delete(ctx, repo.WithByID(db.ID))
	return nil
}

func runMongodbCreate(from, database, dbName, username, password, permission string) error {
	if from == constant.AppResourceRemote {
		return runRemoteMongodbCreate(database, dbName, username, password, permission)
	}
	script, err := buildMongodbCreateScript(dbName, username, password, permission)
	if err != nil {
		return err
	}
	return runMongodbAdminScript(database, script)
}

func runMongodbDelete(from, database, dbName string) error {
	if from == constant.AppResourceRemote {
		return runRemoteMongodbDelete(database, dbName)
	}
	script, err := buildMongodbDeleteScript(dbName)
	if err != nil {
		return err
	}
	return runMongodbAdminScript(database, script)
}

func bindMongodbUser(from, connectionName, dbName, username, password string) error {
	if from == constant.AppResourceRemote {
		return bindRemoteMongodbUser(connectionName, dbName, username, password)
	}
	script, err := buildMongodbBindUserScript(dbName, username, password)
	if err != nil {
		return err
	}
	return runMongodbAdminScript(connectionName, script)
}

func updateMongodbPassword(from, connectionName, dbName, username, password string) error {
	if from == constant.AppResourceRemote {
		return updateRemoteMongodbPasswordOnly(connectionName, dbName, username, password)
	}
	script, err := buildMongodbPasswordScript(dbName, username, password)
	if err != nil {
		return err
	}
	return runMongodbAdminScript(connectionName, script)
}

func runMongodbAdminScript(database, script string) error {
	appInfo, err := appInstallRepo.LoadBaseInfo(constant.AppMongodb, database)
	if err != nil {
		return err
	}
	if appInfo.ContainerName == "" {
		return fmt.Errorf("mongodb container not found for database %s", database)
	}
	return cmd.NewCommandMgr().Run(
		"docker",
		"exec",
		appInfo.ContainerName,
		"mongosh",
		buildMongodbRestoreURI(appInfo.UserName, appInfo.Password),
		"--quiet",
		"--eval",
		script,
	)
}

func runMongodbAdminScriptWithStdout(database, script string) (string, error) {
	appInfo, err := appInstallRepo.LoadBaseInfo(constant.AppMongodb, database)
	if err != nil {
		return "", err
	}
	if appInfo.ContainerName == "" {
		return "", fmt.Errorf("mongodb container not found for database %s", database)
	}
	return cmd.NewCommandMgr().RunWithStdout(
		"docker",
		"exec",
		appInfo.ContainerName,
		"mongosh",
		buildMongodbRestoreURI(appInfo.UserName, appInfo.Password),
		"--quiet",
		"--eval",
		script,
	)
}

func buildMongodbCreateScript(dbName, username, password, permission string) (string, error) {
	dbNameJSON, err := json.Marshal(dbName)
	if err != nil {
		return "", err
	}
	usernameJSON, err := json.Marshal(username)
	if err != nil {
		return "", err
	}
	passwordJSON, err := json.Marshal(password)
	if err != nil {
		return "", err
	}
	permissionJSON, err := json.Marshal(permission)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
const dbName = %s;
const userName = %s;
const password = %s;
const permission = %s;
const targetDb = db.getSiblingDB(dbName);
targetDb.createCollection("_init");
targetDb.createUser({
  user: userName,
  pwd: password,
  roles: [{ role: permission, db: dbName }]
});
`, dbNameJSON, usernameJSON, passwordJSON, permissionJSON)), nil
}

func buildMongodbDeleteScript(dbName string) (string, error) {
	dbNameJSON, err := json.Marshal(dbName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
const dbName = %s;
const targetDb = db.getSiblingDB(dbName);
const dropUsersResult = targetDb.runCommand({ dropAllUsersFromDatabase: 1 });
if (!dropUsersResult || dropUsersResult.ok !== 1) {
  throw new Error("failed to drop users from " + dbName);
}
const dropDatabaseResult = targetDb.runCommand({ dropDatabase: 1 });
if (!dropDatabaseResult || dropDatabaseResult.ok !== 1) {
  throw new Error("failed to drop database " + dbName);
}
`, dbNameJSON)), nil
}

func buildMongodbBindUserScript(dbName, username, password string) (string, error) {
	dbNameJSON, err := json.Marshal(dbName)
	if err != nil {
		return "", err
	}
	usernameJSON, err := json.Marshal(username)
	if err != nil {
		return "", err
	}
	passwordJSON, err := json.Marshal(password)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
const dbName = %s;
const userName = %s;
const password = %s;
const targetDb = db.getSiblingDB(dbName);
const userInfo = targetDb.runCommand({
  usersInfo: userName,
  showCredentials: false,
  showCustomData: false
});
if (!userInfo || userInfo.ok !== 1) {
  throw new Error("failed to load mongodb user " + userName);
}
const roles = [{ role: "readWrite", db: dbName }];
if (Array.isArray(userInfo.users) && userInfo.users.length > 0) {
  const result = targetDb.runCommand({
    updateUser: userName,
    pwd: password,
    roles: roles
  });
  if (!result || result.ok !== 1) {
    throw new Error("failed to update mongodb user " + userName);
  }
} else {
  const result = targetDb.runCommand({
    createUser: userName,
    pwd: password,
    roles: roles
  });
  if (!result || result.ok !== 1) {
    throw new Error("failed to create mongodb user " + userName);
  }
}
`, dbNameJSON, usernameJSON, passwordJSON)), nil
}

func buildMongodbPasswordScript(dbName, username, password string) (string, error) {
	dbNameJSON, err := json.Marshal(dbName)
	if err != nil {
		return "", err
	}
	usernameJSON, err := json.Marshal(username)
	if err != nil {
		return "", err
	}
	passwordJSON, err := json.Marshal(password)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(fmt.Sprintf(`
const dbName = %s;
const userName = %s;
const password = %s;
const targetDb = db.getSiblingDB(dbName);
const result = targetDb.runCommand({
  updateUser: userName,
  pwd: password
});
if (!result || result.ok !== 1) {
  throw new Error("failed to update mongodb user password " + userName);
}
`, dbNameJSON, usernameJSON, passwordJSON)), nil
}

type mongodbSyncItem struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

func loadMongodbDatabases(req dto.MongodbLoadDB) ([]mongodbSyncItem, error) {
	if req.From == constant.AppResourceRemote {
		return loadRemoteMongodbDatabases(req.Database)
	}
	return loadLocalMongodbDatabases(req.Database)
}

func loadLocalMongodbDatabases(database string) ([]mongodbSyncItem, error) {
	script := strings.TrimSpace(`
const systemDbs = new Set(["admin", "config", "local"]);
const result = db.adminCommand({ listDatabases: 1, nameOnly: false });
if (!result || result.ok !== 1) {
  throw new Error("failed to list mongodb databases");
}
const items = result.databases
  .filter(item => !systemDbs.has(item.name))
  .map(item => ({ name: item.name, username: "" }));
print("__1panel_json_begin__");
print(JSON.stringify(items));
print("__1panel_json_end__");
`)
	stdout, err := runMongodbAdminScriptWithStdout(database, script)
	if err != nil {
		return nil, err
	}
	items := make([]mongodbSyncItem, 0)
	jsonResult, err := extractMongodbJSONOutput(stdout)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonResult, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func runRemoteMongodbCreate(database, dbName, username, password, permission string) error {
	info, err := loadRemoteMongodbConnection(database)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	targetDB := client.Database(dbName)
	if err := targetDB.CreateCollection(ctx, "_init"); err != nil {
		return err
	}
	return targetDB.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: username},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{
			bson.D{{Key: "role", Value: permission}, {Key: "db", Value: dbName}},
		}},
	}).Err()
}

func runRemoteMongodbDelete(database, dbName string) error {
	info, err := loadRemoteMongodbConnection(database)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	targetDB := client.Database(dbName)
	if err := targetDB.RunCommand(ctx, bson.D{{Key: "dropAllUsersFromDatabase", Value: 1}}).Err(); err != nil {
		return err
	}
	return targetDB.RunCommand(ctx, bson.D{{Key: "dropDatabase", Value: 1}}).Err()
}

func bindRemoteMongodbUser(connectionName, dbName, username, password string) error {
	info, err := loadRemoteMongodbConnection(connectionName)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	targetDB := client.Database(dbName)
	var userInfo struct {
		Users []struct{} `bson:"users"`
	}
	if err := targetDB.RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: username},
		{Key: "showCredentials", Value: false},
		{Key: "showCustomData", Value: false},
	}).Decode(&userInfo); err != nil {
		return err
	}
	if len(userInfo.Users) > 0 {
		return targetDB.RunCommand(ctx, bson.D{
			{Key: "updateUser", Value: username},
			{Key: "pwd", Value: password},
			{Key: "roles", Value: bson.A{
				bson.D{{Key: "role", Value: "readWrite"}, {Key: "db", Value: dbName}},
			}},
		}).Err()
	}
	return targetDB.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: username},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{
			bson.D{{Key: "role", Value: "readWrite"}, {Key: "db", Value: dbName}},
		}},
	}).Err()
}

func updateRemoteMongodbPasswordOnly(connectionName, dbName, username, password string) error {
	info, err := loadRemoteMongodbConnection(connectionName)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	targetDB := client.Database(dbName)
	return targetDB.RunCommand(ctx, bson.D{
		{Key: "updateUser", Value: username},
		{Key: "pwd", Value: password},
	}).Err()
}

func loadRemoteMongodbDatabases(database string) ([]mongodbSyncItem, error) {
	info, err := loadRemoteMongodbConnection(database)
	if err != nil {
		return nil, err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return nil, err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	adminDB := client.Database("admin")
	var listResult struct {
		Databases []struct {
			Name string `bson:"name"`
		} `bson:"databases"`
	}
	if err := adminDB.RunCommand(ctx, bson.D{
		{Key: "listDatabases", Value: 1},
		{Key: "nameOnly", Value: false},
	}).Decode(&listResult); err != nil {
		return nil, err
	}

	systemDbs := map[string]struct{}{
		"admin":  {},
		"config": {},
		"local":  {},
	}
	items := make([]mongodbSyncItem, 0, len(listResult.Databases))
	for _, databaseItem := range listResult.Databases {
		if _, ok := systemDbs[databaseItem.Name]; ok {
			continue
		}
		items = append(items, mongodbSyncItem{
			Name:     databaseItem.Name,
			Username: "",
		})
	}
	return items, nil
}

func loadMongodbPrivilege(from, connectionName, dbName, username string) (string, error) {
	if from == constant.AppResourceRemote {
		return loadRemoteMongodbPrivilege(connectionName, dbName, username)
	}
	return loadLocalMongodbPrivilege(connectionName, dbName, username)
}

func updateMongodbPrivilege(from, connectionName, dbName, username, permission string) error {
	if from == constant.AppResourceRemote {
		return updateRemoteMongodbPrivilege(connectionName, dbName, username, permission)
	}
	return updateLocalMongodbPrivilege(connectionName, dbName, username, permission)
}

func loadLocalMongodbPrivilege(connectionName, dbName, username string) (string, error) {
	databaseJSON, err := json.Marshal(dbName)
	if err != nil {
		return "", err
	}
	usernameJSON, err := json.Marshal(username)
	if err != nil {
		return "", err
	}
	script := strings.TrimSpace(fmt.Sprintf(`
const dbName = %s;
const userName = %s;
const result = db.getSiblingDB(dbName).runCommand({
  usersInfo: userName,
  showCredentials: false,
  showCustomData: false
});
if (!result || result.ok !== 1) {
  throw new Error("failed to load mongodb user privileges");
}
const roles = Array.isArray(result.users) && result.users.length > 0 ? result.users[0].roles || [] : [];
const permissions = roles.filter(role => role.db === dbName).map(role => role.role);
print("__1panel_json_begin__");
print(JSON.stringify(permissions));
print("__1panel_json_end__");
`, databaseJSON, usernameJSON))
	stdout, err := runMongodbAdminScriptWithStdout(connectionName, script)
	if err != nil {
		return "", err
	}
	var permissions []string
	jsonResult, err := extractMongodbJSONOutput(stdout)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(jsonResult, &permissions); err != nil {
		return "", err
	}
	return normalizeMongodbPrivilege(permissions), nil
}

func updateLocalMongodbPrivilege(connectionName, dbName, username, permission string) error {
	databaseJSON, err := json.Marshal(dbName)
	if err != nil {
		return err
	}
	usernameJSON, err := json.Marshal(username)
	if err != nil {
		return err
	}
	permissionJSON, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	script := strings.TrimSpace(fmt.Sprintf(`
const dbName = %s;
const userName = %s;
const permission = %s;
const targetDb = db.getSiblingDB(dbName);
const userInfo = targetDb.runCommand({
  usersInfo: userName,
  showCredentials: false,
  showCustomData: false
});
if (!userInfo || userInfo.ok !== 1) {
  throw new Error("failed to load mongodb user privileges");
}
if (!Array.isArray(userInfo.users) || userInfo.users.length === 0) {
  throw new Error("mongodb user not found: " + userName);
}
const roles = (userInfo.users[0].roles || []).filter(role => role.db !== dbName);
roles.push({ role: permission, db: dbName });
const result = targetDb.runCommand({
  updateUser: userName,
  roles: roles
});
if (!result || result.ok !== 1) {
  throw new Error("failed to update mongodb user privileges");
}
`, databaseJSON, usernameJSON, permissionJSON))
	return runMongodbAdminScript(connectionName, script)
}

func loadRemoteMongodbPrivilege(connectionName, dbName, username string) (string, error) {
	info, err := loadRemoteMongodbConnection(connectionName)
	if err != nil {
		return "", err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return "", err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	targetDB := client.Database(dbName)
	var result struct {
		Users []struct {
			Roles []struct {
				Role string `bson:"role"`
				DB   string `bson:"db"`
			} `bson:"roles"`
		} `bson:"users"`
	}
	if err := targetDB.RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: username},
		{Key: "showCredentials", Value: false},
		{Key: "showCustomData", Value: false},
	}).Decode(&result); err != nil {
		return "", err
	}

	permissions := make([]string, 0)
	if len(result.Users) > 0 {
		for _, role := range result.Users[0].Roles {
			if role.DB == dbName {
				permissions = append(permissions, role.Role)
			}
		}
	}
	return normalizeMongodbPrivilege(permissions), nil
}

func updateRemoteMongodbPrivilege(connectionName, dbName, username, permission string) error {
	info, err := loadRemoteMongodbConnection(connectionName)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := newRemoteMongodbClient(info)
	if err != nil {
		return err
	}
	defer cancel()
	defer client.Disconnect(ctx)

	targetDB := client.Database(dbName)
	var userInfo struct {
		Users []struct {
			Roles []struct {
				Role string `bson:"role"`
				DB   string `bson:"db"`
			} `bson:"roles"`
		} `bson:"users"`
	}
	if err := targetDB.RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: username},
		{Key: "showCredentials", Value: false},
		{Key: "showCustomData", Value: false},
	}).Decode(&userInfo); err != nil {
		return err
	}
	if len(userInfo.Users) == 0 {
		return fmt.Errorf("mongodb user %s not found in database %s", username, dbName)
	}
	roles := make(bson.A, 0, len(userInfo.Users[0].Roles)+1)
	for _, role := range userInfo.Users[0].Roles {
		if role.DB == dbName {
			continue
		}
		roles = append(roles, bson.D{{Key: "role", Value: role.Role}, {Key: "db", Value: role.DB}})
	}
	roles = append(roles, bson.D{{Key: "role", Value: permission}, {Key: "db", Value: dbName}})
	return targetDB.RunCommand(ctx, bson.D{
		{Key: "updateUser", Value: username},
		{Key: "roles", Value: roles},
	}).Err()
}

func normalizeMongodbPrivilege(permissions []string) string {
	roleMap := map[string]struct{}{
		"dbOwner":   {},
		"read":      {},
		"readWrite": {},
		"userAdmin": {},
	}
	for _, permission := range permissions {
		if _, ok := roleMap[permission]; ok {
			return permission
		}
	}
	return ""
}

func isSupportedMongodbPrivilege(permission string) bool {
	switch permission {
	case "dbOwner", "read", "readWrite", "userAdmin":
		return true
	default:
		return false
	}
}

func extractMongodbJSONOutput(stdout string) ([]byte, error) {
	const (
		beginMark = "__1panel_json_begin__"
		endMark   = "__1panel_json_end__"
	)

	output := strings.TrimSpace(stdout)
	if output == "" {
		return nil, fmt.Errorf("empty mongodb command output")
	}

	beginIndex := strings.Index(output, beginMark)
	endIndex := strings.LastIndex(output, endMark)
	if beginIndex >= 0 && endIndex > beginIndex {
		content := strings.TrimSpace(output[beginIndex+len(beginMark) : endIndex])
		if content != "" {
			return []byte(content), nil
		}
	}

	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		return []byte(line), nil
	}
	return nil, fmt.Errorf("mongodb command output does not contain json result")
}
