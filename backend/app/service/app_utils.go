package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/compose-spec/compose-go/types"
	"github.com/subosito/gotenv"
	"math"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strconv"
	"strings"

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
		portN := int(math.Ceil(port.(float64)))

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
		if authParam.RootPassword != "" {
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
		if err := json.Unmarshal(paramByte, &dbConfig); err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(dbConfig, dto.AppDatabase{}) && dbConfig.ServiceName != "" {
		dbInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithServiceName(dbConfig.ServiceName))
		if err != nil {
			return err
		}
		var resourceId uint
		if dbConfig.DbName != "" && dbConfig.DbUser != "" && dbConfig.Password != "" {
			iMysqlRepo := repo.NewIMysqlRepo()
			oldMysqlDb, _ := iMysqlRepo.Get(commonRepo.WithByName(dbConfig.DbName))
			resourceId = oldMysqlDb.ID
			if oldMysqlDb.ID > 0 {
				if oldMysqlDb.Username != dbConfig.DbUser || oldMysqlDb.Password != dbConfig.Password {
					return buserr.New(constant.ErrDbUserNotValid)
				}
			} else {
				var createMysql dto.MysqlDBCreate
				createMysql.Name = dbConfig.DbName
				createMysql.Username = dbConfig.DbUser
				createMysql.Format = "utf8mb4"
				createMysql.Permission = "%"
				createMysql.Password = dbConfig.Password
				mysqldb, err := NewIMysqlService().Create(ctx, createMysql)
				if err != nil {
					return err
				}
				resourceId = mysqldb.ID
			}
		}
		var installResource model.AppInstallResource
		installResource.ResourceId = resourceId
		installResource.AppInstallId = appInstall.ID
		installResource.LinkId = dbInstall.ID
		installResource.Key = dbInstall.App.Key
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
	if err := deleteLink(ctx, install, true, true); err != nil {
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
	}
	tx, ctx := helper.GetTxAndContext()
	defer tx.Rollback()
	if err := appInstallRepo.Delete(ctx, install); err != nil {
		return err
	}
	if err := deleteLink(ctx, &install, deleteDB, forceDelete); err != nil && !forceDelete {
		return err
	}
	_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType("app"), commonRepo.WithByName(install.App.Key), backupRepo.WithByDetailName(install.Name))
	_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType(install.App.Key))
	if install.App.Key == constant.AppMysql {
		_ = mysqlRepo.DeleteAll(ctx)
	}
	uploadDir := fmt.Sprintf("%s/1panel/uploads/app/%s/%s", global.CONF.System.BaseDir, install.App.Key, install.Name)
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	if deleteBackup {
		localDir, _ := loadLocalDir()
		backupDir := fmt.Sprintf("%s/app/%s/%s", localDir, install.App.Key, install.Name)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		global.LOG.Infof("delete app %s-%s backups successful", install.App.Key, install.Name)
	}
	_ = op.DeleteDir(appDir)
	tx.Commit()
	return nil
}

func deleteLink(ctx context.Context, install *model.AppInstall, deleteDB bool, forceDelete bool) error {
	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	if len(resources) == 0 {
		return nil
	}
	for _, re := range resources {
		mysqlService := NewIMysqlService()
		if re.Key == "mysql" && deleteDB {
			database, _ := mysqlRepo.Get(commonRepo.WithByID(re.ResourceId))
			if reflect.DeepEqual(database, model.DatabaseMysql{}) {
				continue
			}
			if err := mysqlService.Delete(ctx, dto.MysqlDBDelete{
				ID:          database.ID,
				ForceDelete: forceDelete,
			}); err != nil && !forceDelete {
				return err
			}
		}
	}
	return appInstallResourceRepo.DeleteBy(ctx, appInstallResourceRepo.WithAppInstallId(install.ID))
}

func upgradeInstall(installId uint, detailId uint) error {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
	if err != nil {
		return err
	}
	detail, err := appDetailRepo.GetFirst(commonRepo.WithByID(detailId))
	if err != nil {
		return err
	}
	if install.Version == detail.Version {
		return errors.New("two version is same")
	}
	if err := NewIBackupService().AppBackup(dto.CommonBackup{Name: install.App.Key, DetailName: install.Name}); err != nil {
		return err
	}

	detailDir := path.Join(constant.ResourceDir, "apps", install.App.Resource, install.App.Key, detail.Version)
	if install.App.Resource == constant.AppResourceLocal {
		detailDir = path.Join(constant.ResourceDir, "localApps", strings.TrimPrefix(install.App.Key, "local"), "versions", detail.Version)
	}

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cp -rf %s/* %s", detailDir, install.GetPath()))
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if stdout != nil {
			return errors.New(string(stdout))
		}
		return err
	}

	if out, err := compose.Down(install.GetComposePath()); err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	install.DockerCompose = detail.DockerCompose
	install.Version = detail.Version
	install.AppDetailId = detailId

	fileOp := files.NewFileOp()
	if err := fileOp.WriteFile(install.GetComposePath(), strings.NewReader(install.DockerCompose), 0775); err != nil {
		return err
	}
	if out, err := compose.Up(install.GetComposePath()); err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
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
	containerMap[install.ContainerName] = struct{}{}
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
		default:
			envParams[k] = t.(string)
		}
	}
}

func downloadApp(app model.App, appDetail model.AppDetail, appInstall *model.AppInstall, req request.AppInstallCreate) (err error) {
	fileOp := files.NewFileOp()
	appResourceDir := path.Join(constant.AppResourceDir, app.Resource)

	if app.Resource == constant.AppResourceRemote && appDetail.Update {
		appDownloadDir := path.Join(appResourceDir, app.Key)
		if !fileOp.Stat(appDownloadDir) {
			_ = fileOp.CreateDir(appDownloadDir, 0755)
		}
		appVersionDir := path.Join(appDownloadDir, appDetail.Version)
		if !fileOp.Stat(appVersionDir) {
			_ = fileOp.CreateDir(appVersionDir, 0755)
		}
		global.LOG.Infof("download app[%s] from %s", app.Name, appDetail.DownloadUrl)
		filePath := path.Join(appVersionDir, appDetail.Version+".tar.gz")
		if err = fileOp.DownloadFile(appDetail.DownloadUrl, filePath); err != nil {
			appInstall.Status = constant.DownloadErr
			global.LOG.Errorf("download app[%s] error %v", app.Name, err)
			return
		}
		if err = fileOp.Decompress(filePath, appVersionDir, files.TarGz); err != nil {
			global.LOG.Errorf("decompress app[%s] error %v", app.Name, err)
			appInstall.Status = constant.DownloadErr
			return
		}
		_ = fileOp.DeleteFile(filePath)
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

// 处理文件夹权限等问题
func upAppPre(app model.App, appInstall *model.AppInstall) error {
	if app.Key == "nexus" {
		dataPath := path.Join(appInstall.GetPath(), "data")
		if err := files.NewFileOp().Chown(dataPath, 200, 0); err != nil {
			return err
		}
	}
	return nil
}

func getServiceFromInstall(appInstall *model.AppInstall) (service *composeV2.ComposeService, err error) {
	var (
		project *types.Project
		envStr  string
	)
	envStr, err = coverEnvJsonToStr(appInstall.Env)
	if err != nil {
		return
	}
	project, err = composeV2.GetComposeProject(appInstall.Name, appInstall.GetPath(), []byte(appInstall.DockerCompose), []byte(envStr), true)
	if err != nil {
		return
	}
	service, err = composeV2.NewComposeService()
	if err != nil {
		return
	}
	service.SetProject(project)
	return
}

func upApp(appInstall *model.AppInstall) {
	upProject := func(appInstall *model.AppInstall) (err error) {
		if err == nil {
			var composeService *composeV2.ComposeService
			composeService, err = getServiceFromInstall(appInstall)
			if err != nil {
				return err
			}
			err = composeService.ComposeUp()
			if err != nil {
				return err
			}
			return
		} else {
			return
		}
	}
	if err := upProject(appInstall); err != nil {
		appInstall.Status = constant.Error
		appInstall.Message = err.Error()
	} else {
		appInstall.Status = constant.Running
	}
	exist, _ := appInstallRepo.GetFirst(commonRepo.WithByID(appInstall.ID))
	if exist.ID > 0 {
		_ = appInstallRepo.Save(context.Background(), appInstall)
	}
}

func rebuildApp(appInstall model.AppInstall) error {
	dockerComposePath := appInstall.GetComposePath()
	out, err := compose.Down(dockerComposePath)
	if err != nil {
		return handleErr(appInstall, err, out)
	}
	out, err = compose.Up(dockerComposePath)
	if err != nil {
		return handleErr(appInstall, err, out)
	}
	return syncById(appInstall.ID)
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

func handleErr(install model.AppInstall, err error, out string) error {
	reErr := err
	install.Message = err.Error()
	if out != "" {
		install.Message = out
		reErr = errors.New(out)
		install.Status = constant.Error
	}
	_ = appInstallRepo.Save(context.Background(), &install)
	return reErr
}

func handleInstalled(appInstallList []model.AppInstall, updated bool) ([]response.AppInstalledDTO, error) {
	var res []response.AppInstalledDTO
	for _, installed := range appInstallList {
		if updated && installed.App.Type == "php" {
			continue
		}
		installDTO := response.AppInstalledDTO{
			AppInstall: installed,
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
			versions = append(versions, detail.Version)
		}
		versions = common.GetSortedVersions(versions)
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
