package service

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
)

type DBCommonService struct{}

type IDBCommonService interface {
	LoadBaseInfo(req dto.OperationWithNameAndType) (*dto.DBBaseInfo, error)
	LoadDatabaseFile(req dto.OperationWithNameAndType) (string, error)

	UpdateConfByFile(req dto.DBConfUpdateByFile) error
}

func NewIDBCommonService() IDBCommonService {
	return &DBCommonService{}
}

func (u *DBCommonService) LoadBaseInfo(req dto.OperationWithNameAndType) (*dto.DBBaseInfo, error) {
	var data dto.DBBaseInfo
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Name)
	if err != nil {
		return nil, err
	}
	data.ContainerName = app.ContainerName
	data.Name = app.Name
	data.Port = int64(app.Port)

	return &data, nil
}

func (u *DBCommonService) LoadDatabaseFile(req dto.OperationWithNameAndType) (string, error) {
	filePath := ""
	switch req.Type {
	case "mysql-conf":
		filePath = path.Join(global.CONF.System.DataDir, fmt.Sprintf("apps/mysql/%s/conf/my.cnf", req.Name))
	case "mariadb-conf":
		filePath = path.Join(global.CONF.System.DataDir, fmt.Sprintf("apps/mariadb/%s/conf/my.cnf", req.Name))
	case "postgresql-conf":
		filePath = path.Join(global.CONF.System.DataDir, fmt.Sprintf("apps/postgresql/%s/data/postgresql.conf", req.Name))
	case "redis-conf":
		filePath = path.Join(global.CONF.System.DataDir, fmt.Sprintf("apps/redis/%s/conf/redis.conf", req.Name))
	}
	if _, err := os.Stat(filePath); err != nil {
		return "", buserr.New("ErrHttpReqNotFound")
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (u *DBCommonService) UpdateConfByFile(req dto.DBConfUpdateByFile) error {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Database)
	if err != nil {
		return err
	}
	path := ""
	switch req.Type {
	case constant.AppMariaDB, constant.AppMysql:
		path = fmt.Sprintf("%s/%s/%s/conf/my.cnf", constant.AppInstallDir, req.Type, app.Name)
	case constant.AppPostgresql:
		path = fmt.Sprintf("%s/%s/%s/data/postgresql.conf", constant.AppInstallDir, req.Type, app.Name)
	case constant.AppRedis:
		path = fmt.Sprintf("%s/%s/%s/conf/redis.conf", constant.AppInstallDir, req.Type, app.Name)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()
	if _, err := compose.Restart(fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, req.Type, app.Name)); err != nil {
		return err
	}
	return nil
}
