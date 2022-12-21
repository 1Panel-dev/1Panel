package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"

	"github.com/1Panel-dev/1Panel/backend/app/repo"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

type AppInstallService struct {
}

func (a AppInstallService) Page(req request.AppInstalledSearch) (int64, []response.AppInstalledDTO, error) {
	var opts []repo.DBOption

	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}

	total, installs, err := appInstallRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	installDTOs, err := handleInstalled(installs)
	if err != nil {
		return 0, nil, err
	}

	return total, installDTOs, nil
}

func (a AppInstallService) CheckExist(key string) (*response.AppInstalledCheck, error) {
	res := &response.AppInstalledCheck{
		IsExist: false,
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return res, nil
	}
	res.App = app.Name
	appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID))
	if reflect.DeepEqual(appInstall, model.AppInstall{}) {
		return res, nil
	}
	if err := syncById(appInstall.ID); err != nil {
		return nil, err
	}

	res.ContainerName = appInstall.ContainerName
	res.Name = appInstall.Name
	res.Version = appInstall.Version
	res.CreatedAt = appInstall.CreatedAt
	res.Status = appInstall.Status
	res.AppInstallID = appInstall.ID
	res.IsExist = true
	if len(appInstall.Backups) > 0 {
		res.LastBackupAt = appInstall.Backups[0].CreatedAt.Format("2006-01-02 15:04:05")
	}

	return res, nil
}

func (a AppInstallService) LoadPort(key string) (int64, error) {
	app, err := appInstallRepo.LoadBaseInfo(key, "")
	if err != nil {
		return int64(0), nil
	}
	return app.Port, nil
}

func (a AppInstallService) LoadPassword(key string) (string, error) {
	app, err := appInstallRepo.LoadBaseInfo(key, "")
	if err != nil {
		return "", nil
	}
	return app.Password, nil
}

func (a AppInstallService) Search(req request.AppInstalledSearch) ([]response.AppInstalledDTO, error) {
	var installs []model.AppInstall
	var err error
	if req.Type != "" {
		apps, err := appRepo.GetBy(appRepo.WithType(req.Type))
		if err != nil {
			return nil, err
		}
		var ids []uint
		for _, app := range apps {
			ids = append(ids, app.ID)
		}
		installs, err = appInstallRepo.GetBy(appInstallRepo.WithAppIdsIn(ids), appInstallRepo.WithIdNotInWebsite())
		if err != nil {
			return nil, err
		}
	} else {
		installs, err = appInstallRepo.GetBy()
		if err != nil {
			return nil, err
		}
	}

	return handleInstalled(installs)
}

func (a AppInstallService) Operate(req request.AppInstalledOperate) error {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(req.InstallId))
	if err != nil {
		return err
	}

	dockerComposePath := install.GetComposePath()

	switch req.Operate {
	case constant.Up:
		out, err := compose.Up(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		install.Status = constant.Running
	case constant.Down:
		out, err := compose.Stop(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		install.Status = constant.Stopped
	case constant.Restart:
		out, err := compose.Restart(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		install.Status = constant.Running
	case constant.Delete:
		tx, ctx := getTxAndContext()
		if err := deleteAppInstall(ctx, install, req.ForceDelete, req.DeleteBackup); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	case constant.Sync:
		return syncById(install.ID)
	case constant.Backup:
		tx, ctx := getTxAndContext()
		if err := backupInstall(ctx, install); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	case constant.Restore:
		return restoreInstall(install, req.BackupId)
	case constant.Update:
		return updateInstall(install.ID, req.DetailId)
	default:
		return errors.New("operate not support")
	}

	return appInstallRepo.Save(&install)
}

func (a AppInstallService) SyncAll() error {
	allList, err := appInstallRepo.GetBy()
	if err != nil {
		return err
	}
	go func() {
		for _, i := range allList {
			if err := syncById(i.ID); err != nil {
				global.LOG.Errorf("sync install app[%s] error,mgs: %s", i.Name, err.Error())
			}
		}
	}()
	return nil
}

func (a AppInstallService) PageInstallBackups(req request.AppBackupSearch) (int64, []model.AppInstallBackup, error) {
	return appInstallBackupRepo.Page(req.Page, req.PageSize, appInstallBackupRepo.WithAppInstallID(req.AppInstallID))
}

func (a AppInstallService) DeleteBackup(req request.AppBackupDelete) error {
	backups, err := appInstallBackupRepo.GetBy(commonRepo.WithIdsIn(req.Ids))
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()

	var errStr strings.Builder
	for _, backup := range backups {
		dst := path.Join(backup.Path, backup.Name)
		if err := fileOp.DeleteFile(dst); err != nil {
			errStr.WriteString(err.Error())
			continue
		}
		if err := appInstallBackupRepo.Delete(context.TODO(), commonRepo.WithIdsIn(req.Ids)); err != nil {
			errStr.WriteString(err.Error())
		}
	}
	if errStr.String() != "" {
		return errors.New(errStr.String())
	}
	return nil
}

func (a AppInstallService) GetServices(key string) ([]response.AppService, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return nil, err
	}
	installs, err := appInstallRepo.GetBy(appInstallRepo.WithAppId(app.ID), appInstallRepo.WithStatus(constant.Running))
	if err != nil {
		return nil, err
	}
	var res []response.AppService
	for _, install := range installs {
		paramMap := make(map[string]string)
		if install.Param != "" {
			_ = json.Unmarshal([]byte(install.Param), &paramMap)
		}
		res = append(res, response.AppService{
			Label:  install.Name,
			Value:  install.ServiceName,
			Config: paramMap,
		})
	}
	return res, nil
}

func (a AppInstallService) GetUpdateVersions(installId uint) ([]dto.AppVersion, error) {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
	var versions []dto.AppVersion
	if err != nil {
		return versions, err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(install.AppId))
	if err != nil {
		return versions, err
	}
	details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(app.ID))
	if err != nil {
		return versions, err
	}
	for _, detail := range details {
		if common.CompareVersion(detail.Version, install.Version) {
			versions = append(versions, dto.AppVersion{
				Version:  detail.Version,
				DetailId: detail.ID,
			})
		}
	}
	return versions, nil
}

func (a AppInstallService) ChangeAppPort(req request.PortUpdate) error {
	return updateInstallInfoInDB(req.Key, "", "port", true, strconv.FormatInt(req.Port, 10))
}

func (a AppInstallService) DeleteCheck(installId uint) ([]dto.AppResource, error) {
	var res []dto.AppResource
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
	if err != nil {
		return nil, err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appInstall.AppId))
	if err != nil {
		return nil, err
	}
	if app.Type == "website" {
		websites, _ := websiteRepo.GetBy(websiteRepo.WithAppInstallId(appInstall.ID))
		for _, website := range websites {
			res = append(res, dto.AppResource{
				Type: "website",
				Name: website.PrimaryDomain,
			})
		}
	}
	if app.Key == constant.AppNginx {
		websites, _ := websiteRepo.GetBy()
		for _, website := range websites {
			res = append(res, dto.AppResource{
				Type: "website",
				Name: website.PrimaryDomain,
			})
		}
	}
	if app.Type == "runtime" {
		resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(appInstall.ID))
		for _, resource := range resources {
			linkInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(resource.AppInstallId))
			res = append(res, dto.AppResource{
				Type: "app",
				Name: linkInstall.Name,
			})
		}
	}

	if app.Key == constant.AppMysql {
		databases, _ := mysqlRepo.List()
		for _, database := range databases {
			res = append(res, dto.AppResource{
				Type: "database",
				Name: database.Name,
			})
		}
	}

	return res, nil
}

func (a AppInstallService) GetDefaultConfigByKey(key string) (string, error) {
	appInstall, err := getAppInstallByKey(key)
	if err != nil {
		return "", err
	}
	filePath := path.Join(constant.AppResourceDir, appInstall.App.Key, "versions", appInstall.Version, "conf")
	if key == constant.AppMysql {
		filePath = path.Join(filePath, "my.cnf")
	}
	if key == constant.AppRedis {
		filePath = path.Join(filePath, "redis.conf")
	}
	if key == constant.AppNginx {
		filePath = path.Join(filePath, "nginx.conf")
	}
	contentByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(contentByte), nil
}

func syncById(installId uint) error {
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
	if err != nil {
		return err
	}

	containerNames, err := getContainerNames(appInstall)
	if err != nil {
		return err
	}

	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	containers, err := cli.ListContainersByName(containerNames)
	if err != nil {
		return err
	}
	var (
		errorContainers    []string
		notFoundContainers []string
		runningContainers  []string
		exitedContainers   []string
	)

	for _, n := range containers {
		switch n.State {
		case "exited":
			exitedContainers = append(exitedContainers, n.Names[0])
		case "running":
			runningContainers = append(runningContainers, n.Names[0])
		default:
			errorContainers = append(errorContainers, n.Names[0])
		}
	}
	for _, old := range containerNames {
		exist := false
		for _, new := range containers {
			if common.ExistWithStrArray(old, new.Names) {
				exist = true
				break
			}
		}
		if !exist {
			notFoundContainers = append(notFoundContainers, old)
		}
	}

	containerCount := len(containers)
	errCount := len(errorContainers)
	notFoundCount := len(notFoundContainers)
	existedCount := len(exitedContainers)
	normalCount := len(containerNames)
	runningCount := len(runningContainers)

	if containerCount == 0 {
		appInstall.Status = constant.Error
		appInstall.Message = "container is not found"
		return appInstallRepo.Save(&appInstall)
	}
	if errCount == 0 && existedCount == 0 {
		appInstall.Status = constant.Running
		return appInstallRepo.Save(&appInstall)
	}
	if existedCount == normalCount {
		appInstall.Status = constant.Stopped
		return appInstallRepo.Save(&appInstall)
	}
	if errCount == normalCount {
		appInstall.Status = constant.Error
	}
	if runningCount < normalCount {
		appInstall.Status = constant.UnHealthy
	}

	var errMsg strings.Builder
	if errCount > 0 {
		errMsg.Write([]byte(string(rune(errCount)) + " error containers:"))
		for _, e := range errorContainers {
			errMsg.Write([]byte(e))
		}
		errMsg.Write([]byte("\n"))
	}
	if notFoundCount > 0 {
		errMsg.Write([]byte(string(rune(notFoundCount)) + " not found containers:"))
		for _, e := range notFoundContainers {
			errMsg.Write([]byte(e))
		}
		errMsg.Write([]byte("\n"))
	}
	appInstall.Message = errMsg.String()
	return appInstallRepo.Save(&appInstall)
}

func updateInstallInfoInDB(appKey, appName, param string, isRestart bool, value interface{}) error {
	if param != "password" && param != "port" {
		return nil
	}
	appInstall, err := appInstallRepo.LoadBaseInfo(appKey, appName)
	if err != nil {
		return nil
	}
	envPath := fmt.Sprintf("%s/%s/%s/.env", constant.AppInstallDir, appKey, appInstall.Name)
	lineBytes, err := ioutil.ReadFile(envPath)
	if err != nil {
		return err
	}
	envKey := "PANEL_DB_ROOT_PASSWORD="
	if param == "port" {
		envKey = "PANEL_APP_PORT_HTTP="
	}
	files := strings.Split(string(lineBytes), "\n")
	var newFiles []string
	for _, line := range files {
		if strings.HasPrefix(line, envKey) {
			newFiles = append(newFiles, fmt.Sprintf("%s%v", envKey, value))
		} else {
			newFiles = append(newFiles, line)
		}
	}
	file, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(newFiles, "\n"))
	if err != nil {
		return err
	}

	oldVal, newVal := "", ""
	if param == "password" {
		oldVal = fmt.Sprintf("\"PANEL_DB_ROOT_PASSWORD\":\"%v\"", appInstall.Password)
		newVal = fmt.Sprintf("\"PANEL_DB_ROOT_PASSWORD\":\"%v\"", value)
		_ = appInstallRepo.BatchUpdateBy(map[string]interface{}{
			"param": strings.ReplaceAll(appInstall.Param, oldVal, newVal),
			"env":   strings.ReplaceAll(appInstall.Env, oldVal, newVal),
		}, commonRepo.WithByID(appInstall.ID))

	}
	if param == "port" {
		oldVal = fmt.Sprintf("\"PANEL_APP_PORT_HTTP\":%v", appInstall.Port)
		newVal = fmt.Sprintf("\"PANEL_APP_PORT_HTTP\":%v", value)
		_ = appInstallRepo.BatchUpdateBy(map[string]interface{}{
			"param":     strings.ReplaceAll(appInstall.Param, oldVal, newVal),
			"env":       strings.ReplaceAll(appInstall.Env, oldVal, newVal),
			"http_port": value,
		}, commonRepo.WithByID(appInstall.ID))
	}

	ComposeFile := fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, appKey, appInstall.Name)
	stdout, err := compose.Down(ComposeFile)
	if err != nil {
		return errors.New(stdout)
	}
	stdout, err = compose.Up(ComposeFile)
	if err != nil {
		return errors.New(stdout)
	}
	return nil
}
