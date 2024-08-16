package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql/client"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.MysqlDBSearch) (int64, interface{}, error)
	ListDBOption() ([]dto.MysqlOption, error)
	Create(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error)
	BindUser(req dto.BindUser) error
	LoadFromRemote(req dto.MysqlLoadDB) error
	ChangeAccess(info dto.ChangeDBInfo) error
	ChangePassword(info dto.ChangeDBInfo) error
	UpdateVariables(req dto.MysqlVariablesUpdate) error
	UpdateDescription(req dto.UpdateDescription) error
	DeleteCheck(req dto.MysqlDBDeleteCheck) ([]string, error)
	Delete(ctx context.Context, req dto.MysqlDBDelete) error

	LoadStatus(req dto.OperationWithNameAndType) (*dto.MysqlStatus, error)
	LoadVariables(req dto.OperationWithNameAndType) (*dto.MysqlVariables, error)
	LoadRemoteAccess(req dto.OperationWithNameAndType) (bool, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func (u *MysqlService) SearchWithPage(search dto.MysqlDBSearch) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize,
		mysqlRepo.WithByMysqlName(search.Database),
		commonRepo.WithLikeName(search.Info),
		commonRepo.WithOrderRuleBy(search.OrderBy, search.Order),
	)
	var dtoMysqls []dto.MysqlDBInfo
	for _, mysql := range mysqls {
		var item dto.MysqlDBInfo
		if err := copier.Copy(&item, &mysql); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
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
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
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
	if cmd.CheckIllegal(req.Name, req.Username, req.Password, req.Format, req.Permission) {
		return nil, buserr.New(constant.ErrCmdIllegal)
	}

	mysql, _ := mysqlRepo.Get(commonRepo.WithByName(req.Name), mysqlRepo.WithByMysqlName(req.Database), databaseRepo.WithByFrom(req.From))
	if mysql.ID != 0 {
		return nil, constant.ErrRecordExist
	}

	var createItem model.DatabaseMysql
	if err := copier.Copy(&createItem, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	if req.From == "local" && req.Username == "root" {
		return nil, errors.New("Cannot set root as user name")
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
		Username:   req.Username,
		Password:   req.Password,
		Permission: req.Permission,
		Version:    version,
		Timeout:    300,
	}); err != nil {
		return nil, err
	}

	global.LOG.Infof("create database %s successful!", req.Name)
	if err := mysqlRepo.Create(ctx, &createItem); err != nil {
		return nil, err
	}
	return &createItem, nil
}

func (u *MysqlService) BindUser(req dto.BindUser) error {
	if cmd.CheckIllegal(req.Username, req.Password, req.Permission) {
		return buserr.New(constant.ErrCmdIllegal)
	}

	dbItem, err := mysqlRepo.Get(mysqlRepo.WithByMysqlName(req.Database), commonRepo.WithByName(req.DB))
	if err != nil {
		return err
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.CreateUser(client.CreateInfo{
		Name:       dbItem.Name,
		Format:     dbItem.Format,
		Username:   req.Username,
		Password:   req.Password,
		Permission: req.Permission,
		Version:    version,
		Timeout:    300,
	}, false); err != nil {
		return err
	}
	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return fmt.Errorf("decrypt database db password failed, err: %v", err)
	}
	if err := mysqlRepo.Update(dbItem.ID, map[string]interface{}{
		"username":   req.Username,
		"password":   pass,
		"permission": req.Permission,
	}); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) LoadFromRemote(req dto.MysqlLoadDB) error {
	client, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}

	databases, err := mysqlRepo.List(mysqlRepo.WithByMysqlName(req.Database))
	if err != nil {
		return err
	}
	datas, err := client.SyncDB(version)
	if err != nil {
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
				return errors.WithMessage(constant.ErrStructTransform, err.Error())
			}
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

func (u *MysqlService) DeleteCheck(req dto.MysqlDBDeleteCheck) ([]string, error) {
	var appInUsed []string
	db, err := mysqlRepo.Get(commonRepo.WithByID(req.ID))
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

func (u *MysqlService) Delete(ctx context.Context, req dto.MysqlDBDelete) error {
	db, err := mysqlRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil && !req.ForceDelete {
		return err
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Delete(client.DeleteInfo{
		Name:       db.Name,
		Version:    version,
		Username:   db.Username,
		Permission: db.Permission,
		Timeout:    300,
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
		backupDir := path.Join(localDir, fmt.Sprintf("database/%s/%s/%s", req.Type, db.MysqlName, db.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType(req.Type), commonRepo.WithByName(req.Database), backupRepo.WithByDetailName(db.Name))
		global.LOG.Infof("delete database %s-%s backups successful", req.Database, db.Name)
	}

	_ = mysqlRepo.Delete(ctx, commonRepo.WithByID(db.ID))
	return nil
}

func (u *MysqlService) ChangePassword(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	var (
		mysqlData    model.DatabaseMysql
		passwordInfo client.PasswordChangeInfo
	)
	passwordInfo.Password = req.Value
	passwordInfo.Timeout = 300
	passwordInfo.Version = version

	if req.ID != 0 {
		mysqlData, err = mysqlRepo.Get(commonRepo.WithByID(req.ID))
		if err != nil {
			return err
		}
		passwordInfo.Name = mysqlData.Name
		passwordInfo.Username = mysqlData.Username
		passwordInfo.Permission = mysqlData.Permission
	} else {
		passwordInfo.Username = "root"
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
			appRess, _ = appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(app.ID), appInstallResourceRepo.WithResourceId(mysqlData.ID))
		} else {
			appRess, _ = appInstallResourceRepo.GetBy(appInstallResourceRepo.WithResourceId(mysqlData.ID))
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

			global.LOG.Infof("start to update mysql password used by app %s-%s", appModel.Key, appInstall.Name)
			if err := updateInstallInfoInDB(appModel.Key, appInstall.Name, "user-password", req.Value); err != nil {
				return err
			}
		}
		global.LOG.Info("execute password change sql successful")
		pass, err := encrypt.StringEncrypt(req.Value)
		if err != nil {
			return fmt.Errorf("decrypt database db password failed, err: %v", err)
		}
		_ = mysqlRepo.Update(mysqlData.ID, map[string]interface{}{"password": pass})
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

func (u *MysqlService) ChangeAccess(req dto.ChangeDBInfo) error {
	if cmd.CheckIllegal(req.Value) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	cli, version, err := LoadMysqlClientByFrom(req.Database)
	if err != nil {
		return err
	}
	defer cli.Close()
	var (
		mysqlData  model.DatabaseMysql
		accessInfo client.AccessChangeInfo
	)
	accessInfo.Permission = req.Value
	accessInfo.Timeout = 300
	accessInfo.Version = version

	if req.ID != 0 {
		mysqlData, err = mysqlRepo.Get(commonRepo.WithByID(req.ID))
		if err != nil {
			return err
		}
		accessInfo.Name = mysqlData.Name
		accessInfo.Username = mysqlData.Username
		accessInfo.Password = mysqlData.Password
		accessInfo.OldPermission = mysqlData.Permission
	} else {
		accessInfo.Username = "root"
	}
	if err := cli.ChangeAccess(accessInfo); err != nil {
		return err
	}

	if mysqlData.ID != 0 {
		_ = mysqlRepo.Update(mysqlData.ID, map[string]interface{}{"permission": req.Value})
	}

	return nil
}

func (u *MysqlService) UpdateVariables(req dto.MysqlVariablesUpdate) error {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
	if err != nil {
		return err
	}
	var files []string

	path := fmt.Sprintf("%s/%s/%s/conf/my.cnf", constant.AppInstallDir, req.Type, app.Name)
	lineBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	files = strings.Split(string(lineBytes), "\n")

	group := "[mysqld]"
	for _, info := range req.Variables {
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
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(files, "\n"))
	if err != nil {
		return err
	}

	if _, err := compose.Restart(fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, req.Type, app.Name)); err != nil {
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
	rows, err := executeSqlForRows(app.ContainerName, app.Key, app.Password, "show master status;")
	if err != nil {
		rows, err = executeSqlForRows(app.ContainerName, app.Key, app.Password, "SHOW BINARY LOG STATUS;")
		if err != nil {
			return nil, err
		}
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

func executeSqlForMaps(containerName, dbType, password, command string) (map[string]string, error) {
	cmd := exec.Command("docker", "exec", containerName, dbType, "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
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
	cmd := exec.Command("docker", "exec", containerName, dbType, "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return nil, errors.New(stdStr)
	}
	return strings.Split(stdStr, "\n"), nil
}

func updateMyCnf(oldFiles []string, group string, param string, value interface{}) []string {
	isOn := false
	hasGroup := false
	hasKey := false
	regItem, _ := regexp.Compile(`\[*\]`)
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
	databaseItem, err := databaseRepo.Get(commonRepo.WithByName(database))
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
