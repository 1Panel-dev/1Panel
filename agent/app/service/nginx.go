package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/subosito/gotenv"

	"github.com/1Panel-dev/1Panel/agent/utils/compose"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

type NginxService struct {
}

type INginxService interface {
	GetNginxConfig() (*response.NginxFile, error)
	GetConfigByScope(req request.NginxScopeReq) ([]response.NginxParam, error)
	UpdateConfigByScope(req request.NginxConfigUpdate) error
	GetStatus() (response.NginxStatus, error)
	UpdateConfigFile(req request.NginxConfigFileUpdate) error

	Build(req request.NginxBuildReq) error
	GetModules() (*response.NginxBuildConfig, error)
	UpdateModule(req request.NginxModuleUpdate) error

	OperateDefaultHTTPs(req request.NginxDefaultHTTPSUpdate) error
	GetDefaultHttpsStatus() (*response.NginxConfigRes, error)
}

func NewINginxService() INginxService {
	return &NginxService{}
}

func (n NginxService) GetNginxConfig() (*response.NginxFile, error) {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	configPath := path.Join(global.Dir.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, nginxModuleConfDir, "nginx.conf")
	byteContent, err := files.NewFileOp().GetContent(configPath)
	if err != nil {
		return nil, err
	}
	return &response.NginxFile{Content: string(byteContent)}, nil
}

func (n NginxService) GetConfigByScope(req request.NginxScopeReq) ([]response.NginxParam, error) {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil, nil
	}
	return getNginxParamsByKeys(constant.NginxScopeHttp, keys, nil)
}

func (n NginxService) UpdateConfigByScope(req request.NginxConfigUpdate) error {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil
	}
	return updateNginxConfig(constant.NginxScopeHttp, getNginxParams(req.Params, keys), nil)
}

func (n NginxService) GetStatus() (response.NginxStatus, error) {
	httpPort, _, err := getAppInstallPort(constant.AppOpenresty)
	if err != nil {
		return response.NginxStatus{}, err
	}
	url := "http://127.0.0.1/nginx_status"
	if httpPort != 80 {
		url = fmt.Sprintf("http://127.0.0.1:%v/nginx_status", httpPort)
	}
	res, err := http.Get(url)
	if err != nil || res.StatusCode > 300 {
		return response.NginxStatus{}, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return response.NginxStatus{}, err
	}
	var status response.NginxStatus
	resArray := strings.Split(string(content), " ")
	active, err := strconv.Atoi(resArray[2])
	if err == nil {
		status.Active = active
	}
	accepts, err := strconv.Atoi(resArray[7])
	if err == nil {
		status.Accepts = accepts
	}
	handled, err := strconv.Atoi(resArray[8])
	if err == nil {
		status.Handled = handled
	}
	requests, err := strconv.Atoi(resArray[9])
	if err == nil {
		status.Requests = requests
	}
	reading, err := strconv.Atoi(resArray[11])
	if err == nil {
		status.Reading = reading
	}
	writing, err := strconv.Atoi(resArray[13])
	if err == nil {
		status.Writing = writing
	}
	waiting, err := strconv.Atoi(resArray[15])
	if err == nil {
		status.Waiting = waiting
	}
	return status, nil
}

func (n NginxService) UpdateConfigFile(req request.NginxConfigFileUpdate) error {
	fileOp := files.NewFileOp()
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	filePath := path.Join(global.Dir.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, nginxModuleConfDir, "nginx.conf")
	if req.Backup {
		backupPath := path.Join(path.Dir(filePath), "bak")
		if !fileOp.Stat(backupPath) {
			if err := fileOp.CreateDir(backupPath, constant.DirPerm); err != nil {
				return err
			}
		}
		newFile := path.Join(backupPath, "nginx.bak"+"-"+time.Now().Format("2006-01-02-15-04-05"))
		if err := fileOp.Copy(filePath, backupPath); err != nil {
			return err
		}
		if err := fileOp.Rename(path.Join(backupPath, "nginx.conf"), newFile); err != nil {
			return err
		}
	}
	oldContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(filePath, strings.NewReader(req.Content), constant.DirPerm); err != nil {
		return err
	}
	if status, err := checkContainerStatus(nginxInstall.ContainerName); err == nil && status != "running" {
		if out, err := compose.DownAndUp(nginxInstall.GetComposePath()); err != nil {
			_ = fileOp.SaveFile(filePath, string(oldContent), constant.DirPerm)
			return fmt.Errorf("nginx restart failed: %v", out)
		} else {
			return nginxCheckAndReload(string(oldContent), filePath, nginxInstall.ContainerName)
		}
	}
	return nginxCheckAndReload(string(oldContent), filePath, nginxInstall.ContainerName)
}

func (n NginxService) Build(req request.NginxBuildReq) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	taskName := task.GetTaskName(nginxInstall.Name, task.TaskBuild, task.TaskScopeApp)
	if err = task.CheckTaskIsExecuting(taskName); err != nil {
		return err
	}
	if err = task.CheckScopeTaskIsExecuting(task.TaskScopeApp, nginxInstall.ID); err != nil {
		return err
	}
	buildPath := path.Join(nginxInstall.GetPath(), nginxModuleBuildDir)
	if !files.NewFileOp().Stat(buildPath) {
		return buserr.New("ErrBuildDirNotFound")
	}

	buildTask, err := task.NewTaskWithOps(nginxInstall.Name, task.TaskBuild, task.TaskScopeApp, req.TaskID, nginxInstall.ID)
	if err != nil {
		return err
	}
	buildTask.AddSubTaskWithOps("", func(t *task.Task) error {
		return executeNginxModuleBuild(nginxInstall, req.Modules, req.Force, req.Mirror, t, true)
	}, nil, 0, 120*time.Minute)

	go func() {
		_ = buildTask.Execute()
	}()
	return nil
}

func (n NginxService) GetModules() (*response.NginxBuildConfig, error) {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	modules, err := loadNginxModules(nginxInstall)
	if err != nil {
		return nil, err
	}
	target, targetWarning, targetErr := resolveNginxModuleTarget(nginxInstall)
	if targetWarning != "" {
		global.LOG.Warn(targetWarning)
	}
	var resList []response.NginxModule
	for _, module := range modules {
		buildStatus := nginxModuleStatusPending
		loadStatus := nginxModuleLoadDisabled
		compatibility := "unknown"
		var artifacts []dto.NginxModuleArtifact
		if module.BuildMode == nginxModuleBuildStatic {
			buildStatus = nginxModuleStatusReady
			compatibility = "static"
			if module.Enable {
				loadStatus = nginxModuleLoadEnabled
			}
		} else if targetErr == nil {
			if build := findCurrentNginxModuleBuild(module, target); build != nil {
				buildStatus = build.Status
				artifacts = build.Artifacts
				if build.Status == nginxModuleStatusReady {
					compatibility = "compatible"
					if module.Enable {
						loadStatus = nginxModuleLoadEnabled
					}
				}
			} else if latestBuild := findLatestNginxModuleBuild(module, target); latestBuild != nil {
				compatibility = "stale"
				artifacts = latestBuild.Artifacts
				if module.Enable {
					loadStatus = nginxModuleLoadEnabled
				}
			}
		}
		if module.BuildMode != nginxModuleBuildStatic && module.LastError != "" {
			buildStatus = nginxModuleStatusFailed
		}
		resList = append(resList, response.NginxModule{
			Name:          module.Name,
			Custom:        module.Custom,
			Script:        module.Script,
			Packages:      strings.Join(module.Packages, ","),
			Params:        module.Params,
			Enable:        module.Enable,
			BuildMode:     module.BuildMode,
			Provider:      module.Provider,
			LoadOrder:     module.LoadOrder,
			BuildStatus:   buildStatus,
			LoadStatus:    loadStatus,
			Compatibility: compatibility,
			Artifacts:     artifacts,
			LastError:     module.LastError,
		})
	}
	envs, err := gotenv.Read(nginxInstall.GetEnvPath())
	if err != nil {
		return nil, err
	}

	return &response.NginxBuildConfig{
		Mirror:           envs["CONTAINER_PACKAGE_URL"],
		DynamicSupported: nginxModuleDynamicSupported(nginxInstall),
		Modules:          resList,
	}, nil
}

func applyNginxModuleUpdate(modules []dto.NginxModule, req request.NginxModuleUpdate) ([]dto.NginxModule, *dto.NginxModule, error) {
	switch req.Operate {
	case nginxModuleOperateCreate:
		if err := validateNginxModuleBuildMode(dto.NginxModule{Name: req.Name, BuildMode: req.BuildMode}); err != nil {
			return nil, nil, err
		}
		for _, module := range modules {
			if module.Name == req.Name {
				return nil, nil, buserr.New("ErrNameIsExist")
			}
		}
		modules = append(modules, dto.NginxModule{
			Name:      req.Name,
			Custom:    true,
			Script:    req.Script,
			Packages:  strings.Split(req.Packages, ","),
			Params:    req.Params,
			Enable:    req.Enable,
			BuildMode: req.BuildMode,
			Provider:  req.Provider,
			LoadOrder: req.LoadOrder,
		})
	case nginxModuleOperateUpdate:
		for i := range modules {
			if modules[i].Name != req.Name {
				continue
			}
			if modules[i].Custom {
				if err := validateNginxModuleBuildMode(dto.NginxModule{Name: req.Name, BuildMode: req.BuildMode}); err != nil {
					return nil, nil, err
				}
			} else {
				modules[i].Enable = req.Enable
				return modules, nil, nil
			}
			modules[i].Enable = req.Enable
			modules[i].Script = req.Script
			modules[i].Packages = strings.Split(req.Packages, ",")
			modules[i].Params = req.Params
			modules[i].BuildMode = req.BuildMode
			modules[i].Provider = req.Provider
			modules[i].LoadOrder = req.LoadOrder
			return modules, nil, nil
		}
		return nil, nil, fmt.Errorf("OpenResty module %s not found", req.Name)
	case nginxModuleOperateDelete:
		for i := range modules {
			if modules[i].Name != req.Name {
				continue
			}
			if !modules[i].Custom {
				return nil, nil, fmt.Errorf("built-in OpenResty module %s cannot be deleted", req.Name)
			}
			deleted := modules[i]
			modules = append(modules[:i], modules[i+1:]...)
			return modules, &deleted, nil
		}
		return nil, nil, fmt.Errorf("OpenResty module %s not found", req.Name)
	default:
		return nil, nil, fmt.Errorf("unsupported OpenResty module operation %q", req.Operate)
	}
	return modules, nil, nil
}

func (n NginxService) UpdateModule(req request.NginxModuleUpdate) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if err = task.CheckScopeTaskIsExecuting(task.TaskScopeApp, nginxInstall.ID); err != nil {
		return err
	}
	modules, err := loadNginxModules(nginxInstall)
	if err != nil {
		return err
	}
	oldModules := cloneNginxModules(modules)
	modules, deletedModule, err := applyNginxModuleUpdate(modules, req)
	if err != nil {
		return err
	}
	if err = saveNginxModules(nginxInstall, modules); err != nil {
		return err
	}
	if err = reconcileDynamicNginxModuleConfig(nginxInstall, modules, true); err != nil {
		_ = saveNginxModules(nginxInstall, oldModules)
		_ = reconcileDynamicNginxModuleConfig(nginxInstall, oldModules, false)
		return err
	}
	if deletedModule != nil {
		_ = removeNginxModuleArtifacts(nginxInstall, *deletedModule)
	}
	return nil
}

func (n NginxService) OperateDefaultHTTPs(req request.NginxDefaultHTTPSUpdate) error {
	appInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	websites, _ := websiteRepo.List()
	hasDefaultWebsite := false
	for _, website := range websites {
		if website.DefaultServer {
			hasDefaultWebsite = true
			break
		}
	}
	defaultConfigPath := path.Join(appInstall.GetPath(), nginxModuleConfDir, "default", "00.default.conf")
	content, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return err
	}
	switch req.Operate {
	case "enable":
		if req.SSLRejectHandshake {
			defaultWebsite, _ := websiteRepo.GetFirst(websiteRepo.WithDefaultServer())
			if defaultWebsite.ID > 0 {
				return buserr.New("ErrDefaultWebsite")
			}
		}
		if err := handleSSLConfig(&appInstall, hasDefaultWebsite, req.SSLRejectHandshake); err != nil {
			return err
		}
	case "disable":
		defaultConfig, err := parser.NewStringParser(string(content)).Parse()
		if err != nil {
			return err
		}
		defaultConfig.FilePath = defaultConfigPath
		defaultServer := defaultConfig.FindServers()[0]
		defaultServer.RemoveListen(fmt.Sprintf("%d", appInstall.HttpsPort))
		defaultServer.RemoveListen(fmt.Sprintf("[::]:%d", appInstall.HttpsPort))
		defaultServer.RemoveDirective("include", []string{"/usr/local/openresty/nginx/conf/ssl/root_ssl.conf"})
		defaultServer.RemoveDirective("http2", []string{"on"})
		defaultServer.RemoveDirective("ssl_reject_handshake", []string{"on"})
		if err = nginx.WriteConfig(defaultConfig, nginx.IndentedStyle); err != nil {
			return err
		}
	}
	return nginxCheckAndReload(string(content), defaultConfigPath, appInstall.ContainerName)
}

func (n NginxService) GetDefaultHttpsStatus() (*response.NginxConfigRes, error) {
	appInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	defaultConfigPath := path.Join(appInstall.GetPath(), nginxModuleConfDir, "default", "00.default.conf")
	content, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return nil, err
	}
	defaultConfig, err := parser.NewStringParser(string(content)).Parse()
	if err != nil {
		return nil, err
	}
	defaultConfig.FilePath = defaultConfigPath
	defaultServer := defaultConfig.FindServers()[0]
	res := &response.NginxConfigRes{}
	for _, directive := range defaultServer.GetDirectives() {
		if directive.GetName() == "include" && directive.GetParameters()[0] == "/usr/local/openresty/nginx/conf/ssl/root_ssl.conf" {
			res.Https = true
		}
		if directive.GetName() == "ssl_reject_handshake" && directive.GetParameters()[0] == "on" {
			res.Https = true
			res.SSLRejectHandshake = true
		}
	}
	return res, nil
}
