package service

import (
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/cmd"
	"github.com/1Panel-dev/1Panel/utils/compose"
	"github.com/1Panel-dev/1Panel/utils/files"
	"github.com/joho/godotenv"
	"path"
	"strconv"
)

type DatabaseOp string

var (
	Add    DatabaseOp = "add"
	Delete DatabaseOp = "delete"
)

func execDockerCommand(database model.Database, container model.AppContainer, op DatabaseOp) error {
	var auth dto.AuthParam
	var dbConfig dto.AppDatabase
	dbConfig.Password = database.Password
	dbConfig.DbUser = database.Username
	dbConfig.DbName = database.Dbname
	json.Unmarshal([]byte(container.Auth), &auth)
	execConfig := dto.ContainerExec{
		ContainerName: container.ContainerName,
		Auth:          auth,
		DbParam:       dbConfig,
	}
	_, err := cmd.Exec(getSqlStr(database.Key, op, execConfig))
	if err != nil {
		return err
	}
	return nil
}

func getSqlStr(key string, operate DatabaseOp, exec dto.ContainerExec) string {
	var str string
	param := exec.DbParam
	switch key {
	case "mysql":
		if operate == Add {
			str = fmt.Sprintf("docker exec -i  %s  mysql -uroot -p%s  -e \"CREATE USER '%s'@'%%' IDENTIFIED BY '%s';\" -e \"create database %s;\" -e \"GRANT ALL ON %s.* TO '%s'@'%%';\"",
				exec.ContainerName, exec.Auth.RootPassword, param.DbUser, param.Password, param.DbName, param.DbName, param.DbUser)
		}
		if operate == Delete {
			str = fmt.Sprintf("docker exec -i  %s  mysql -uroot -p%s   -e \"drop database %s;\"  -e \"drop user %s;\" ",
				exec.ContainerName, exec.Auth.RootPassword, param.DbName, param.DbUser)
		}
	}
	return str
}

func copyAppData(key, version, installName string, params map[string]interface{}) (composeFilePath string, err error) {
	resourceDir := path.Join(global.CONF.System.ResourceDir, "apps", key, version)
	installDir := path.Join(global.CONF.System.AppDir, key)
	installVersionDir := path.Join(installDir, version)
	fileOp := files.NewFileOp()
	if err = fileOp.Copy(resourceDir, installVersionDir); err != nil {
		return
	}
	appDir := path.Join(installDir, installName)
	if err = fileOp.Rename(installVersionDir, appDir); err != nil {
		return
	}
	composeFilePath = path.Join(appDir, "docker-compose.yml")
	envPath := path.Join(appDir, ".env")

	envParams := make(map[string]string, len(params))
	for k, v := range params {
		switch t := v.(type) {
		case string:
			envParams[k] = t
		case float64:
			envParams[k] = strconv.FormatFloat(t, 'f', -1, 32)
		default:
			envParams[k] = t.(string)
		}
	}
	if err = godotenv.Write(envParams, envPath); err != nil {
		return
	}
	return
}

func upApp(composeFilePath string, appInstall model.AppInstall) {
	out, err := compose.Up(composeFilePath)
	if err != nil {
		if out != "" {
			appInstall.Message = out
		} else {
			appInstall.Message = err.Error()
		}
		appInstall.Status = constant.Error
		_ = appInstallRepo.Save(appInstall)
	} else {
		appInstall.Status = constant.Running
		_ = appInstallRepo.Save(appInstall)
	}
}

func getAppDetails(details []model.AppDetail, versions []string) map[string]model.AppDetail {
	appDetails := make(map[string]model.AppDetail, len(details))
	for _, old := range details {
		old.Status = constant.AppTakeDown
		appDetails[old.Version] = old
	}

	for _, v := range versions {
		detail, ok := appDetails[v]
		if ok {
			detail.Status = constant.AppNormal
			appDetails[v] = detail
		} else {
			appDetails[v] = model.AppDetail{
				Version: v,
				Status:  constant.AppNormal,
			}
		}
	}
	return appDetails
}

func getApps(oldApps []model.App, items []dto.AppDefine) map[string]model.App {
	apps := make(map[string]model.App, len(oldApps))
	for _, old := range oldApps {
		old.Status = constant.AppTakeDown
		apps[old.Key] = old
	}
	for _, item := range items {
		app, ok := apps[item.Key]
		if !ok {
			app = model.App{}
		}
		app.Name = item.Name
		app.Key = item.Key
		app.ShortDesc = item.ShortDesc
		app.Author = item.Author
		app.Source = item.Source
		app.Type = item.Type
		app.CrossVersionUpdate = item.CrossVersionUpdate
		app.Required = item.GetRequired()
		app.Status = constant.AppNormal
		apps[item.Key] = app
	}
	return apps
}
