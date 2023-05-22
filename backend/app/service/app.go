package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

type AppService struct {
}

type IAppService interface {
	PageApp(req request.AppSearch) (interface{}, error)
	GetAppTags() ([]response.TagDTO, error)
	GetApp(key string) (*response.AppDTO, error)
	GetAppDetail(appId uint, version, appType string) (response.AppDetailDTO, error)
	Install(ctx context.Context, req request.AppInstallCreate) (*model.AppInstall, error)
	SyncAppListFromRemote() error
	GetAppUpdate() (*response.AppUpdateRes, error)
	GetAppDetailByID(id uint) (*response.AppDetailDTO, error)
	SyncAppListFromLocal()
}

func NewIAppService() IAppService {
	return &AppService{}
}

func (a AppService) PageApp(req request.AppSearch) (interface{}, error) {
	var opts []repo.DBOption
	opts = append(opts, appRepo.OrderByRecommend())
	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}
	if req.Type != "" {
		opts = append(opts, appRepo.WithType(req.Type))
	}
	if req.Recommend {
		opts = append(opts, appRepo.GetRecommend())
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
	var appDTOs []*response.AppDTO
	for _, a := range apps {
		appDTO := &response.AppDTO{
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

	return res, nil
}

func (a AppService) GetAppTags() ([]response.TagDTO, error) {
	tags, err := tagRepo.All()
	if err != nil {
		return nil, err
	}
	var res []response.TagDTO
	for _, tag := range tags {
		res = append(res, response.TagDTO{
			Tag: tag,
		})
	}
	return res, nil
}

func (a AppService) GetApp(key string) (*response.AppDTO, error) {
	var appDTO response.AppDTO
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return nil, err
	}
	appDTO.App = app
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

func (a AppService) GetAppDetail(appId uint, version, appType string) (response.AppDetailDTO, error) {
	var (
		appDetailDTO response.AppDetailDTO
		opts         []repo.DBOption
	)
	opts = append(opts, appDetailRepo.WithAppId(appId), appDetailRepo.WithVersion(version))
	detail, err := appDetailRepo.GetFirst(opts...)
	if err != nil {
		return appDetailDTO, err
	}
	appDetailDTO.AppDetail = detail
	appDetailDTO.Enable = true

	if appType == "runtime" {
		app, err := appRepo.GetFirst(commonRepo.WithByID(appId))
		if err != nil {
			return appDetailDTO, err
		}
		fileOp := files.NewFileOp()
		buildPath := path.Join(constant.AppResourceDir, app.Key, "versions", detail.Version, "build")
		paramsPath := path.Join(buildPath, "config.json")
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
		composePath := path.Join(buildPath, "docker-compose.yml")
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
	} else {
		paramMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(detail.Params), &paramMap); err != nil {
			return appDetailDTO, err
		}
		appDetailDTO.Params = paramMap
	}

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
	return res, nil
}

func (a AppService) Install(ctx context.Context, req request.AppInstallCreate) (appInstall *model.AppInstall, err error) {
	if err = docker.CreateDefaultDockerNetwork(); err != nil {
		err = buserr.WithDetail(constant.Err1PanelNetworkFailed, err.Error(), nil)
		return
	}
	if list, _ := appInstallRepo.ListBy(commonRepo.WithByName(req.Name)); len(list) > 0 {
		err = buserr.New(constant.ErrNameIsExist)
		return
	}
	var (
		httpPort  int
		httpsPort int
		appDetail model.AppDetail
		app       model.App
	)
	httpPort, err = checkPort("PANEL_APP_PORT_HTTP", req.Params)
	if err != nil {
		return
	}
	httpsPort, err = checkPort("PANEL_APP_PORT_HTTPS", req.Params)
	if err != nil {
		return
	}
	appDetail, err = appDetailRepo.GetFirst(commonRepo.WithByID(req.AppDetailId))
	if err != nil {
		return
	}
	app, err = appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	if err != nil {
		return
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
	if !ok {
		err = buserr.New(constant.ErrFileParse)
		return
	}
	servicesMap := value.(map[string]interface{})
	changeKeys := make(map[string]string, len(servicesMap))
	index := 0
	for k := range servicesMap {
		serviceName := k + "-" + common.RandStr(4)
		changeKeys[k] = serviceName
		containerName := constant.ContainerPrefix + k + "-" + common.RandStr(4)
		if req.Advanced && req.ContainerName != "" {
			containerName = req.ContainerName
		}
		if index > 0 {
			continue
		}
		req.Params[constant.ContainerName] = containerName
		appInstall.ServiceName = serviceName
		appInstall.ContainerName = containerName
		index++
	}
	for k, v := range changeKeys {
		servicesMap[v] = servicesMap[k]
		delete(servicesMap, k)
	}

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
	if err = upAppPre(app, appInstall); err != nil {
		return
	}
	go func() {
		if err = copyData(app, appDetail, appInstall, req); err != nil {
			if appInstall.Status == constant.Installing {
				appInstall.Status = constant.Error
				appInstall.Message = err.Error()
			}
			_ = appInstallRepo.Save(context.Background(), appInstall)
			return
		}
		go func() {
			_, _ = http.Get(appDetail.DownloadCallBackUrl)
		}()
		upApp(appInstall)
	}()
	go updateToolApp(appInstall)
	return
}

func (a AppService) GetAppUpdate() (*response.AppUpdateRes, error) {
	res := &response.AppUpdateRes{
		CanUpdate: false,
	}
	setting, err := NewISettingService().GetSettingInfo()
	if err != nil {
		return nil, err
	}
	versionUrl := fmt.Sprintf("%s/%s/1panel.json", global.CONF.System.AppRepo, global.CONF.System.Mode)
	versionRes, err := http.Get(versionUrl)
	if err != nil {
		return nil, err
	}
	defer versionRes.Body.Close()
	body, err := io.ReadAll(versionRes.Body)
	if err != nil {
		return nil, err
	}
	list := &dto.AppList{}
	if err = json.Unmarshal(body, list); err != nil {
		return nil, err
	}
	res.AppStoreLastModified = list.LastModified
	res.List = *list

	appStoreLastModified, _ := strconv.Atoi(setting.AppStoreLastModified)
	if setting.AppStoreLastModified == "" || list.LastModified > appStoreLastModified {
		res.CanUpdate = true
		return res, err
	}
	return res, nil
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
			global.LOG.Errorf("sync app failed %v", err)
		}
	}()

	global.LOG.Infof("start sync local apps...")
	dirEntries, err = os.ReadDir(localAppDir)
	if err != nil {
		return
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			appDir := path.Join(localAppDir, dirEntry.Name())
			appDirEntries, err := os.ReadDir(appDir)
			if err != nil || len(appDirEntries) == 0 {
				continue
			}
			configYamlPath := path.Join(appDir, "data.yml")
			if !fileOp.Stat(configYamlPath) {
				continue
			}
			iconPath := path.Join(appDir, "logo.png")
			if !fileOp.Stat(iconPath) {
				continue
			}
			configYamlByte, err := fileOp.GetContent(configYamlPath)
			if err != nil {
				continue
			}
			localAppDefine := dto.LocalAppAppDefine{}
			if err := yaml.Unmarshal(configYamlByte, &localAppDefine); err != nil {
				continue
			}
			app := localAppDefine.AppProperty
			app.Resource = constant.AppResourceLocal
			app.Status = constant.AppNormal
			app.Recommend = 9999
			app.TagsKey = append(app.TagsKey, "Local")
			app.Key = "local" + app.Key
			readMePath := path.Join(appDir, "README.md")
			if fileOp.Stat(configYamlPath) {
				readMeByte, err := fileOp.GetContent(readMePath)
				if err == nil {
					app.ReadMe = string(readMeByte)
				}
			}

			iconByte, _ := fileOp.GetContent(iconPath)
			if iconByte != nil {
				iconStr := base64.StdEncoding.EncodeToString(iconByte)
				app.Icon = iconStr
			}
			var appDetails []model.AppDetail
			for _, appDirEntry := range appDirEntries {
				if appDirEntry.IsDir() {
					appDetail := model.AppDetail{
						Version: appDirEntry.Name(),
						Status:  constant.AppNormal,
					}
					versionDir := path.Join(appDir, appDirEntry.Name())
					dockerComposePath := path.Join(versionDir, "docker-compose.yml")
					if !fileOp.Stat(dockerComposePath) {
						continue
					}
					dockerComposeByte, _ := fileOp.GetContent(dockerComposePath)
					if dockerComposeByte == nil {
						continue
					}
					appDetail.DockerCompose = string(dockerComposeByte)
					paramPath := path.Join(versionDir, "data.yml")
					if !fileOp.Stat(paramPath) {
						continue
					}
					paramByte, _ := fileOp.GetContent(paramPath)
					if paramByte == nil {
						continue
					}
					appParamConfig := dto.LocalAppParam{}
					if err := yaml.Unmarshal(paramByte, &appParamConfig); err != nil {
						continue
					}
					dataJson, err := json.Marshal(appParamConfig.AppParams)
					if err != nil {
						continue
					}
					appDetail.Params = string(dataJson)
					appDetails = append(appDetails, appDetail)
				}
			}
			app.Details = appDetails
			localApps = append(localApps, app)
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
				}
				app.Details[i] = newDetail
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
		if err := appRepo.BatchCreate(ctx, newApps); err != nil {
			return
		}
	}
	for _, update := range updateApps {
		if err := appRepo.Save(ctx, &update); err != nil {
			return
		}
	}
	if len(deleteApps) > 0 {
		if err := appRepo.BatchDelete(ctx, deleteApps); err != nil {
			return
		}
		if err := appDetailRepo.DeleteByAppIds(ctx, deleteAppIds); err != nil {
			return
		}
	}

	if err := appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
		return
	}
	var ()

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
		if err := appDetailRepo.BatchCreate(ctx, newAppDetails); err != nil {
			return
		}
	}

	for _, updateAppDetail := range updateDetails {
		if err := appDetailRepo.Update(ctx, updateAppDetail); err != nil {
			return
		}
	}

	if len(deleteAppDetails) > 0 {
		if err := appDetailRepo.BatchDelete(ctx, deleteAppDetails); err != nil {
			return
		}
	}

	if len(oldAppIds) > 0 {
		if err := appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
			return
		}
	}

	if len(appTags) > 0 {
		if err := appTagRepo.BatchCreate(ctx, appTags); err != nil {
			return
		}
	}
	tx.Commit()
	global.LOG.Infof("sync local apps success")
}
func (a AppService) SyncAppListFromRemote() error {
	updateRes, err := a.GetAppUpdate()
	if err != nil {
		return err
	}
	if !updateRes.CanUpdate {
		return nil
	}
	var (
		tags      []*model.Tag
		appTags   []*model.AppTag
		list      = updateRes.List
		oldAppIds []uint
	)
	for _, t := range list.Extra.Tags {
		tags = append(tags, &model.Tag{
			Key:  t.Key,
			Name: t.Name,
		})
	}
	oldApps, err := appRepo.GetBy(appRepo.WithResource(constant.AppResourceRemote))
	if err != nil {
		return err
	}
	for _, old := range oldApps {
		oldAppIds = append(oldAppIds, old.ID)
	}

	baseRemoteUrl := fmt.Sprintf("%s/%s/1panel", global.CONF.System.AppRepo, global.CONF.System.Mode)
	appsMap := getApps(oldApps, list.Apps)
	for _, l := range list.Apps {
		app := appsMap[l.AppProperty.Key]
		iconRes, err := http.Get(l.Icon)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(iconRes.Body)
		if err != nil {
			return err
		}
		iconStr := base64.StdEncoding.EncodeToString(body)
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

			dockerComposeUrl := fmt.Sprintf("%s/%s/%s/%s", baseRemoteUrl, app.Key, version, "docker-compose.yml")
			composeRes, err := http.Get(dockerComposeUrl)
			if err != nil {
				return err
			}
			bodyContent, err := io.ReadAll(composeRes.Body)
			if err != nil {
				return err
			}
			detail.DockerCompose = string(bodyContent)

			paramByte, _ := json.Marshal(v.AppForm)
			detail.Params = string(paramByte)
			detail.DownloadUrl = v.DownloadUrl
			detail.DownloadCallBackUrl = v.DownloadCallBackUrl
			if v.LastModified > detail.LastModified {
				detail.Update = true
				detail.LastModified = v.LastModified
			}
			detailsMap[version] = detail
		}
		var newDetails []model.AppDetail
		for _, detail := range detailsMap {
			newDetails = append(newDetails, detail)
		}
		app.Details = newDetails
		appsMap[l.AppProperty.Key] = app
	}

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
		if err := appRepo.BatchCreate(ctx, addAppArray); err != nil {
			return err
		}
	}
	if len(deleteAppArray) > 0 {
		if err := appRepo.BatchDelete(ctx, deleteAppArray); err != nil {
			return err
		}
		if err := appDetailRepo.DeleteByAppIds(ctx, deleteIds); err != nil {
			return err
		}
	}
	if err := tagRepo.DeleteAll(ctx); err != nil {
		return err
	}
	if len(tags) > 0 {
		if err := tagRepo.BatchCreate(ctx, tags); err != nil {
			return err
		}
		for _, t := range tags {
			tagMap[t.Key] = t.ID
		}
	}
	for _, update := range updateAppArray {
		if err := appRepo.Save(ctx, &update); err != nil {
			return err
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
					deleteDetails = append(deleteDetails, d)
				} else {
					updateDetails = append(updateDetails, d)
				}
			}
		}
	}
	if len(addDetails) > 0 {
		if err := appDetailRepo.BatchCreate(ctx, addDetails); err != nil {
			return err
		}
	}
	if len(deleteDetails) > 0 {
		if err := appDetailRepo.BatchDelete(ctx, addDetails); err != nil {
			return err
		}
	}
	for _, u := range updateDetails {
		if err := appDetailRepo.Update(ctx, u); err != nil {
			return err
		}
	}

	if len(oldAppIds) > 0 {
		if err := appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
			return err
		}
	}

	if len(appTags) > 0 {
		if err := appTagRepo.BatchCreate(ctx, appTags); err != nil {
			return err
		}
	}
	tx.Commit()
	if err := NewISettingService().Update("AppStoreLastModified", strconv.Itoa(list.LastModified)); err != nil {
		return err
	}
	return nil
}
