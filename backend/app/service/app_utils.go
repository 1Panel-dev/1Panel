package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/utils/env"

	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	composeV2 "github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

type DatabaseOp string

var (
	Add    DatabaseOp = "add"
	Delete DatabaseOp = "delete"
)

func checkPort(key string, params map[string]interface{}) (int, error) {
	port, ok := params[key]
	if ok {
		portN := 0
		var err error
		switch p := port.(type) {
		case string:
			portN, err = strconv.Atoi(p)
			if err != nil {
				return portN, nil
			}
		case float64:
			portN = int(math.Ceil(p))
		case int:
			portN = p
		}

		oldInstalled, _ := appInstallRepo.ListBy(appInstallRepo.WithPort(portN))
		if len(oldInstalled) > 0 {
			var apps []string
			for _, install := range oldInstalled {
				apps = append(apps, install.App.Name)
			}
			return portN, buserr.WithMap(constant.ErrPortInOtherApp, map[string]interface{}{"port": portN, "apps": apps}, nil)
		}
		if common.ScanPort(portN) {
			return portN, buserr.WithDetail(constant.ErrPortInUsed, portN, nil)
		} else {
			return portN, nil
		}
	}
	return 0, nil
}

func checkPortExist(port int) error {
	errMap := make(map[string]interface{})
	errMap["port"] = port
	appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithPort(port))
	if appInstall.ID > 0 {
		errMap["type"] = i18n.GetMsgByKey("TYPE_APP")
		errMap["name"] = appInstall.Name
		return buserr.WithMap("ErrPortExist", errMap, nil)
	}
	runtime, _ := runtimeRepo.GetFirst(runtimeRepo.WithPort(port))
	if runtime != nil {
		errMap["type"] = i18n.GetMsgByKey("TYPE_RUNTIME")
		errMap["name"] = runtime.Name
		return buserr.WithMap("ErrPortExist", errMap, nil)
	}
	domain, _ := websiteDomainRepo.GetFirst(websiteDomainRepo.WithPort(port))
	if domain.ID > 0 {
		errMap["type"] = i18n.GetMsgByKey("TYPE_DOMAIN")
		errMap["name"] = domain.Domain
		return buserr.WithMap("ErrPortExist", errMap, nil)
	}
	if common.ScanPort(port) {
		return buserr.WithDetail(constant.ErrPortInUsed, port, nil)
	}
	return nil
}

var DatabaseKeys = map[string]uint{
	constant.AppMysql:      3306,
	constant.AppMariaDB:    3306,
	constant.AppPostgresql: 5432,
	constant.AppPostgres:   5432,
	constant.AppMongodb:    27017,
	constant.AppRedis:      6379,
	constant.AppMemcached:  11211,
}

var ToolKeys = map[string]uint{
	"minio": 9001,
}

func createLink(ctx context.Context, app model.App, appInstall *model.AppInstall, params map[string]interface{}) error {
	var dbConfig dto.AppDatabase
	if DatabaseKeys[app.Key] > 0 {
		database := &model.Database{
			AppInstallID: appInstall.ID,
			Name:         appInstall.Name,
			Type:         app.Key,
			Version:      appInstall.Version,
			From:         "local",
			Address:      appInstall.ServiceName,
			Port:         DatabaseKeys[app.Key],
		}
		detail, err := appDetailRepo.GetFirst(commonRepo.WithByID(appInstall.AppDetailId))
		if err != nil {
			return err
		}

		formFields := &dto.AppForm{}
		if err := json.Unmarshal([]byte(detail.Params), formFields); err != nil {
			return err
		}
		for _, form := range formFields.FormFields {
			if form.EnvKey == "PANEL_APP_PORT_HTTP" {
				portFloat, ok := form.Default.(float64)
				if ok {
					database.Port = uint(int(portFloat))
				}
				break
			}
		}

		switch app.Key {
		case constant.AppMysql, constant.AppMariaDB, constant.AppPostgresql, constant.AppMongodb:
			if password, ok := params["PANEL_DB_ROOT_PASSWORD"]; ok {
				if password != "" {
					database.Password = password.(string)
					if app.Key == "mysql" || app.Key == "mariadb" {
						database.Username = "root"
					}
					if rootUser, ok := params["PANEL_DB_ROOT_USER"]; ok {
						database.Username = rootUser.(string)
					}
					authParam := dto.AuthParam{
						RootPassword: password.(string),
						RootUser:     database.Username,
					}
					authByte, err := json.Marshal(authParam)
					if err != nil {
						return err
					}
					appInstall.Param = string(authByte)

				}
			}
		case constant.AppRedis:
			if password, ok := params["PANEL_REDIS_ROOT_PASSWORD"]; ok {
				if password != "" {
					authParam := dto.RedisAuthParam{
						RootPassword: password.(string),
					}
					authByte, err := json.Marshal(authParam)
					if err != nil {
						return err
					}
					appInstall.Param = string(authByte)
				}
				database.Password = password.(string)
			}
		}
		if err := databaseRepo.Create(ctx, database); err != nil {
			return err
		}
	}
	if ToolKeys[app.Key] > 0 {
		if app.Key == "minio" {
			authParam := dto.MinioAuthParam{}
			if password, ok := params["PANEL_MINIO_ROOT_PASSWORD"]; ok {
				authParam.RootPassword = password.(string)
			}
			if rootUser, ok := params["PANEL_MINIO_ROOT_USER"]; ok {
				authParam.RootUser = rootUser.(string)
			}
			authByte, err := json.Marshal(authParam)
			if err != nil {
				return err
			}
			appInstall.Param = string(authByte)
		}
	}

	if app.Type == "website" || app.Type == "tool" {
		paramByte, err := json.Marshal(params)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(paramByte, &dbConfig); err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(dbConfig, dto.AppDatabase{}) && dbConfig.ServiceName != "" {
		hostName := params["PANEL_DB_HOST_NAME"]
		if hostName == nil || hostName.(string) == "" {
			return nil
		}
		database, _ := databaseRepo.Get(commonRepo.WithByName(hostName.(string)))
		if database.ID == 0 {
			return nil
		}
		var resourceId uint
		if dbConfig.DbName != "" && dbConfig.DbUser != "" && dbConfig.Password != "" {
			switch database.Type {
			case constant.AppPostgresql, constant.AppPostgres:
				iPostgresqlRepo := repo.NewIPostgresqlRepo()
				oldPostgresqlDb, _ := iPostgresqlRepo.Get(commonRepo.WithByName(dbConfig.DbName), iPostgresqlRepo.WithByFrom(constant.ResourceLocal))
				resourceId = oldPostgresqlDb.ID
				if oldPostgresqlDb.ID > 0 {
					if oldPostgresqlDb.Username != dbConfig.DbUser || oldPostgresqlDb.Password != dbConfig.Password {
						return buserr.New(constant.ErrDbUserNotValid)
					}
				} else {
					var createPostgresql dto.PostgresqlDBCreate
					createPostgresql.Name = dbConfig.DbName
					createPostgresql.Username = dbConfig.DbUser
					createPostgresql.Database = database.Name
					createPostgresql.Format = "UTF8"
					createPostgresql.Password = dbConfig.Password
					createPostgresql.From = database.From
					createPostgresql.SuperUser = true
					pgdb, err := NewIPostgresqlService().Create(ctx, createPostgresql)
					if err != nil {
						return err
					}
					resourceId = pgdb.ID
				}
			case constant.AppMysql, constant.AppMariaDB:
				iMysqlRepo := repo.NewIMysqlRepo()
				oldMysqlDb, _ := iMysqlRepo.Get(commonRepo.WithByName(dbConfig.DbName), iMysqlRepo.WithByFrom(constant.ResourceLocal))
				resourceId = oldMysqlDb.ID
				if oldMysqlDb.ID > 0 {
					if oldMysqlDb.Username != dbConfig.DbUser || oldMysqlDb.Password != dbConfig.Password {
						return buserr.New(constant.ErrDbUserNotValid)
					}
				} else {
					var createMysql dto.MysqlDBCreate
					createMysql.Name = dbConfig.DbName
					createMysql.Username = dbConfig.DbUser
					createMysql.Database = database.Name
					createMysql.Format = "utf8mb4"
					createMysql.Permission = "%"
					createMysql.Password = dbConfig.Password
					createMysql.From = database.From
					mysqldb, err := NewIMysqlService().Create(ctx, createMysql)
					if err != nil {
						return err
					}
					resourceId = mysqldb.ID
				}
			}

		}
		var installResource model.AppInstallResource
		installResource.ResourceId = resourceId
		installResource.AppInstallId = appInstall.ID
		if database.AppInstallID > 0 {
			installResource.LinkId = database.AppInstallID
		} else {
			installResource.LinkId = database.ID
		}
		installResource.Key = database.Type
		installResource.From = database.From
		if err := appInstallResourceRepo.Create(ctx, &installResource); err != nil {
			return err
		}
	}
	return nil
}

func handleAppInstallErr(ctx context.Context, install *model.AppInstall) error {
	op := files.NewFileOp()
	appDir := install.GetPath()
	dir, _ := os.Stat(appDir)
	if dir != nil {
		_, _ = compose.Down(install.GetComposePath())
		if err := op.DeleteDir(appDir); err != nil {
			return err
		}
	}
	if err := deleteLink(ctx, install, true, true, true); err != nil {
		return err
	}
	return nil
}

func deleteAppInstall(install model.AppInstall, deleteBackup bool, forceDelete bool, deleteDB bool) error {
	op := files.NewFileOp()
	appDir := install.GetPath()
	dir, _ := os.Stat(appDir)
	if dir != nil {
		out, err := compose.Down(install.GetComposePath())
		if err != nil && !forceDelete {
			return handleErr(install, err, out)
		}
		if err = runScript(&install, "uninstall"); err != nil {
			_, _ = compose.Up(install.GetComposePath())
			return err
		}
	}
	tx, ctx := helper.GetTxAndContext()
	defer tx.Rollback()
	if err := appInstallRepo.Delete(ctx, install); err != nil {
		return err
	}
	if err := deleteLink(ctx, &install, deleteDB, forceDelete, deleteBackup); err != nil && !forceDelete {
		return err
	}

	if DatabaseKeys[install.App.Key] > 0 {
		_ = databaseRepo.Delete(ctx, databaseRepo.WithAppInstallID(install.ID))
	}

	switch install.App.Key {
	case constant.AppOpenresty:
		websites, _ := websiteRepo.List()
		for _, website := range websites {
			if website.AppInstallID > 0 {
				websiteAppInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
				if websiteAppInstall.AppId > 0 {
					websiteApp, _ := appRepo.GetFirst(commonRepo.WithByID(websiteAppInstall.AppId))
					if websiteApp.Type == constant.RuntimePHP {
						go func() {
							_, _ = compose.Down(install.GetComposePath())
							_ = op.DeleteDir(install.GetPath())
						}()
						_ = appInstallRepo.Delete(ctx, websiteAppInstall)
					}
				}
			}
		}
		_ = websiteRepo.DeleteAll(ctx)
		_ = websiteDomainRepo.DeleteAll(ctx)
	case constant.AppMysql, constant.AppMariaDB:
		_ = mysqlRepo.Delete(ctx, mysqlRepo.WithByMysqlName(install.Name))
	case constant.AppPostgresql:
		_ = postgresqlRepo.Delete(ctx, postgresqlRepo.WithByPostgresqlName(install.Name))
	}

	_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType("app"), commonRepo.WithByName(install.App.Key), backupRepo.WithByDetailName(install.Name))
	uploadDir := path.Join(global.CONF.System.BaseDir, fmt.Sprintf("1panel/uploads/app/%s/%s", install.App.Key, install.Name))
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	if deleteBackup {
		localDir, _ := loadLocalDir()
		backupDir := path.Join(localDir, fmt.Sprintf("app/%s/%s", install.App.Key, install.Name))
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		global.LOG.Infof("delete app %s-%s backups successful", install.App.Key, install.Name)
	}
	_ = op.DeleteDir(appDir)
	tx.Commit()
	return nil
}

func deleteLink(ctx context.Context, install *model.AppInstall, deleteDB bool, forceDelete bool, deleteBackup bool) error {
	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	if len(resources) == 0 {
		return nil
	}
	for _, re := range resources {
		if deleteDB {
			switch re.Key {
			case constant.AppMysql, constant.AppMariaDB:
				mysqlService := NewIMysqlService()
				database, _ := mysqlRepo.Get(commonRepo.WithByID(re.ResourceId))
				if reflect.DeepEqual(database, model.DatabaseMysql{}) {
					continue
				}
				if err := mysqlService.Delete(ctx, dto.MysqlDBDelete{
					ID:           database.ID,
					ForceDelete:  forceDelete,
					DeleteBackup: deleteBackup,
					Type:         re.Key,
					Database:     database.MysqlName,
				}); err != nil && !forceDelete {
					return err
				}
			case constant.AppPostgresql:
				pgsqlService := NewIPostgresqlService()
				database, _ := postgresqlRepo.Get(commonRepo.WithByID(re.ResourceId))
				if reflect.DeepEqual(database, model.DatabasePostgresql{}) {
					continue
				}
				if err := pgsqlService.Delete(ctx, dto.PostgresqlDBDelete{
					ID:           database.ID,
					ForceDelete:  forceDelete,
					DeleteBackup: deleteBackup,
					Type:         re.Key,
					Database:     database.PostgresqlName,
				}); err != nil {
					return err
				}
			}
		}

	}
	return appInstallResourceRepo.DeleteBy(ctx, appInstallResourceRepo.WithAppInstallId(install.ID))
}

func upgradeInstall(installID uint, detailID uint, backup, pullImage bool) error {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(installID))
	if err != nil {
		return err
	}
	detail, err := appDetailRepo.GetFirst(commonRepo.WithByID(detailID))
	if err != nil {
		return err
	}
	if install.Version == detail.Version {
		return errors.New("two version is same")
	}
	install.Status = constant.Upgrading

	go func() {
		if backup {
			if err = NewIBackupService().AppBackup(dto.CommonBackup{Name: install.App.Key, DetailName: install.Name}); err != nil {
				global.LOG.Errorf(i18n.GetMsgWithMap("ErrAppBackup", map[string]interface{}{"name": install.Name, "err": err.Error()}))
			}
		}
		var upErr error

		defer func() {
			if upErr != nil {
				if backup {
					if err := NewIBackupService().AppRecover(dto.CommonRecover{Name: install.App.Key, DetailName: install.Name, Type: "app", Source: constant.ResourceLocal}); err != nil {
						global.LOG.Errorf("recover app [%s] [%s] failed %v", install.App.Key, install.Name, err)
					}
				}
				install.Status = constant.UpgradeErr
				install.Message = upErr.Error()
				_ = appInstallRepo.Save(context.Background(), &install)
			}
		}()

		fileOp := files.NewFileOp()
		detailDir := path.Join(constant.ResourceDir, "apps", install.App.Resource, install.App.Key, detail.Version)
		if install.App.Resource == constant.AppResourceRemote {
			if upErr = downloadApp(install.App, detail, &install); upErr != nil {
				return
			}
			if detail.DockerCompose == "" {
				composeDetail, err := fileOp.GetContent(path.Join(detailDir, "docker-compose.yml"))
				if err != nil {
					upErr = err
					return
				}
				detail.DockerCompose = string(composeDetail)
				_ = appDetailRepo.Update(context.Background(), detail)
			}
			go func() {
				_, _ = http.Get(detail.DownloadCallBackUrl)
			}()
		}

		if install.App.Resource == constant.AppResourceLocal {
			detailDir = path.Join(constant.ResourceDir, "apps", "local", strings.TrimPrefix(install.App.Key, "local"), detail.Version)
		}

		command := exec.Command("/bin/bash", "-c", fmt.Sprintf("cp -rn %s/* %s || true", detailDir, install.GetPath()))
		stdout, _ := command.CombinedOutput()
		if stdout != nil {
			global.LOG.Infof("upgrade app [%s] [%s] cp file log : %s ", install.App.Key, install.Name, string(stdout))
		}
		sourceScripts := path.Join(detailDir, "scripts")
		if fileOp.Stat(sourceScripts) {
			dstScripts := path.Join(install.GetPath(), "scripts")
			_ = fileOp.DeleteDir(dstScripts)
			_ = fileOp.CreateDir(dstScripts, 0755)
			scriptCmd := exec.Command("cp", "-rf", sourceScripts+"/.", dstScripts+"/")
			_, _ = scriptCmd.CombinedOutput()
		}

		composeMap := make(map[string]interface{})
		if upErr = yaml.Unmarshal([]byte(detail.DockerCompose), &composeMap); upErr != nil {
			return
		}
		value, ok := composeMap["services"]
		if !ok {
			upErr = buserr.New(constant.ErrFileParse)
			return
		}
		servicesMap := value.(map[string]interface{})
		if len(servicesMap) == 1 {
			index := 0
			oldServiceName := ""
			for k := range servicesMap {
				oldServiceName = k
				index++
				if index > 0 {
					break
				}
			}
			servicesMap[install.ServiceName] = servicesMap[oldServiceName]
			if install.ServiceName != oldServiceName {
				delete(servicesMap, oldServiceName)
			}
		}
		envs := make(map[string]interface{})
		if upErr = json.Unmarshal([]byte(install.Env), &envs); upErr != nil {
			return
		}
		config := getAppCommonConfig(envs)
		if config.ContainerName == "" {
			config.ContainerName = install.ContainerName
			envs[constant.ContainerName] = install.ContainerName
		}
		config.Advanced = true
		if upErr = addDockerComposeCommonParam(composeMap, install.ServiceName, config, envs); upErr != nil {
			return
		}
		paramByte, upErr := json.Marshal(envs)
		if upErr != nil {
			return
		}
		install.Env = string(paramByte)
		composeByte, upErr := yaml.Marshal(composeMap)
		if upErr != nil {
			return
		}

		install.DockerCompose = string(composeByte)
		install.Version = detail.Version
		install.AppDetailId = detailID

		if out, err := compose.Down(install.GetComposePath()); err != nil {
			if out != "" {
				upErr = errors.New(out)
				return
			}
			return
		}
		envParams := make(map[string]string, len(envs))
		handleMap(envs, envParams)
		if err = env.Write(envParams, install.GetEnvPath()); err != nil {
			return
		}

		if err = runScript(&install, "upgrade"); err != nil {
			return
		}

		content, err := fileOp.GetContent(install.GetEnvPath())
		if err != nil {
			upErr = err
			return
		}

		if pullImage {
			images, err := composeV2.GetDockerComposeImages(install.Name, content, []byte(detail.DockerCompose))
			if err != nil {
				upErr = err
				return
			}
			dockerCli, err := composeV2.NewClient()
			if err != nil {
				upErr = err
				return
			}
			for _, image := range images {
				if err = dockerCli.PullImage(image, true); err != nil {
					upErr = buserr.WithNameAndErr("ErrDockerPullImage", "", err)
					return
				}
			}
		}

		if upErr = fileOp.WriteFile(install.GetComposePath(), strings.NewReader(install.DockerCompose), 0775); upErr != nil {
			return
		}
		if out, err := compose.Up(install.GetComposePath()); err != nil {
			if out != "" {
				upErr = errors.New(out)
				return
			}
			upErr = err
			return
		}
		install.Status = constant.Running
		_ = appInstallRepo.Save(context.Background(), &install)
	}()

	return appInstallRepo.Save(context.Background(), &install)
}

func getContainerNames(install model.AppInstall) ([]string, error) {
	envStr, err := coverEnvJsonToStr(install.Env)
	if err != nil {
		return nil, err
	}
	project, err := composeV2.GetComposeProject(install.Name, install.GetPath(), []byte(install.DockerCompose), []byte(envStr), true)
	if err != nil {
		return nil, err
	}
	containerMap := make(map[string]struct{})
	for _, service := range project.AllServices() {
		if service.ContainerName == "${CONTAINER_NAME}" || service.ContainerName == "" {
			continue
		}
		containerMap[service.ContainerName] = struct{}{}
	}
	var containerNames []string
	for k := range containerMap {
		containerNames = append(containerNames, k)
	}
	if len(containerNames) == 0 {
		containerNames = append(containerNames, install.ContainerName)
	}
	return containerNames, nil
}

func coverEnvJsonToStr(envJson string) (string, error) {
	envMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(envJson), &envMap)
	newEnvMap := make(map[string]string, len(envMap))
	handleMap(envMap, newEnvMap)
	envStr, err := gotenv.Marshal(newEnvMap)
	if err != nil {
		return "", err
	}
	return envStr, nil
}

func checkLimit(app model.App) error {
	if app.Limit > 0 {
		installs, err := appInstallRepo.ListBy(appInstallRepo.WithAppId(app.ID))
		if err != nil {
			return err
		}
		if len(installs) >= app.Limit {
			return buserr.New(constant.ErrAppLimit)
		}
	}
	return nil
}

func checkRequiredAndLimit(app model.App) error {
	if err := checkLimit(app); err != nil {
		return err
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
		case uint:
			envParams[k] = strconv.Itoa(int(t))
		case int:
			envParams[k] = strconv.Itoa(t)
		case []interface{}:
			strArray := make([]string, len(t))
			for i := range t {
				strArray[i] = strings.ToLower(fmt.Sprintf("%v", t[i]))
			}
			envParams[k] = strings.Join(strArray, ",")
		default:
			envParams[k] = t.(string)
		}
	}
}

func downloadApp(app model.App, appDetail model.AppDetail, appInstall *model.AppInstall) (err error) {
	if app.IsLocalApp() {
		//本地应用,不去官网下载
		return nil
	}
	appResourceDir := path.Join(constant.AppResourceDir, app.Resource)
	appDownloadDir := app.GetAppResourcePath()
	appVersionDir := path.Join(appDownloadDir, appDetail.Version)
	fileOp := files.NewFileOp()
	if !appDetail.Update && fileOp.Stat(appVersionDir) {
		return
	}
	if !fileOp.Stat(appDownloadDir) {
		_ = fileOp.CreateDir(appDownloadDir, 0755)
	}
	if !fileOp.Stat(appVersionDir) {
		_ = fileOp.CreateDir(appVersionDir, 0755)
	}
	global.LOG.Infof("download app[%s] from %s", app.Name, appDetail.DownloadUrl)
	filePath := path.Join(appVersionDir, app.Key+"-"+appDetail.Version+".tar.gz")

	defer func() {
		if err != nil {
			if appInstall != nil {
				appInstall.Status = constant.DownloadErr
				appInstall.Message = err.Error()
			}
		}
	}()

	if err = fileOp.DownloadFile(appDetail.DownloadUrl, filePath); err != nil {
		global.LOG.Errorf("download app[%s] error %v", app.Name, err)
		return
	}
	if err = fileOp.Decompress(filePath, appResourceDir, files.SdkTarGz); err != nil {
		global.LOG.Errorf("decompress app[%s] error %v", app.Name, err)
		return
	}
	_ = fileOp.DeleteFile(filePath)
	appDetail.Update = false
	_ = appDetailRepo.Update(context.Background(), appDetail)
	return
}

func copyData(app model.App, appDetail model.AppDetail, appInstall *model.AppInstall, req request.AppInstallCreate) (err error) {
	fileOp := files.NewFileOp()
	appResourceDir := path.Join(constant.AppResourceDir, app.Resource)

	if app.Resource == constant.AppResourceRemote {
		err = downloadApp(app, appDetail, appInstall)
		if err != nil {
			return
		}
		go func() {
			_, _ = http.Get(appDetail.DownloadCallBackUrl)
		}()
	}
	appKey := app.Key
	installAppDir := path.Join(constant.AppInstallDir, app.Key)
	if app.Resource == constant.AppResourceLocal {
		appResourceDir = constant.LocalAppResourceDir
		appKey = strings.TrimPrefix(app.Key, "local")
		installAppDir = path.Join(constant.LocalAppInstallDir, appKey)
	}
	resourceDir := path.Join(appResourceDir, appKey, appDetail.Version)

	if !fileOp.Stat(installAppDir) {
		if err = fileOp.CreateDir(installAppDir, 0755); err != nil {
			return
		}
	}
	appDir := path.Join(installAppDir, req.Name)
	if fileOp.Stat(appDir) {
		if err = fileOp.DeleteDir(appDir); err != nil {
			return
		}
	}
	if err = fileOp.Copy(resourceDir, installAppDir); err != nil {
		return
	}
	versionDir := path.Join(installAppDir, appDetail.Version)
	if err = fileOp.Rename(versionDir, appDir); err != nil {
		return
	}
	envPath := path.Join(appDir, ".env")

	envParams := make(map[string]string, len(req.Params))
	handleMap(req.Params, envParams)
	if err = env.Write(envParams, envPath); err != nil {
		return
	}
	if err := fileOp.WriteFile(appInstall.GetComposePath(), strings.NewReader(appInstall.DockerCompose), 0755); err != nil {
		return err
	}
	return
}

func runScript(appInstall *model.AppInstall, operate string) error {
	workDir := appInstall.GetPath()
	scriptPath := ""
	switch operate {
	case "init":
		scriptPath = path.Join(workDir, "scripts", "init.sh")
	case "upgrade":
		scriptPath = path.Join(workDir, "scripts", "upgrade.sh")
	case "uninstall":
		scriptPath = path.Join(workDir, "scripts", "uninstall.sh")
	}
	if !files.NewFileOp().Stat(scriptPath) {
		return nil
	}
	out, err := cmd.ExecScript(scriptPath, workDir)
	if err != nil {
		if out != "" {
			errMsg := fmt.Sprintf("run script %s error %s", scriptPath, out)
			global.LOG.Error(errMsg)
			return errors.New(errMsg)
		}
		return err
	}
	return nil
}

func checkContainerNameIsExist(containerName, appDir string) (bool, error) {
	client, err := composeV2.NewDockerClient()
	if err != nil {
		return false, err
	}
	var options dockerTypes.ContainerListOptions
	list, err := client.ContainerList(context.Background(), options)
	if err != nil {
		return false, err
	}
	for _, container := range list {
		if containerName == container.Names[0][1:] {
			if workDir, ok := container.Labels[composeWorkdirLabel]; ok {
				if workDir != appDir {
					return true, nil
				}
			} else {
				return true, nil
			}
		}

	}
	return false, nil
}

func upApp(appInstall *model.AppInstall, pullImages bool) {
	upProject := func(appInstall *model.AppInstall) (err error) {
		if err == nil {
			var (
				out    string
				errMsg string
			)
			if pullImages && appInstall.App.Type != "php" {
				out, err = compose.Pull(appInstall.GetComposePath())
				if err != nil {
					if out != "" {
						if strings.Contains(out, "no such host") {
							errMsg = i18n.GetMsgByKey("ErrNoSuchHost") + ":"
						}
						if strings.Contains(out, "timeout") {
							errMsg = i18n.GetMsgByKey("ErrImagePullTimeOut") + ":"
						}
						appInstall.Message = errMsg + out
					}
					return err
				}
			}

			out, err = compose.Up(appInstall.GetComposePath())
			if err != nil {
				if out != "" {
					appInstall.Message = errMsg + out
				}
				return err
			}
			return
		} else {
			return
		}
	}
	if err := upProject(appInstall); err != nil {
		appInstall.Status = constant.Error
	} else {
		appInstall.Status = constant.Running
	}
	exist, _ := appInstallRepo.GetFirst(commonRepo.WithByID(appInstall.ID))
	if exist.ID > 0 {
		_ = appInstallRepo.Save(context.Background(), appInstall)
	}
}

func rebuildApp(appInstall model.AppInstall) error {
	appInstall.Status = constant.Rebuilding
	_ = appInstallRepo.Save(context.Background(), &appInstall)
	go func() {
		dockerComposePath := appInstall.GetComposePath()
		out, err := compose.Down(dockerComposePath)
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}
		fileOp := files.NewFileOp()
		envByte, err := fileOp.GetContent(appInstall.GetEnvPath())
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}
		composeByte, err := fileOp.GetContent(dockerComposePath)
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}

		project, err := composeV2.GetComposeProject(appInstall.Name, appInstall.GetPath(), composeByte, envByte, true)
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}
		if err = composeV2.UpComposeProject(project); err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}

		appInstall.Status = constant.Running
		_ = appInstallRepo.Save(context.Background(), &appInstall)
	}()
	return nil
}

func getAppDetails(details []model.AppDetail, versions []dto.AppConfigVersion) map[string]model.AppDetail {
	appDetails := make(map[string]model.AppDetail, len(details))
	for _, old := range details {
		old.Status = constant.AppTakeDown
		appDetails[old.Version] = old
	}
	for _, v := range versions {
		version := v.Name
		detail, ok := appDetails[version]
		if ok {
			detail.Status = constant.AppNormal
			appDetails[version] = detail
		} else {
			appDetails[version] = model.AppDetail{
				Version: version,
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
		config := item.AppProperty
		key := config.Key
		app, ok := apps[key]
		if !ok {
			app = model.App{}
		}
		app.Resource = constant.AppResourceRemote
		app.Name = item.Name
		app.Limit = config.Limit
		app.Key = key
		app.ShortDescZh = config.ShortDescZh
		app.ShortDescEn = config.ShortDescEn
		app.Website = config.Website
		app.Document = config.Document
		app.Github = config.Github
		app.Type = config.Type
		app.CrossVersionUpdate = config.CrossVersionUpdate
		app.Status = constant.AppNormal
		app.LastModified = item.LastModified
		app.ReadMe = item.ReadMe
		apps[key] = app
	}
	return apps
}

func handleLocalAppDetail(versionDir string, appDetail *model.AppDetail) error {
	fileOp := files.NewFileOp()
	dockerComposePath := path.Join(versionDir, "docker-compose.yml")
	if !fileOp.Stat(dockerComposePath) {
		return buserr.WithName(constant.ErrFileNotFound, "docker-compose.yml")
	}
	dockerComposeByte, _ := fileOp.GetContent(dockerComposePath)
	if dockerComposeByte == nil {
		return buserr.WithName(constant.ErrFileParseApp, "docker-compose.yml")
	}
	appDetail.DockerCompose = string(dockerComposeByte)
	paramPath := path.Join(versionDir, "data.yml")
	if !fileOp.Stat(paramPath) {
		return buserr.WithName(constant.ErrFileNotFound, "data.yml")
	}
	paramByte, _ := fileOp.GetContent(paramPath)
	if paramByte == nil {
		return buserr.WithName(constant.ErrFileNotFound, "data.yml")
	}
	appParamConfig := dto.LocalAppParam{}
	if err := yaml.Unmarshal(paramByte, &appParamConfig); err != nil {
		return buserr.WithMap(constant.ErrFileParseApp, map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
	}
	dataJson, err := json.Marshal(appParamConfig.AppParams)
	if err != nil {
		return buserr.WithMap(constant.ErrFileParseApp, map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
	}
	appDetail.Params = string(dataJson)
	return nil
}

func handleLocalApp(appDir string) (app *model.App, err error) {
	fileOp := files.NewFileOp()
	configYamlPath := path.Join(appDir, "data.yml")
	if !fileOp.Stat(configYamlPath) {
		err = buserr.WithName(constant.ErrFileNotFound, "data.yml")
		return
	}
	iconPath := path.Join(appDir, "logo.png")
	if !fileOp.Stat(iconPath) {
		err = buserr.WithName(constant.ErrFileNotFound, "logo.png")
		return
	}
	configYamlByte, err := fileOp.GetContent(configYamlPath)
	if err != nil {
		err = buserr.WithMap(constant.ErrFileParseApp, map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
		return
	}
	localAppDefine := dto.LocalAppAppDefine{}
	if err = yaml.Unmarshal(configYamlByte, &localAppDefine); err != nil {
		err = buserr.WithMap(constant.ErrFileParseApp, map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
		return
	}
	app = &localAppDefine.AppProperty
	app.Resource = constant.AppResourceLocal
	app.Status = constant.AppNormal
	app.Recommend = 9999
	app.TagsKey = append(app.TagsKey, "Local")
	app.Key = "local" + app.Key
	readMePath := path.Join(appDir, "README.md")
	readMeByte, err := fileOp.GetContent(readMePath)
	if err == nil {
		app.ReadMe = string(readMeByte)
	}
	iconByte, _ := fileOp.GetContent(iconPath)
	if iconByte != nil {
		iconStr := base64.StdEncoding.EncodeToString(iconByte)
		app.Icon = iconStr
	}
	return
}

func handleErr(install model.AppInstall, err error, out string) error {
	reErr := err
	install.Message = err.Error()
	if out != "" {
		install.Message = out
		reErr = errors.New(out)
	}
	install.Status = constant.Error
	_ = appInstallRepo.Save(context.Background(), &install)
	return reErr
}

func handleInstalled(appInstallList []model.AppInstall, updated bool) ([]response.AppInstalledDTO, error) {
	var res []response.AppInstalledDTO
	for _, installed := range appInstallList {
		if updated && (installed.App.Type == "php" || installed.Status == constant.Installing || (installed.App.Key == constant.AppMysql && installed.Version == "5.6.51")) {
			continue
		}
		if err := syncAppInstallStatus(&installed); err != nil {
			global.LOG.Error("sync app install status error : ", err)
		}

		installDTO := response.AppInstalledDTO{
			AppInstall: installed,
			Path:       installed.GetPath(),
		}
		app, err := appRepo.GetFirst(commonRepo.WithByID(installed.AppId))
		if err != nil {
			return nil, err
		}
		details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(app.ID))
		if err != nil {
			return nil, err
		}
		var versions []string
		for _, detail := range details {
			if detail.IgnoreUpgrade || installed.Version == "latest" {
				continue
			}
			if common.IsCrossVersion(installed.Version, detail.Version) && !app.CrossVersionUpdate {
				continue
			}
			versions = append(versions, detail.Version)
		}
		versions = common.GetSortedVersions(versions)
		if len(versions) == 0 {
			if !updated {
				installDTO.CanUpdate = false
				res = append(res, installDTO)
			}
			continue
		}
		lastVersion := versions[0]
		if common.IsCrossVersion(installed.Version, lastVersion) {
			installDTO.CanUpdate = app.CrossVersionUpdate
		} else {
			installDTO.CanUpdate = common.CompareVersion(lastVersion, installed.Version)
		}
		if updated {
			if installDTO.CanUpdate {
				res = append(res, installDTO)
			}
		} else {
			res = append(res, installDTO)
		}
	}
	return res, nil
}

func getAppInstallByKey(key string) (model.AppInstall, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return model.AppInstall{}, err
	}
	appInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return model.AppInstall{}, err
	}
	return appInstall, nil
}

func getAppInstallPort(key string) (httpPort, httpsPort int, err error) {
	install, err := getAppInstallByKey(key)
	if err != nil {
		return
	}
	httpPort = install.HttpPort
	httpsPort = install.HttpsPort
	return
}

func updateToolApp(installed *model.AppInstall) {
	tooKey, ok := dto.AppToolMap[installed.App.Key]
	if !ok {
		return
	}
	toolInstall, _ := getAppInstallByKey(tooKey)
	if reflect.DeepEqual(toolInstall, model.AppInstall{}) {
		return
	}
	paramMap := make(map[string]string)
	_ = json.Unmarshal([]byte(installed.Param), &paramMap)
	envMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(toolInstall.Env), &envMap)
	if password, ok := paramMap["PANEL_DB_ROOT_PASSWORD"]; ok {
		envMap["PANEL_DB_ROOT_PASSWORD"] = password
	}
	if _, ok := envMap["PANEL_REDIS_HOST"]; ok {
		envMap["PANEL_REDIS_HOST"] = installed.ServiceName
	}
	if _, ok := envMap["PANEL_DB_HOST"]; ok {
		envMap["PANEL_DB_HOST"] = installed.ServiceName
	}

	envPath := path.Join(toolInstall.GetPath(), ".env")
	contentByte, err := json.Marshal(envMap)
	if err != nil {
		global.LOG.Errorf("update tool app [%s] error : %s", toolInstall.Name, err.Error())
		return
	}
	envFileMap := make(map[string]string)
	handleMap(envMap, envFileMap)
	if err = env.Write(envFileMap, envPath); err != nil {
		global.LOG.Errorf("update tool app [%s] error : %s", toolInstall.Name, err.Error())
		return
	}
	toolInstall.Env = string(contentByte)
	if err := appInstallRepo.Save(context.Background(), &toolInstall); err != nil {
		global.LOG.Errorf("update tool app [%s] error : %s", toolInstall.Name, err.Error())
		return
	}
	if out, err := compose.Down(toolInstall.GetComposePath()); err != nil {
		global.LOG.Errorf("update tool app [%s] error : %s", toolInstall.Name, out)
		return
	}
	if out, err := compose.Up(toolInstall.GetComposePath()); err != nil {
		global.LOG.Errorf("update tool app [%s] error : %s", toolInstall.Name, out)
		return
	}
}

func addDockerComposeCommonParam(composeMap map[string]interface{}, serviceName string, req request.AppContainerConfig, params map[string]interface{}) error {
	services, serviceValid := composeMap["services"].(map[string]interface{})
	if !serviceValid {
		return buserr.New(constant.ErrFileParse)
	}
	service, serviceExist := services[serviceName]
	if !serviceExist {
		return buserr.New(constant.ErrFileParse)
	}
	serviceValue := service.(map[string]interface{})
	deploy := map[string]interface{}{
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpus":   "${CPUS}",
				"memory": "${MEMORY_LIMIT}",
			},
		},
	}
	serviceValue["deploy"] = deploy

	ports, ok := serviceValue["ports"].([]interface{})
	if ok {
		for i, port := range ports {
			portStr, portOK := port.(string)
			if !portOK {
				continue
			}
			portArray := strings.Split(portStr, ":")
			if len(portArray) == 2 {
				portArray = append([]string{"${HOST_IP}"}, portArray...)
			}
			ports[i] = strings.Join(portArray, ":")
		}
		serviceValue["ports"] = ports
	}

	params[constant.CPUS] = "0"
	params[constant.MemoryLimit] = "0"
	if req.Advanced {
		if req.CpuQuota > 0 {
			params[constant.CPUS] = req.CpuQuota
		}
		if req.MemoryLimit > 0 {
			params[constant.MemoryLimit] = strconv.FormatFloat(req.MemoryLimit, 'f', -1, 32) + req.MemoryUnit
		}
	}
	_, portExist := serviceValue["ports"].([]interface{})
	if portExist {
		allowHost := "127.0.0.1"
		if req.Advanced && req.AllowPort {
			allowHost = ""
		}
		params[constant.HostIP] = allowHost
	}
	services[serviceName] = serviceValue
	return nil
}

func getAppCommonConfig(envs map[string]interface{}) request.AppContainerConfig {
	config := request.AppContainerConfig{}

	if hostIp, ok := envs[constant.HostIP]; ok {
		config.AllowPort = hostIp.(string) != "127.0.0.1"
	} else {
		config.AllowPort = true
	}
	if cpuCore, ok := envs[constant.CPUS]; ok {
		numStr, ok := cpuCore.(string)
		if ok {
			num, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				config.CpuQuota = num
			}
		} else {
			num64, flOk := cpuCore.(float64)
			if flOk {
				config.CpuQuota = num64
			}
		}
	} else {
		config.CpuQuota = 0
	}
	if memLimit, ok := envs[constant.MemoryLimit]; ok {
		re := regexp.MustCompile(`(\d+)([A-Za-z]+)`)
		matches := re.FindStringSubmatch(memLimit.(string))
		if len(matches) == 3 {
			num, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				unit := matches[2]
				config.MemoryLimit = num
				config.MemoryUnit = unit
			}
		}
	} else {
		config.MemoryLimit = 0
		config.MemoryUnit = "M"
	}

	if containerName, ok := envs[constant.ContainerName]; ok {
		config.ContainerName = containerName.(string)
	}

	return config
}

func isHostModel(dockerCompose string) bool {
	composeMap := make(map[string]interface{})
	_ = yaml.Unmarshal([]byte(dockerCompose), &composeMap)
	services, serviceValid := composeMap["services"].(map[string]interface{})
	if !serviceValid {
		return false
	}
	for _, service := range services {
		serviceValue := service.(map[string]interface{})
		if value, ok := serviceValue["network_mode"]; ok && value == "host" {
			return true
		}
	}
	return false
}
