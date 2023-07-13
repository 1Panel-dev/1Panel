package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	ListDBName() ([]string, error)
	Create(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error)
	ChangeAccess(info dto.ChangeDBInfo) error
	ChangePassword(info dto.ChangeDBInfo) error
	UpdateVariables(updates []dto.MysqlVariablesUpdate) error
	UpdateConfByFile(info dto.MysqlConfUpdateByFile) error
	UpdateDescription(req dto.UpdateDescription) error
	DeleteCheck(id uint) ([]string, error)
	Delete(ctx context.Context, req dto.MysqlDBDelete) error
	LoadStatus() (*dto.MysqlStatus, error)
	LoadVariables() (*dto.MysqlVariables, error)
	LoadBaseInfo() (*dto.DBBaseInfo, error)
	LoadRemoteAccess() (bool, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func (u *MysqlService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Info), commonRepo.WithOrderRuleBy(search.OrderBy, search.Order))
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

func (u *MysqlService) ListDBName() ([]string, error) {
	mysqls, err := mysqlRepo.List()
	var dbNames []string
	for _, mysql := range mysqls {
		dbNames = append(dbNames, mysql.Name)
	}
	return dbNames, err
}

var formatMap = map[string]string{
	"utf8":    "utf8_general_ci",
	"utf8mb4": "utf8mb4_general_ci",
	"gbk":     "gbk_chinese_ci",
	"big5":    "big5_chinese_ci",
}

func (u *MysqlService) Create(ctx context.Context, req dto.MysqlDBCreate) (*model.DatabaseMysql, error) {
	if req.Username == "root" {
		return nil, errors.New("Cannot set root as user name")
	}
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return nil, err
	}
	mysql, _ := mysqlRepo.Get(commonRepo.WithByName(req.Name))
	if mysql.ID != 0 {
		return nil, constant.ErrRecordExist
	}
	if err := copier.Copy(&mysql, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	createSql := fmt.Sprintf("create database `%s` default character set %s collate %s", req.Name, req.Format, formatMap[req.Format])
	if err := excSQL(app.ContainerName, app.Password, createSql); err != nil {
		if strings.Contains(err.Error(), "ERROR 1007") {
			return nil, buserr.New(constant.ErrDatabaseIsExist)
		}
		return nil, err
	}
	if err := u.createUser(app.ContainerName, app.Password, app.Version, req); err != nil {
		return nil, err
	}

	global.LOG.Infof("create database %s successful!", req.Name)
	mysql.MysqlName = app.Name
	if err := mysqlRepo.Create(ctx, &mysql); err != nil {
		return nil, err
	}
	return &mysql, nil
}

func (u *MysqlService) UpdateDescription(req dto.UpdateDescription) error {
	return mysqlRepo.Update(req.ID, map[string]interface{}{"description": req.Description})
}

func (u *MysqlService) DeleteCheck(id uint) ([]string, error) {
	var appInUsed []string
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return appInUsed, err
	}

	db, err := mysqlRepo.Get(commonRepo.WithByID(id))
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
	return appInUsed, nil
}

func (u *MysqlService) Delete(ctx context.Context, req dto.MysqlDBDelete) error {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil && !req.ForceDelete {
		return err
	}

	db, err := mysqlRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil && !req.ForceDelete {
		return err
	}

	if strings.HasPrefix(app.Version, "5.6") {
		if err := excSQL(app.ContainerName, app.Password, fmt.Sprintf("drop user '%s'@'%s'", db.Username, db.Permission)); err != nil && !req.ForceDelete {
			return err
		}
	} else {
		if err := excSQL(app.ContainerName, app.Password, fmt.Sprintf("drop user if exists '%s'@'%s'", db.Username, db.Permission)); err != nil && !req.ForceDelete {
			return err
		}
	}
	if err := excSQL(app.ContainerName, app.Password, fmt.Sprintf("drop database if exists `%s`", db.Name)); err != nil && !req.ForceDelete {
		return err
	}
	global.LOG.Info("execute delete database sql successful, now start to drop uploads and records")

	uploadDir := fmt.Sprintf("%s/1panel/uploads/database/mysql/%s/%s", global.CONF.System.BaseDir, app.Name, db.Name)
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	if req.DeleteBackup {
		localDir, err := loadLocalDir()
		if err != nil && !req.ForceDelete {
			return err
		}
		backupDir := fmt.Sprintf("%s/database/mysql/%s/%s", localDir, db.MysqlName, db.Name)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		global.LOG.Infof("delete database %s-%s backups successful", app.Name, db.Name)
	}
	_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType("mysql"), commonRepo.WithByName(app.Name), backupRepo.WithByDetailName(db.Name))

	_ = mysqlRepo.Delete(ctx, commonRepo.WithByID(db.ID))
	return nil
}

func (u *MysqlService) ChangePassword(info dto.ChangeDBInfo) error {
	var (
		mysql model.DatabaseMysql
		err   error
	)
	if info.ID != 0 {
		mysql, err = mysqlRepo.Get(commonRepo.WithByID(info.ID))
		if err != nil {
			return err
		}
	}
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}

	passwordChangeCMD := fmt.Sprintf("set password for '%s'@'%s' = password('%s')", mysql.Username, mysql.Permission, info.Value)
	if !strings.HasPrefix(app.Version, "5.7") && !strings.HasPrefix(app.Version, "5.6") {
		passwordChangeCMD = fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED WITH mysql_native_password BY '%s';", mysql.Username, mysql.Permission, info.Value)
	}
	if info.ID != 0 {
		appRess, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(app.ID), appInstallResourceRepo.WithResourceId(mysql.ID))
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
			if err := updateInstallInfoInDB(appModel.Key, appInstall.Name, "user-password", true, info.Value); err != nil {
				return err
			}
		}
		if err := excuteSql(app.ContainerName, app.Password, passwordChangeCMD); err != nil {
			return err
		}
		global.LOG.Info("excute password change sql successful")
		_ = mysqlRepo.Update(mysql.ID, map[string]interface{}{"password": info.Value})
		return nil
	}

	hosts, err := excuteSqlForRows(app.ContainerName, app.Password, "select host from mysql.user where user='root';")
	if err != nil {
		return err
	}
	for _, host := range hosts {
		if host == "%" || host == "localhost" {
			passwordRootChangeCMD := fmt.Sprintf("set password for 'root'@'%s' = password('%s')", host, info.Value)
			if !strings.HasPrefix(app.Version, "5.7") && !strings.HasPrefix(app.Version, "5.6") {
				passwordRootChangeCMD = fmt.Sprintf("alter user 'root'@'%s' identified with mysql_native_password BY '%s';", host, info.Value)
			}
			if err := excuteSql(app.ContainerName, app.Password, passwordRootChangeCMD); err != nil {
				return err
			}
		}
	}
	if err := updateInstallInfoInDB("mysql", "", "password", false, info.Value); err != nil {
		return err
	}
	if err := updateInstallInfoInDB("phpmyadmin", "", "password", true, info.Value); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) ChangeAccess(info dto.ChangeDBInfo) error {
	var (
		mysql model.DatabaseMysql
		err   error
	)
	if info.ID != 0 {
		mysql, err = mysqlRepo.Get(commonRepo.WithByID(info.ID))
		if err != nil {
			return err
		}
		if info.Value == mysql.Permission {
			return nil
		}
	}
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}
	if info.ID == 0 {
		mysql.Name = "*"
		mysql.Username = "root"
		mysql.Permission = "%"
		mysql.Password = app.Password
	}

	if info.Value != mysql.Permission {
		var userlist []string
		if strings.Contains(mysql.Permission, ",") {
			userlist = strings.Split(mysql.Permission, ",")
		} else {
			userlist = append(userlist, mysql.Permission)
		}
		for _, user := range userlist {
			if len(user) != 0 {
				if strings.HasPrefix(app.Version, "5.6") {
					if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop user '%s'@'%s'", mysql.Username, user)); err != nil {
						return err
					}
				} else {
					if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop user if exists '%s'@'%s'", mysql.Username, user)); err != nil {
						return err
					}
				}
			}
		}
		if info.ID == 0 {
			return nil
		}
	}

	if err := u.createUser(app.ContainerName, app.Password, app.Version, dto.MysqlDBCreate{
		Username:   mysql.Username,
		Name:       mysql.Name,
		Permission: info.Value,
		Password:   mysql.Password,
	}); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, "flush privileges"); err != nil {
		return err
	}
	if info.ID == 0 {
		return nil
	}

	_ = mysqlRepo.Update(mysql.ID, map[string]interface{}{"permission": info.Value})

	return nil
}

func (u *MysqlService) UpdateConfByFile(info dto.MysqlConfUpdateByFile) error {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/mysql/%s/conf/my.cnf", constant.AppInstallDir, app.Name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(info.File)
	write.Flush()
	if _, err := compose.Restart(fmt.Sprintf("%s/mysql/%s/docker-compose.yml", constant.AppInstallDir, app.Name)); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) UpdateVariables(updates []dto.MysqlVariablesUpdate) error {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}
	var files []string

	path := fmt.Sprintf("%s/mysql/%s/conf/my.cnf", constant.AppInstallDir, app.Name)
	lineBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	files = strings.Split(string(lineBytes), "\n")

	group := "[mysqld]"
	for _, info := range updates {
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

	if _, err := compose.Restart(fmt.Sprintf("%s/mysql/%s/docker-compose.yml", constant.AppInstallDir, app.Name)); err != nil {
		return err
	}

	return nil
}

func (u *MysqlService) LoadBaseInfo() (*dto.DBBaseInfo, error) {
	var data dto.DBBaseInfo
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return nil, err
	}
	data.ContainerName = app.ContainerName
	data.Name = app.Name
	data.Port = int64(app.Port)

	return &data, nil
}

func (u *MysqlService) LoadRemoteAccess() (bool, error) {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return false, err
	}
	hosts, err := excuteSqlForRows(app.ContainerName, app.Password, "select host from mysql.user where user='root';")
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

func (u *MysqlService) LoadVariables() (*dto.MysqlVariables, error) {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return nil, err
	}
	variableMap, err := excuteSqlForMaps(app.ContainerName, app.Password, "show global variables;")
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

func (u *MysqlService) LoadStatus() (*dto.MysqlStatus, error) {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return nil, err
	}

	statusMap, err := excuteSqlForMaps(app.ContainerName, app.Password, "show global status;")
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
		info.Run = time.Unix(time.Now().Unix()-int64(uptime), 0).Format("2006-01-02 15:04:05")
	} else {
		if value, ok := statusMap["Uptime"]; ok {
			uptime, _ := strconv.Atoi(value)
			info.Run = time.Unix(time.Now().Unix()-int64(uptime), 0).Format("2006-01-02 15:04:05")
		}
	}

	info.File = "OFF"
	info.Position = "OFF"
	rows, err := excuteSqlForRows(app.ContainerName, app.Password, "show master status;")
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

func (u *MysqlService) createUser(container, password, version string, req dto.MysqlDBCreate) error {
	var userlist []string
	if strings.Contains(req.Permission, ",") {
		ips := strings.Split(req.Permission, ",")
		for _, ip := range ips {
			if len(ip) != 0 {
				userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", req.Username, ip))
			}
		}
	} else {
		userlist = append(userlist, fmt.Sprintf("'%s'@'%s'", req.Username, req.Permission))
	}

	for _, user := range userlist {
		if err := excSQL(container, password, fmt.Sprintf("create user %s identified by '%s';", user, req.Password)); err != nil {
			if strings.Contains(err.Error(), "ERROR 1396") {
				handleCreateError(container, password, req.Name, userlist, false)
				return buserr.New(constant.ErrUserIsExist)
			}
			handleCreateError(container, password, req.Name, userlist, true)
			return err
		}
		grantStr := fmt.Sprintf("grant all privileges on `%s`.* to %s", req.Name, user)
		if req.Name == "*" {
			grantStr = fmt.Sprintf("grant all privileges on *.* to %s", user)
		}
		if strings.HasPrefix(version, "5.7") || strings.HasPrefix(version, "5.6") {
			grantStr = fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, req.Password)
		}
		if err := excSQL(container, password, grantStr); err != nil {
			handleCreateError(container, password, req.Name, userlist, true)
			return err
		}
	}
	return nil
}
func handleCreateError(contaienr, password, dbName string, userlist []string, dropUser bool) {
	_ = excSQL(contaienr, password, fmt.Sprintf("drop database `%s`", dbName))
	if dropUser {
		for _, user := range userlist {
			if err := excSQL(contaienr, password, fmt.Sprintf("drop user if exists %s", user)); err != nil {
				global.LOG.Errorf("drop user failed, err: %v", err)
			}
		}
	}
}

func excuteSqlForMaps(containerName, password, command string) (map[string]string, error) {
	cmd := exec.Command("docker", "exec", containerName, "mysql", "-uroot", "-p"+password, "-e", command)
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

func excuteSqlForRows(containerName, password, command string) ([]string, error) {
	cmd := exec.Command("docker", "exec", containerName, "mysql", "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return nil, errors.New(stdStr)
	}
	return strings.Split(stdStr, "\n"), nil
}

func excuteSql(containerName, password, command string) error {
	cmd := exec.Command("docker", "exec", containerName, "mysql", "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
	}
	return nil
}

func excSQL(containerName, password, command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "exec", containerName, "mysql", "-uroot", "-p"+password, "-e", command)
	stdout, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return buserr.WithDetail(constant.ErrExecTimeOut, containerName, nil)
	}
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
	}
	return nil
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
