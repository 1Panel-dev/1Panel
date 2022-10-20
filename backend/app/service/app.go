package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

type AppService struct {
}

func (a AppService) PageApp(req dto.AppRequest) (interface{}, error) {

	var opts []repo.DBOption
	opts = append(opts, commonRepo.WithOrderBy("name"))
	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}
	if len(req.Tags) != 0 {
		tags, err := tagRepo.GetByKeys(req.Tags)
		if err != nil {
			return nil, err
		}
		var tagIds []uint
		for _, t := range tags {
			tagIds = append(tagIds, t.ID)
		}
		appTags, err := appTagRepo.GetByTagIds(tagIds)
		if err != nil {
			return nil, err
		}
		var appIds []uint
		for _, t := range appTags {
			appIds = append(appIds, t.AppId)
		}

		opts = append(opts, commonRepo.WithIdsIn(appIds))
	}
	var res dto.AppRes
	total, apps, err := appRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, err
	}
	var appDTOs []*dto.AppDTO
	for _, a := range apps {
		appDTO := &dto.AppDTO{
			App: a,
		}
		appDTOs = append(appDTOs, appDTO)
		appTags, err := appTagRepo.GetByAppId(a.ID)
		if err != nil {
			continue
		}
		var tagIds []uint
		for _, at := range appTags {
			tagIds = append(tagIds, at.TagId)
		}
		tags, err := tagRepo.GetByIds(tagIds)
		if err != nil {
			continue
		}
		appDTO.Tags = tags
	}
	res.Items = appDTOs
	res.Total = total
	tags, err := tagRepo.All()
	if err != nil {
		return nil, err
	}
	res.Tags = tags

	return res, nil
}

func (a AppService) GetApp(id uint) (dto.AppDTO, error) {
	var appDTO dto.AppDTO
	app, err := appRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return appDTO, err
	}
	appDTO.App = app
	details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(app.ID))
	if err != nil {
		return appDTO, err
	}
	var versionsRaw []string
	for _, detail := range details {
		versionsRaw = append(versionsRaw, detail.Version)
	}

	appDTO.Versions = common.GetSortedVersions(versionsRaw)

	return appDTO, nil
}

func (a AppService) PageInstalled(req dto.AppInstalledRequest) (int64, []dto.AppInstalled, error) {
	total, installed, err := appInstallRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	var installDTOs []dto.AppInstalled
	for _, in := range installed {
		installDto := dto.AppInstalled{
			AppInstall: in,
		}
		installDTOs = append(installDTOs, installDto)
	}

	return total, installDTOs, nil
}

func (a AppService) GetAppDetail(appId uint, version string) (dto.AppDetailDTO, error) {

	var (
		appDetailDTO dto.AppDetailDTO
		opts         []repo.DBOption
	)

	opts = append(opts, appDetailRepo.WithAppId(appId), appDetailRepo.WithVersion(version))
	detail, err := appDetailRepo.GetFirst(opts...)
	if err != nil {
		return appDetailDTO, err
	}
	paramMap := make(map[string]interface{})
	_ = json.Unmarshal([]byte(detail.Params), &paramMap)
	appDetailDTO.AppDetail = detail
	appDetailDTO.Params = paramMap
	return appDetailDTO, nil
}

func (a AppService) PageInstallBackups(req dto.AppBackupRequest) (int64, []model.AppInstallBackup, error) {
	return appInstallBackupRepo.Page(req.Page, req.PageSize, appInstallBackupRepo.WithAppInstallID(req.AppInstallID))
}

func (a AppService) OperateInstall(req dto.AppInstallOperate) error {
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
		return deleteAppInstall(install)
	case dto.Sync:
		return a.SyncInstalled(install.ID)
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

func (a AppService) Install(name string, appDetailId uint, params map[string]interface{}) error {

	httpPort, err := checkPort("PANEL_APP_PORT_HTTP", params)
	if err != nil {
		return fmt.Errorf("%d port is in used", httpPort)
	}
	httpsPort, err := checkPort("PANEL_APP_PORT_HTTPS", params)
	if err != nil {
		return fmt.Errorf("%d port is in used", httpPort)
	}

	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(appDetailId))
	if err != nil {
		return err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	if err != nil {
		return err
	}

	if err := checkRequiredAndLimit(app); err != nil {
		return err
	}
	if err := copyAppData(app.Key, appDetail.Version, name, params); err != nil {
		return err
	}

	paramByte, err := json.Marshal(params)
	if err != nil {
		return err
	}
	appInstall := model.AppInstall{
		Name:        name,
		AppId:       appDetail.AppId,
		AppDetailId: appDetail.ID,
		Version:     appDetail.Version,
		Status:      constant.Installing,
		Env:         string(paramByte),
		HttpPort:    httpPort,
		HttpsPort:   httpsPort,
		App:         app,
	}

	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(appDetail.DockerCompose), &composeMap); err != nil {
		return err
	}
	servicesMap := composeMap["services"].(map[string]interface{})
	changeKeys := make(map[string]string, len(servicesMap))
	for k, v := range servicesMap {
		serviceName := k + "-" + common.RandStr(4)
		changeKeys[k] = serviceName
		value := v.(map[string]interface{})
		containerName := constant.ContainerPrefix + k + "-" + common.RandStr(4)
		value["container_name"] = containerName
		appInstall.ServiceName = serviceName
		appInstall.ContainerName = containerName
	}
	for k, v := range changeKeys {
		servicesMap[v] = servicesMap[k]
		delete(servicesMap, k)
	}
	composeByte, err := yaml.Marshal(composeMap)
	if err != nil {
		return err
	}
	appInstall.DockerCompose = string(composeByte)

	fileOp := files.NewFileOp()
	if err := fileOp.WriteFile(appInstall.GetComposePath(), strings.NewReader(string(composeByte)), 0775); err != nil {
		return err
	}

	tx, ctx := getTxAndContext()
	if err := appInstallRepo.Create(ctx, &appInstall); err != nil {
		tx.Rollback()
		return err
	}
	if err := createLink(ctx, app, &appInstall, params); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	go upApp(appInstall.GetComposePath(), appInstall)
	return nil
}

func (a AppService) SyncAllInstalled() error {
	allList, err := appInstallRepo.GetBy()
	if err != nil {
		return err
	}
	go func() {
		for _, i := range allList {
			if err := a.SyncInstalled(i.ID); err != nil {
				global.LOG.Errorf("sync install app[%s] error,mgs: %s", i.Name, err.Error())
			}
		}
	}()
	return nil
}

func (a AppService) DeleteBackup(req dto.AppBackupDeleteRequest) error {

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

func (a AppService) SyncInstalled(installId uint) error {
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

func (a AppService) SyncAppList() error {
	if err := getAppFromOss(); err != nil {
		global.LOG.Errorf("get app from oss  error: %s", err.Error())
		return err
	}

	appDir := constant.AppResourceDir
	iconDir := path.Join(appDir, "icons")
	listFile := path.Join(appDir, "list.json")

	content, err := os.ReadFile(listFile)
	if err != nil {
		return err
	}
	list := &dto.AppList{}
	if err := json.Unmarshal(content, list); err != nil {
		return err
	}

	var (
		tags    []*model.Tag
		appTags []*model.AppTag
	)

	for _, t := range list.Tags {
		tags = append(tags, &model.Tag{
			Key:  t.Key,
			Name: t.Name,
		})
	}

	oldApps, err := appRepo.GetBy()
	if err != nil {
		return err
	}
	appsMap := getApps(oldApps, list.Items)

	for _, l := range list.Items {

		app := appsMap[l.Key]
		icon, err := os.ReadFile(path.Join(iconDir, l.Icon))
		if err != nil {
			global.LOG.Errorf("get [%s] icon error: %s", l.Name, err.Error())
			continue
		}
		iconStr := base64.StdEncoding.EncodeToString(icon)
		app.Icon = iconStr
		app.TagsKey = l.Tags

		versions := l.Versions
		detailsMap := getAppDetails(app.Details, versions)

		for _, v := range versions {
			detail := detailsMap[v]
			detailPath := path.Join(appDir, l.Key, v)
			if _, err := os.Stat(detailPath); err != nil {
				global.LOG.Errorf("get [%s] folder error: %s", detailPath, err.Error())
				continue
			}
			readmeStr, err := os.ReadFile(path.Join(detailPath, "README.md"))
			if err != nil {
				global.LOG.Errorf("get [%s] README error: %s", detailPath, err.Error())
			}
			detail.Readme = string(readmeStr)
			dockerComposeStr, err := os.ReadFile(path.Join(detailPath, "docker-compose.yml"))
			if err != nil {
				global.LOG.Errorf("get [%s] docker-compose.yml error: %s", detailPath, err.Error())
				continue
			}
			detail.DockerCompose = string(dockerComposeStr)
			paramStr, err := os.ReadFile(path.Join(detailPath, "config.json"))
			if err != nil {
				global.LOG.Errorf("get [%s] form.json error: %s", detailPath, err.Error())
			}
			detail.Params = string(paramStr)
			detailsMap[v] = detail
		}
		var newDetails []model.AppDetail
		for _, v := range detailsMap {
			newDetails = append(newDetails, v)
		}
		app.Details = newDetails
		appsMap[l.Key] = app
	}

	var (
		addAppArray []model.App
		updateArray []model.App
	)
	tagMap := make(map[string]uint, len(tags))
	for _, v := range appsMap {
		if v.ID == 0 {
			addAppArray = append(addAppArray, v)
		} else {
			updateArray = append(updateArray, v)
		}
	}

	tx := global.DB.Begin()
	ctx := context.WithValue(context.Background(), "db", tx)

	if len(addAppArray) > 0 {
		if err := appRepo.BatchCreate(ctx, addAppArray); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tagRepo.DeleteAll(ctx); err != nil {
		tx.Rollback()
		return err
	}
	if len(tags) > 0 {
		if err := tagRepo.BatchCreate(ctx, tags); err != nil {
			tx.Rollback()
			return err
		}
		for _, t := range tags {
			tagMap[t.Key] = t.ID
		}
	}
	for _, update := range updateArray {
		if err := appRepo.Save(ctx, &update); err != nil {
			tx.Rollback()
			return err
		}
	}

	apps := append(addAppArray, updateArray...)

	var (
		addDetails    []model.AppDetail
		updateDetails []model.AppDetail
	)
	for _, a := range apps {
		for _, t := range a.TagsKey {
			tagId, ok := tagMap[t]
			if ok {
				appTags = append(appTags, &model.AppTag{
					AppId: a.ID,
					TagId: tagId,
				})
			}
		}

		for _, d := range a.Details {
			d.AppId = a.ID
			if d.ID == 0 {
				addDetails = append(addDetails, d)
			} else {
				updateDetails = append(updateDetails, d)
			}
		}
	}

	if len(addDetails) > 0 {
		if err := appDetailRepo.BatchCreate(ctx, addDetails); err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, u := range updateDetails {
		if err := appDetailRepo.Update(ctx, u); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := appTagRepo.DeleteAll(ctx); err != nil {
		tx.Rollback()
		return err
	}

	if len(appTags) > 0 {
		if err := appTagRepo.BatchCreate(ctx, appTags); err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	go syncCanUpdate()
	return nil
}

func syncCanUpdate() {
	apps, err := appRepo.GetBy()
	if err != nil {
		global.LOG.Errorf("sync update app error: %s", err.Error())
	}
	for _, app := range apps {
		details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(app.ID))
		if err != nil {
			global.LOG.Errorf("sync update app error: %s", err.Error())
		}
		var versions []string
		for _, detail := range details {
			versions = append(versions, detail.Version)
		}
		versions = common.GetSortedVersions(versions)
		lastVersion := versions[0]

		var updateDetailIds []uint
		for _, detail := range details {
			if lastVersion == detail.Version {
				continue
			}
			if common.CompareVersion(lastVersion, detail.Version) && (app.CrossVersionUpdate || !common.IsCrossVersion(detail.Version, lastVersion)) {
				updateDetailIds = append(updateDetailIds, detail.ID)
			}
		}
		if err := appDetailRepo.BatchUpdateBy(map[string]interface{}{"last_version": ""}); err != nil {
			global.LOG.Errorf("sync update app error: %s", err.Error())
		}

		if err := appInstallRepo.BatchUpdateBy(map[string]interface{}{"can_update": 0}); err != nil {
			global.LOG.Errorf("sync update app error: %s", err.Error())
		}
		if len(updateDetailIds) > 0 {
			if err := appDetailRepo.BatchUpdateBy(map[string]interface{}{"last_version": lastVersion}, commonRepo.WithIdsIn(updateDetailIds)); err != nil {
				global.LOG.Errorf("sync update app error: %s", err.Error())
			}
			if err := appInstallRepo.BatchUpdateBy(map[string]interface{}{"can_update": 1}, appInstallRepo.WithDetailIdsIn(updateDetailIds)); err != nil {
				global.LOG.Errorf("sync update app error: %s", err.Error())
			}
		}

	}
}

func (a AppService) GetServices(key string) ([]dto.AppService, error) {
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

func (a AppService) GetUpdateVersions(installId uint) ([]dto.AppVersion, error) {
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
