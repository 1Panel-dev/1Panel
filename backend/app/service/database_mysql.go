package service

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.SearchDBWithPage) (int64, interface{}, error)
	SearchBacpupsWithPage(search dto.SearchBackupsWithPage) (int64, interface{}, error)
	Create(mysqlDto dto.MysqlDBCreate) error
	ChangeInfo(info dto.ChangeDBInfo) error
	UpdateVariables(variables dto.MysqlVariablesUpdate) error

	Backup(db dto.BackupDB) error
	Recover(db dto.RecoverDB) error

	Delete(version string, ids []uint) error
	LoadStatus(version string) (*dto.MysqlStatus, error)
	LoadVariables(version string) (*dto.MysqlVariables, error)
	LoadRunningVersion() ([]string, error)
	LoadBaseInfo(version string) (*dto.DBBaseInfo, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func (u *MysqlService) SearchWithPage(search dto.SearchDBWithPage) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize, mysqlRepo.WithByVersion(search.Version))
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

func (u *MysqlService) SearchBacpupsWithPage(search dto.SearchBackupsWithPage) (int64, interface{}, error) {
	app, err := mysqlRepo.LoadBaseInfoByVersion(search.Version)
	if err != nil {
		return 0, nil, err
	}
	searchDto := dto.BackupSearch{
		Type:       "database-mysql",
		PageInfo:   search.PageInfo,
		Name:       app.Name,
		DetailName: search.DBName,
	}

	return NewIBackupService().SearchRecordWithPage(searchDto)
}

func (u *MysqlService) LoadRunningVersion() ([]string, error) {
	return mysqlRepo.LoadRunningVersion()
}

func (u *MysqlService) Create(mysqlDto dto.MysqlDBCreate) error {
	if mysqlDto.Username == "root" {
		return errors.New("Cannot set root as user name")
	}
	mysql, _ := mysqlRepo.Get(commonRepo.WithByName(mysqlDto.Name))
	if mysql.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&mysql, &mysqlDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	app, err := mysqlRepo.LoadBaseInfoByVersion(mysqlDto.Version)
	if err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("create database if not exists %s character set=%s", mysqlDto.Name, mysqlDto.Format)); err != nil {
		return err
	}
	tmpPermission := mysqlDto.Permission
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("create user if not exists '%s'@'%s' identified by '%s';", mysqlDto.Name, tmpPermission, mysqlDto.Password)); err != nil {
		return err
	}
	grantStr := fmt.Sprintf("grant all privileges on %s.* to '%s'@'%s'", mysqlDto.Name, mysqlDto.Username, tmpPermission)
	if mysqlDto.Version == "5.7.39" {
		grantStr = fmt.Sprintf("%s identified by '%s' with grant option;", grantStr, mysqlDto.Password)
	}
	if err := excuteSql(app.ContainerName, app.Password, grantStr); err != nil {
		return err
	}
	if err := mysqlRepo.Create(&mysql); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) Backup(db dto.BackupDB) error {
	app, err := mysqlRepo.LoadBaseInfoByVersion(db.Version)
	if err != nil {
		return err
	}

	backupDir := fmt.Sprintf("%s/%s/%s/", constant.DatabaseDir, app.Name, db.DBName)
	if _, err := os.Stat(backupDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(backupDir, os.ModePerm); err != nil {
			return err
		}
	}
	backupName := fmt.Sprintf("%s%s_%s.sql.gz", backupDir, db.DBName, time.Now().Format("20060102150405"))
	outfile, _ := os.OpenFile(backupName, os.O_RDWR|os.O_CREATE, 0755)
	cmd := exec.Command("docker", "exec", app.ContainerName, "mysqldump", "-uroot", "-p"+app.Password, db.DBName)
	gzipCmd := exec.Command("gzip", "-cf")
	gzipCmd.Stdin, _ = cmd.StdoutPipe()
	gzipCmd.Stdout = outfile
	_ = gzipCmd.Start()
	_ = cmd.Run()
	_ = gzipCmd.Wait()

	if err := backupRepo.CreateRecord(&model.BackupRecord{
		Type:       "database-mysql",
		Name:       app.Name,
		DetailName: db.DBName,
		Source:     "LOCAL",
		FileDir:    backupDir,
		FileName:   strings.ReplaceAll(backupName, backupDir, ""),
	}); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
	}
	return nil
}

func (u *MysqlService) Recover(db dto.RecoverDB) error {
	app, err := mysqlRepo.LoadBaseInfoByVersion(db.Version)
	if err != nil {
		return err
	}
	gzipFile, err := os.Open(db.BackupName)
	if err != nil {
		fmt.Println(err)
	}
	defer gzipFile.Close()
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		fmt.Println(err)
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

func (u *MysqlService) Delete(version string, ids []uint) error {
	app, err := mysqlRepo.LoadBaseInfoByVersion(version)
	if err != nil {
		return err
	}

	dbs, err := mysqlRepo.List(commonRepo.WithIdsIn(ids))
	if err != nil {
		return err
	}

	for _, db := range dbs {
		if len(db.Name) != 0 {
			if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop user if exists '%s'@'%s'", db.Name, db.Permission)); err != nil {
				return err
			}
			if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("drop database if exists %s", db.Name)); err != nil {
				return err
			}
		}
		_ = mysqlRepo.Delete(commonRepo.WithByID(db.ID))
	}
	return nil
}

func (u *MysqlService) ChangeInfo(info dto.ChangeDBInfo) error {
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
	app, err := mysqlRepo.LoadBaseInfoByVersion(info.Version)
	if err != nil {
		return err
	}
	if info.Operation == "password" {
		if info.ID != 0 {
			if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set password for %s@%s = password('%s')", mysql.Username, mysql.Permission, info.Value)); err != nil {
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
				if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set password for root@'%s' = password('%s')", host, info.Value)); err != nil {
					return err
				}
			}
		}
		_ = mysqlRepo.UpdateMysqlConf(app.ID, map[string]interface{}{
			"param": strings.ReplaceAll(app.Param, app.Password, info.Value),
			"env":   strings.ReplaceAll(app.Env, app.Password, info.Value),
		})
		return nil
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
	if mysql.Version == "5.7.39" {
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

func (u *MysqlService) UpdateVariables(variables dto.MysqlVariablesUpdate) error {
	app, err := mysqlRepo.LoadBaseInfoByVersion(variables.Version)
	if err != nil {
		return err
	}

	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL key_buffer_size=%d", variables.KeyBufferSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL query_cache_size=%d", variables.QueryCacheSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL tmp_table_size=%d", variables.TmpTableSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL innodb_buffer_pool_size=%d", variables.InnodbBufferPoolSize)); err != nil {
		return err
	}
	// if  err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL innodb_log_buffer_size=%d", variables.InnodbLogBufferSize)); err != nil {
	// 	return err
	// }
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL sort_buffer_size=%d", variables.SortBufferSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL read_buffer_size=%d", variables.ReadBufferSize)); err != nil {
		return err
	}

	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL read_rnd_buffer_size=%d", variables.ReadRndBufferSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL join_buffer_size=%d", variables.JoinBufferSize)); err != nil {
		return err
	}
	// if  err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL thread_stack=%d", variables.ThreadStack)); err != nil {
	// 	return err
	// }
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL binlog_cache_size=%d", variables.BinlogCachSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL thread_cache_size=%d", variables.ThreadCacheSize)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL table_open_cache=%d", variables.TableOpenCache)); err != nil {
		return err
	}
	if err := excuteSql(app.ContainerName, app.Password, fmt.Sprintf("set GLOBAL max_connections=%d", variables.MaxConnections)); err != nil {
		return err
	}

	return nil
}

func (u *MysqlService) LoadBaseInfo(version string) (*dto.DBBaseInfo, error) {
	var data dto.DBBaseInfo
	app, err := mysqlRepo.LoadBaseInfoByVersion(version)
	if err != nil {
		return nil, err
	}
	data.Name = app.Name
	data.Port = int64(app.Port)
	data.Password = app.Password

	hosts, err := excuteSqlForRows(app.ContainerName, app.Password, "select host from mysql.user where user='root';")
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		if host == "%" {
			data.RemoteConn = true
			break
		}
	}
	return &data, nil
}

func (u *MysqlService) LoadVariables(version string) (*dto.MysqlVariables, error) {
	app, err := mysqlRepo.LoadBaseInfoByVersion(version)
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

func (u *MysqlService) LoadStatus(version string) (*dto.MysqlStatus, error) {
	app, err := mysqlRepo.LoadBaseInfoByVersion(version)
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
