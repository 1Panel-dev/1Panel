package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/1Panel-dev/1Panel/agent/app/task"

	"github.com/docker/docker/api/types"

	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"github.com/docker/docker/api/types/container"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/agent/utils/env"

	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	composeV2 "github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
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

		oldInstalled, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithPort(portN))
		if len(oldInstalled) > 0 {
			var apps []string
			for _, install := range oldInstalled {
				apps = append(apps, install.App.Name)
			}
			return portN, buserr.WithMap("ErrPortInOtherApp", map[string]interface{}{"port": portN, "apps": apps}, nil)
		}
		if common.ScanPort(portN) {
			return portN, buserr.WithDetail("ErrPortInUsed", portN, nil)
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
	runtime, _ := runtimeRepo.GetFirst(context.Background(), runtimeRepo.WithPort(port))
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
		return buserr.WithDetail("ErrPortInUsed", port, nil)
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

func createLink(ctx context.Context, installTask *task.Task, app model.App, appInstall *model.AppInstall, params map[string]interface{}) error {
	deleteAppLink := func(t *task.Task) {
		del := dto.DelAppLink{
			Ctx:         ctx,
			Install:     appInstall,
			ForceDelete: true,
		}
		_ = deleteLink(del)
	}
	var dbConfig dto.AppDatabase
	if DatabaseKeys[app.Key] > 0 {
		handleDataBaseApp := func(task *task.Task) error {
			database := &model.Database{
				AppInstallID: appInstall.ID,
				Name:         appInstall.Name,
				Type:         app.Key,
				Version:      appInstall.Version,
				From:         "local",
				Address:      appInstall.ServiceName,
				Port:         DatabaseKeys[app.Key],
			}
			detail, err := appDetailRepo.GetFirst(repo.WithByID(appInstall.AppDetailId))
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
					authParam := dto.RedisAuthParam{
						RootPassword: "",
					}
					if password != "" {
						authParam.RootPassword = password.(string)
						database.Password = password.(string)
					}
					authByte, err := json.Marshal(authParam)
					if err != nil {
						return err
					}
					appInstall.Param = string(authByte)
				}
			}
			return databaseRepo.Create(ctx, database)
		}
		installTask.AddSubTask(i18n.GetMsgByKey("HandleDatabaseApp"), handleDataBaseApp, deleteAppLink)
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
		createAppDataBase := func(rootTask *task.Task) error {
			hostName := params["PANEL_DB_HOST_NAME"]
			if hostName == nil || hostName.(string) == "" {
				return nil
			}
			database, _ := databaseRepo.Get(repo.WithByName(hostName.(string)))
			if database.ID == 0 {
				return nil
			}
			var resourceId uint
			if dbConfig.DbName != "" && dbConfig.DbUser != "" && dbConfig.Password != "" {
				switch database.Type {
				case constant.AppPostgresql, constant.AppPostgres:
					oldPostgresqlDb, _ := postgresqlRepo.Get(repo.WithByName(dbConfig.DbName), repo.WithByFrom(constant.ResourceLocal))
					resourceId = oldPostgresqlDb.ID
					if oldPostgresqlDb.ID > 0 {
						if oldPostgresqlDb.Username != dbConfig.DbUser || oldPostgresqlDb.Password != dbConfig.Password {
							return buserr.New("ErrDbUserNotValid")
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
					oldMysqlDb, _ := mysqlRepo.Get(repo.WithByName(dbConfig.DbName), repo.WithByFrom(constant.ResourceLocal))
					resourceId = oldMysqlDb.ID
					if oldMysqlDb.ID > 0 {
						if oldMysqlDb.Username != dbConfig.DbUser || oldMysqlDb.Password != dbConfig.Password {
							return buserr.New("ErrDbUserNotValid")
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
			return appInstallResourceRepo.Create(ctx, &installResource)
		}
		installTask.AddSubTask(task.GetTaskName(dbConfig.DbName, task.TaskCreate, task.TaskScopeDatabase), createAppDataBase, deleteAppLink)
	}
	return nil
}

func deleteAppInstall(deleteReq request.AppInstallDelete) error {
	install := deleteReq.Install
	op := files.NewFileOp()
	appDir := install.GetPath()

	uninstallTask, err := task.NewTaskWithOps(install.Name, task.TaskUninstall, task.TaskScopeApp, deleteReq.TaskID, install.ID)
	if err != nil {
		return err
	}

	uninstall := func(t *task.Task) error {
		install.Status = constant.StatusUninstalling
		_ = appInstallRepo.Save(context.Background(), &install)
		dir, _ := os.Stat(appDir)
		if dir != nil {
			logStr := i18n.GetMsgByKey("Stop") + i18n.GetMsgByKey("App")
			t.Log(logStr)

			out, err := compose.Down(install.GetComposePath())
			if err != nil && !deleteReq.ForceDelete {
				return handleErr(install, err, out)
			}
			t.LogSuccess(logStr)
			if err = runScript(t, &install, "uninstall"); err != nil {
				_, _ = compose.Up(install.GetComposePath())
				return err
			}
			if deleteReq.DeleteImage {
				delImageStr := i18n.GetMsgByKey("TaskDelete") + i18n.GetMsgByKey("Image")
				content, err := op.GetContent(install.GetEnvPath())
				if err != nil {
					return err
				}
				images, err := composeV2.GetDockerComposeImagesV2(content, []byte(install.DockerCompose))
				if err != nil {
					return err
				}
				client, err := docker.NewClient()
				if err != nil {
					return err
				}
				defer client.Close()
				for _, image := range images {
					imageID, err := client.GetImageIDByName(image)
					if err == nil {
						imgStr := delImageStr + image
						t.Log(imgStr)

						if err = client.DeleteImage(imageID); err != nil {
							t.LogFailedWithErr(imgStr, err)
							continue
						}
						t.LogSuccess(delImageStr + image)
					}
				}
			}
		}
		tx, ctx := helper.GetTxAndContext()
		defer tx.Rollback()
		if err = appInstallRepo.Delete(ctx, install); err != nil {
			return err
		}

		resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
		if deleteReq.DeleteDB && len(resources) > 0 {
			del := dto.DelAppLink{
				Ctx:         ctx,
				Install:     &install,
				ForceDelete: deleteReq.ForceDelete,
				Task:        uninstallTask,
			}
			t.LogWithOps(task.TaskDelete, i18n.GetMsgByKey("Database"))
			if err = deleteLink(del); err != nil {
				t.LogFailedWithOps(task.TaskDelete, i18n.GetMsgByKey("Database"), err)
				if !deleteReq.ForceDelete {
					return err
				}
			}
			t.LogSuccessWithOps(task.TaskDelete, i18n.GetMsgByKey("Database"))
		}

		if DatabaseKeys[install.App.Key] > 0 {
			_ = databaseRepo.Delete(ctx, databaseRepo.WithAppInstallID(install.ID))
		}

		switch install.App.Key {
		case constant.AppMysql, constant.AppMariaDB:
			_ = mysqlRepo.Delete(ctx, mysqlRepo.WithByMysqlName(install.Name))
		case constant.AppPostgresql:
			_ = postgresqlRepo.Delete(ctx, postgresqlRepo.WithByPostgresqlName(install.Name))
		}

		_ = backupRepo.DeleteRecord(ctx, repo.WithByType("app"), repo.WithByName(install.App.Key), repo.WithByDetailName(install.Name))
		uploadDir := path.Join(global.Dir.BaseDir, fmt.Sprintf("1panel/uploads/app/%s/%s", install.App.Key, install.Name))
		if _, err := os.Stat(uploadDir); err == nil {
			_ = os.RemoveAll(uploadDir)
		}
		if deleteReq.DeleteBackup {
			backupDir := path.Join(global.Dir.LocalBackupDir, fmt.Sprintf("app/%s/%s", install.App.Key, install.Name))
			if _, err = os.Stat(backupDir); err == nil {
				t.LogWithOps(task.TaskDelete, i18n.GetMsgByKey("TaskBackup"))
				_ = os.RemoveAll(backupDir)
				t.LogSuccessWithOps(task.TaskDelete, i18n.GetMsgByKey("TaskBackup"))
			}
		}
		_ = op.DeleteDir(appDir)
		parentDir := filepath.Dir(appDir)
		entries, err := os.ReadDir(parentDir)
		if err == nil && len(entries) == 0 {
			_ = op.DeleteDir(parentDir)
		}
		tx.Commit()
		return nil
	}
	uninstallTask.AddSubTask(task.GetTaskName(install.Name, task.TaskUninstall, task.TaskScopeApp), uninstall, nil)
	go func() {
		if err := uninstallTask.Execute(); err != nil && !deleteReq.ForceDelete {
			install.Status = constant.StatusError
			_ = appInstallRepo.Save(context.Background(), &install)
		}
	}()
	return nil
}

func deleteLink(del dto.DelAppLink) error {
	install := del.Install
	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	if len(resources) == 0 {
		return nil
	}
	for _, re := range resources {
		switch re.Key {
		case constant.AppMysql, constant.AppMariaDB:
			mysqlService := NewIMysqlService()
			database, _ := mysqlRepo.Get(repo.WithByID(re.ResourceId))
			if reflect.DeepEqual(database, model.DatabaseMysql{}) {
				continue
			}
			if err := mysqlService.Delete(del.Ctx, dto.MysqlDBDelete{
				ID:           database.ID,
				ForceDelete:  del.ForceDelete,
				DeleteBackup: true,
				Type:         re.Key,
				Database:     database.MysqlName,
			}); err != nil && !del.ForceDelete {
				return err
			}
		case constant.AppPostgresql:
			pgsqlService := NewIPostgresqlService()
			database, _ := postgresqlRepo.Get(repo.WithByID(re.ResourceId))
			if reflect.DeepEqual(database, model.DatabasePostgresql{}) {
				continue
			}
			if err := pgsqlService.Delete(del.Ctx, dto.PostgresqlDBDelete{
				ID:           database.ID,
				ForceDelete:  del.ForceDelete,
				DeleteBackup: true,
				Type:         re.Key,
				Database:     database.PostgresqlName,
			}); err != nil {
				return err
			}
		}
	}
	return appInstallResourceRepo.DeleteBy(del.Ctx, appInstallResourceRepo.WithAppInstallId(install.ID))
}

func getUpgradeCompose(install model.AppInstall, detail model.AppDetail) (string, error) {
	if detail.DockerCompose == "" {
		return "", nil
	}
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(detail.DockerCompose), &composeMap); err != nil {
		return "", err
	}
	value, ok := composeMap["services"]
	if !ok || value == nil {
		return "", buserr.New("ErrFileParse")
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
	if err := json.Unmarshal([]byte(install.Env), &envs); err != nil {
		return "", err
	}
	config := getAppCommonConfig(envs)
	if config.ContainerName == "" {
		config.ContainerName = install.ContainerName
		envs[constant.ContainerName] = install.ContainerName
	}
	config.Advanced = true
	if err := addDockerComposeCommonParam(composeMap, install.ServiceName, config, envs); err != nil {
		return "", err
	}
	paramByte, err := json.Marshal(envs)
	if err != nil {
		return "", err
	}
	install.Env = string(paramByte)
	composeByte, err := yaml.Marshal(composeMap)
	if err != nil {
		return "", err
	}
	return string(composeByte), nil
}

func upgradeInstall(req request.AppInstallUpgrade) error {
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.InstallID))
	if err != nil {
		return err
	}
	detail, err := appDetailRepo.GetFirst(repo.WithByID(req.DetailID))
	if err != nil {
		return err
	}
	if install.Version == detail.Version {
		return errors.New("two version is same")
	}

	upgradeTask, err := task.NewTaskWithOps(install.Name, task.TaskUpgrade, task.TaskScopeApp, req.TaskID, install.ID)
	if err != nil {
		return err
	}
	install.Status = constant.StatusUpgrading

	var (
		upErr      error
		backupFile string
	)
	backUpApp := func(t *task.Task) error {
		if req.Backup {
			backupService := NewIBackupService()
			backupRecordService := NewIBackupRecordService()
			fileName := fmt.Sprintf("upgrade_backup_%s_%s.tar.gz", install.Name, time.Now().Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5))
			backupRecord, err := backupService.AppBackup(dto.CommonBackup{Name: install.App.Key, DetailName: install.Name, FileName: fileName})
			if err == nil {
				backups, _ := backupRecordService.ListAppRecords(install.App.Key, install.Name, "upgrade_backup")
				if len(backups) > 3 {
					backupsToDelete := backups[:len(backups)-3]
					var deleteIDs []uint
					for _, backup := range backupsToDelete {
						deleteIDs = append(deleteIDs, backup.ID)
					}
					_ = backupRecordService.BatchDeleteRecord(deleteIDs)
				}
				backupFile = path.Join(global.Dir.LocalBackupDir, backupRecord.FileDir, backupRecord.FileName)
			} else {
				return buserr.WithNameAndErr("ErrAppBackup", install.Name, err)
			}
		}
		return nil
	}
	upgradeTask.AddSubTask(task.GetTaskName(install.Name, task.TaskBackup, task.TaskScopeApp), backUpApp, nil)

	upgradeApp := func(t *task.Task) error {
		fileOp := files.NewFileOp()
		detailDir := path.Join(global.Dir.ResourceDir, "apps", install.App.Resource, install.App.Key, detail.Version)
		if install.App.Resource == constant.AppResourceRemote {
			if err = downloadApp(install.App, detail, &install, t.Logger); err != nil {
				return err
			}
			if detail.DockerCompose == "" {
				composeDetail, err := fileOp.GetContent(path.Join(detailDir, "docker-compose.yml"))
				if err != nil {
					return err
				}
				detail.DockerCompose = string(composeDetail)
				_ = appDetailRepo.Update(context.Background(), detail)
			}
			go func() {
				RequestDownloadCallBack(detail.DownloadCallBackUrl)
			}()
		}
		if install.App.Resource == constant.AppResourceLocal {
			detailDir = path.Join(global.Dir.ResourceDir, "apps", "local", strings.TrimPrefix(install.App.Key, "local"), detail.Version)
		}

		content, err := fileOp.GetContent(install.GetEnvPath())
		if err != nil {
			return err
		}
		dockerCLi, err := docker.NewClient()
		if req.PullImage {
			images, err := composeV2.GetDockerComposeImagesV2(content, []byte(detail.DockerCompose))
			if err != nil {
				return err
			}
			for _, image := range images {
				t.Log(i18n.GetWithName("PullImageStart", image))
				if err = dockerCLi.PullImageWithProcess(t, image); err != nil {
					err = buserr.WithNameAndErr("ErrDockerPullImage", "", err)
					return err
				}
				t.LogSuccess(i18n.GetMsgByKey("PullImage"))
			}
		}

		command := exec.Command("/bin/bash", "-c", fmt.Sprintf("cp -rn %s/* %s || true", detailDir, install.GetPath()))
		stdout, _ := command.CombinedOutput()
		if stdout != nil {
			t.Logger.Printf("upgrade app [%s] [%s] cp file log : %s ", install.App.Key, install.Name, string(stdout))
		}
		sourceScripts := path.Join(detailDir, "scripts")
		if fileOp.Stat(sourceScripts) {
			dstScripts := path.Join(install.GetPath(), "scripts")
			_ = fileOp.DeleteDir(dstScripts)
			_ = fileOp.CreateDir(dstScripts, constant.DirPerm)
			scriptCmd := exec.Command("cp", "-rf", sourceScripts+"/.", dstScripts+"/")
			_, _ = scriptCmd.CombinedOutput()
		}

		var newCompose string
		if req.DockerCompose == "" {
			newCompose, err = getUpgradeCompose(install, detail)
			if err != nil {
				return err
			}
		} else {
			newCompose = req.DockerCompose
		}

		install.DockerCompose = newCompose
		install.Version = detail.Version
		install.AppDetailId = req.DetailID

		if out, err := compose.Down(install.GetComposePath()); err != nil {
			if out != "" {
				upErr = errors.New(out)
				return upErr
			}
			return err
		}
		envs := make(map[string]interface{})
		if err = json.Unmarshal([]byte(install.Env), &envs); err != nil {
			return err
		}
		envParams := make(map[string]string, len(envs))
		handleMap(envs, envParams)
		if install.App.Key == "openresty" && install.App.Resource == "remote" && !common.CompareVersion(install.Version, "1.25.3.2-0-1") {
			t.Log(i18n.GetMsgByKey("MoveSiteDir"))
			siteDir := path.Join(global.Dir.DataDir, "www")
			envParams["WEBSITE_DIR"] = siteDir
			oldSiteDir := path.Join(install.GetPath(), "www")
			t.Log(i18n.GetWithName("MoveSiteToDir", siteDir))
			if err := fileOp.CopyDir(oldSiteDir, global.Dir.DataDir); err != nil {
				t.Log(i18n.GetMsgByKey("ErrMoveSiteDir"))
				return err
			}
			newConfDir := path.Join(global.Dir.DataDir, "www", "conf.d")
			_ = fileOp.CreateDir(newConfDir, constant.DirPerm)
			oldConfDir := path.Join(install.GetPath(), "conf/conf.d")
			items, err := os.ReadDir(oldConfDir)
			if err != nil {
				return err
			}
			for _, item := range items {
				itemPath := path.Join(oldConfDir, item.Name())
				if item.IsDir() {
					_ = fileOp.Mv(itemPath, newConfDir)
				}
				if item.Name() != "default.conf" && item.Name() != "00.default.conf" {
					_ = fileOp.Mv(itemPath, newConfDir)
				}
			}
			nginxConfPath := path.Join(install.GetPath(), "conf", "nginx.conf")
			parse, err := parser.NewParser(nginxConfPath)
			if err != nil {
				return err
			}
			config, err := parse.Parse()
			if err != nil {
				return err
			}
			config.FilePath = nginxConfPath
			httpDirective := config.FindHttp()
			httpDirective.UpdateDirective("include", []string{"/usr/local/openresty/nginx/conf/default/*.conf"})

			if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
				return buserr.WithErr("ErrUpdateBuWebsite", err)
			}
			t.Log(i18n.GetMsgByKey("DeleteRuntimePHP"))
			_ = fileOp.DeleteDir(path.Join(global.Dir.RuntimeDir, "php"))
			websites, _ := websiteRepo.List(repo.WithByType("runtime"))
			for _, website := range websites {
				runtime, _ := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
				if runtime != nil && runtime.Type == "php" {
					website.Type = constant.Static
					website.RuntimeID = 0
					_ = websiteRepo.SaveWithoutCtx(&website)
				}
			}
			_ = runtimeRepo.DeleteBy(repo.WithByType("php"))
			t.Log(i18n.GetMsgByKey("MoveSiteDirSuccess"))
		}

		if err = env.Write(envParams, install.GetEnvPath()); err != nil {
			return err
		}

		if err = runScript(t, &install, "upgrade"); err != nil {
			return err
		}

		if err = fileOp.WriteFile(install.GetComposePath(), strings.NewReader(install.DockerCompose), constant.FilePerm); err != nil {
			return err
		}

		logStr := fmt.Sprintf("%s %s", i18n.GetMsgByKey("Run"), i18n.GetMsgByKey("App"))
		t.Log(logStr)
		if out, err := compose.Up(install.GetComposePath()); err != nil {
			if out != "" {
				return errors.New(out)
			}
			return err
		}
		t.LogSuccess(logStr)
		install.Status = constant.StatusRunning
		return appInstallRepo.Save(context.Background(), &install)
	}

	rollBackApp := func(t *task.Task) {
		if req.Backup {
			t.Log(i18n.GetWithName("AppRecover", install.Name))
			if err := NewIBackupService().AppRecover(dto.CommonRecover{Name: install.App.Key, DetailName: install.Name, Type: "app", DownloadAccountID: 1, File: backupFile}); err != nil {
				t.LogFailedWithErr(i18n.GetWithName("AppRecover", install.Name), err)
				return
			}
			t.LogSuccess(i18n.GetWithName("AppRecover", install.Name))
			return
		}
	}

	upgradeTask.AddSubTask(task.GetTaskName(install.Name, task.TaskScopeApp, task.TaskUpgrade), upgradeApp, rollBackApp)

	go func() {
		err = upgradeTask.Execute()
		if err != nil {
			existInstall, _ := appInstallRepo.GetFirst(repo.WithByID(req.InstallID))
			if existInstall.ID > 0 && existInstall.Status != constant.StatusRunning {
				existInstall.Status = constant.StatusUpgradeErr
				existInstall.Message = err.Error()
				_ = appInstallRepo.Save(context.Background(), &existInstall)
			}
		}
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
		installs, err := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(app.ID))
		if err != nil {
			return err
		}
		if len(installs) >= app.Limit {
			return buserr.New("ErrAppLimit")
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
		case map[string]interface{}:
			handleMap(t, envParams)
		}
	}
}

func downloadApp(app model.App, appDetail model.AppDetail, appInstall *model.AppInstall, logger *log.Logger) (err error) {
	if app.IsLocalApp() {
		return nil
	}
	appResourceDir := path.Join(global.Dir.AppResourceDir, app.Resource)
	appDownloadDir := app.GetAppResourcePath()
	appVersionDir := path.Join(appDownloadDir, appDetail.Version)
	fileOp := files.NewFileOp()
	if !appDetail.Update && fileOp.Stat(appVersionDir) {
		return
	}
	if !fileOp.Stat(appDownloadDir) {
		_ = fileOp.CreateDir(appDownloadDir, constant.DirPerm)
	}
	if !fileOp.Stat(appVersionDir) {
		_ = fileOp.CreateDir(appVersionDir, constant.DirPerm)
	}
	if logger == nil {
		global.LOG.Infof("download app[%s] from %s", app.Name, appDetail.DownloadUrl)
	} else {
		logger.Printf("download app[%s] from %s", app.Name, appDetail.DownloadUrl)
	}

	filePath := path.Join(appVersionDir, app.Key+"-"+appDetail.Version+".tar.gz")

	defer func() {
		if err != nil {
			if appInstall != nil {
				appInstall.Status = constant.StatusDownloadErr
				appInstall.Message = err.Error()
			}
		}
	}()

	if err = fileOp.DownloadFile(appDetail.DownloadUrl, filePath); err != nil {
		if logger == nil {
			global.LOG.Errorf("download app[%s] error %v", app.Name, err)
		} else {
			logger.Printf("download app[%s] error %v", app.Name, err)
		}
		return
	}
	if err = fileOp.Decompress(filePath, appResourceDir, files.SdkTarGz, ""); err != nil {
		if logger == nil {
			global.LOG.Errorf("decompress app[%s] error %v", app.Name, err)
		} else {
			logger.Printf("decompress app[%s] error %v", app.Name, err)
		}
		return
	}
	_ = fileOp.DeleteFile(filePath)
	appDetail.Update = false
	_ = appDetailRepo.Update(context.Background(), appDetail)
	return
}

func copyData(task *task.Task, app model.App, appDetail model.AppDetail, appInstall *model.AppInstall, req request.AppInstallCreate) (err error) {
	fileOp := files.NewFileOp()
	appResourceDir := path.Join(global.Dir.AppResourceDir, app.Resource)

	if app.Resource == constant.AppResourceRemote {
		err = downloadApp(app, appDetail, appInstall, task.Logger)
		if err != nil {
			return
		}
		go func() {
			RequestDownloadCallBack(appDetail.DownloadCallBackUrl)
		}()
	}
	appKey := app.Key
	installAppDir := path.Join(global.Dir.AppInstallDir, app.Key)
	if app.Resource == constant.AppResourceLocal {
		appResourceDir = global.Dir.LocalAppResourceDir
		appKey = strings.TrimPrefix(app.Key, "local")
		installAppDir = path.Join(global.Dir.LocalAppInstallDir, appKey)
	}
	if app.Resource == constant.AppResourceCustom {
		appResourceDir = path.Join(global.Dir.AppResourceDir, "custom")
	}
	resourceDir := path.Join(appResourceDir, appKey, appDetail.Version)

	if !fileOp.Stat(installAppDir) {
		if err = fileOp.CreateDir(installAppDir, constant.DirPerm); err != nil {
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
	if err := fileOp.WriteFile(appInstall.GetComposePath(), strings.NewReader(appInstall.DockerCompose), constant.DirPerm); err != nil {
		return err
	}
	return
}

func runScript(task *task.Task, appInstall *model.AppInstall, operate string) error {
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
	logStr := i18n.GetWithName("ExecShell", operate)
	task.LogStart(logStr)
	out, err := cmd.ExecScript(scriptPath, workDir)
	if err != nil {
		if out != "" {
			err = errors.New(out)
		}
		task.LogFailedWithErr(logStr, err)
		return err
	}
	task.LogSuccess(logStr)
	return nil
}

func checkContainerNameIsExist(containerName, appDir string) (bool, error) {
	client, err := composeV2.NewDockerClient()
	if err != nil {
		return false, err
	}
	defer client.Close()
	var options container.ListOptions
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

func upApp(task *task.Task, appInstall *model.AppInstall, pullImages bool) error {
	upProject := func(appInstall *model.AppInstall) (err error) {
		var (
			out    string
			errMsg string
		)
		if pullImages && appInstall.App.Type != "php" {
			projectName := strings.ToLower(appInstall.Name)
			envByte, err := files.NewFileOp().GetContent(appInstall.GetEnvPath())
			if err != nil {
				return err
			}
			images, err := composeV2.GetDockerComposeImages(projectName, envByte, []byte(appInstall.DockerCompose))
			if err != nil {
				return err
			}
			imagePrefix := xpack.GetImagePrefix()
			dockerCLi, err := docker.NewClient()
			if err != nil {
				return err
			}
			for _, image := range images {
				if imagePrefix != "" {
					lastSlashIndex := strings.LastIndex(image, "/")
					if lastSlashIndex != -1 {
						image = image[lastSlashIndex+1:]
					}
					image = imagePrefix + "/" + image
				}

				task.Log(i18n.GetWithName("PullImageStart", image))
				if err = dockerCLi.PullImageWithProcess(task, image); err != nil {
					errOur := err.Error()
					if errOur != "" {
						if strings.Contains(errOur, "no such host") {
							errMsg = i18n.GetMsgByKey("ErrNoSuchHost") + ":"
						}
						if strings.Contains(errOur, "timeout") {
							errMsg = i18n.GetMsgByKey("ErrImagePullTimeOut") + ":"
						}
					}
					appInstall.Message = errMsg + errOur
					installErr := errors.New(appInstall.Message)
					task.LogFailedWithErr(i18n.GetMsgByKey("PullImage"), installErr)
					return installErr
				} else {
					task.Log(i18n.GetMsgByKey("PullImageSuccess"))
				}
			}
		}
		logStr := fmt.Sprintf("%s %s", i18n.GetMsgByKey("Run"), i18n.GetMsgByKey("App"))
		task.Log(logStr)
		out, err = compose.Up(appInstall.GetComposePath())
		if err != nil {
			if out != "" {
				appInstall.Message = errMsg + out
				err = errors.New(out)
			}
			task.LogFailedWithErr(logStr, err)
			return err
		}
		task.LogSuccess(logStr)
		return
	}
	exist, _ := appInstallRepo.GetFirst(repo.WithByID(appInstall.ID))
	if exist.ID > 0 {
		containerNames, err := getContainerNames(*appInstall)
		if err == nil {
			if len(containerNames) > 0 {
				appInstall.ContainerName = strings.Join(containerNames, ",")
			}
			_ = appInstallRepo.Save(context.Background(), appInstall)
		}
	}
	if err := upProject(appInstall); err != nil {
		if appInstall.Message == "" {
			appInstall.Message = err.Error()
		}
		appInstall.Status = constant.StatusUpErr
		_ = appInstallRepo.Save(context.Background(), appInstall)
		return err
	} else {
		appInstall.Status = constant.StatusRunning
		_ = appInstallRepo.Save(context.Background(), appInstall)
		return nil
	}
}

func rebuildApp(appInstall model.AppInstall) error {
	appInstall.Status = constant.StatusRebuilding
	_ = appInstallRepo.Save(context.Background(), &appInstall)
	go func() {
		dockerComposePath := appInstall.GetComposePath()
		out, err := compose.Down(dockerComposePath)
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}
		out, err = compose.Up(appInstall.GetComposePath())
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}
		containerNames, err := getContainerNames(appInstall)
		if err != nil {
			_ = handleErr(appInstall, err, out)
			return
		}
		appInstall.ContainerName = strings.Join(containerNames, ",")

		appInstall.Status = constant.StatusRunning
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

func getApps(oldApps []model.App, items []dto.AppDefine, systemVersion string, task *task.Task) map[string]model.App {
	apps := make(map[string]model.App, len(oldApps))
	for _, old := range oldApps {
		old.Status = constant.AppTakeDown
		apps[old.Key] = old
	}
	for _, item := range items {
		if item.AppProperty.Version > 0 && common.CompareVersion(strconv.FormatFloat(item.AppProperty.Version, 'f', -1, 64), systemVersion) {
			task.Log(i18n.GetWithName("AppVersionNotMatch", item.Name))
			continue
		}
		config := item.AppProperty
		key := config.Key
		app, ok := apps[key]
		if !ok {
			app = model.App{}
		}
		app.RequiredPanelVersion = config.Version
		app.Resource = constant.AppResourceRemote
		app.Name = item.Name
		app.Limit = config.Limit
		app.Key = key
		app.ShortDescZh = config.ShortDescZh
		app.ShortDescEn = config.ShortDescEn
		description, _ := json.Marshal(config.Description)
		app.Description = string(description)
		app.Website = config.Website
		app.Document = config.Document
		app.Github = config.Github
		app.Type = config.Type
		app.CrossVersionUpdate = config.CrossVersionUpdate
		app.Status = constant.AppNormal
		app.LastModified = item.LastModified
		app.ReadMe = item.ReadMe
		app.MemoryRequired = config.MemoryRequired
		app.Architectures = strings.Join(config.Architectures, ",")
		app.GpuSupport = config.GpuSupport
		apps[key] = app
	}
	return apps
}

func handleLocalAppDetail(versionDir string, appDetail *model.AppDetail) error {
	fileOp := files.NewFileOp()
	dockerComposePath := path.Join(versionDir, "docker-compose.yml")
	if !fileOp.Stat(dockerComposePath) {
		return buserr.WithName("ErrFileNotFound", "docker-compose.yml")
	}
	dockerComposeByte, _ := fileOp.GetContent(dockerComposePath)
	if dockerComposeByte == nil {
		return buserr.WithName("ErrFileParseApp", "docker-compose.yml")
	}
	appDetail.DockerCompose = string(dockerComposeByte)
	paramPath := path.Join(versionDir, "data.yml")
	if !fileOp.Stat(paramPath) {
		return buserr.WithName("ErrFileNotFound", "data.yml")
	}
	paramByte, _ := fileOp.GetContent(paramPath)
	if paramByte == nil {
		return buserr.WithName("ErrFileNotFound", "data.yml")
	}
	appParamConfig := dto.LocalAppParam{}
	if err := yaml.Unmarshal(paramByte, &appParamConfig); err != nil {
		return buserr.WithMap("ErrFileParseApp", map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
	}
	dataJson, err := json.Marshal(appParamConfig.AppParams)
	if err != nil {
		return buserr.WithMap("ErrFileParseApp", map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
	}
	var appParam dto.AppForm
	if err = json.Unmarshal(dataJson, &appParam); err != nil {
		return buserr.WithMap("ErrFileParseApp", map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
	}
	for _, formField := range appParam.FormFields {
		if strings.Contains(formField.EnvKey, " ") {
			return buserr.WithName("ErrAppParamKey", formField.EnvKey)
		}
	}

	var dataMap map[string]interface{}
	err = yaml.Unmarshal(paramByte, &dataMap)
	if err != nil {
		return buserr.WithMap("ErrFileParseApp", map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
	}

	additionalProperties, _ := dataMap["additionalProperties"].(map[string]interface{})
	formFieldsInterface, ok := additionalProperties["formFields"]
	if ok {
		formFields, ok := formFieldsInterface.([]interface{})
		if !ok {
			return buserr.WithName("ErrAppParamKey", "formFields")
		}
		for _, item := range formFields {
			field := item.(map[string]interface{})
			for key, value := range field {
				if value == nil {
					return buserr.WithName("ErrAppParamKey", key)
				}
			}
		}
	}

	appDetail.Params = string(dataJson)
	return nil
}

func handleLocalApp(appDir string) (app *model.App, err error) {
	fileOp := files.NewFileOp()
	configYamlPath := path.Join(appDir, "data.yml")
	if !fileOp.Stat(configYamlPath) {
		err = buserr.WithName("ErrFileNotFound", "data.yml")
		return
	}
	iconPath := path.Join(appDir, "logo.png")
	if !fileOp.Stat(iconPath) {
		err = buserr.WithName("ErrFileNotFound", "logo.png")
		return
	}
	configYamlByte, err := fileOp.GetContent(configYamlPath)
	if err != nil {
		err = buserr.WithMap("ErrFileParseApp", map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
		return
	}
	localAppDefine := dto.LocalAppAppDefine{}
	if err = yaml.Unmarshal(configYamlByte, &localAppDefine); err != nil {
		err = buserr.WithMap("ErrFileParseApp", map[string]interface{}{"name": "data.yml", "err": err.Error()}, err)
		return
	}
	appDefine := localAppDefine.AppProperty
	app = &model.App{}
	app.Name = appDefine.Name
	app.TagsKey = append(appDefine.Tags, "Local")
	app.Type = appDefine.Type
	app.CrossVersionUpdate = appDefine.CrossVersionUpdate
	app.Limit = appDefine.Limit
	app.Recommend = appDefine.Recommend
	app.Website = appDefine.Website
	app.Github = appDefine.Github
	app.Document = appDefine.Document
	if appDefine.ShortDescZh != "" {
		appDefine.Description.Zh = appDefine.ShortDescZh
	}
	if appDefine.ShortDescEn != "" {
		appDefine.Description.En = appDefine.ShortDescEn
	}
	desc, _ := json.Marshal(appDefine.Description)
	app.Description = string(desc)
	app.Key = "local" + appDefine.Key

	app.Resource = constant.AppResourceLocal
	app.Status = constant.AppNormal
	app.Recommend = 9999
	readMeByte, err := fileOp.GetContent(path.Join(appDir, "README.md"))
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
	install.Status = constant.StatusUpErr
	_ = appInstallRepo.Save(context.Background(), &install)
	return reErr
}

func doNotNeedSync(installed model.AppInstall) bool {
	return installed.Status == constant.StatusInstalling || installed.Status == constant.StatusRebuilding || installed.Status == constant.StatusUpgrading ||
		installed.Status == constant.StatusSyncing || installed.Status == constant.StatusUninstalling || installed.Status == constant.StatusInstallErr
}

func synAppInstall(containers map[string]types.Container, appInstall *model.AppInstall, force bool) {
	oldStatus := appInstall.Status
	containerNames := strings.Split(appInstall.ContainerName, ",")
	if len(containers) == 0 {
		if appInstall.Status == constant.StatusUpErr && !force {
			return
		}
		appInstall.Status = constant.StatusError
		appInstall.Message = buserr.WithName("ErrContainerNotFound", strings.Join(containerNames, ",")).Error()
		_ = appInstallRepo.Save(context.Background(), appInstall)
		return
	}
	notFoundNames := make([]string, 0)
	exitNames := make([]string, 0)
	exitedCount := 0
	pausedCount := 0
	runningCount := 0
	restartingCount := 0
	total := len(containerNames)
	for _, name := range containerNames {
		if con, ok := containers["/"+name]; ok {
			switch con.State {
			case "exited":
				exitedCount++
				exitNames = append(exitNames, name)
			case "running":
				runningCount++
			case "restarting":
				restartingCount++
			case "paused":
				pausedCount++
			}
		} else {
			notFoundNames = append(notFoundNames, name)
		}
	}
	switch {
	case exitedCount == total:
		appInstall.Status = constant.StatusStopped
	case runningCount == total:
		appInstall.Status = constant.StatusRunning
		if oldStatus == constant.StatusRunning {
			return
		}
	case restartingCount == total:
		appInstall.Status = constant.StatusRestarting
	case pausedCount == total:
		appInstall.Status = constant.StatusPaused
	case len(notFoundNames) == total:
		if appInstall.Status == constant.StatusUpErr && !force {
			return
		}
		appInstall.Status = constant.StatusError
		appInstall.Message = buserr.WithName("ErrContainerNotFound", strings.Join(notFoundNames, ",")).Error()
	default:
		var msg string
		if exitedCount > 0 {
			msg = buserr.WithName("ErrContainerMsg", strings.Join(exitNames, ",")).Error()
		}
		if len(notFoundNames) > 0 {
			msg += buserr.WithName("ErrContainerNotFound", strings.Join(notFoundNames, ",")).Error()
		}
		if msg == "" {
			msg = buserr.New("ErrAppWarn").Error()
		}
		appInstall.Message = msg
		appInstall.Status = constant.StatusUnHealthy
	}
	_ = appInstallRepo.Save(context.Background(), appInstall)
}

func handleInstalled(appInstallList []model.AppInstall, updated bool, sync bool) ([]response.AppInstallDTO, error) {
	var (
		res           []response.AppInstallDTO
		containersMap map[string]types.Container
	)
	if sync {
		cli, err := docker.NewClient()
		if err != nil {
			return nil, err
		}
		defer cli.Close()
		containers, err := cli.ListAllContainers()
		if err != nil {
			return nil, err
		}
		containersMap = make(map[string]types.Container, len(containers))
		for _, contain := range containers {
			containersMap[contain.Names[0]] = contain
		}
	}

	for _, installed := range appInstallList {
		if updated && ignoreUpdate(installed) {
			continue
		}
		if sync && !doNotNeedSync(installed) {
			synAppInstall(containersMap, &installed, false)
		}

		installDTO := response.AppInstallDTO{
			ID:          installed.ID,
			Name:        installed.Name,
			AppID:       installed.AppId,
			AppDetailID: installed.AppDetailId,
			Version:     installed.Version,
			Status:      installed.Status,
			Message:     installed.Message,
			HttpPort:    installed.HttpPort,
			HttpsPort:   installed.HttpsPort,
			Icon:        installed.App.Icon,
			AppName:     installed.App.Name,
			AppKey:      installed.App.Key,
			AppType:     installed.App.Type,
			Path:        installed.GetPath(),
			CreatedAt:   installed.CreatedAt,
			WebUI:       installed.WebUI,
			App: response.AppDetail{
				Github:   installed.App.Github,
				Website:  installed.App.Website,
				Document: installed.App.Document,
			},
		}
		if updated {
			installDTO.DockerCompose = installed.DockerCompose
		}
		app, err := appRepo.GetFirst(repo.WithByID(installed.AppId))
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
		if app.Key == constant.AppMysql {
			for _, version := range versions {
				majorVersion := getMajorVersion(installed.Version)
				if !strings.HasPrefix(version, majorVersion) {
					continue
				} else {
					lastVersion = version
					break
				}
			}
		}
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
		return buserr.New("ErrFileParse")
	}
	imagePreFix := xpack.GetImagePrefix()
	if imagePreFix != "" {
		for _, service := range services {
			serviceValue := service.(map[string]interface{})
			if image, ok := serviceValue["image"]; ok {
				imageStr := image.(string)
				lastSlashIndex := strings.LastIndex(imageStr, "/")
				if lastSlashIndex != -1 {
					imageStr = imageStr[lastSlashIndex+1:]
				}
				imageStr = imagePreFix + "/" + imageStr
				serviceValue["image"] = imageStr
			}
		}
	}

	service, serviceExist := services[serviceName]
	if !serviceExist {
		return buserr.New("ErrFileParse")
	}
	serviceValue := service.(map[string]interface{})

	deploy := map[string]interface{}{}
	if de, ok := serviceValue["deploy"]; ok {
		deploy = de.(map[string]interface{})
	}
	resource := map[string]interface{}{}
	if res, ok := deploy["resources"]; ok {
		resource = res.(map[string]interface{})
	}
	resource["limits"] = map[string]interface{}{
		"cpus":   "${CPUS}",
		"memory": "${MEMORY_LIMIT}",
	}
	if req.GpuConfig {
		resource["reservations"] = map[string]interface{}{
			"devices": []map[string]interface{}{
				{
					"driver":       "nvidia",
					"count":        "all",
					"capabilities": []string{"gpu"},
				},
			},
		}
	}
	deploy["resources"] = resource
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

func getMajorVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}

func ignoreUpdate(installed model.AppInstall) bool {
	if installed.App.Type == "php" || installed.Status == constant.StatusInstalling {
		return true
	}
	if installed.App.Key == constant.AppMysql {
		majorVersion := getMajorVersion(installed.Version)
		appDetails, _ := appDetailRepo.GetBy(appDetailRepo.WithAppId(installed.App.ID))
		for _, appDetail := range appDetails {
			if strings.HasPrefix(appDetail.Version, majorVersion) && common.CompareVersion(appDetail.Version, installed.Version) {
				return false
			}
		}
		return true
	}
	return false
}

func RequestDownloadCallBack(downloadCallBackUrl string) {
	if downloadCallBackUrl == "" {
		return
	}
	_, _, _ = req_helper.HandleRequest(downloadCallBackUrl, http.MethodGet, constant.TimeOut5s)
}

func getAppTags(appID uint, lang string) ([]response.TagDTO, error) {
	appTags, err := appTagRepo.GetByAppId(appID)
	if err != nil {
		return nil, err
	}
	var tagIds []uint
	for _, at := range appTags {
		tagIds = append(tagIds, at.TagId)
	}
	tags, err := tagRepo.GetByIds(tagIds)
	if err != nil {
		return nil, err
	}
	var res []response.TagDTO
	for _, t := range tags {
		if t.Name != "" {
			tagDTO := response.TagDTO{
				ID:   t.ID,
				Key:  t.Key,
				Name: t.Name,
			}
			res = append(res, tagDTO)
		} else {
			var translations = make(map[string]string)
			_ = json.Unmarshal([]byte(t.Translations), &translations)
			if name, ok := translations[lang]; ok {
				tagDTO := response.TagDTO{
					ID:   t.ID,
					Key:  t.Key,
					Name: name,
				}
				res = append(res, tagDTO)
			}
		}
	}
	return res, nil
}

func handleOpenrestyFile(appInstall *model.AppInstall) error {
	websites, _ := websiteRepo.List()
	if len(websites) == 0 {
		return nil
	}
	hasDefaultWebsite := false
	for _, website := range websites {
		if website.DefaultServer {
			hasDefaultWebsite = true
			break
		}
	}
	if hasDefaultWebsite {
		installDir := appInstall.GetPath()
		defaultConfigPath := path.Join(installDir, "conf", "default", "00.default.conf")
		fileOp := files.NewFileOp()
		content, err := fileOp.GetContent(defaultConfigPath)
		if err != nil {
			return err
		}
		newContent := strings.ReplaceAll(string(content), "default_server", "")
		if err := fileOp.WriteFile(defaultConfigPath, strings.NewReader(newContent), constant.FilePerm); err != nil {
			return err
		}
	}
	return createAllWebsitesWAFConfig(websites)
}
