package service

import (
	"database/sql"
	"encoding/json"
	"fmt"

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
	Delete(ids []uint) error
	LoadStatus(version string) (*dto.MysqlStatus, error)
	LoadConf(version string) (*dto.MysqlConf, error)
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
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
	mysql, _ := mysqlRepo.Get(commonRepo.WithByName(mysqlDto.Name))
	if mysql.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&mysql, &mysqlDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := mysqlRepo.Create(&mysql); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) Delete(ids []uint) error {
	if len(ids) == 1 {
		mysql, _ := mysqlRepo.Get(commonRepo.WithByID(ids[0]))
		if mysql.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return mysqlRepo.Delete(commonRepo.WithByID(ids[0]))
	}
	return mysqlRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *MysqlService) LoadConf(version string) (*dto.MysqlConf, error) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "localhost", 2306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("show variables")
	if err != nil {
		return nil, err
	}
	variableMap := make(map[string]string)
	for rows.Next() {
		var variableName, variableValue string
		if err := rows.Scan(&variableName, &variableValue); err != nil {
			continue
		}
		variableMap[variableName] = variableValue
	}

	var info dto.MysqlConf
	arr, err := json.Marshal(variableMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)
	return &info, nil
}

func (u *MysqlService) LoadStatus(version string) (*dto.MysqlStatus, error) {
	connArgs := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8", "root", "Calong@2015", "localhost", 2306)
	db, err := sql.Open("mysql", connArgs)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("show status")
	if err != nil {
		return nil, err
	}
	variableMap := make(map[string]string)
	for rows.Next() {
		var variableName, variableValue string
		if err := rows.Scan(&variableName, &variableValue); err != nil {
			continue
		}
		variableMap[variableName] = variableValue
	}

	var info dto.MysqlStatus
	arr, err := json.Marshal(variableMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)
	return &info, nil
}
