package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/app/repo"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/common"
	"github.com/1Panel-dev/1Panel/utils/compose"
	"github.com/1Panel-dev/1Panel/utils/files"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"math"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
)

type AppService struct {
}

func (a AppService) Page(req dto.AppRequest) (interface{}, error) {

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
	details, err := appDetailRepo.GetByAppId(context.WithValue(context.Background(), "db", global.DB), id)
	if err != nil {
		return appDTO, err
	}
	var versionsRaw []string
	for _, detail := range details {
		versionsRaw = append(versionsRaw, detail.Version)
	}

	sort.Slice(versionsRaw, func(i, j int) bool {
		return common.CompareVersion(versionsRaw[i], versionsRaw[j])
	})
	appDTO.Versions = versionsRaw

	return appDTO, nil
}

func (a AppService) PageInstalled(req dto.AppInstalledRequest) (int64, []dto.AppInstalled, error) {
	total, installed, err := appInstallRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	installDTOs := []dto.AppInstalled{}
	for _, in := range installed {
		installDto := dto.AppInstalled{
			AppInstall: in,
			AppName:    in.App.Name,
			Icon:       in.App.Icon,
		}
		installDTOs = append(installDTOs, installDto)
	}

	return total, installDTOs, nil
}

func (a AppService) GetAppDetail(appId uint, version string) (dto.AppDetailDTO, error) {

	var (
		appDetailDTO dto.AppDetailDTO
	)

	var opts []repo.DBOption
	opts = append(opts, appDetailRepo.WithAppId(appId), appDetailRepo.WithVersion(version))
	detail, err := appDetailRepo.GetAppDetail(opts...)
	if err != nil {
		return appDetailDTO, err
	}
	paramMap := make(map[string]interface{})
	json.Unmarshal([]byte(detail.Params), &paramMap)
	appDetailDTO.AppDetail = detail
	appDetailDTO.Params = paramMap
	return appDetailDTO, nil
}

func (a AppService) Operate(req dto.AppInstallOperate) error {
	appInstall, err := appInstallRepo.GetBy(commonRepo.WithByID(req.InstallId))
	if err != nil {
		return err
	}
	if len(appInstall) == 0 {
		return errors.New("req not found")
	}

	install := appInstall[0]
	dockerComposePath := path.Join(global.CONF.System.AppDir, install.App.Key, install.ContainerName, "docker-compose.yml")
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
		op := files.NewFileOp()
		appDir := path.Join(global.CONF.System.AppDir, install.App.Key, install.ContainerName)
		dir, _ := os.Stat(appDir)
		if dir == nil {
			return appInstallRepo.Delete(commonRepo.WithByID(install.ID))
		}
		out, err := compose.Down(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		out, err = compose.Rmf(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		_ = op.DeleteDir(appDir)
		_ = appInstallRepo.Delete(commonRepo.WithByID(install.ID))
		return nil
	default:
		return errors.New("operate not support")
	}

	return appInstallRepo.Save(install)
}

func handleErr(install model.AppInstall, err error, out string) error {
	reErr := err
	install.Message = err.Error()
	if out != "" {
		install.Message = out
		reErr = errors.New(out)
	}
	_ = appInstallRepo.Save(install)
	return reErr
}

func (a AppService) Install(name string, appDetailId uint, params map[string]interface{}) error {

	port, ok := params["PORT"]
	if ok {
		port := int(math.Floor(port.(float64)))
		if common.ScanPort(string(port)) {
			return errors.New("port is  in used")
		}
	}

	appDetail, err := appDetailRepo.GetAppDetail(commonRepo.WithByID(appDetailId))
	if err != nil {
		return err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	paramByte, err := json.Marshal(params)
	if err != nil {
		return err
	}
	containerName := constant.ContainerPrefix + app.Key + "-" + common.RandStr(6)
	appInstall := model.AppInstall{
		Name:          name,
		AppId:         appDetail.AppId,
		AppDetailId:   appDetail.ID,
		Version:       appDetail.Version,
		Status:        constant.Installing,
		Params:        string(paramByte),
		ContainerName: containerName,
		Message:       "",
	}
	resourceDir := path.Join(global.CONF.System.ResourceDir, "apps", app.Key, appDetail.Version)
	installDir := path.Join(global.CONF.System.AppDir, app.Key)
	installAppDir := path.Join(installDir, appDetail.Version)
	op := files.NewFileOp()
	if err := op.Copy(resourceDir, installAppDir); err != nil {
		return err
	}
	containerNameDir := path.Join(installDir, containerName)
	if err := op.Rename(installAppDir, containerNameDir); err != nil {
		return err
	}
	composeFilePath := path.Join(containerNameDir, "docker-compose.yml")
	envPath := path.Join(containerNameDir, ".env")

	envParams := make(map[string]string, len(params))
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
	envParams["CONTAINER_NAME"] = containerName
	if err := godotenv.Write(envParams, envPath); err != nil {
		return err
	}
	if err := appInstallRepo.Create(&appInstall); err != nil {
		return err
	}
	go upApp(composeFilePath, appInstall)
	return nil
}

func upApp(composeFilePath string, appInstall model.AppInstall) {
	out, err := compose.Up(composeFilePath)
	if err != nil {
		if out != "" {
			appInstall.Message = out
		} else {
			appInstall.Message = err.Error()
		}
		appInstall.Status = constant.Error
		_ = appInstallRepo.Save(appInstall)
	} else {
		appInstall.Status = constant.Running
		_ = appInstallRepo.Save(appInstall)
	}
}

func (a AppService) SyncInstalled() error {

	return nil
}

func (a AppService) Sync() error {
	//TODO 从 oss 拉取最新列表
	var appConfig model.AppConfig
	appConfig.OssPath = global.CONF.System.AppOss

	appDir := path.Join(global.CONF.System.ResourceDir, "apps")
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
	appConfig.Version = list.Version
	appConfig.CanUpdate = false

	var (
		tags       []*model.Tag
		addApps    []*model.App
		updateApps []*model.App
		appTags    []*model.AppTag
	)

	for _, t := range list.Tags {
		tags = append(tags, &model.Tag{
			Key:  t.Key,
			Name: t.Name,
		})
	}

	db := global.DB
	dbCtx := context.WithValue(context.Background(), "db", db)
	for _, l := range list.Items {
		icon, err := os.ReadFile(path.Join(iconDir, l.Icon))
		if err != nil {
			global.LOG.Errorf("get [%s] icon error: %s", l.Name, err.Error())
			continue
		}
		iconStr := base64.StdEncoding.EncodeToString(icon)
		app := &model.App{
			Name:      l.Name,
			Key:       l.Key,
			ShortDesc: l.ShortDesc,
			Author:    l.Author,
			Source:    l.Source,
			Icon:      iconStr,
			Type:      l.Type,
		}
		app.TagsKey = l.Tags
		old, _ := appRepo.GetByKey(dbCtx, l.Key)
		if reflect.DeepEqual(old, &model.App{}) {
			addApps = append(addApps, app)
		} else {
			app.ID = old.ID
			updateApps = append(updateApps, app)
		}

		versions := l.Versions
		for _, v := range versions {
			detail := &model.AppDetail{
				Version: v,
			}
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
			paramStr, err := os.ReadFile(path.Join(detailPath, "params.json"))
			if err != nil {
				global.LOG.Errorf("get [%s] form.json error: %s", detailPath, err.Error())
			}
			detail.Params = string(paramStr)
			app.Details = append(app.Details, detail)
		}
	}
	tx := global.DB.Begin()
	ctx := context.WithValue(context.Background(), "db", tx)
	if len(addApps) > 0 {
		if err := appRepo.BatchCreate(ctx, addApps); err != nil {
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
	}

	tagMap := make(map[string]uint, len(tags))

	for _, t := range tags {
		tagMap[t.Key] = t.ID
	}

	for _, a := range updateApps {
		if err := appRepo.Save(ctx, a); err != nil {
			tx.Rollback()
			return err
		}
	}
	apps := append(addApps, updateApps...)

	var (
		appDetails []*model.AppDetail
		appIds     []uint
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
			appDetails = append(appDetails, d)
		}
		appIds = append(appIds, a.ID)
	}

	if err := appDetailRepo.DeleteByAppIds(ctx, appIds); err != nil {
		tx.Rollback()
		return err
	}

	if len(appDetails) > 0 {
		if err := appDetailRepo.BatchCreate(ctx, appDetails); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := appTagRepo.DeleteByAppIds(ctx, appIds); err != nil {
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
	return nil
}
