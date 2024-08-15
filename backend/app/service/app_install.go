package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/files"
	httpUtil "github.com/1Panel-dev/1Panel/backend/utils/http"
	"github.com/docker/docker/api/types"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/backend/utils/env"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/joho/godotenv"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"

	"github.com/1Panel-dev/1Panel/backend/app/repo"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/pkg/errors"
)

type AppInstallService struct {
}

type IAppInstallService interface {
	Page(req request.AppInstalledSearch) (int64, []response.AppInstallDTO, error)
	CheckExist(req request.AppInstalledInfo) (*response.AppInstalledCheck, error)
	LoadPort(req dto.OperationWithNameAndType) (int64, error)
	LoadConnInfo(req dto.OperationWithNameAndType) (response.DatabaseConn, error)
	SearchForWebsite(req request.AppInstalledSearch) ([]response.AppInstallDTO, error)
	Operate(req request.AppInstalledOperate) error
	Update(req request.AppInstalledUpdate) error
	IgnoreUpgrade(req request.AppInstalledIgnoreUpgrade) error
	SyncAll(systemInit bool) error
	GetServices(key string) ([]response.AppService, error)
	GetUpdateVersions(req request.AppUpdateVersion) ([]dto.AppVersion, error)
	GetParams(id uint) (*response.AppConfig, error)
	ChangeAppPort(req request.PortUpdate) error
	GetDefaultConfigByKey(key, name string) (string, error)
	DeleteCheck(installId uint) ([]dto.AppResource, error)

	GetInstallList() ([]dto.AppInstallInfo, error)
}

func NewIAppInstalledService() IAppInstallService {
	return &AppInstallService{}
}

func (a *AppInstallService) GetInstallList() ([]dto.AppInstallInfo, error) {
	var datas []dto.AppInstallInfo
	appInstalls, err := appInstallRepo.ListBy()
	if err != nil {
		return nil, err
	}
	for _, install := range appInstalls {
		datas = append(datas, dto.AppInstallInfo{ID: install.ID, Key: install.App.Key, Name: install.Name})
	}
	return datas, nil
}

func (a *AppInstallService) Page(req request.AppInstalledSearch) (int64, []response.AppInstallDTO, error) {
	var (
		opts     []repo.DBOption
		total    int64
		installs []model.AppInstall
		err      error
	)

	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}
	if len(req.Tags) != 0 {
		tags, err := tagRepo.GetByKeys(req.Tags)
		if err != nil {
			return 0, nil, err
		}
		var tagIds []uint
		for _, t := range tags {
			tagIds = append(tagIds, t.ID)
		}
		appTags, err := appTagRepo.GetByTagIds(tagIds)
		if err != nil {
			return 0, nil, err
		}
		var appIds []uint
		for _, t := range appTags {
			appIds = append(appIds, t.AppId)
		}

		opts = append(opts, appInstallRepo.WithAppIdsIn(appIds))
	}

	if req.Update {
		installs, err = appInstallRepo.ListBy(opts...)
		if err != nil {
			return 0, nil, err
		}
	} else {
		total, installs, err = appInstallRepo.Page(req.Page, req.PageSize, opts...)
		if err != nil {
			return 0, nil, err
		}
	}

	installDTOs, err := handleInstalled(installs, req.Update, req.Sync)
	if err != nil {
		return 0, nil, err
	}
	if req.Update {
		total = int64(len(installDTOs))
	}

	return total, installDTOs, nil
}

func (a *AppInstallService) CheckExist(req request.AppInstalledInfo) (*response.AppInstalledCheck, error) {
	res := &response.AppInstalledCheck{
		IsExist: false,
	}

	app, err := appRepo.GetFirst(appRepo.WithKey(req.Key))
	if err != nil {
		return res, nil
	}
	res.App = app.Name

	var appInstall model.AppInstall
	if len(req.Name) == 0 {
		appInstall, _ = appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID))
	} else {
		appInstall, _ = appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID), commonRepo.WithByName(req.Name))
	}

	if reflect.DeepEqual(appInstall, model.AppInstall{}) {
		return res, nil
	}
	if err = syncAppInstallStatus(&appInstall, false); err != nil {
		return nil, err
	}

	res.ContainerName = appInstall.ContainerName
	res.Name = appInstall.Name
	res.Version = appInstall.Version
	res.CreatedAt = appInstall.CreatedAt
	res.Status = appInstall.Status
	res.AppInstallID = appInstall.ID
	res.IsExist = true
	res.InstallPath = path.Join(constant.AppInstallDir, appInstall.App.Key, appInstall.Name)
	res.HttpPort = appInstall.HttpPort
	res.HttpsPort = appInstall.HttpsPort

	return res, nil
}

func (a *AppInstallService) LoadPort(req dto.OperationWithNameAndType) (int64, error) {
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Name)
	if err != nil {
		return int64(0), nil
	}
	return app.Port, nil
}

func (a *AppInstallService) LoadConnInfo(req dto.OperationWithNameAndType) (response.DatabaseConn, error) {
	var data response.DatabaseConn
	app, err := appInstallRepo.LoadBaseInfo(req.Type, req.Name)
	if err != nil {
		return data, nil
	}
	data.Status = app.Status
	data.Username = app.UserName
	data.Password = app.Password
	data.ServiceName = app.ServiceName
	data.Port = app.Port
	data.ContainerName = app.ContainerName
	return data, nil
}

func (a *AppInstallService) SearchForWebsite(req request.AppInstalledSearch) ([]response.AppInstallDTO, error) {
	var (
		installs []model.AppInstall
		err      error
		opts     []repo.DBOption
	)
	if req.Type != "" {
		apps, err := appRepo.GetBy(appRepo.WithType(req.Type))
		if err != nil {
			return nil, err
		}
		var ids []uint
		for _, app := range apps {
			ids = append(ids, app.ID)
		}
		if req.Unused {
			opts = append(opts, appInstallRepo.WithIdNotInWebsite())
		}
		opts = append(opts, appInstallRepo.WithAppIdsIn(ids))
		installs, err = appInstallRepo.ListBy(opts...)
		if err != nil {
			return nil, err
		}
	} else {
		installs, err = appInstallRepo.ListBy()
		if err != nil {
			return nil, err
		}
	}

	return handleInstalled(installs, false, true)
}

func (a *AppInstallService) Operate(req request.AppInstalledOperate) error {
	install, err := appInstallRepo.GetFirstByCtx(context.Background(), commonRepo.WithByID(req.InstallId))
	if err != nil {
		return err
	}
	if !req.ForceDelete && !files.NewFileOp().Stat(install.GetPath()) {
		return buserr.New(constant.ErrInstallDirNotFound)
	}
	dockerComposePath := install.GetComposePath()
	switch req.Operate {
	case constant.Rebuild:
		return rebuildApp(install)
	case constant.Start:
		out, err := compose.Start(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		return syncAppInstallStatus(&install, false)
	case constant.Stop:
		out, err := compose.Stop(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		return syncAppInstallStatus(&install, false)
	case constant.Restart:
		out, err := compose.Restart(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		return syncAppInstallStatus(&install, false)
	case constant.Delete:
		if err := deleteAppInstall(install, req.DeleteBackup, req.ForceDelete, req.DeleteDB); err != nil && !req.ForceDelete {
			return err
		}
		return nil
	case constant.Sync:
		return syncAppInstallStatus(&install, true)
	case constant.Upgrade:
		upgradeReq := request.AppInstallUpgrade{
			InstallID:     install.ID,
			DetailID:      req.DetailId,
			Backup:        req.Backup,
			PullImage:     req.PullImage,
			DockerCompose: req.DockerCompose,
		}
		return upgradeInstall(upgradeReq)
	case constant.Reload:
		return opNginx(install.ContainerName, constant.NginxReload)
	default:
		return errors.New("operate not support")
	}
}

func (a *AppInstallService) Update(req request.AppInstalledUpdate) error {
	installed, err := appInstallRepo.GetFirst(commonRepo.WithByID(req.InstallId))
	if err != nil {
		return err
	}
	changePort := false
	port, ok := req.Params["PANEL_APP_PORT_HTTP"]
	if ok {
		portN := int(math.Ceil(port.(float64)))
		if portN != installed.HttpPort {
			changePort = true
			httpPort, err := checkPort("PANEL_APP_PORT_HTTP", req.Params)
			if err != nil {
				return err
			}
			installed.HttpPort = httpPort
		}
	}
	ports, ok := req.Params["PANEL_APP_PORT_HTTPS"]
	if ok {
		portN := int(math.Ceil(ports.(float64)))
		if portN != installed.HttpsPort {
			httpsPort, err := checkPort("PANEL_APP_PORT_HTTPS", req.Params)
			if err != nil {
				return err
			}
			installed.HttpsPort = httpsPort
		}
	}

	backupDockerCompose := installed.DockerCompose
	if req.Advanced {
		composeMap := make(map[string]interface{})
		if req.EditCompose {
			if err = yaml.Unmarshal([]byte(req.DockerCompose), &composeMap); err != nil {
				return err
			}
		} else {
			if err = yaml.Unmarshal([]byte(installed.DockerCompose), &composeMap); err != nil {
				return err
			}
		}
		if err = addDockerComposeCommonParam(composeMap, installed.ServiceName, req.AppContainerConfig, req.Params); err != nil {
			return err
		}
		composeByte, err := yaml.Marshal(composeMap)
		if err != nil {
			return err
		}
		installed.DockerCompose = string(composeByte)
		if req.ContainerName == "" {
			req.Params[constant.ContainerName] = installed.ContainerName
		} else {
			req.Params[constant.ContainerName] = req.ContainerName
			if installed.ContainerName != req.ContainerName {
				exist, _ := appInstallRepo.GetFirst(appInstallRepo.WithContainerName(req.ContainerName), appInstallRepo.WithIDNotIs(installed.ID))
				if exist.ID > 0 {
					return buserr.New(constant.ErrContainerName)
				}
				containerExist, err := checkContainerNameIsExist(req.ContainerName, installed.GetPath())
				if err != nil {
					return err
				}
				if containerExist {
					return buserr.New(constant.ErrContainerName)
				}
				installed.ContainerName = req.ContainerName
			}
		}
	}

	envPath := path.Join(installed.GetPath(), ".env")
	oldEnvMaps, err := godotenv.Read(envPath)
	if err != nil {
		return err
	}
	backupEnvMaps := oldEnvMaps
	handleMap(req.Params, oldEnvMaps)
	paramByte, err := json.Marshal(oldEnvMaps)
	if err != nil {
		return err
	}
	installed.Env = string(paramByte)
	if err := env.Write(oldEnvMaps, envPath); err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	_ = fileOp.WriteFile(installed.GetComposePath(), strings.NewReader(installed.DockerCompose), 0755)
	if err := rebuildApp(installed); err != nil {
		_ = env.Write(backupEnvMaps, envPath)
		_ = fileOp.WriteFile(installed.GetComposePath(), strings.NewReader(backupDockerCompose), 0755)
		return err
	}
	installed.Status = constant.Running
	_ = appInstallRepo.Save(context.Background(), &installed)

	website, _ := websiteRepo.GetFirst(websiteRepo.WithAppInstallId(installed.ID))
	if changePort && website.ID != 0 && website.Status == constant.Running {
		go func() {
			nginxInstall, err := getNginxFull(&website)
			if err != nil {
				global.LOG.Errorf(buserr.WithErr(constant.ErrUpdateBuWebsite, err).Error())
				return
			}
			config := nginxInstall.SiteConfig.Config
			servers := config.FindServers()
			if len(servers) == 0 {
				global.LOG.Errorf(buserr.WithErr(constant.ErrUpdateBuWebsite, errors.New("nginx config is not valid")).Error())
				return
			}
			server := servers[0]
			proxy := fmt.Sprintf("http://127.0.0.1:%d", installed.HttpPort)
			server.UpdateRootProxy([]string{proxy})

			if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
				global.LOG.Errorf(buserr.WithErr(constant.ErrUpdateBuWebsite, err).Error())
				return
			}
			if err := nginxCheckAndReload(nginxInstall.SiteConfig.OldContent, config.FilePath, nginxInstall.Install.ContainerName); err != nil {
				global.LOG.Errorf(buserr.WithErr(constant.ErrUpdateBuWebsite, err).Error())
				return
			}
		}()
	}
	return nil
}

func (a *AppInstallService) IgnoreUpgrade(req request.AppInstalledIgnoreUpgrade) error {
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(req.DetailID))
	if err != nil {
		return err
	}
	appDetail.IgnoreUpgrade = req.Operate == "ignore"
	return appDetailRepo.Update(context.Background(), appDetail)
}

func (a *AppInstallService) SyncAll(systemInit bool) error {
	allList, err := appInstallRepo.ListBy()
	if err != nil {
		return err
	}
	for _, i := range allList {
		if i.Status == constant.Installing || i.Status == constant.Upgrading || i.Status == constant.Rebuilding {
			if systemInit {
				i.Status = constant.Error
				i.Message = "1Panel restart causes the task to terminate"
				_ = appInstallRepo.Save(context.Background(), &i)
			}
			continue
		}
		if !systemInit {
			if err = syncAppInstallStatus(&i, false); err != nil {
				global.LOG.Errorf("sync install app[%s] error,mgs: %s", i.Name, err.Error())
			}
		}
	}
	return nil
}

func (a *AppInstallService) GetServices(key string) ([]response.AppService, error) {
	var res []response.AppService
	if DatabaseKeys[key] > 0 {
		if key == constant.AppPostgres {
			key = constant.AppPostgresql
		}
		dbs, _ := databaseRepo.GetList(commonRepo.WithByType(key))
		if len(dbs) == 0 {
			return res, nil
		}
		for _, db := range dbs {
			service := response.AppService{
				Label: db.Name,
				Value: db.Name,
			}
			if db.AppInstallID > 0 {
				install, err := appInstallRepo.GetFirst(commonRepo.WithByID(db.AppInstallID))
				if err != nil {
					return nil, err
				}
				paramMap := make(map[string]string)
				if install.Param != "" {
					_ = json.Unmarshal([]byte(install.Param), &paramMap)
				}
				service.Config = paramMap
				service.From = constant.AppResourceLocal
			} else {
				service.From = constant.AppResourceRemote
			}
			res = append(res, service)
		}
	} else {
		app, err := appRepo.GetFirst(appRepo.WithKey(key))
		if err != nil {
			return nil, err
		}
		installs, err := appInstallRepo.ListBy(appInstallRepo.WithAppId(app.ID), appInstallRepo.WithStatus(constant.Running))
		if err != nil {
			return nil, err
		}
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
	}
	return res, nil
}

func (a *AppInstallService) GetUpdateVersions(req request.AppUpdateVersion) ([]dto.AppVersion, error) {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(req.AppInstallID))
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
		if detail.IgnoreUpgrade {
			continue
		}
		if common.IsCrossVersion(install.Version, detail.Version) && !app.CrossVersionUpdate {
			continue
		}
		if common.CompareVersion(detail.Version, install.Version) {
			var newCompose string
			if req.UpdateVersion != "" && req.UpdateVersion == detail.Version && detail.DockerCompose == "" && !app.IsLocalApp() {
				filename := filepath.Base(detail.DownloadUrl)
				dockerComposeUrl := fmt.Sprintf("%s%s", strings.TrimSuffix(detail.DownloadUrl, filename), "docker-compose.yml")
				statusCode, composeRes, err := httpUtil.HandleGet(dockerComposeUrl, http.MethodGet, constant.TimeOut20s)
				if err != nil {
					return versions, err
				}
				if statusCode > 200 {
					return versions, err
				}
				detail.DockerCompose = string(composeRes)
				_ = appDetailRepo.Update(context.Background(), detail)
			}
			newCompose, err = getUpgradeCompose(install, detail)
			if err != nil {
				return versions, err
			}
			if app.Key == constant.AppMysql {
				majorVersion := getMajorVersion(install.Version)
				if !strings.HasPrefix(detail.Version, majorVersion) {
					continue
				}
			}
			versions = append(versions, dto.AppVersion{
				Version:       detail.Version,
				DetailId:      detail.ID,
				DockerCompose: newCompose,
			})
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return common.CompareVersion(versions[i].Version, versions[j].Version)
	})
	return versions, nil
}

func (a *AppInstallService) ChangeAppPort(req request.PortUpdate) error {
	if common.ScanPort(int(req.Port)) {
		return buserr.WithDetail(constant.ErrPortInUsed, req.Port, nil)
	}

	appInstall, err := appInstallRepo.LoadBaseInfo(req.Key, req.Name)
	if err != nil {
		return nil
	}

	if err := updateInstallInfoInDB(req.Key, req.Name, "port", strconv.FormatInt(req.Port, 10)); err != nil {
		return nil
	}

	appRess, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(appInstall.ID))
	for _, appRes := range appRess {
		appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(appRes.AppInstallId))
		if err != nil {
			return err
		}
		if _, err := compose.Restart(fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, appInstall.App.Key, appInstall.Name)); err != nil {
			global.LOG.Errorf("docker-compose restart %s[%s] failed, err: %v", appInstall.App.Key, appInstall.Name, err)
		}
	}

	return nil
}

func (a *AppInstallService) DeleteCheck(installID uint) ([]dto.AppResource, error) {
	var res []dto.AppResource
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(installID))
	if err != nil {
		return nil, err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appInstall.AppId))
	if err != nil {
		return nil, err
	}
	websites, _ := websiteRepo.GetBy(websiteRepo.WithAppInstallId(appInstall.ID))
	for _, website := range websites {
		res = append(res, dto.AppResource{
			Type: "website",
			Name: website.PrimaryDomain,
		})
	}
	if app.Key == constant.AppOpenresty {
		websites, _ := websiteRepo.GetBy()
		for _, website := range websites {
			res = append(res, dto.AppResource{
				Type: "website",
				Name: website.PrimaryDomain,
			})
		}
	}
	if app.Type == constant.Runtime {
		resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithLinkId(appInstall.ID), commonRepo.WithByFrom(constant.AppResourceLocal))
		for _, resource := range resources {
			linkInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(resource.AppInstallId))
			res = append(res, dto.AppResource{
				Type: "app",
				Name: linkInstall.Name,
			})
		}
	}
	return res, nil
}

func (a *AppInstallService) GetDefaultConfigByKey(key, name string) (string, error) {
	baseInfo, err := appInstallRepo.LoadBaseInfo(key, name)
	if err != nil {
		return "", err
	}

	fileOp := files.NewFileOp()
	filePath := path.Join(constant.AppResourceDir, "remote", baseInfo.Key, baseInfo.Version, "conf")
	if !fileOp.Stat(filePath) {
		filePath = path.Join(constant.AppResourceDir, baseInfo.Key, "versions", baseInfo.Version, "conf")
	}
	if !fileOp.Stat(filePath) {
		return "", buserr.New(constant.ErrPathNotFound)
	}

	if key == constant.AppMysql || key == constant.AppMariaDB {
		filePath = path.Join(filePath, "my.cnf")
	}
	if key == constant.AppRedis {
		filePath = path.Join(filePath, "redis.conf")
	}
	if key == constant.AppOpenresty {
		filePath = path.Join(filePath, "nginx.conf")
	}
	contentByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(contentByte), nil
}

func (a *AppInstallService) GetParams(id uint) (*response.AppConfig, error) {
	var (
		params  []response.AppParam
		appForm dto.AppForm
		envs    = make(map[string]interface{})
		res     response.AppConfig
	)
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	detail, err := appDetailRepo.GetFirst(commonRepo.WithByID(install.AppDetailId))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(detail.Params), &appForm); err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(install.Env), &envs); err != nil {
		return nil, err
	}
	for _, form := range appForm.FormFields {
		if v, ok := envs[form.EnvKey]; ok {
			appParam := response.AppParam{
				Edit:     false,
				Key:      form.EnvKey,
				Rule:     form.Rule,
				Type:     form.Type,
				Multiple: form.Multiple,
			}
			if form.Edit {
				appParam.Edit = true
			}
			appParam.LabelZh = form.LabelZh
			appParam.LabelEn = form.LabelEn
			appParam.Value = v
			if form.Type == "service" {
				appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithServiceName(v.(string)))
				appParam.ShowValue = appInstall.Name
			} else if form.Type == "select" {
				if form.Multiple {
					if v == "" {
						appParam.Value = []string{}
					} else {
						if str, ok := v.(string); ok {
							appParam.Value = strings.Split(str, ",")
						}
					}
				} else {
					for _, fv := range form.Values {
						if fv.Value == v {
							appParam.ShowValue = fv.Label
							break
						}
					}
				}
				appParam.Values = form.Values
			} else if form.Type == "apps" {
				if m, ok := form.Child.(map[string]interface{}); ok {
					result := make(map[string]string)
					for key, value := range m {
						if strVal, ok := value.(string); ok {
							result[key] = strVal
						}
					}
					if envKey, ok := result["envKey"]; ok {
						serviceName := envs[envKey]
						if serviceName != nil {
							appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithServiceName(serviceName.(string)))
							appParam.ShowValue = appInstall.Name
						}
					}
				}
			}
			params = append(params, appParam)
		} else {
			params = append(params, response.AppParam{
				Edit:     form.Edit,
				Key:      form.EnvKey,
				Rule:     form.Rule,
				Type:     form.Type,
				LabelZh:  form.LabelZh,
				LabelEn:  form.LabelEn,
				Value:    form.Default,
				Values:   form.Values,
				Multiple: form.Multiple,
			})
		}
	}

	config := getAppCommonConfig(envs)
	config.DockerCompose = install.DockerCompose
	res.Params = params
	if config.ContainerName == "" {
		config.ContainerName = install.ContainerName
	}
	res.AppContainerConfig = config
	res.HostMode = isHostModel(install.DockerCompose)
	return &res, nil
}

func syncAppInstallStatus(appInstall *model.AppInstall, force bool) error {
	if appInstall.Status == constant.Installing || appInstall.Status == constant.Rebuilding || appInstall.Status == constant.Upgrading {
		return nil
	}
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	var (
		containers     []types.Container
		containersMap  map[string]types.Container
		containerNames = strings.Split(appInstall.ContainerName, ",")
	)
	containers, err = cli.ListContainersByName(containerNames)
	if err != nil {
		return err
	}
	containersMap = make(map[string]types.Container)
	for _, con := range containers {
		containersMap[con.Names[0]] = con
	}
	synAppInstall(containersMap, appInstall, force)
	return nil
}

func updateInstallInfoInDB(appKey, appName, param string, value interface{}) error {
	if param != "password" && param != "port" && param != "user-password" {
		return nil
	}
	appInstall, err := appInstallRepo.LoadBaseInfo(appKey, appName)
	if err != nil {
		return nil
	}
	envPath := fmt.Sprintf("%s/%s/.env", appInstall.AppPath, appInstall.Name)
	lineBytes, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	envKey := ""
	switch param {
	case "password":
		if appKey == "mysql" || appKey == "mariadb" || appKey == "postgresql" {
			envKey = "PANEL_DB_ROOT_PASSWORD="
		} else {
			envKey = "PANEL_REDIS_ROOT_PASSWORD="
		}
	case "port":
		envKey = "PANEL_APP_PORT_HTTP="
	default:
		envKey = "PANEL_DB_USER_PASSWORD="
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
		if appKey == "redis" {
			oldVal = fmt.Sprintf("\"PANEL_REDIS_ROOT_PASSWORD\":\"%v\"", appInstall.Password)
			newVal = fmt.Sprintf("\"PANEL_REDIS_ROOT_PASSWORD\":\"%v\"", value)
		}
		_ = appInstallRepo.BatchUpdateBy(map[string]interface{}{
			"param": strings.ReplaceAll(appInstall.Param, oldVal, newVal),
			"env":   strings.ReplaceAll(appInstall.Env, oldVal, newVal),
		}, commonRepo.WithByID(appInstall.ID))
	}
	if param == "user-password" {
		oldVal = fmt.Sprintf("\"PANEL_DB_USER_PASSWORD\":\"%v\"", appInstall.UserPassword)
		newVal = fmt.Sprintf("\"PANEL_DB_USER_PASSWORD\":\"%v\"", value)
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
