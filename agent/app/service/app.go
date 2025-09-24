package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"gopkg.in/yaml.v3"
)

type AppService struct {
}

type IAppService interface {
	PageApp(ctx *gin.Context, req request.AppSearch) (interface{}, error)
	GetAppTags(ctx *gin.Context) ([]response.TagDTO, error)
	GetApp(ctx *gin.Context, key string) (*response.AppDTO, error)
	GetAppDetail(appId uint, version, appType string) (response.AppDetailDTO, error)
	Install(req request.AppInstallCreate) (*model.AppInstall, error)
	SyncAppListFromRemote(taskID string) error
	GetAppUpdate() (*response.AppUpdateRes, error)
	GetAppDetailByID(id uint) (*response.AppDetailDTO, error)
	SyncAppListFromLocal(taskID string)
}

func NewIAppService() IAppService {
	return &AppService{}
}

func (a AppService) PageApp(ctx *gin.Context, req request.AppSearch) (interface{}, error) {
	var opts []repo.DBOption
	opts = append(opts, appRepo.OrderByRecommend())
	if req.Name != "" {
		opts = append(opts, appRepo.WithByLikeName(strings.TrimSpace(req.Name)))
	}
	if req.Type != "" {
		opts = append(opts, appRepo.WithType(req.Type))
	}
	if req.Recommend {
		opts = append(opts, appRepo.GetRecommend())
	}
	if req.Resource != "" && req.Resource != "all" {
		opts = append(opts, appRepo.WithResource(req.Resource))
	}

	if req.ShowCurrentArch {
		info, err := NewIDashboardService().LoadOsInfo()
		if err != nil {
			return nil, err
		}
		opts = append(opts, appRepo.WithArch(info.KernelArch))
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
		opts = append(opts, repo.WithByIDs(appIds))
	}
	var res response.AppRes

	total, apps, err := appRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, err
	}
	appDTOs := make([]*response.AppItem, 0)
	info := &dto.SettingInfo{}
	if req.Type == "php" {
		info, _ = NewISettingService().GetSettingInfo()
	}
	lang := strings.ToLower(common.GetLang(ctx))
	for _, ap := range apps {
		if req.Type == "php" {
			if !global.CONF.Base.IsOffLine && (ap.RequiredPanelVersion == 0 || !common.CompareAppVersion(common.GetSystemVersion(info.SystemVersion), fmt.Sprintf("%f", ap.RequiredPanelVersion))) {
				continue
			}
		}
		appDTO := &response.AppItem{
			ID:          ap.ID,
			Name:        ap.Name,
			Key:         ap.Key,
			Type:        ap.Type,
			Icon:        ap.Icon,
			Resource:    ap.Resource,
			Limit:       ap.Limit,
			Website:     ap.Website,
			Github:      ap.Github,
			GpuSupport:  ap.GpuSupport,
			Recommend:   ap.Recommend,
			Description: ap.GetDescription(ctx),
		}
		appDTOs = append(appDTOs, appDTO)
		tags, err := getAppTags(ap.ID, lang)
		if err != nil {
			continue
		}
		appDTO.Tags = tags
		if ap.Type == constant.RuntimePHP || ap.Type == constant.RuntimeGo || ap.Type == constant.RuntimeNode || ap.Type == constant.RuntimePython || ap.Type == constant.RuntimeJava || ap.Type == constant.RuntimeDotNet {
			details, _ := appDetailRepo.GetBy(appDetailRepo.WithAppId(ap.ID))
			var ids []uint
			if len(details) == 0 {
				continue
			}
			for _, d := range details {
				ids = append(ids, d.ID)
			}
			runtimes, _ := runtimeRepo.List(runtimeRepo.WithDetailIdsIn(ids))
			appDTO.Installed = len(runtimes) > 0
		} else {
			installs, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(ap.ID))
			appDTO.Installed = len(installs) > 0
		}
	}
	res.Items = appDTOs
	res.Total = total

	return res, nil
}

func (a AppService) GetAppTags(ctx *gin.Context) ([]response.TagDTO, error) {
	tags, err := tagRepo.All()
	if err != nil {
		return nil, err
	}
	var res []response.TagDTO
	lang := strings.ToLower(common.GetLang(ctx))
	for _, tag := range tags {
		tagDTO := response.TagDTO{
			ID:  tag.ID,
			Key: tag.Key,
		}
		var translations = make(map[string]string)
		_ = json.Unmarshal([]byte(tag.Translations), &translations)
		if name, ok := translations[lang]; ok {
			tagDTO.Name = name
		}
		res = append(res, tagDTO)
	}
	return res, nil
}

func (a AppService) GetApp(ctx *gin.Context, key string) (*response.AppDTO, error) {
	var appDTO response.AppDTO
	if key == "postgres" {
		key = "postgresql"
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return nil, err
	}
	appDTO.App = app
	appDTO.App.Description = app.GetDescription(ctx)
	details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(app.ID))
	if err != nil {
		return nil, err
	}
	var versionsRaw []string
	hasLatest := false
	latestVersion := ""
	for _, detail := range details {
		if strings.Contains(detail.Version, "latest") {
			hasLatest = true
			latestVersion = detail.Version
			continue
		}
		if key == "openresty" && !common.CompareAppVersion(detail.Version, "1.27") {
			continue
		}
		versionsRaw = append(versionsRaw, detail.Version)
	}
	appDTO.Versions = common.GetSortedVersions(versionsRaw)
	if hasLatest {
		appDTO.Versions = append([]string{latestVersion}, appDTO.Versions...)
	}
	tags, err := getAppTags(app.ID, strings.ToLower(common.GetLang(ctx)))
	if err != nil {
		return nil, err
	}
	appDTO.Tags = tags
	return &appDTO, nil
}

func (a AppService) GetAppDetail(appID uint, version, appType string) (response.AppDetailDTO, error) {
	var (
		appDetailDTO response.AppDetailDTO
		opts         []repo.DBOption
	)
	opts = append(opts, appDetailRepo.WithAppId(appID), appDetailRepo.WithVersion(version))
	detail, err := appDetailRepo.GetFirst(opts...)
	if err != nil {
		return appDetailDTO, err
	}
	appDetailDTO.AppDetail = detail
	appDetailDTO.Enable = true

	if appType == "runtime" {
		app, err := appRepo.GetFirst(repo.WithByID(appID))
		if err != nil {
			return appDetailDTO, err
		}
		fileOp := files.NewFileOp()

		versionPath := filepath.Join(app.GetAppResourcePath(), detail.Version)
		if !fileOp.Stat(versionPath) || detail.Update {
			if err = downloadApp(app, detail, nil, nil); err != nil && !fileOp.Stat(versionPath) {
				return appDetailDTO, err
			}
		}
		switch app.Type {
		case constant.RuntimePHP:
			paramsPath := filepath.Join(versionPath, "data.yml")
			if !fileOp.Stat(paramsPath) {
				return appDetailDTO, buserr.WithDetail("ErrFileNotExist", paramsPath, nil)
			}
			param, err := fileOp.GetContent(paramsPath)
			if err != nil {
				return appDetailDTO, err
			}
			paramMap := make(map[string]interface{})
			if err = yaml.Unmarshal(param, &paramMap); err != nil {
				return appDetailDTO, err
			}
			appDetailDTO.Params = paramMap["additionalProperties"]
			composePath := filepath.Join(versionPath, "docker-compose.yml")
			if !fileOp.Stat(composePath) {
				return appDetailDTO, buserr.WithDetail("ErrFileNotExist", composePath, nil)
			}
			compose, err := fileOp.GetContent(composePath)
			if err != nil {
				return appDetailDTO, err
			}
			composeMap := make(map[string]interface{})
			if err := yaml.Unmarshal(compose, &composeMap); err != nil {
				return appDetailDTO, err
			}
			if service, ok := composeMap["services"]; ok {
				servicesMap := service.(map[string]interface{})
				for k := range servicesMap {
					appDetailDTO.Image = k
				}
			}
		}
	} else {
		paramMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(detail.Params), &paramMap); err != nil {
			return appDetailDTO, err
		}
		appDetailDTO.Params = paramMap
	}

	if appDetailDTO.DockerCompose == "" {
		filename := filepath.Base(appDetailDTO.DownloadUrl)
		dockerComposeUrl := fmt.Sprintf("%s%s", strings.TrimSuffix(appDetailDTO.DownloadUrl, filename), "docker-compose.yml")
		statusCode, composeRes, err := req_helper.HandleRequest(dockerComposeUrl, http.MethodGet, constant.TimeOut20s)
		if err != nil {
			return appDetailDTO, buserr.WithDetail("ErrGetCompose", err.Error(), err)
		}
		if statusCode > 200 {
			return appDetailDTO, buserr.WithDetail("ErrGetCompose", string(composeRes), err)
		}
		detail.DockerCompose = string(composeRes)
		_ = appDetailRepo.Update(context.Background(), detail)
		appDetailDTO.DockerCompose = string(composeRes)
	}

	appDetailDTO.HostMode = isHostModel(appDetailDTO.DockerCompose)

	app, err := appRepo.GetFirst(repo.WithByID(detail.AppId))
	if err != nil {
		return appDetailDTO, err
	}
	if err := checkLimit(app); err != nil {
		appDetailDTO.Enable = false
	}
	appDetailDTO.Architectures = app.Architectures
	appDetailDTO.MemoryRequired = app.MemoryRequired
	appDetailDTO.GpuSupport = app.GpuSupport
	return appDetailDTO, nil
}
func (a AppService) GetAppDetailByID(id uint) (*response.AppDetailDTO, error) {
	res := &response.AppDetailDTO{}
	appDetail, err := appDetailRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	res.AppDetail = appDetail
	paramMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(appDetail.Params), &paramMap); err != nil {
		return nil, err
	}
	res.Params = paramMap
	res.HostMode = isHostModel(appDetail.DockerCompose)
	return res, nil
}

func (a AppService) Install(req request.AppInstallCreate) (appInstall *model.AppInstall, err error) {
	if err = docker.CreateDefaultDockerNetwork(); err != nil {
		err = buserr.WithDetail("Err1PanelNetworkFailed", err.Error(), nil)
		return
	}
	if list, _ := appInstallRepo.ListBy(context.Background(), repo.WithByLowerName(req.Name)); len(list) > 0 {
		err = buserr.New("ErrAppNameExist")
		return
	}
	var (
		httpPort  int
		httpsPort int
		appDetail model.AppDetail
		app       model.App
	)
	appDetail, err = appDetailRepo.GetFirst(repo.WithByID(req.AppDetailId))
	if err != nil {
		return
	}
	app, err = appRepo.GetFirst(repo.WithByID(appDetail.AppId))
	if err != nil {
		return
	}
	if DatabaseKeys[app.Key] > 0 {
		if existDatabases, _ := databaseRepo.GetList(repo.WithByName(req.Name)); len(existDatabases) > 0 {
			err = buserr.New("ErrRemoteExist")
			return
		}
	}
	if hostName, ok := req.Params["PANEL_DB_HOST"]; ok {
		database, _ := databaseRepo.Get(repo.WithByName(hostName.(string)))
		if database.AppInstallID > 0 {
			databaseInstall, _ := appInstallRepo.GetFirst(repo.WithByID(database.AppInstallID))
			if databaseInstall.Status != constant.StatusRunning {
				return nil, buserr.WithName("ErrAppIsDown", databaseInstall.Name)
			}
		}
	}
	for key := range req.Params {
		if !strings.Contains(key, "PANEL_APP_PORT") {
			continue
		}
		var port int
		port, err = checkPort(key, req.Params)
		if err != nil {
			return
		}
		if key == "PANEL_APP_PORT_HTTP" {
			httpPort = port
		}
		if key == "PANEL_APP_PORT_HTTPS" {
			httpsPort = port
		}
	}

	if err = checkRequiredAndLimit(app); err != nil {
		return
	}

	appInstall = &model.AppInstall{
		Name:        req.Name,
		AppId:       appDetail.AppId,
		AppDetailId: appDetail.ID,
		Version:     appDetail.Version,
		Status:      constant.StatusInstalling,
		HttpPort:    httpPort,
		HttpsPort:   httpsPort,
		App:         app,
	}
	composeMap := make(map[string]interface{})
	if req.EditCompose {
		if err = yaml.Unmarshal([]byte(req.DockerCompose), &composeMap); err != nil {
			return
		}
	} else {
		if err = yaml.Unmarshal([]byte(appDetail.DockerCompose), &composeMap); err != nil {
			return
		}
	}

	value, ok := composeMap["services"]
	if !ok || value == nil {
		err = buserr.New("ErrFileParse")
		return
	}
	servicesMap := value.(map[string]interface{})
	containerName := constant.ContainerPrefix + app.Key + "-" + common.RandStr(4)
	if req.Advanced && req.ContainerName != "" {
		containerName = req.ContainerName
		appInstalls, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithContainerName(containerName))
		if len(appInstalls) > 0 {
			err = buserr.New("ErrContainerName")
			return
		}
		containerExist := false
		containerExist, err = checkContainerNameIsExist(req.ContainerName, appInstall.GetPath())
		if err != nil {
			return
		}
		if containerExist {
			err = buserr.New("ErrContainerName")
			return
		}
	}
	req.Params[constant.ContainerName] = containerName
	appInstall.ContainerName = containerName

	index := 0
	serviceName := ""
	for k := range servicesMap {
		serviceName = k
		if index > 0 {
			continue
		}
		index++
	}
	if app.Limit == 0 && appInstall.Name != serviceName && len(servicesMap) == 1 {
		servicesMap[appInstall.Name] = servicesMap[serviceName]
		delete(servicesMap, serviceName)
		serviceName = appInstall.Name
	}
	appInstall.ServiceName = serviceName

	if err = addDockerComposeCommonParam(composeMap, appInstall.ServiceName, req.AppContainerConfig, req.Params); err != nil {
		return
	}
	var (
		composeByte []byte
		paramByte   []byte
	)

	composeByte, err = yaml.Marshal(composeMap)
	if err != nil {
		return
	}
	appInstall.DockerCompose = string(composeByte)

	if hostName, ok := req.Params["PANEL_DB_HOST"]; ok {
		database, _ := databaseRepo.Get(repo.WithByName(hostName.(string)))
		if !reflect.DeepEqual(database, model.Database{}) {
			req.Params["PANEL_DB_HOST"] = database.Address
			req.Params["PANEL_DB_PORT"] = database.Port
			req.Params["PANEL_DB_HOST_NAME"] = hostName
			req.Params["DATABASE_NAME"] = database.Name
		}
	}
	if app.Key == "openresty" {
		req.Params["CONTAINER_PACKAGE_URL"] = "http://archive.ubuntu.com/ubuntu/"
		req.Params["RESTY_ADD_PACKAGE_BUILDDEPS"] = ""
		req.Params["RESTY_CONFIG_OPTIONS_MORE"] = ""
	}
	paramByte, err = json.Marshal(req.Params)
	if err != nil {
		return
	}
	appInstall.Env = string(paramByte)

	if err = appInstallRepo.Create(context.Background(), appInstall); err != nil {
		return
	}

	installTask, err := task.NewTaskWithOps(appInstall.Name, task.TaskInstall, task.TaskScopeApp, req.TaskID, appInstall.ID)
	if err != nil {
		return
	}

	if err = createLink(context.Background(), installTask, app, appInstall, req.Params); err != nil {
		return
	}

	installApp := func(t *task.Task) error {
		if err = copyData(t, app, appDetail, appInstall, req); err != nil {
			return err
		}
		if err = runScript(t, appInstall, "init"); err != nil {
			return err
		}
		if app.Key == "openresty" {
			if err = handleSiteDir(app, appDetail, req, t); err != nil {
				return err
			}
			if err = handleOpenrestyFile(appInstall); err != nil {
				return err
			}
		}
		if err = upApp(t, appInstall, req.PullImage); err != nil {
			return err
		}
		updateToolApp(appInstall)
		return nil
	}

	handleAppStatus := func(t *task.Task) {
		appInstall.Status = constant.StatusUpErr
		appInstall.Message = installTask.Task.ErrorMsg
		_ = appInstallRepo.Save(context.Background(), appInstall)
	}

	installTask.AddSubTask(task.GetTaskName(appInstall.Name, task.TaskInstall, task.TaskScopeApp), installApp, handleAppStatus)

	go func() {
		if taskErr := installTask.Execute(); taskErr != nil {
			appInstall.Status = constant.StatusInstallErr
			appInstall.Message = taskErr.Error()
			if strings.Contains(taskErr.Error(), "Timeout") && strings.Contains(taskErr.Error(), "Pulling") {
				appInstall.Message = buserr.New("PullImageTimeout").Error() + appInstall.Message
			}
			_ = appInstallRepo.Save(context.Background(), appInstall)
		}
	}()

	return
}

func (a AppService) SyncAppListFromLocal(TaskID string) {
	var (
		err        error
		dirEntries []os.DirEntry
		localApps  []model.App
	)

	syncTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("LocalApp"), task.TaskSync, task.TaskScopeAppStore, TaskID, 0)
	if err != nil {
		global.LOG.Errorf("Create sync task failed %v", err)
		return
	}

	syncTask.AddSubTask(task.GetTaskName(i18n.GetMsgByKey("LocalApp"), task.TaskSync, task.TaskScopeAppStore), func(t *task.Task) (err error) {
		fileOp := files.NewFileOp()
		localAppDir := global.Dir.LocalAppResourceDir
		if !fileOp.Stat(localAppDir) {
			return nil
		}
		dirEntries, err = os.ReadDir(localAppDir)
		if err != nil {
			return
		}
		for _, dirEntry := range dirEntries {
			if dirEntry.IsDir() {
				appDir := filepath.Join(localAppDir, dirEntry.Name())
				appDirEntries, err := os.ReadDir(appDir)
				if err != nil {
					t.Log(i18n.GetWithNameAndErr("ErrAppDirNull", dirEntry.Name(), err))
					continue
				}
				app, err := handleLocalApp(appDir)
				if err != nil {
					t.Log(i18n.GetWithNameAndErr("LocalAppErr", dirEntry.Name(), err))
					continue
				}
				var appDetails []model.AppDetail
				for _, appDirEntry := range appDirEntries {
					if appDirEntry.IsDir() {
						appDetail := model.AppDetail{
							Version: appDirEntry.Name(),
							Status:  constant.AppNormal,
						}
						versionDir := filepath.Join(appDir, appDirEntry.Name())
						if err = handleLocalAppDetail(versionDir, &appDetail); err != nil {
							t.Log(i18n.GetMsgWithMap("LocalAppVersionErr", map[string]interface{}{"name": app.Name, "version": appDetail.Version, "err": err.Error()}))
							continue
						}
						appDetails = append(appDetails, appDetail)
					}
				}
				if len(appDetails) > 0 {
					app.Details = appDetails
					localApps = append(localApps, *app)
				} else {
					t.Log(i18n.GetWithName("LocalAppVersionNull", app.Name))
				}
			}
		}

		var (
			newApps    []model.App
			deleteApps []model.App
			updateApps []model.App
			oldAppIds  []uint

			deleteAppIds     []uint
			deleteAppDetails []model.AppDetail
			newAppDetails    []model.AppDetail
			updateDetails    []model.AppDetail

			appTags []*model.AppTag
		)

		oldApps, _ := appRepo.GetBy(appRepo.WithResource(constant.AppResourceLocal))
		apps := make(map[string]model.App, len(oldApps))
		for _, old := range oldApps {
			old.Status = constant.AppTakeDown
			apps[old.Key] = old
		}
		for _, app := range localApps {
			if oldApp, ok := apps[app.Key]; ok {
				app.ID = oldApp.ID
				appDetails := make(map[string]model.AppDetail, len(oldApp.Details))
				for _, old := range oldApp.Details {
					old.Status = constant.AppTakeDown
					appDetails[old.Version] = old
				}
				for i, newDetail := range app.Details {
					version := newDetail.Version
					newDetail.Status = constant.AppNormal
					newDetail.AppId = app.ID
					oldDetail, exist := appDetails[version]
					if exist {
						newDetail.ID = oldDetail.ID
						delete(appDetails, version)
					}
					app.Details[i] = newDetail
				}
				for _, v := range appDetails {
					app.Details = append(app.Details, v)
				}
			}
			app.TagsKey = append(app.TagsKey, constant.AppResourceLocal)
			apps[app.Key] = app
		}

		for _, app := range apps {
			if app.ID == 0 {
				newApps = append(newApps, app)
			} else {
				oldAppIds = append(oldAppIds, app.ID)
				if app.Status == constant.AppTakeDown {
					installs, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(app.ID))
					if len(installs) > 0 {
						updateApps = append(updateApps, app)
						continue
					}
					deleteAppIds = append(deleteAppIds, app.ID)
					deleteApps = append(deleteApps, app)
					deleteAppDetails = append(deleteAppDetails, app.Details...)
				} else {
					updateApps = append(updateApps, app)
				}
			}

		}

		tags, _ := tagRepo.All()
		tagMap := make(map[string]uint, len(tags))
		for _, tag := range tags {
			tagMap[tag.Key] = tag.ID
		}

		tx, ctx := getTxAndContext()
		defer tx.Rollback()
		if len(newApps) > 0 {
			if err = appRepo.BatchCreate(ctx, newApps); err != nil {
				return
			}
		}
		for _, update := range updateApps {
			if err = appRepo.Save(ctx, &update); err != nil {
				return
			}
		}
		if len(deleteApps) > 0 {
			if err = appRepo.BatchDelete(ctx, deleteApps); err != nil {
				return
			}
			if err = appDetailRepo.DeleteByAppIds(ctx, deleteAppIds); err != nil {
				return
			}
		}

		if err = appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
			return
		}
		for _, newApp := range newApps {
			if newApp.ID > 0 {
				for _, detail := range newApp.Details {
					detail.AppId = newApp.ID
					newAppDetails = append(newAppDetails, detail)
				}
			}
		}
		for _, update := range updateApps {
			for _, detail := range update.Details {
				if detail.ID == 0 {
					detail.AppId = update.ID
					newAppDetails = append(newAppDetails, detail)
				} else {
					if detail.Status == constant.AppNormal {
						updateDetails = append(updateDetails, detail)
					} else {
						deleteAppDetails = append(deleteAppDetails, detail)
					}
				}
			}
		}

		allApps := append(newApps, updateApps...)
		for _, app := range allApps {
			for _, t := range app.TagsKey {
				tagId, ok := tagMap[t]
				if ok {
					appTags = append(appTags, &model.AppTag{
						AppId: app.ID,
						TagId: tagId,
					})
				}
			}
		}

		if len(newAppDetails) > 0 {
			if err = appDetailRepo.BatchCreate(ctx, newAppDetails); err != nil {
				return
			}
		}

		for _, updateAppDetail := range updateDetails {
			if err = appDetailRepo.Update(ctx, updateAppDetail); err != nil {
				return
			}
		}

		if len(deleteAppDetails) > 0 {
			if err = appDetailRepo.BatchDelete(ctx, deleteAppDetails); err != nil {
				return
			}
		}

		if len(oldAppIds) > 0 {
			if err = appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
				return
			}
		}

		if len(appTags) > 0 {
			if err = appTagRepo.BatchCreate(ctx, appTags); err != nil {
				return
			}
		}
		tx.Commit()
		global.LOG.Infof("Synchronization of local applications completed")
		return nil
	}, nil)
	go func() {
		_ = syncTask.Execute()
	}()
}

func (a AppService) GetAppUpdate() (*response.AppUpdateRes, error) {
	res := &response.AppUpdateRes{
		CanUpdate: false,
	}

	versionUrl := fmt.Sprintf("%s/%s/1panel.json.version.txt", global.CONF.RemoteURL.AppRepo, global.CONF.Base.Mode)
	_, versionRes, err := req_helper.HandleRequest(versionUrl, http.MethodGet, constant.TimeOut20s)
	if err != nil {
		return nil, err
	}
	lastModifiedStr := string(versionRes)
	lastModified, err := strconv.Atoi(lastModifiedStr)
	if err != nil {
		return nil, err
	}
	setting, err := NewISettingService().GetSettingInfo()
	if err != nil {
		return nil, err
	}
	if setting.AppStoreSyncStatus == constant.StatusSyncing {
		res.IsSyncing = true
		return res, nil
	}

	appStoreLastModified, _ := strconv.Atoi(setting.AppStoreLastModified)
	res.AppStoreLastModified = appStoreLastModified
	if setting.AppStoreLastModified == "" || lastModified != appStoreLastModified {
		res.CanUpdate = true
		return res, err
	}
	apps, _ := appRepo.GetBy(appRepo.WithResource(constant.AppResourceRemote))
	for _, app := range apps {
		if app.Icon == "" {
			res.CanUpdate = true
			return res, err
		}
	}

	list, err := getAppList()
	if err != nil {
		return res, err
	}
	if list.Extra.Version != "" && setting.SystemVersion != list.Extra.Version && !common.CompareVersion(setting.SystemVersion, list.Extra.Version) {
		global.LOG.Errorf("The current version %s is too low to synchronize with the App Store. The minimum required version is %s", setting.SystemVersion, list.Extra.Version)
		return nil, buserr.New("ErrVersionTooLow")
	}
	res.AppList = list
	return res, nil
}

func getAppFromRepo(downloadPath string) error {
	downloadUrl := downloadPath
	global.LOG.Infof("[AppStore] download file from %s", downloadUrl)
	fileOp := files.NewFileOp()
	packagePath := filepath.Join(global.Dir.ResourceDir, filepath.Base(downloadUrl))
	if err := files.DownloadFileWithProxy(downloadUrl, packagePath); err != nil {
		return err
	}

	if err := fileOp.Decompress(packagePath, global.Dir.ResourceDir, files.SdkZip, ""); err != nil {
		return err
	}
	defer func() {
		_ = fileOp.DeleteFile(packagePath)
	}()
	return nil
}

func getAppList() (*dto.AppList, error) {
	list := &dto.AppList{}
	if err := getAppFromRepo(fmt.Sprintf("%s/%s/1panel.json.zip", global.CONF.RemoteURL.AppRepo, global.CONF.Base.Mode)); err != nil {
		return nil, err
	}
	listFile := filepath.Join(global.Dir.ResourceDir, "1panel.json")
	content, err := os.ReadFile(listFile)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(content, list); err != nil {
		return nil, err
	}
	return list, nil
}

var InitTypes = map[string]struct{}{
	"runtime": {},
	"php":     {},
	"node":    {},
}

func deleteCustomApp() {
	var appIDS []uint
	installs, _ := appInstallRepo.ListBy(context.Background())
	for _, install := range installs {
		appIDS = append(appIDS, install.AppId)
	}
	var ops []repo.DBOption
	ops = append(ops, repo.WithByIDNotIn(appIDS))
	if len(appIDS) > 0 {
		ops = append(ops, repo.WithByIDNotIn(appIDS))
	}
	apps, _ := appRepo.GetBy(ops...)
	var deleteIDS []uint
	for _, app := range apps {
		if app.Resource == constant.AppResourceCustom {
			deleteIDS = append(deleteIDS, app.ID)
		}
	}
	_ = appRepo.DeleteByIDs(context.Background(), deleteIDS)
	_ = appDetailRepo.DeleteByAppIds(context.Background(), deleteIDS)
}

func (a AppService) SyncAppListFromRemote(taskID string) (err error) {
	if xpack.IsUseCustomApp() {
		return nil
	}
	syncTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("App"), task.TaskSync, task.TaskScopeAppStore, taskID, 0)
	if err != nil {
		return err
	}
	syncTask.AddSubTask(task.GetTaskName(i18n.GetMsgByKey("App"), task.TaskSync, task.TaskScopeAppStore), func(t *task.Task) (err error) {
		updateRes, err := a.GetAppUpdate()
		if err != nil {
			return err
		}
		if !updateRes.CanUpdate {
			if updateRes.IsSyncing {
				t.Log(i18n.GetMsgByKey("AppStoreIsSyncing"))
				return nil
			}
			t.Log(i18n.GetMsgByKey("AppStoreIsUpToDate"))
			return nil
		}
		list := &dto.AppList{}
		if updateRes.AppList == nil {
			list, err = getAppList()
			if err != nil {
				return err
			}
		} else {
			list = updateRes.AppList
		}
		settingService := NewISettingService()
		_ = settingService.Update("AppStoreSyncStatus", constant.StatusSyncing)

		setting, err := settingService.GetSettingInfo()
		if err != nil {
			return err
		}
		var (
			appTags   []*model.AppTag
			oldAppIds []uint
		)
		if err = SyncTags(list.Extra); err != nil {
			return err
		}
		deleteCustomApp()
		oldApps, err := appRepo.GetBy(appRepo.WithNotLocal())
		if err != nil {
			return err
		}
		for _, old := range oldApps {
			oldAppIds = append(oldAppIds, old.ID)
		}

		baseRemoteUrl := fmt.Sprintf("%s/%s/1panel", global.CONF.RemoteURL.AppRepo, global.CONF.Base.Mode)

		appsMap := getApps(oldApps, list.Apps, setting.SystemVersion, t)

		t.LogStart(i18n.GetMsgByKey("SyncAppDetail"))
		for _, l := range list.Apps {
			app, ok := appsMap[l.AppProperty.Key]
			if !ok {
				continue
			}
			iconStr := ""
			_, iconRes, err := req_helper.HandleRequest(l.Icon, http.MethodGet, constant.TimeOut20s)
			if err == nil {
				if !strings.Contains(string(iconRes), "<xml>") {
					iconStr = base64.StdEncoding.EncodeToString(iconRes)
				}
			}
			app.Icon = iconStr
			app.TagsKey = l.AppProperty.Tags
			if l.AppProperty.Recommend > 0 {
				app.Recommend = l.AppProperty.Recommend
			} else {
				app.Recommend = 9999
			}
			app.ReadMe = l.ReadMe
			app.LastModified = l.LastModified
			versions := l.Versions
			detailsMap := getAppDetails(app.Details, versions)
			for _, v := range versions {
				version := v.Name
				detail := detailsMap[version]
				versionUrl := fmt.Sprintf("%s/%s/%s", baseRemoteUrl, app.Key, version)
				paramByte, _ := json.Marshal(v.AppForm)
				var appForm dto.AppForm
				_ = json.Unmarshal(paramByte, &appForm)
				if appForm.SupportVersion > 0 && common.CompareVersion(strconv.FormatFloat(appForm.SupportVersion, 'f', -1, 64), setting.SystemVersion) {
					delete(detailsMap, version)
					continue
				}
				if _, ok := InitTypes[app.Type]; ok {
					dockerComposeUrl := fmt.Sprintf("%s/%s", versionUrl, "docker-compose.yml")
					_, composeRes, err := req_helper.HandleRequest(dockerComposeUrl, http.MethodGet, constant.TimeOut20s)
					if err == nil {
						detail.DockerCompose = string(composeRes)
					}
				} else {
					detail.DockerCompose = ""
				}

				detail.Params = string(paramByte)
				detail.DownloadUrl = fmt.Sprintf("%s/%s", versionUrl, app.Key+"-"+version+".tar.gz")
				detail.DownloadCallBackUrl = v.DownloadCallBackUrl
				detail.Update = true
				detail.LastModified = v.LastModified
				detailsMap[version] = detail
			}
			var newDetails []model.AppDetail
			for _, detail := range detailsMap {
				newDetails = append(newDetails, detail)
			}
			app.Details = newDetails
			appsMap[l.AppProperty.Key] = app
		}
		t.LogSuccess(i18n.GetMsgByKey("SyncAppDetail"))

		tags, _ := tagRepo.All()
		var (
			addAppArray    []model.App
			updateAppArray []model.App
			deleteAppArray []model.App
			deleteIds      []uint
			tagMap         = make(map[string]uint, len(tags))
		)

		for _, v := range appsMap {
			if v.ID == 0 {
				addAppArray = append(addAppArray, v)
			} else {
				if v.Status == constant.AppTakeDown {
					installs, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(v.ID))
					if len(installs) > 0 {
						updateAppArray = append(updateAppArray, v)
						continue
					}
					deleteAppArray = append(deleteAppArray, v)
					deleteIds = append(deleteIds, v.ID)
				} else {
					updateAppArray = append(updateAppArray, v)
				}
			}
		}

		tx, ctx := getTxAndContext()
		defer func() {
			if err != nil {
				tx.Rollback()
				return
			}
		}()
		if len(addAppArray) > 0 {
			if err = appRepo.BatchCreate(ctx, addAppArray); err != nil {
				return
			}
		}
		if len(deleteAppArray) > 0 {
			if err = appRepo.BatchDelete(ctx, deleteAppArray); err != nil {
				return
			}
			if err = appDetailRepo.DeleteByAppIds(ctx, deleteIds); err != nil {
				return
			}
		}
		for _, tag := range tags {
			tagMap[tag.Key] = tag.ID
		}
		for _, update := range updateAppArray {
			if err = appRepo.Save(ctx, &update); err != nil {
				return
			}
		}
		apps := append(addAppArray, updateAppArray...)

		var (
			addDetails    []model.AppDetail
			updateDetails []model.AppDetail
			deleteDetails []model.AppDetail
		)
		for _, app := range apps {
			for _, tag := range app.TagsKey {
				tagId, ok := tagMap[tag]
				if ok {
					exist, _ := appTagRepo.GetFirst(ctx, appTagRepo.WithByTagID(tagId), appTagRepo.WithByAppID(app.ID))
					if exist == nil {
						appTags = append(appTags, &model.AppTag{
							AppId: app.ID,
							TagId: tagId,
						})
					}
				}
			}
			for _, d := range app.Details {
				d.AppId = app.ID
				if d.ID == 0 {
					addDetails = append(addDetails, d)
				} else {
					if d.Status == constant.AppTakeDown {
						runtime, _ := runtimeRepo.GetFirst(ctx, runtimeRepo.WithDetailId(d.ID))
						if runtime != nil {
							updateDetails = append(updateDetails, d)
							continue
						}
						installs, _ := appInstallRepo.ListBy(ctx, appInstallRepo.WithDetailIdsIn([]uint{d.ID}))
						if len(installs) > 0 {
							updateDetails = append(updateDetails, d)
							continue
						}
						deleteDetails = append(deleteDetails, d)
					} else {
						updateDetails = append(updateDetails, d)
					}
				}
			}
		}
		if len(addDetails) > 0 {
			if err = appDetailRepo.BatchCreate(ctx, addDetails); err != nil {
				return
			}
		}
		if len(deleteDetails) > 0 {
			if err = appDetailRepo.BatchDelete(ctx, deleteDetails); err != nil {
				return
			}
		}
		for _, u := range updateDetails {
			if err = appDetailRepo.Update(ctx, u); err != nil {
				return
			}
		}

		if len(oldAppIds) > 0 {
			if err = appTagRepo.DeleteByAppIds(ctx, deleteIds); err != nil {
				return
			}
		}

		if len(appTags) > 0 {
			if err = appTagRepo.BatchCreate(ctx, appTags); err != nil {
				return
			}
		}
		tx.Commit()

		_ = settingService.Update("AppStoreSyncStatus", constant.StatusSyncSuccess)
		_ = settingService.Update("AppStoreLastModified", strconv.Itoa(list.LastModified))
		return nil
	}, nil)

	go func() {
		if err = syncTask.Execute(); err != nil {
			_ = NewISettingService().Update("AppStoreLastModified", "0")
			_ = NewISettingService().Update("AppStoreSyncStatus", constant.StatusError)
		}
	}()

	return nil
}
