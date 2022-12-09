package service

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.PageInfo) (int64, interface{}, error)
	ListDBName() ([]string, error)
	Create(mysqlDto dto.MysqlDBCreate) error
	ChangeAccess(info dto.ChangeDBInfo) error
	ChangePassword(info dto.ChangeDBInfo) error
	UpdateVariables(updatas []dto.MysqlVariablesUpdate) error
	UpdateConfByFile(info dto.MysqlConfUpdateByFile) error

	RecoverByUpload(req dto.UploadRecover) error
	Backup(db dto.BackupDB) error
	Recover(db dto.RecoverDB) error

	DeleteCheck(id uint) ([]string, error)
	Delete(id uint) error
	LoadStatus() (*dto.MysqlStatus, error)
	LoadVariables() (*dto.MysqlVariables, error)
	LoadBaseInfo() (*dto.DBBaseInfo, error)
	LoadRemoteAccess() (bool, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func (u *MysqlService) SearchWithPage(search dto.PageInfo) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize)
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

func (u *MysqlService) RecoverByUpload(req dto.UploadRecover) error {
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	file := req.FileDir + "/" + req.FileName
	if !strings.HasSuffix(req.FileName, ".sql") && !strings.HasSuffix(req.FileName, ".gz") {
		fileOp := files.NewFileOp()
		fileNameItem := time.Now().Format("20060102150405")
		dstDir := fmt.Sprintf("%s/database/mysql/%s/upload/tmp/%s", localDir, req.MysqlName, fileNameItem)
		if _, err := os.Stat(dstDir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
				if err != nil {
					return fmt.Errorf("mkdir %s failed, err: %v", dstDir, err)
				}
			}
		}
		var compressType files.CompressType
		switch {
		case strings.HasSuffix(req.FileName, ".tar.gz"), strings.HasSuffix(req.FileName, ".tgz"):
			compressType = files.TarGz
		case strings.HasSuffix(req.FileName, ".zip"):
			compressType = files.Zip
		}
		if err := fileOp.Decompress(req.FileDir+"/"+req.FileName, dstDir, compressType); err != nil {
			_ = os.RemoveAll(dstDir)
			return err
		}
		hasTestSql := false
		_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "test.sql" {
				hasTestSql = true
				file = path
			}
			return nil
		})
		if !hasTestSql {
			_ = os.RemoveAll(dstDir)
			return fmt.Errorf("no such file named test.sql in %s, err: %v", req.FileName, err)
		}
		defer func() {
			_ = os.RemoveAll(dstDir)
		}()
	}

	fi, _ := os.Open(file)
	defer fi.Close()
	cmd := exec.Command("docker", "exec", "-i", app.ContainerName, "mysql", "-uroot", "-p"+app.Password, req.DBName)
	if strings.HasSuffix(req.FileName, ".gz") {
		gzipFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer gzipFile.Close()
		gzipReader, err := gzip.NewReader(gzipFile)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		cmd.Stdin = gzipReader
	} else {
		cmd.Stdin = fi
	}
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
	}
	return nil
}

func (u *MysqlService) ListDBName() ([]string, error) {
	mysqls, err := mysqlRepo.List()
	var dbNames []string
	for _, mysql := range mysqls {
		dbNames = append(dbNames, mysql.Name)
	}
	return dbNames, err
}

func (u *MysqlService) Create(mysqlDto dto.MysqlDBCreate) error {
	if mysqlDto.Username == "root" {
		return errors.New("Cannot set root as user name")
	}
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}
	mysql, _ := mysqlRepo.Get(commonRepo.WithByName(mysqlDto.Name))
	if mysql.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&mysql, &mysqlDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("create database if not exists %s character set=%s", mysqlDto.Name, mysqlDto.Format)); err != nil {
		return err
	}
	tmpPermission := mysqlDto.Permission
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("create user if not exists '%s'@'%s' identified by '%s';", mysqlDto.Username, tmpPermission, mysqlDto.Password)); err != nil {
		return err
	}
	grantStr := fmt.Sprintf("grant all privileges on %s.* to '%s'@'%s'", mysqlDto.Name, mysqlDto.Username, tmpPermission)
	if app.Version == "5.7.39" {
		grantStr = fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, mysqlDto.Password)
	}
	if err := excuteSql(app.ContainerName, app.Password, grantStr); err != nil {
		return err
	}
	mysql.MysqlName = app.Name
	if err := mysqlRepo.Create(context.TODO(), &mysql); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) Backup(db dto.BackupDB) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	backupDir := fmt.Sprintf("database/mysql/%s/%s", db.MysqlName, db.DBName)
	fileName := fmt.Sprintf("%s_%s.sql.gz", db.DBName, time.Now().Format("20060102150405"))
	if err := backupMysql("LOCAL", localDir, backupDir, db.MysqlName, db.DBName, fileName); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) Recover(db dto.RecoverDB) error {
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}
	gzipFile, err := os.Open(db.BackupName)
	if err != nil {
		return err
	}
	defer gzipFile.Close()
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	cmd := exec.Command("docker", "exec", "-i", app.ContainerName, "mysql", "-uroot", "-p"+app.Password, db.DBName)
	cmd.Stdin = gzipReader
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
	}
	return nil
}

func (u *MysqlService) DeleteCheck(id uint) ([]string, error) {
	var appInUsed []string
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return appInUsed, err
	}

	db, err := mysqlRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return appInUsed, err
	}

	apps, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithResourceId(app.ID), appInstallResourceRepo.WithLinkId(db.ID))
	for _, app := range apps {
		appInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(app.AppInstallId))
		if appInstall.ID != 0 {
			appInUsed = append(appInUsed, appInstall.Name)
		}
	}
	return appInUsed, nil
}

func (u *MysqlService) Delete(id uint) error {
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}

	db, err := mysqlRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return err
	}

	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop user if exists '%s'@'%s'", db.Name, db.Permission)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop database if exists %s", db.Name)); err != nil {
		return err
	}

	uploadDir := fmt.Sprintf("%s/uploads/%s/mysql/%s", constant.DefaultDataDir, app.Name, db.Name)
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}

	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	backupDir := fmt.Sprintf("%s/database/mysql/%s/%s", localDir, db.MysqlName, db.Name)
	if _, err := os.Stat(backupDir); err == nil {
		_ = os.RemoveAll(backupDir)
	}
	_ = backupRepo.DeleteRecord(commonRepo.WithByType("database-mysql"), commonRepo.WithByName(app.Name), backupRepo.WithByDetailName(db.Name))

	_ = mysqlRepo.Delete(context.Background(), commonRepo.WithByID(db.ID))
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
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}

	passwordChangeCMD := fmt.Sprintf("set password for '%s'@'%s' = password('%s')", mysql.Username, mysql.Permission, info.Value)
	if app.Version != "5.7.39" {
		passwordChangeCMD = fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED WITH mysql_native_password BY '%s';", mysql.Username, mysql.Permission, info.Value)
	}
	if info.ID != 0 {
		if err := excuteSql(app.ContainerName, app.Password, passwordChangeCMD); err != nil {
			return err
		}
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
			if app.Version != "5.7.39" {
				passwordRootChangeCMD = fmt.Sprintf("ALTER USER 'root'@'%s' IDENTIFIED WITH mysql_native_password BY '%s';", host, info.Value)
			}
			if err := excuteSql(app.ContainerName, app.Password, passwordRootChangeCMD); err != nil {
				return err
			}
		}
	}
	updateInstallInfoInDB("mysql", "password", info.Value)
	updateInstallInfoInDB("phpmyadmin", "password", info.Value)
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
	}
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
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
		if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop user if exists '%s'@'%s'", mysql.Username, mysql.Permission)); err != nil {
			return err
		}
		if info.ID == 0 {
			return nil
		}
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("create user if not exists '%s'@'%s' identified by '%s';", mysql.Username, info.Value, mysql.Password)); err != nil {
		return err
	}
	grantStr := fmt.Sprintf("grant all privileges on %s.* to '%s'@'%s'", mysql.Name, mysql.Username, info.Value)
	if app.Version == "5.7.39" {
		grantStr = fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, mysql.Password)
	}
	if err := excuteSql(app.ContainerName, app.Password, grantStr); err != nil {
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
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
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

func (u *MysqlService) UpdateVariables(updatas []dto.MysqlVariablesUpdate) error {
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}
	var files []string

	path := fmt.Sprintf("%s/mysql/%s/conf/my.cnf", constant.AppInstallDir, app.Name)
	lineBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		files = strings.Split(string(lineBytes), "\n")
	}
	group := "[mysqld]"
	for _, info := range updatas {
		if app.Version != "5.7.39" {
			if info.Param == "query_cache_size" {
				continue
			}
		}

		files = updateMyCnf(files, group, info.Param, loadSizeUnit(info.Value))
	}
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
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
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return nil, err
	}
	data.ContainerName = app.ContainerName
	data.Name = app.Name
	data.Port = int64(app.Port)

	return &data, nil
}

func (u *MysqlService) LoadRemoteAccess() (bool, error) {
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
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
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
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
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
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

func backupMysql(backupType, baseDir, backupDir, mysqlName, dbName, fileName string) error {
	app, err := appInstallRepo.LoadBaseInfoByKey("mysql")
	if err != nil {
		return err
	}

	fullDir := baseDir + "/" + backupDir
	if _, err := os.Stat(fullDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(fullDir, os.ModePerm); err != nil {
			if err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", fullDir, err)
			}
		}
	}
	outfile, _ := os.OpenFile(fullDir+"/"+fileName, os.O_RDWR|os.O_CREATE, 0755)
	cmd := exec.Command("docker", "exec", app.ContainerName, "mysqldump", "-uroot", "-p"+app.Password, dbName)
	gzipCmd := exec.Command("gzip", "-cf")
	gzipCmd.Stdin, _ = cmd.StdoutPipe()
	gzipCmd.Stdout = outfile
	_ = gzipCmd.Start()
	_ = cmd.Run()
	_ = gzipCmd.Wait()

	record := &model.BackupRecord{
		Type:       "database-mysql",
		Name:       app.Name,
		DetailName: dbName,
		Source:     backupType,
		BackupType: backupType,
		FileDir:    backupDir,
		FileName:   fileName,
	}
	if baseDir != constant.TmpDir || backupType == "LOCAL" {
		record.Source = "LOCAL"
		record.FileDir = fullDir
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
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
		if strings.HasPrefix(line, param) || strings.HasPrefix(line, "# "+param) {
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

func loadSizeUnit(value int64) string {
	if value > 1048576 {
		return fmt.Sprintf("%dM", value/1048576)
	}
	if value > 1024 {
		return fmt.Sprintf("%dK", value/1024)
	}
	return fmt.Sprintf("%d", value)
}
