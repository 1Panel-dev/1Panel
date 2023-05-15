package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"gopkg.in/yaml.v3"
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
	if err = yaml.Unmarshal([]byte(appDetail.DockerCompose), &composeMap); err != nil {
		return
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
		if index > 0 {
			continue
		}
		req.Params["CONTAINER_NAME"] = containerName
		appInstall.ServiceName = serviceName
		appInstall.ContainerName = containerName
		index++
	}
	for k, v := range changeKeys {
		servicesMap[v] = servicesMap[k]
		delete(servicesMap, k)
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
		if err = downloadApp(app, appDetail, appInstall, req); err != nil {
			_ = appInstallRepo.Save(ctx, appInstall)
			return
		}
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
	res.CanUpdate = true
	return res, nil
}

func (a AppService) SyncAppListFromLocal() {
	//fileOp := files.NewFileOp()
	//appDir := constant.LocalAppResourceDir
	//listFile := path.Join(appDir, "list.json")
	//if !fileOp.Stat(listFile) {
	//	return
	//}
	//global.LOG.Infof("start sync local apps...")
	//content, err := fileOp.GetContent(listFile)
	//if err != nil {
	//	global.LOG.Errorf("get list.json content failed %s", err.Error())
	//	return
	//}
	//list := &dto.AppList{}
	//if err := json.Unmarshal(content, list); err != nil {
	//	global.LOG.Errorf("unmarshal list.json failed %s", err.Error())
	//	return
	//}
	//oldApps, _ := appRepo.GetBy(appRepo.WithResource(constant.AppResourceLocal))
	//appsMap := getApps(oldApps, list.Apps, true)
	//for _, l := range list.Apps {
	//	localKey := "local" + l.Config.Key
	//	app := appsMap[localKey]
	//	icon, err := os.ReadFile(path.Join(appDir, l.Config.Key, "metadata", "logo.png"))
	//	if err != nil {
	//		global.LOG.Errorf("get [%s] icon error: %s", l.Name, err.Error())
	//		continue
	//	}
	//	iconStr := base64.StdEncoding.EncodeToString(icon)
	//	app.Icon = iconStr
	//	app.TagsKey = append(l.Tags, "Local")
	//	app.Recommend = 9999
	//	versions := l.Versions
	//	detailsMap := getAppDetails(app.Details, versions)
	//
	//	for _, v := range versions {
	//		detail := detailsMap[v]
	//		detailPath := path.Join(appDir, l.Key, "versions", v)
	//		if _, err := os.Stat(detailPath); err != nil {
	//			global.LOG.Errorf("get [%s] folder error: %s", detailPath, err.Error())
	//			continue
	//		}
	//		readmeStr, err := os.ReadFile(path.Join(detailPath, "README.md"))
	//		if err != nil {
	//			global.LOG.Errorf("get [%s] README error: %s", detailPath, err.Error())
	//		}
	//		detail.Readme = string(readmeStr)
	//		dockerComposeStr, err := os.ReadFile(path.Join(detailPath, "docker-compose.yml"))
	//		if err != nil {
	//			global.LOG.Errorf("get [%s] docker-compose.yml error: %s", detailPath, err.Error())
	//			continue
	//		}
	//		detail.DockerCompose = string(dockerComposeStr)
	//		paramStr, err := os.ReadFile(path.Join(detailPath, "config.json"))
	//		if err != nil {
	//			global.LOG.Errorf("get [%s] form.json error: %s", detailPath, err.Error())
	//		}
	//		detail.Params = string(paramStr)
	//		detailsMap[v] = detail
	//	}
	//	var newDetails []model.AppDetail
	//	for _, v := range detailsMap {
	//		newDetails = append(newDetails, v)
	//	}
	//	app.Details = newDetails
	//	appsMap[localKey] = app
	//}
	//var (
	//	addAppArray []model.App
	//	updateArray []model.App
	//	appIds      []uint
	//)
	//for _, v := range appsMap {
	//	if v.ID == 0 {
	//		addAppArray = append(addAppArray, v)
	//	} else {
	//		updateArray = append(updateArray, v)
	//		appIds = append(appIds, v.ID)
	//	}
	//}
	//tx, ctx := getTxAndContext()
	//if len(addAppArray) > 0 {
	//	if err := appRepo.BatchCreate(ctx, addAppArray); err != nil {
	//		tx.Rollback()
	//		return
	//	}
	//}
	//for _, update := range updateArray {
	//	if err := appRepo.Save(ctx, &update); err != nil {
	//		tx.Rollback()
	//		return
	//	}
	//}
	//if err := appTagRepo.DeleteByAppIds(ctx, appIds); err != nil {
	//	tx.Rollback()
	//	return
	//}
	//apps := append(addAppArray, updateArray...)
	//var (
	//	addDetails    []model.AppDetail
	//	updateDetails []model.AppDetail
	//	appTags       []*model.AppTag
	//)
	//tags, _ := tagRepo.All()
	//tagMap := make(map[string]uint, len(tags))
	//for _, app := range tags {
	//	tagMap[app.Key] = app.ID
	//}
	//for _, a := range apps {
	//	for _, t := range a.TagsKey {
	//		tagId, ok := tagMap[t]
	//		if ok {
	//			appTags = append(appTags, &model.AppTag{
	//				AppId: a.ID,
	//				TagId: tagId,
	//			})
	//		}
	//	}
	//	for _, d := range a.Details {
	//		d.AppId = a.ID
	//		if d.ID == 0 {
	//			addDetails = append(addDetails, d)
	//		} else {
	//			updateDetails = append(updateDetails, d)
	//		}
	//	}
	//}
	//if len(addDetails) > 0 {
	//	if err := appDetailRepo.BatchCreate(ctx, addDetails); err != nil {
	//		tx.Rollback()
	//		return
	//	}
	//}
	//for _, u := range updateDetails {
	//	if err := appDetailRepo.Update(ctx, u); err != nil {
	//		tx.Rollback()
	//		return
	//	}
	//}
	//if len(appTags) > 0 {
	//	if err := appTagRepo.BatchCreate(ctx, appTags); err != nil {
	//		tx.Rollback()
	//		return
	//	}
	//}
	//tx.Commit()
	//global.LOG.Infof("sync local apps success")
}
func (a AppService) SyncAppListFromRemote() error {
	updateRes, err := a.GetAppUpdate()
	if err != nil {
		return err
	}
	if !updateRes.CanUpdate {
		//global.LOG.Infof("The latest version is [%s] The app store is already up to date", updateRes.Version)
		return nil
	}
	var (
		tags    []*model.Tag
		appTags []*model.AppTag
		list    = updateRes.List
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
	baseRemoteUrl := fmt.Sprintf("%s/%s/1panel", global.CONF.System.AppRepo, global.CONF.System.Mode)
	appsMap := getApps(oldApps, list.Apps, false)
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
	if len(addAppArray) > 0 {
		if err := appRepo.BatchCreate(ctx, addAppArray); err != nil {
			tx.Rollback()
			return err
		}
	}
	if len(deleteAppArray) > 0 {
		if err := appRepo.BatchDelete(ctx, deleteAppArray); err != nil {
			tx.Rollback()
			return err
		}
		if err := appDetailRepo.DeleteByAppIds(ctx, deleteIds); err != nil {
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
	for _, update := range updateAppArray {
		if err := appRepo.Save(ctx, &update); err != nil {
			tx.Rollback()
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
	if err := NewISettingService().Update("AppStoreLastModified", strconv.Itoa(list.LastModified)); err != nil {
		return err
	}
	return nil
}
