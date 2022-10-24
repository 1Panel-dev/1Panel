package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Create(mysqlDto dto.MysqlDBCreate) error
	ChangeInfo(info dto.ChangeDBInfo) error
	Delete(ids []uint) error
	LoadStatus(version string) (*dto.MysqlStatus, error)
	LoadVariables(version string) (*dto.MysqlVariables, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func newDatabaseClient() (*sql.DB, error) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "localhost", 2306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func handleSql(db *sql.DB, query string) (map[string]string, error) {
	variableMap := make(map[string]string)
	rows, err := db.Query(query)
	if err != nil {
		return variableMap, err
	}

	for rows.Next() {
		var variableName, variableValue string
		if err := rows.Scan(&variableName, &variableValue); err != nil {
			return variableMap, err
		}
		variableMap[variableName] = variableValue
	}
	return variableMap, err
}

func (u *MysqlService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Info))
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
	sql, err := newDatabaseClient()
	if err != nil {
		return err
	}
	defer sql.Close()
	if _, err := sql.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET=%s", mysqlDto.Name, mysqlDto.Format)); err != nil {
		return err
	}
	tmpPermission := mysqlDto.Permission
	if _, err := sql.Exec(fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s';", mysqlDto.Name, tmpPermission, mysqlDto.Password)); err != nil {
		return err
	}
	grantStr := fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%s'", mysqlDto.Name, mysqlDto.Username, tmpPermission)
	if mysqlDto.Version == "5.7.39" {
		grantStr = fmt.Sprintf("%s IDENTIFIED BY '%s' WITH GRANT OPTION;", grantStr, mysqlDto.Password)
	}
	if _, err := sql.Exec(grantStr); err != nil {
		return err
	}
	if err := mysqlRepo.Create(&mysql); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) Delete(ids []uint) error {
	dbClient, err := newDatabaseClient()
	if err != nil {
		return err
	}
	defer dbClient.Close()
	dbs, err := mysqlRepo.List(commonRepo.WithIdsIn(ids))
	if err != nil {
		return err
	}

	for _, db := range dbs {
		if len(db.Name) != 0 {
			if _, err := dbClient.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'@'%s'", db.Name, db.Permission)); err != nil {
				return err
			}
			if _, err := dbClient.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", db.Name)); err != nil {
				return err
			}
		}
		_ = mysqlRepo.Delete(commonRepo.WithByID(db.ID))
	}
	return nil
}

func (u *MysqlService) ChangeInfo(info dto.ChangeDBInfo) error {
	mysql, err := mysqlRepo.Get(commonRepo.WithByID(info.ID))
	if err != nil {
		return err
	}
	db, err := newDatabaseClient()
	if err != nil {
		return err
	}
	defer db.Close()
	if info.Operation == "password" {
		if _, err := db.Exec(fmt.Sprintf("SET PASSWORD FOR %s@%s = password('%s')", mysql.Username, mysql.Permission, info.Value)); err != nil {
			return err
		}
		_ = mysqlRepo.Update(mysql.ID, map[string]interface{}{"password": info.Value})
		return nil
	}

	if _, err := db.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'@'%s'", mysql.Name, mysql.Permission)); err != nil {
		return err
	}
	grantStr := fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%s'", mysql.Name, mysql.Username, info.Value)
	if mysql.Version == "5.7.39" {
		grantStr = fmt.Sprintf("%s IDENTIFIED BY '%s' WITH GRANT OPTION;", grantStr, mysql.Password)
	}
	if _, err := db.Exec(grantStr); err != nil {
		return err
	}
	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		return err
	}
	_ = mysqlRepo.Update(mysql.ID, map[string]interface{}{"permission": info.Value})

	return nil
}

func (u *MysqlService) LoadVariables(version string) (*dto.MysqlVariables, error) {
	db, err := newDatabaseClient()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	variableMap, err := handleSql(db, "SHOW	VARIABLES")
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
	db, err := newDatabaseClient()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	globalMap, err := handleSql(db, "SHOW GLOBAL STATUS")
	if err != nil {
		return nil, err
	}
	var info dto.MysqlStatus
	arr, err := json.Marshal(globalMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)

	if value, ok := globalMap["Run"]; ok {
		uptime, _ := strconv.Atoi(value)
		info.Run = time.Unix(time.Now().Unix()-int64(uptime), 0).Format("2006-01-02 15:04:05")
	} else {
		if value, ok := globalMap["Uptime"]; ok {
			uptime, _ := strconv.Atoi(value)
			info.Run = time.Unix(time.Now().Unix()-int64(uptime), 0).Format("2006-01-02 15:04:05")
		}
	}

	rows, err := db.Query("SHOW MASTER STATUS")
	if err != nil {
		return &info, err
	}
	masterRows := make([]string, 5)
	for rows.Next() {
		if err := rows.Scan(&masterRows[0], &masterRows[1], &masterRows[2], &masterRows[3], &masterRows[4]); err != nil {
			return &info, err
		}
	}
	info.File = masterRows[0]
	if len(masterRows[0]) == 0 {
		info.File = "OFF"
	}
	info.Position = masterRows[1]
	if len(masterRows[1]) == 0 {
		info.Position = "OFF"
	}

	return &info, nil
}
