package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	http2 "github.com/1Panel-dev/1Panel/backend/utils/http"
	httpUtil "github.com/1Panel-dev/1Panel/backend/utils/http"
	"github.com/1Panel-dev/1Panel/backend/utils/xpack"
	"gopkg.in/yaml.v3"
)

type AppService struct {
}

type IAppService interface {
	PageApp(ctx *gin.Context, req request.AppSearch) (interface{}, error)
	GetAppTags(ctx *gin.Context) ([]response.TagDTO, error)
	GetApp(ctx *gin.Context, key string) (*response.AppDTO, error)
	GetAppDetail(appId uint, version, appType string) (response.AppDetailDTO, error)
	Install(ctx context.Context, req request.AppInstallCreate) (*model.AppInstall, error)
	SyncAppListFromRemote() error
	GetAppUpdate() (*response.AppUpdateRes, error)
	GetAppDetailByID(id uint) (*response.AppDetailDTO, error)
	SyncAppListFromLocal()
	GetIgnoredApp() ([]response.IgnoredApp, error)
}

func NewIAppService() IAppService {
	return &AppService{}
}

func (a AppService) PageApp(ctx *gin.Context, req request.AppSearch) (interface{}, error) {
	var opts []repo.DBOption
	opts = append(opts, appRepo.OrderByRecommend())
	if req.Name != "" {
		opts = append(opts, appRepo.WithLikeName(req.Name))
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
	var res response.AppRes
	total, apps, err := appRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, err
	}
	var appDTOs []*response.AppItem
	lang := strings.ToLower(common.GetLang(ctx))
	for _, ap := range apps {
		appDTO := &response.AppItem{
			ID:       ap.ID,
			Name:     ap.Name,
			Key:      ap.Key,
			Type:     ap.Type,
			Icon:     ap.Icon,
			Resource: ap.Resource,
			Limit:    ap.Limit,
		}
		appDTO.Description = ap.GetDescription(ctx)
		appDTOs = append(appDTOs, appDTO)
		appTags, err := appTagRepo.GetByAppId(ap.ID)
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
		for _, t := range tags {
			if t.Name != "" {
				tagDTO := response.TagDTO{
					ID:   t.ID,
					Key:  t.Key,
					Name: t.Name,
				}
				appDTO.Tags = append(appDTO.Tags, tagDTO)
			} else {
				var translations = make(map[string]string)
				_ = json.Unmarshal([]byte(t.Translations), &translations)
				if name, ok := translations[lang]; ok {
					tagDTO := response.TagDTO{
						ID:   t.ID,
						Key:  t.Key,
						Name: name,
					}
					appDTO.Tags = append(appDTO.Tags, tagDTO)
				}
			}
		}
		installs, _ := appInstallRepo.ListBy(appInstallRepo.WithAppId(ap.ID))
		appDTO.Installed = len(installs) > 0
	}
	res.Items = appDTOs
	res.Total = total

	return res, nil
}

func (a AppService) GetAppTags(ctx *gin.Context) ([]response.TagDTO, error) {
	tags, _ := tagRepo.All()
	res := make([]response.TagDTO, 0)
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
	for _, detail := range details {
		versionsRaw = append(versionsRaw, detail.Version)
	}
	appDTO.Versions = common.GetSortedVersions(versionsRaw)

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
		app, err := appRepo.GetFirst(commonRepo.WithByID(appID))
		if err != nil {
			return appDetailDTO, err
		}
		fileOp := files.NewFileOp()

		versionPath := filepath.Join(app.GetAppResourcePath(), detail.Version)
		if !fileOp.Stat(versionPath) || detail.Update {
			if err = downloadApp(app, detail, nil); err != nil && !fileOp.Stat(versionPath) {
				return appDetailDTO, err
			}
		}
		switch app.Type {
		case constant.RuntimePHP:
			buildPath := filepath.Join(versionPath, "build")
			paramsPath := filepath.Join(buildPath, "config.json")
			if !fileOp.Stat(paramsPath) {
				return appDetailDTO, buserr.New(constant.ErrFileNotExist)
			}
			param, err := fileOp.GetContent(paramsPath)
			if err != nil {
				return appDetailDTO, err
			}
			paramMap := make(map[string]interface{})
			if err := json.Unmarshal(param, &paramMap); err != nil {
				return appDetailDTO, err
			}
			appDetailDTO.Params = paramMap
			composePath := filepath.Join(buildPath, "docker-compose.yml")
			if !fileOp.Stat(composePath) {
				return appDetailDTO, buserr.New(constant.ErrFileNotExist)
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
		statusCode, composeRes, err := httpUtil.HandleGet(dockerComposeUrl, http.MethodGet, constant.TimeOut20s)
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

	app, err := appRepo.GetFirst(commonRepo.WithByID(detail.AppId))
	if err != nil {
		return appDetailDTO, err
	}
	if err := checkLimit(app); err != nil {
		appDetailDTO.Enable = false
	}
	return appDetailDTO, nil
}
func (a AppService) GetAppDetailByID(id uint) (*response.AppDetailDTO, error) {
	res := &response.AppDetailDTO{}
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(id))
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

func (a AppService) GetIgnoredApp() ([]response.IgnoredApp, error) {
	var res []response.IgnoredApp
	details, _ := appDetailRepo.GetBy(appDetailRepo.WithIgnored())
	if len(details) == 0 {
		return res, nil
	}
	for _, detail := range details {
		app, err := appRepo.GetFirst(commonRepo.WithByID(detail.AppId))
		if err != nil {
			return nil, err
		}
		res = append(res, response.IgnoredApp{
			Name:     app.Name,
			Version:  detail.Version,
			DetailID: detail.ID,
			Icon:     app.Icon,
		})
	}
	return res, nil
}

func (a AppService) Install(ctx context.Context, req request.AppInstallCreate) (appInstall *model.AppInstall, err error) {
	if err = docker.CreateDefaultDockerNetwork(); err != nil {
		err = buserr.WithDetail(constant.Err1PanelNetworkFailed, err.Error(), nil)
		return
	}
	if list, _ := appInstallRepo.ListBy(commonRepo.WithByName(req.Name)); len(list) > 0 {
		err = buserr.New(constant.ErrAppNameExist)
		return
	}
	var (
		httpPort  int
		httpsPort int
		appDetail model.AppDetail
		app       model.App
	)
	appDetail, err = appDetailRepo.GetFirst(commonRepo.WithByID(req.AppDetailId))
	if err != nil {
		return
	}
	app, err = appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	if err != nil {
		return
	}
	if DatabaseKeys[app.Key] > 0 {
		if existDatabases, _ := databaseRepo.GetList(commonRepo.WithByName(req.Name)); len(existDatabases) > 0 {
			err = buserr.New(constant.ErrRemoteExist)
			return
		}
	}
	for key := range req.Params {
		if !strings.Contains(key, "PANEL_APP_PORT") {
			continue
		}
		var port int
		if port, err = checkPort(key, req.Params); err == nil {
			if key == "PANEL_APP_PORT_HTTP" {
				httpPort = port
			}
			if key == "PANEL_APP_PORT_HTTPS" {
				httpsPort = port
			}
		} else {
			return
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
		Status:      constant.Installing,
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
		err = buserr.New(constant.ErrFileParse)
		return
	}
	servicesMap := value.(map[string]interface{})
	containerName := constant.ContainerPrefix + app.Key + "-" + common.RandStr(4)
	if req.Advanced && req.ContainerName != "" {
		containerName = req.ContainerName
		appInstalls, _ := appInstallRepo.ListBy(appInstallRepo.WithContainerName(containerName))
		if len(appInstalls) > 0 {
			err = buserr.New(constant.ErrContainerName)
			return
		}
		containerExist := false
		containerExist, err = checkContainerNameIsExist(req.ContainerName, appInstall.GetPath())
		if err != nil {
			return
		}
		if containerExist {
			err = buserr.New(constant.ErrContainerName)
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

	defer func() {
		if err != nil {
			hErr := handleAppInstallErr(ctx, appInstall)
			if hErr != nil {
				global.LOG.Errorf("delete app dir error %s", hErr.Error())
			}
		}
	}()
	if hostName, ok := req.Params["PANEL_DB_HOST"]; ok {
		database, _ := databaseRepo.Get(commonRepo.WithByName(hostName.(string)))
		if !reflect.DeepEqual(database, model.Database{}) {
			req.Params["PANEL_DB_HOST"] = database.Address
			req.Params["PANEL_DB_PORT"] = database.Port
			req.Params["PANEL_DB_HOST_NAME"] = hostName
		}
	}
	paramByte, err = json.Marshal(req.Params)
	if err != nil {
		return
	}
	appInstall.Env = string(paramByte)

	if err = appInstallRepo.Create(ctx, appInstall); err != nil {
		return
	}
	if err = createLink(ctx, app, appInstall, req.Params); err != nil {
		return
	}
	go func() {
		defer func() {
			if err != nil {
				appInstall.Status = constant.UpErr
				appInstall.Message = err.Error()
				_ = appInstallRepo.Save(context.Background(), appInstall)
			}
		}()
		if err = copyData(app, appDetail, appInstall, req); err != nil {
			return
		}
		if err = runScript(appInstall, "init"); err != nil {
			return
		}
		upApp(appInstall, req.PullImage)
	}()
	go updateToolApp(appInstall)
	return
}

func (a AppService) SyncAppListFromLocal() {
	fileOp := files.NewFileOp()
	localAppDir := constant.LocalAppResourceDir
	if !fileOp.Stat(localAppDir) {
		return
	}
	var (
		err        error
		dirEntries []os.DirEntry
		localApps  []model.App
	)

	defer func() {
		if err != nil {
			global.LOG.Errorf("Sync local app failed %v", err)
		}
	}()

	global.LOG.Infof("Starting local application synchronization ...")
	dirEntries, err = os.ReadDir(localAppDir)
	if err != nil {
		return
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			appDir := filepath.Join(localAppDir, dirEntry.Name())
			appDirEntries, err := os.ReadDir(appDir)
			if err != nil {
				global.LOG.Errorf(i18n.GetMsgWithMap("ErrAppDirNull", map[string]interface{}{"name": dirEntry.Name(), "err": err.Error()}))
				continue
			}
			app, err := handleLocalApp(appDir)
			if err != nil {
				global.LOG.Errorf(i18n.GetMsgWithMap("LocalAppErr", map[string]interface{}{"name": dirEntry.Name(), "err": err.Error()}))
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
						global.LOG.Errorf(i18n.GetMsgWithMap("LocalAppVersionErr", map[string]interface{}{"name": app.Name, "version": appDetail.Version, "err": err.Error()}))
						continue
					}
					appDetails = append(appDetails, appDetail)
				}
			}
			if len(appDetails) > 0 {
				app.Details = appDetails
				localApps = append(localApps, *app)
			} else {
				global.LOG.Errorf(i18n.GetMsgWithMap("LocalAppVersionNull", map[string]interface{}{"name": app.Name}))
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
				installs, _ := appInstallRepo.ListBy(appInstallRepo.WithAppId(app.ID))
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
}

func (a AppService) GetAppUpdate() (*response.AppUpdateRes, error) {
	res := &response.AppUpdateRes{
		CanUpdate: false,
	}

	versionUrl := fmt.Sprintf("%s/%s/1panel.json.version.txt", global.CONF.System.AppRepo, global.CONF.System.Mode)
	_, versionRes, err := http2.HandleGet(versionUrl, http.MethodGet, constant.TimeOut20s)
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
	if setting.AppStoreSyncStatus == constant.Syncing {
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
		global.LOG.Errorf("The current version is too low to synchronize with the App Store. The minimum required version is %s", list.Extra.Version)
		return nil, buserr.New("ErrVersionTooLow")
	}
	res.AppList = list
	return res, nil
}

func getAppFromRepo(downloadPath string) error {
	downloadUrl := downloadPath
	global.LOG.Infof("[AppStore] download file from %s", downloadUrl)
	fileOp := files.NewFileOp()
	packagePath := filepath.Join(constant.ResourceDir, filepath.Base(downloadUrl))
	if err := fileOp.DownloadFileWithProxy(downloadUrl, packagePath); err != nil {
		return err
	}
	if err := fileOp.Decompress(packagePath, constant.ResourceDir, files.SdkZip, ""); err != nil {
		return err
	}
	defer func() {
		_ = fileOp.DeleteFile(packagePath)
	}()
	return nil
}

func getAppList() (*dto.AppList, error) {
	list := &dto.AppList{}
	if err := getAppFromRepo(fmt.Sprintf("%s/%s/1panel.json.zip", global.CONF.System.AppRepo, global.CONF.System.Mode)); err != nil {
		return nil, err
	}
	listFile := filepath.Join(constant.ResourceDir, "1panel.json")
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

func (a AppService) SyncAppListFromRemote() (err error) {
	global.LOG.Infof("Starting synchronization with App Store...")
	updateRes, err := a.GetAppUpdate()
	if err != nil {
		return err
	}
	if !updateRes.CanUpdate {
		if updateRes.IsSyncing {
			global.LOG.Infof("AppStore is Syncing!")
			return
		}
		global.LOG.Infof("The App Store is at the latest version")
		return
	}

	list := &dto.AppList{}
	if updateRes.AppList == nil {
		list, err = getAppList()
		if err != nil {
			return
		}
	} else {
		list = updateRes.AppList
	}
	settingService := NewISettingService()
	_ = settingService.Update("AppStoreSyncStatus", constant.Syncing)

	defer func() {
		if err != nil {
			_ = settingService.Update("AppStoreSyncStatus", constant.SyncFailed)
			global.LOG.Errorf("App Store synchronization failed %v", err)
		}
	}()

	setting, err := settingService.GetSettingInfo()
	if err != nil {
		return err
	}
	var (
		tags      []*model.Tag
		appTags   []*model.AppTag
		oldAppIds []uint
	)
	for _, t := range list.Extra.Tags {
		translations, _ := json.Marshal(t.Locales)
		tags = append(tags, &model.Tag{
			Key:          t.Key,
			Translations: string(translations),
			Sort:         t.Sort,
		})
	}
	oldApps, err := appRepo.GetBy(appRepo.WithResource(constant.AppResourceRemote))
	if err != nil {
		return
	}
	for _, old := range oldApps {
		oldAppIds = append(oldAppIds, old.ID)
	}

	transport := xpack.LoadRequestTransport()
	baseRemoteUrl := fmt.Sprintf("%s/%s/1panel", global.CONF.System.AppRepo, global.CONF.System.Mode)
	appsMap := getApps(oldApps, list.Apps)

	global.LOG.Infof("Starting synchronization of application details...")
	for _, l := range list.Apps {
		app := appsMap[l.AppProperty.Key]

		if l.AppProperty.Version > 0 && common.CompareVersion(strconv.FormatFloat(l.AppProperty.Version, 'f', -1, 64), setting.SystemVersion) {
			delete(appsMap, l.AppProperty.Key)
			continue
		}

		_, iconRes, err := httpUtil.HandleGetWithTransport(l.Icon, http.MethodGet, transport, constant.TimeOut20s)
		if err != nil {
			return err
		}
		iconStr := ""
		if !strings.Contains(string(iconRes), "<xml>") {
			iconStr = base64.StdEncoding.EncodeToString(iconRes)
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
				_, composeRes, err := httpUtil.HandleGetWithTransport(dockerComposeUrl, http.MethodGet, transport, constant.TimeOut20s)
				if err != nil {
					return err
				}
				detail.DockerCompose = string(composeRes)
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

	global.LOG.Infof("Synchronization of application details Success")

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
				installs, _ := appInstallRepo.ListBy(appInstallRepo.WithAppId(v.ID))
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
	defer tx.Rollback()
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
	if err = tagRepo.DeleteAll(ctx); err != nil {
		return
	}
	if len(tags) > 0 {
		if err = tagRepo.BatchCreate(ctx, tags); err != nil {
			return
		}
		for _, t := range tags {
			tagMap[t.Key] = t.ID
		}
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
		for _, t := range app.TagsKey {
			tagId, ok := tagMap[t]
			if ok {
				appTags = append(appTags, &model.AppTag{
					AppId: app.ID,
					TagId: tagId,
				})
			}
		}
		for _, d := range app.Details {
			d.AppId = app.ID
			if d.ID == 0 {
				addDetails = append(addDetails, d)
			} else {
				if d.Status == constant.AppTakeDown {
					runtime, _ := runtimeRepo.GetFirst(runtimeRepo.WithDetailId(d.ID))
					if runtime != nil {
						updateDetails = append(updateDetails, d)
						continue
					}
					installs, _ := appInstallRepo.ListBy(appInstallRepo.WithDetailIdsIn([]uint{d.ID}))
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

	_ = settingService.Update("AppStoreSyncStatus", constant.SyncSuccess)
	_ = settingService.Update("AppStoreLastModified", strconv.Itoa(list.LastModified))

	global.LOG.Infof("Synchronization with the App Store was successful!")
	return
}
