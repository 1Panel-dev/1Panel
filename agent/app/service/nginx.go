package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	cmd2 "github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/subosito/gotenv"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

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
	ClearProxyCache() error

	Build(req request.NginxBuildReq) error
	GetModules() (*response.NginxBuildConfig, error)
	UpdateModule(req request.NginxModuleUpdate) error
}

func NewINginxService() INginxService {
	return &NginxService{}
}

func (n NginxService) GetNginxConfig() (*response.NginxFile, error) {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	configPath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "nginx.conf")
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
	if err != nil {
		return response.NginxStatus{}, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return response.NginxStatus{}, err
	}
	var status response.NginxStatus
	resArray := strings.Split(string(content), " ")
	status.Active = resArray[2]
	status.Accepts = resArray[7]
	status.Handled = resArray[8]
	status.Requests = resArray[9]
	status.Reading = resArray[11]
	status.Writing = resArray[13]
	status.Waiting = resArray[15]
	return status, nil
}

func (n NginxService) UpdateConfigFile(req request.NginxConfigFileUpdate) error {
	fileOp := files.NewFileOp()
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	filePath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "conf", "nginx.conf")
	if req.Backup {
		backupPath := path.Join(path.Dir(filePath), "bak")
		if !fileOp.Stat(backupPath) {
			if err := fileOp.CreateDir(backupPath, 0755); err != nil {
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
	if err = fileOp.WriteFile(filePath, strings.NewReader(req.Content), 0644); err != nil {
		return err
	}
	if status, err := checkContainerStatus(nginxInstall.ContainerName); err == nil && status != "running" {
		if out, err := compose.DownAndUp(nginxInstall.GetComposePath()); err != nil {
			_ = fileOp.SaveFile(filePath, string(oldContent), 0644)
			return fmt.Errorf("nginx restart failed: %v", out)
		} else {
			return nginxCheckAndReload(string(oldContent), filePath, nginxInstall.ContainerName)
		}
	}
	return nginxCheckAndReload(string(oldContent), filePath, nginxInstall.ContainerName)
}

func (n NginxService) ClearProxyCache() error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	cacheDir := path.Join(nginxInstall.GetPath(), "www/common/proxy/proxy_cache_dir")
	fileOp := files.NewFileOp()
	if fileOp.Stat(cacheDir) {
		if err = fileOp.CleanDir(cacheDir); err != nil {
			return err
		}
		_, err = compose.Restart(nginxInstall.GetComposePath())
		if err != nil {
			return err
		}
	}
	return nil
}

func (n NginxService) Build(req request.NginxBuildReq) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	buildPath := path.Join(nginxInstall.GetPath(), "build")
	if !fileOp.Stat(buildPath) {
		return buserr.New("ErrBuildDirNotFound")
	}
	moduleConfigPath := path.Join(buildPath, "module.json")
	moduleContent, err := fileOp.GetContent(moduleConfigPath)
	if err != nil {
		return err
	}
	var (
		modules         []dto.NginxModule
		addModuleParams []string
		addPackages     []string
	)
	if len(moduleContent) > 0 {
		_ = json.Unmarshal(moduleContent, &modules)
		bashFile, err := os.OpenFile(path.Join(buildPath, "tmp", "pre.sh"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer bashFile.Close()
		bashFileWriter := bufio.NewWriter(bashFile)
		for _, module := range modules {
			if !module.Enable {
				continue
			}
			_, err = bashFileWriter.WriteString(module.Script + "\n")
			if err != nil {
				return err
			}
			addModuleParams = append(addModuleParams, module.Params)
			addPackages = append(addPackages, module.Packages...)
		}
		err = bashFileWriter.Flush()
		if err != nil {
			return err
		}
	}
	envs, err := gotenv.Read(nginxInstall.GetEnvPath())
	if err != nil {
		return err
	}
	envs["CONTAINER_PACKAGE_URL"] = req.Mirror
	envs["RESTY_CONFIG_OPTIONS_MORE"] = ""
	envs["RESTY_ADD_PACKAGE_BUILDDEPS"] = ""
	if len(addModuleParams) > 0 {
		envs["RESTY_CONFIG_OPTIONS_MORE"] = strings.Join(addModuleParams, " ")
	}
	if len(addPackages) > 0 {
		envs["RESTY_ADD_PACKAGE_BUILDDEPS"] = strings.Join(addPackages, " ")
	}
	_ = gotenv.Write(envs, nginxInstall.GetEnvPath())

	buildTask, err := task.NewTaskWithOps(nginxInstall.Name, task.TaskBuild, task.TaskScopeApp, req.TaskID, nginxInstall.ID)
	if err != nil {
		return err
	}
	buildTask.AddSubTask("", func(t *task.Task) error {
		if err = cmd2.ExecWithLogFile(fmt.Sprintf("docker compose -f %s build", nginxInstall.GetComposePath()), 15*time.Minute, t.Task.LogFile); err != nil {
			return err
		}
		_, err = compose.DownAndUp(nginxInstall.GetComposePath())
		return err
	}, nil)

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
	fileOp := files.NewFileOp()
	var modules []dto.NginxModule
	moduleConfigPath := path.Join(nginxInstall.GetPath(), "build", "module.json")
	if !fileOp.Stat(moduleConfigPath) {
		return nil, nil
	}
	moduleContent, err := fileOp.GetContent(moduleConfigPath)
	if err != nil {
		return nil, err
	}
	if len(moduleContent) > 0 {
		_ = json.Unmarshal(moduleContent, &modules)
	}
	var resList []response.NginxModule
	for _, module := range modules {
		resList = append(resList, response.NginxModule{
			Name:     module.Name,
			Script:   module.Script,
			Packages: strings.Join(module.Packages, ","),
			Params:   module.Params,
			Enable:   module.Enable,
		})
	}
	envs, err := gotenv.Read(nginxInstall.GetEnvPath())
	if err != nil {
		return nil, err
	}

	return &response.NginxBuildConfig{
		Mirror:  envs["CONTAINER_PACKAGE_URL"],
		Modules: resList,
	}, nil
}

func (n NginxService) UpdateModule(req request.NginxModuleUpdate) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	var (
		modules []dto.NginxModule
	)
	moduleConfigPath := path.Join(nginxInstall.GetPath(), "build", "module.json")
	if !fileOp.Stat(moduleConfigPath) {
		_ = fileOp.CreateFile(moduleConfigPath)
	}
	moduleContent, err := fileOp.GetContent(moduleConfigPath)
	if err != nil {
		return err
	}
	if len(moduleContent) > 0 {
		_ = json.Unmarshal(moduleContent, &modules)
	}

	switch req.Operate {
	case "create":
		for _, module := range modules {
			if module.Name == req.Name {
				return buserr.New("ErrNameIsExist")
			}
		}
		modules = append(modules, dto.NginxModule{
			Name:     req.Name,
			Script:   req.Script,
			Packages: strings.Split(req.Packages, ","),
			Params:   req.Params,
			Enable:   true,
		})
	case "update":
		for i, module := range modules {
			if module.Name == req.Name {
				modules[i].Script = req.Script
				modules[i].Packages = strings.Split(req.Packages, ",")
				modules[i].Params = req.Params
				modules[i].Enable = req.Enable
				break
			}
		}
	case "delete":
		for i, module := range modules {
			if module.Name == req.Name {
				modules = append(modules[:i], modules[i+1:]...)
				break
			}
		}
	}
	moduleByte, err := json.Marshal(modules)
	if err != nil {
		return err
	}
	return fileOp.SaveFileWithByte(moduleConfigPath, moduleByte, 0644)
}
