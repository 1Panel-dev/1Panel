package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/utils/cmd"
	"github.com/1Panel-dev/1Panel/utils/common"
	"github.com/1Panel-dev/1Panel/utils/compose"
	"github.com/1Panel-dev/1Panel/utils/files"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"math"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type DatabaseOp string

var (
	Add    DatabaseOp = "add"
	Delete DatabaseOp = "delete"
)

func execDockerCommand(database model.Database, dbInstall model.AppInstall, op DatabaseOp) error {
	var auth dto.AuthParam
	var dbConfig dto.AppDatabase
	dbConfig.Password = database.Password
	dbConfig.DbUser = database.Username
	dbConfig.DbName = database.Dbname
	json.Unmarshal([]byte(dbInstall.Param), &auth)
	execConfig := dto.ContainerExec{
		ContainerName: dbInstall.ContainerName,
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
			str = fmt.Sprintf("docker exec -i  %s  mysql -uroot -p%s  -e \"CREATE USER '%s'@'%%' IDENTIFIED BY '%s';\" -e \"create database %s;\" -e \"GRANT ALL ON %s.* TO '%s'@'%%' IDENTIFIED BY '%s';\"",
				exec.ContainerName, exec.Auth.RootPassword, param.DbUser, param.Password, param.DbName, param.DbName, param.DbUser, param.Password)
		}
		if operate == Delete {
			str = fmt.Sprintf("docker exec -i  %s  mysql -uroot -p%s   -e \"drop database %s;\"  -e \"drop user %s;\" ",
				exec.ContainerName, exec.Auth.RootPassword, param.DbName, param.DbUser)
		}
	}
	return str
}

func checkPort(key string, params map[string]interface{}) (int, error) {

	port, ok := params[key]
	if ok {
		portN := int(math.Ceil(port.(float64)))
		if common.ScanPort(portN) {
			return portN, errors.New("port is in used")
		} else {
			return portN, nil
		}
	}
	return 0, nil
}

func createLink(ctx context.Context, app model.App, appInstall *model.AppInstall, params map[string]interface{}) error {
	var dbConfig dto.AppDatabase
	if app.Type == "runtime" {
		var authParam dto.AuthParam
		paramByte, err := json.Marshal(params)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(paramByte, &authParam); err != nil {
			return err
		}
		authByte, err := json.Marshal(authParam)
		if err != nil {
			return err
		}
		appInstall.Param = string(authByte)
	}
	if app.Type == "website" {
		paramByte, err := json.Marshal(params)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(paramByte, &dbConfig); err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(dbConfig, dto.AppDatabase{}) {
		dbInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithServiceName(dbConfig.ServiceName))
		if err != nil {
			return err
		}
		var database model.Database
		database.Dbname = dbConfig.DbName
		database.Username = dbConfig.DbUser
		database.Password = dbConfig.Password
		database.AppInstallId = dbInstall.ID
		database.Key = dbInstall.App.Key
		if err := dataBaseRepo.Create(ctx, &database); err != nil {
			return err
		}
		var installResource model.AppInstallResource
		installResource.ResourceId = database.ID
		installResource.AppInstallId = appInstall.ID
		installResource.LinkId = dbInstall.ID
		installResource.Key = dbInstall.App.Key
		if err := appInstallResourceRepo.Create(ctx, &installResource); err != nil {
			return err
		}
		if err := execDockerCommand(database, dbInstall, Add); err != nil {
			return err
		}
	}

	return nil
}

func deleteLink(ctx context.Context, install *model.AppInstall) error {
	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	if len(resources) == 0 {
		return nil
	}
	for _, re := range resources {
		if re.Key == "mysql" {
			database, _ := dataBaseRepo.GetFirst(commonRepo.WithByID(re.ResourceId))
			if reflect.DeepEqual(database, model.Database{}) {
				continue
			}
			appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(database.AppInstallId))
			if err != nil {
				return nil
			}
			if err := execDockerCommand(database, appInstall, Delete); err != nil {
				return err
			}
			if err := dataBaseRepo.DeleteBy(ctx, commonRepo.WithByID(database.ID)); err != nil {
				return err
			}
		}
	}
	return appInstallResourceRepo.DeleteBy(ctx, appInstallResourceRepo.WithAppInstallId(install.ID))
}

func updateInstall(installId uint, detailId uint) error {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
	if err != nil {
		return err
	}
	detail, err := appDetailRepo.GetFirst(commonRepo.WithByID(detailId))
	if err != nil {
		return err
	}
	if install.Version == detail.Version {
		return errors.New("two version is save")
	}
	tx, ctx := getTxAndContext()
	if err := backupInstall(ctx, install); err != nil {
		return err
	}
	tx.Commit()
	if _, err = compose.Down(install.GetComposePath()); err != nil {
		return err
	}
	install.DockerCompose = detail.DockerCompose
	install.Version = detail.Version
	fileOp := files.NewFileOp()
	if err := fileOp.WriteFile(install.GetComposePath(), strings.NewReader(install.DockerCompose), 0775); err != nil {
		return err
	}
	if _, err = compose.Up(install.GetComposePath()); err != nil {
		return err
	}
	return appInstallRepo.Save(&install)
}

func backupInstall(ctx context.Context, install model.AppInstall) error {
	var backup model.AppInstallBackup
	appPath := install.GetPath()
	backupDir := path.Join(constant.BackupDir, install.App.Key, install.Name)
	fileOp := files.NewFileOp()
	now := time.Now()
	day := now.Format("2006-01-02-15-04")
	fileName := fmt.Sprintf("%s-%s-%s%s", install.Name, install.Version, day, ".tar.gz")
	if err := fileOp.Compress([]string{appPath}, backupDir, fileName, files.TarGz); err != nil {
		return err
	}
	backup.Name = fileName
	backup.Path = backupDir
	backup.AppInstallId = install.ID
	backup.AppDetailId = install.AppDetailId
	backup.Param = install.Param

	if err := appInstallBackupRepo.Create(ctx, backup); err != nil {
		return err
	}
	return nil
}

func restoreInstall(install model.AppInstall, backup model.AppInstallBackup) error {
	if _, err := compose.Down(install.GetComposePath()); err != nil {
		return err
	}
	installKeyDir := path.Join(constant.AppInstallDir, install.App.Key)
	installDir := path.Join(installKeyDir, install.Name)
	backupFile := path.Join(backup.Path, backup.Name)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(backupFile) {
		return errors.New(fmt.Sprintf("%s file is not exist", backup.Name))
	}
	backDir := installDir + "_back"
	if fileOp.Stat(backDir) {
		if err := fileOp.DeleteDir(backDir); err != nil {
			return err
		}
	}
	if err := fileOp.Rename(installDir, backDir); err != nil {
		return err
	}
	if err := fileOp.Decompress(backupFile, installKeyDir, files.TarGz); err != nil {
		return err
	}
	composeContent, err := os.ReadFile(install.GetComposePath())
	if err != nil {
		return err
	}
	install.DockerCompose = string(composeContent)
	envContent, err := os.ReadFile(path.Join(installDir, ".env"))
	if err != nil {
		return err
	}
	install.Env = string(envContent)
	envMaps, err := godotenv.Unmarshal(string(envContent))
	if err != nil {
		return err
	}
	install.HttpPort = 0
	httpPort, ok := envMaps["PANEL_APP_PORT_HTTP"]
	if ok {
		httpPortN, _ := strconv.Atoi(httpPort)
		install.HttpPort = httpPortN
	}
	install.HttpsPort = 0
	httpsPort, ok := envMaps["PANEL_APP_PORT_HTTPS"]
	if ok {
		httpsPortN, _ := strconv.Atoi(httpsPort)
		install.HttpsPort = httpsPortN
	}

	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return err
	}
	servicesMap := composeMap["services"].(map[string]interface{})
	for k, v := range servicesMap {
		install.ServiceName = k
		value := v.(map[string]interface{})
		install.ContainerName = value["container_name"].(string)
	}

	install.Param = backup.Param
	_ = fileOp.DeleteDir(backDir)
	if out, err := compose.Up(install.GetComposePath()); err != nil {
		return handleErr(install, err, out)
	}
	install.AppDetailId = backup.AppDetailId
	install.Status = constant.Running
	return appInstallRepo.Save(&install)
}

func getContainerNames(install model.AppInstall) ([]string, error) {
	composeMap := install.DockerCompose
	envMap := make(map[string]string)
	_ = json.Unmarshal([]byte(install.Env), &envMap)
	project, err := compose.GetComposeProject([]byte(composeMap), envMap)
	if err != nil {
		return nil, err
	}
	var containerNames []string
	for _, service := range project.AllServices() {
		containerNames = append(containerNames, service.ContainerName)
	}
	return containerNames, nil
}

func checkRequiredAndLimit(app model.App) error {

	if app.Limit > 0 {
		installs, err := appInstallRepo.GetBy(appInstallRepo.WithAppId(app.ID))
		if err != nil {
			return err
		}
		if len(installs) >= app.Limit {
			return errors.New(fmt.Sprintf("app install limit %d", app.Limit))
		}
	}

	if app.Required != "" {
		var requiredArray []string
		if err := json.Unmarshal([]byte(app.Required), &requiredArray); err != nil {
			return err
		}
		for _, key := range requiredArray {
			if key == "" {
				continue
			}
			requireApp, err := appRepo.GetFirst(appRepo.WithKey(key))
			if err != nil {
				return err
			}
			details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(requireApp.ID))
			if err != nil {
				return err
			}
			var detailIds []uint
			for _, d := range details {
				detailIds = append(detailIds, d.ID)
			}

			_, err = appInstallRepo.GetFirst(appInstallRepo.WithDetailIdsIn(detailIds))
			if err != nil {
				return errors.New(fmt.Sprintf("%s is required", requireApp.Key))
			}
		}
	}

	return nil
}

func handleMap(params map[string]interface{}, envParams map[string]string) {
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
}

func copyAppData(key, version, installName string, params map[string]interface{}) (err error) {
	resourceDir := path.Join(constant.AppResourceDir, key, version)
	installDir := path.Join(constant.AppInstallDir, key)
	installVersionDir := path.Join(installDir, version)
	fileOp := files.NewFileOp()
	if err = fileOp.Copy(resourceDir, installVersionDir); err != nil {
		return
	}
	appDir := path.Join(installDir, installName)
	if err = fileOp.Rename(installVersionDir, appDir); err != nil {
		return
	}
	envPath := path.Join(appDir, ".env")

	envParams := make(map[string]string, len(params))
	handleMap(params, envParams)
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
		_ = appInstallRepo.Save(&appInstall)
	} else {
		appInstall.Status = constant.Running
		_ = appInstallRepo.Save(&appInstall)
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

func handleErr(install model.AppInstall, err error, out string) error {
	reErr := err
	install.Message = err.Error()
	if out != "" {
		install.Message = out
		reErr = errors.New(out)
	}
	_ = appInstallRepo.Save(&install)
	return reErr
}
