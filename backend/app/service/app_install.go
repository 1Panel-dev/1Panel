package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"math"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

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
	Page(req request.AppInstalledSearch) (int64, []response.AppInstalledDTO, error)
	CheckExist(key string) (*response.AppInstalledCheck, error)
	LoadPort(key string) (int64, error)
	LoadConnInfo(key string) (response.DatabaseConn, error)
	SearchForWebsite(req request.AppInstalledSearch) ([]response.AppInstalledDTO, error)
	Operate(req request.AppInstalledOperate) error
	Update(req request.AppInstalledUpdate) error
	SyncAll(systemInit bool) error
	GetServices(key string) ([]response.AppService, error)
	GetUpdateVersions(installId uint) ([]dto.AppVersion, error)
	GetParams(id uint) ([]response.AppParam, error)
	ChangeAppPort(req request.PortUpdate) error
	GetDefaultConfigByKey(key string) (string, error)
	DeleteCheck(installId uint) ([]dto.AppResource, error)
}

func NewIAppInstalledService() IAppInstallService {
	return &AppInstallService{}
}

func (a *AppInstallService) Page(req request.AppInstalledSearch) (int64, []response.AppInstalledDTO, error) {
	var opts []repo.DBOption

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

	total, installs, err := appInstallRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	installDTOs, err := handleInstalled(installs, req.Update)
	if err != nil {
		return 0, nil, err
	}

	return total, installDTOs, nil
}

func (a *AppInstallService) CheckExist(key string) (*response.AppInstalledCheck, error) {
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
	appInstall, _ = appInstallRepo.GetFirst(commonRepo.WithByID(appInstall.ID))

	res.ContainerName = appInstall.ContainerName
	res.Name = appInstall.Name
	res.Version = appInstall.Version
	res.CreatedAt = appInstall.CreatedAt
	res.Status = appInstall.Status
	res.AppInstallID = appInstall.ID
	res.IsExist = true
	res.InstallPath = path.Join(constant.AppInstallDir, app.Key, appInstall.Name)

	return res, nil
}

func (a *AppInstallService) LoadPort(key string) (int64, error) {
	app, err := appInstallRepo.LoadBaseInfo(key, "")
	if err != nil {
		return int64(0), nil
	}
	return app.Port, nil
}

func (a *AppInstallService) LoadConnInfo(key string) (response.DatabaseConn, error) {
	var data response.DatabaseConn
	app, err := appInstallRepo.LoadBaseInfo(key, "")
	if err != nil {
		return data, nil
	}
	data.Password = app.Password
	data.ServiceName = app.ServiceName
	data.Port = app.Port
	return data, nil
}

func (a *AppInstallService) SearchForWebsite(req request.AppInstalledSearch) ([]response.AppInstalledDTO, error) {
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

	return handleInstalled(installs, false)
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
		return syncById(install.ID)
	case constant.Stop:
		out, err := compose.Stop(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		return syncById(install.ID)
	case constant.Restart:
		out, err := compose.Restart(dockerComposePath)
		if err != nil {
			return handleErr(install, err, out)
		}
		return syncById(install.ID)
	case constant.Delete:
		if err := deleteAppInstall(install, req.DeleteBackup, req.ForceDelete, req.DeleteDB); err != nil && !req.ForceDelete {
			return err
		}
		return nil
	case constant.Sync:
		return syncById(install.ID)
	case constant.Upgrade:
		return upgradeInstall(install.ID, req.DetailId)
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
	var (
		oldPorts []int
		newPorts []int
	)
	port, ok := req.Params["PANEL_APP_PORT_HTTP"]
	if ok {
		portN := int(math.Ceil(port.(float64)))
		if portN != installed.HttpPort {
			oldPorts = append(oldPorts, installed.HttpPort)
			changePort = true
			httpPort, err := checkPort("PANEL_APP_PORT_HTTP", req.Params)
			if err != nil {
				return err
			}
			installed.HttpPort = httpPort
			newPorts = append(newPorts, httpPort)
		}
	}
	ports, ok := req.Params["PANEL_APP_PORT_HTTPS"]
	if ok {
		portN := int(math.Ceil(ports.(float64)))
		if portN != installed.HttpsPort {
			oldPorts = append(oldPorts, installed.HttpsPort)
			httpsPort, err := checkPort("PANEL_APP_PORT_HTTPS", req.Params)
			if err != nil {
				return err
			}
			installed.HttpsPort = httpsPort
			newPorts = append(newPorts, httpsPort)
		}
	}

	envPath := path.Join(installed.GetPath(), ".env")
	oldEnvMaps, err := godotenv.Read(envPath)
	if err != nil {
		return err
	}
	handleMap(req.Params, oldEnvMaps)
	paramByte, err := json.Marshal(oldEnvMaps)
	if err != nil {
		return err
	}
	installed.Env = string(paramByte)
	if err := env.Write(oldEnvMaps, envPath); err != nil {
		return err
	}
	_ = appInstallRepo.Save(context.Background(), &installed)

	if err := rebuildApp(installed); err != nil {
		return err
	}
	website, _ := websiteRepo.GetFirst(websiteRepo.WithAppInstallId(installed.ID))
	if changePort && website.ID != 0 && website.Status == constant.Running {
		nginxInstall, err := getNginxFull(&website)
		if err != nil {
			return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
		}
		config := nginxInstall.SiteConfig.Config
		servers := config.FindServers()
		if len(servers) == 0 {
			return buserr.WithErr(constant.ErrUpdateBuWebsite, errors.New("nginx config is not valid"))
		}
		server := servers[0]
		proxy := fmt.Sprintf("http://127.0.0.1:%d", installed.HttpPort)
		server.UpdateRootProxy([]string{proxy})

		if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
			return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
		}
		if err := nginxCheckAndReload(nginxInstall.SiteConfig.OldContent, config.FilePath, nginxInstall.Install.ContainerName); err != nil {
			return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
		}
	}
	if changePort {
		go func() {
			_ = OperateFirewallPort(oldPorts, newPorts)
		}()
	}
	return nil
}

func (a *AppInstallService) SyncAll(systemInit bool) error {
	allList, err := appInstallRepo.ListBy()
	if err != nil {
		return err
	}
	for _, i := range allList {
		if i.Status == constant.Installing {
			if systemInit {
				i.Status = constant.Error
				i.Message = "System restart causes application exception"
				_ = appInstallRepo.Save(context.Background(), &i)
			}
			continue
		}
		if err := syncById(i.ID); err != nil {
			global.LOG.Errorf("sync install app[%s] error,mgs: %s", i.Name, err.Error())
		}
	}
	return nil
}

func (a *AppInstallService) GetServices(key string) ([]response.AppService, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(key))
	if err != nil {
		return nil, err
	}
	installs, err := appInstallRepo.ListBy(appInstallRepo.WithAppId(app.ID), appInstallRepo.WithStatus(constant.Running))
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

func (a *AppInstallService) GetUpdateVersions(installId uint) ([]dto.AppVersion, error) {
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

func (a *AppInstallService) ChangeAppPort(req request.PortUpdate) error {
	if common.ScanPort(int(req.Port)) {
		return buserr.WithDetail(constant.ErrPortInUsed, req.Port, nil)
	}

	appInstall, err := appInstallRepo.LoadBaseInfo(req.Key, req.Name)
	if err != nil {
		return nil
	}

	if err := updateInstallInfoInDB(req.Key, "", "port", true, strconv.FormatInt(req.Port, 10)); err != nil {
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

	if err := OperateFirewallPort([]int{int(appInstall.Port)}, []int{int(req.Port)}); err != nil {
		global.LOG.Errorf("allow firewall failed, err: %v", err)
	}

	return nil
}

func (a *AppInstallService) DeleteCheck(installId uint) ([]dto.AppResource, error) {
	var res []dto.AppResource
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
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
	return res, nil
}

func (a *AppInstallService) GetDefaultConfigByKey(key string) (string, error) {
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
	if key == constant.AppOpenresty {
		filePath = path.Join(filePath, "nginx.conf")
	}
	contentByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(contentByte), nil
}

func (a *AppInstallService) GetParams(id uint) ([]response.AppParam, error) {
	var (
		res     []response.AppParam
		appForm dto.AppForm
		envs    = make(map[string]interface{})
	)
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	detail, err := appDetailRepo.GetFirst(commonRepo.WithByID(install.AppDetailId))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(detail.Params), &appForm); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(install.Env), &envs); err != nil {
		return nil, err
	}
	for _, form := range appForm.FormFields {
		if v, ok := envs[form.EnvKey]; ok {
			appParam := response.AppParam{
				Edit: false,
				Key:  form.EnvKey,
				Rule: form.Rule,
				Type: form.Type,
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
				for _, fv := range form.Values {
					if fv.Value == v {
						appParam.ShowValue = fv.Label
						break
					}
				}
				appParam.Values = form.Values
			}
			res = append(res, appParam)
		}
	}
	return res, nil
}

func syncById(installId uint) error {
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(installId))
	if err != nil {
		return err
	}
	if appInstall.Status == constant.Installing {
		return nil
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
		return appInstallRepo.Save(context.Background(), &appInstall)
	}
	if errCount == 0 && existedCount == 0 {
		appInstall.Status = constant.Running
		return appInstallRepo.Save(context.Background(), &appInstall)
	}
	if existedCount == normalCount {
		appInstall.Status = constant.Stopped
		return appInstallRepo.Save(context.Background(), &appInstall)
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
	return appInstallRepo.Save(context.Background(), &appInstall)
}

func updateInstallInfoInDB(appKey, appName, param string, isRestart bool, value interface{}) error {
	if param != "password" && param != "port" && param != "user-password" {
		return nil
	}
	appInstall, err := appInstallRepo.LoadBaseInfo(appKey, appName)
	if err != nil {
		return nil
	}
	envPath := fmt.Sprintf("%s/%s/%s/.env", constant.AppInstallDir, appKey, appInstall.Name)
	lineBytes, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	envKey := ""
	switch param {
	case "password":
		envKey = "PANEL_DB_ROOT_PASSWORD="
	case "port":
		envKey = "PANEL_APP_PORT_HTTP="
	case "user-password":
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
