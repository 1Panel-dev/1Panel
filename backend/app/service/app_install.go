package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

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

func (a AppInstallService) Page(req dto.AppInstalledRequest) (int64, []dto.AppInstalled, error) {
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

func (a AppInstallService) CheckExist(key string) (*dto.CheckInstalled, error) {
	res := &dto.CheckInstalled{
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

func (a AppInstallService) Search(req dto.AppInstalledRequest) ([]dto.AppInstalled, error) {
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
		installs, err = appInstallRepo.GetBy(appInstallRepo.WithAppIdsIn(ids))
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

func (a AppInstallService) Operate(req dto.AppInstallOperate) error {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(req.InstallId))
	if err != nil {
		return err
	}

	dockerComposePath := install.GetComposePath()

	switch req.Operate {
	case dto.Up:
		out, err := compose.Up(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		install.Status = constant.Running
	case dto.Down:
		out, err := compose.Down(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		install.Status = constant.Stopped
	case dto.Restart:
		out, err := compose.Restart(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		install.Status = constant.Running
	case dto.Delete:
		tx, ctx := getTxAndContext()
		if err := deleteAppInstall(ctx, install); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	case dto.Sync:
		return syncById(install.ID)
	case dto.Backup:
		tx, ctx := getTxAndContext()
		if err := backupInstall(ctx, install); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	case dto.Restore:
		return restoreInstall(install, req.BackupId)
	case dto.Update:
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

func (a AppInstallService) PageInstallBackups(req dto.AppBackupRequest) (int64, []model.AppInstallBackup, error) {
	return appInstallBackupRepo.Page(req.Page, req.PageSize, appInstallBackupRepo.WithAppInstallID(req.AppInstallID))
}

func (a AppInstallService) DeleteBackup(req dto.AppBackupDeleteRequest) error {

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

func (a AppInstallService) GetServices(key string) ([]dto.AppService, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return nil, err
	}
	installs, err := appInstallRepo.GetBy(appInstallRepo.WithAppId(app.ID), appInstallRepo.WithStatus(constant.Running))
	if err != nil {
		return nil, err
	}
	var res []dto.AppService
	for _, install := range installs {
		res = append(res, dto.AppService{
			Label: install.Name,
			Value: install.ServiceName,
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

func (a AppInstallService) ChangeAppPort(req dto.PortUpdate) error {
	var (
		files    []string
		newFiles []string
	)
	app, err := appInstallRepo.LoadBaseInfoByKey(req.Key)
	if err != nil {
		return err
	}

	ComposeDir := fmt.Sprintf("%s/%s/%s", constant.AppInstallDir, req.Key, req.Name)
	ComposeFile := fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, req.Key, req.Name)
	path := fmt.Sprintf("%s/.env", ComposeDir)
	lineBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		files = strings.Split(string(lineBytes), "\n")
	}
	for _, line := range files {
		if strings.HasPrefix(line, "PANEL_APP_PORT_HTTP=") {
			newFiles = append(newFiles, fmt.Sprintf("PANEL_APP_PORT_HTTP=%v", req.Port))
		} else {
			newFiles = append(newFiles, line)
		}
	}
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(newFiles, "\n"))
	if err != nil {
		return err
	}

	if err := mysqlRepo.UpdateDatabaseInfo(app.ID, map[string]interface{}{
		"env": strings.ReplaceAll(app.Env, strconv.FormatInt(app.Port, 10), strconv.FormatInt(req.Port, 10)),
	}); err != nil {
		return err
	}
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
	)

	for _, n := range containers {
		if n.State != "running" {
			errorContainers = append(errorContainers, n.Names[0])
		} else {
			runningContainers = append(runningContainers, n.Names[0])
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
	normalCount := len(containerNames)
	runningCount := len(runningContainers)

	if containerCount == 0 {
		appInstall.Status = constant.Error
		appInstall.Message = "container is not found"
		return appInstallRepo.Save(&appInstall)
	}
	if errCount == 0 && notFoundCount == 0 {
		appInstall.Status = constant.Running
		return appInstallRepo.Save(&appInstall)
	}
	if errCount == normalCount {
		appInstall.Status = constant.Error
	}
	if notFoundCount == normalCount {
		appInstall.Status = constant.Stopped
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
