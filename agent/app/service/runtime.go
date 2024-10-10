package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"gopkg.in/ini.v1"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	cmd2 "github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/env"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
	"github.com/subosito/gotenv"
)

type RuntimeService struct {
}

type IRuntimeService interface {
	Page(req request.RuntimeSearch) (int64, []response.RuntimeDTO, error)
	Create(create request.RuntimeCreate) (*model.Runtime, error)
	Delete(delete request.RuntimeDelete) error
	Update(req request.RuntimeUpdate) error
	Get(id uint) (res *response.RuntimeDTO, err error)
	GetNodePackageRunScript(req request.NodePackageReq) ([]response.PackageScripts, error)
	OperateRuntime(req request.RuntimeOperate) error
	GetNodeModules(req request.NodeModuleReq) ([]response.NodeModule, error)
	OperateNodeModules(req request.NodeModuleOperateReq) error
	SyncForRestart() error
	SyncRuntimeStatus() error
	DeleteCheck(installID uint) ([]dto.AppResource, error)

	GetPHPExtensions(runtimeID uint) (response.PHPExtensionRes, error)
	InstallPHPExtension(req request.PHPExtensionInstallReq) error
	UnInstallPHPExtension(req request.PHPExtensionInstallReq) error
	GetPHPConfig(id uint) (*response.PHPConfig, error)
	UpdatePHPConfig(req request.PHPConfigUpdate) (err error)
	UpdatePHPConfigFile(req request.PHPFileUpdate) error
	GetPHPConfigFile(req request.PHPFileReq) (*response.FileInfo, error)
	UpdateFPMConfig(req request.FPMConfig) error
	GetFPMConfig(id uint) (*request.FPMConfig, error)

	GetSupervisorProcess(id uint) ([]response.SupervisorProcessConfig, error)
	OperateSupervisorProcess(req request.PHPSupervisorProcessConfig) error
	OperateSupervisorProcessFile(req request.PHPSupervisorProcessFileReq) (string, error)
}

func NewRuntimeService() IRuntimeService {
	return &RuntimeService{}
}

func (r *RuntimeService) Create(create request.RuntimeCreate) (*model.Runtime, error) {
	var (
		opts []repo.DBOption
	)
	if create.Name != "" {
		opts = append(opts, commonRepo.WithByLikeName(create.Name))
	}
	if create.Type != "" {
		opts = append(opts, commonRepo.WithByType(create.Type))
	}
	exist, _ := runtimeRepo.GetFirst(opts...)
	if exist != nil {
		return nil, buserr.New(constant.ErrNameIsExist)
	}
	fileOp := files.NewFileOp()

	switch create.Type {
	case constant.RuntimePHP:
		if create.Resource == constant.ResourceLocal {
			runtime := &model.Runtime{
				Name:     create.Name,
				Resource: create.Resource,
				Type:     create.Type,
				Version:  create.Version,
				Status:   constant.RuntimeNormal,
			}
			return nil, runtimeRepo.Create(context.Background(), runtime)
		}
		exist, _ = runtimeRepo.GetFirst(runtimeRepo.WithImage(create.Image))
		if exist != nil {
			return nil, buserr.New(constant.ErrImageExist)
		}
		portValue, _ := create.Params["PANEL_APP_PORT_HTTP"]
		if err := checkPortExist(int(portValue.(float64))); err != nil {
			return nil, err
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython:
		if !fileOp.Stat(create.CodeDir) {
			return nil, buserr.New(constant.ErrPathNotFound)
		}
		create.Install = true
		if err := checkPortExist(create.Port); err != nil {
			return nil, err
		}
		for _, export := range create.ExposedPorts {
			if err := checkPortExist(export.HostPort); err != nil {
				return nil, err
			}
		}
	}
	containerName, ok := create.Params["CONTAINER_NAME"]
	if !ok {
		return nil, buserr.New("ErrContainerNameIsNull")
	}
	if err := checkContainerName(containerName.(string)); err != nil {
		return nil, err
	}

	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(create.AppDetailID))
	if err != nil {
		return nil, err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	if err != nil {
		return nil, err
	}

	appVersionDir := filepath.Join(app.GetAppResourcePath(), appDetail.Version)
	if !fileOp.Stat(appVersionDir) || appDetail.Update {
		if err = downloadApp(app, appDetail, nil, nil); err != nil {
			return nil, err
		}
	}

	runtime := &model.Runtime{
		Name:          create.Name,
		AppDetailID:   create.AppDetailID,
		Type:          create.Type,
		Image:         create.Image,
		Resource:      create.Resource,
		Version:       create.Version,
		ContainerName: containerName.(string),
	}

	switch create.Type {
	case constant.RuntimePHP:
		runtime.Port = int(create.Params["PANEL_APP_PORT_HTTP"].(float64))
		if err = handlePHP(create, runtime, fileOp, appVersionDir); err != nil {
			return nil, err
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython:
		runtime.Port = int(create.Params["port"].(float64))
		if err = handleNodeAndJava(create, runtime, fileOp, appVersionDir); err != nil {
			return nil, err
		}
	}
	if err := runtimeRepo.Create(context.Background(), runtime); err != nil {
		return nil, err
	}
	return runtime, nil
}

func (r *RuntimeService) Page(req request.RuntimeSearch) (int64, []response.RuntimeDTO, error) {
	var (
		opts []repo.DBOption
		res  []response.RuntimeDTO
	)
	if req.Name != "" {
		opts = append(opts, commonRepo.WithByLikeName(req.Name))
	}
	if req.Status != "" {
		opts = append(opts, runtimeRepo.WithStatus(req.Status))
	}
	if req.Type != "" {
		opts = append(opts, commonRepo.WithByType(req.Type))
	}
	total, runtimes, err := runtimeRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	for _, runtime := range runtimes {
		runtimeDTO := response.NewRuntimeDTO(runtime)
		runtimeDTO.Params = make(map[string]interface{})
		envs, err := gotenv.Unmarshal(runtime.Env)
		if err != nil {
			return 0, nil, err
		}
		for k, v := range envs {
			runtimeDTO.Params[k] = v
		}
		res = append(res, runtimeDTO)
	}
	return total, res, nil
}

func (r *RuntimeService) DeleteCheck(runTimeId uint) ([]dto.AppResource, error) {
	var res []dto.AppResource
	websites, _ := websiteRepo.GetBy(websiteRepo.WithRuntimeID(runTimeId))
	for _, website := range websites {
		res = append(res, dto.AppResource{
			Type: "website",
			Name: website.PrimaryDomain,
		})
	}
	return res, nil
}

func (r *RuntimeService) Delete(runtimeDelete request.RuntimeDelete) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(runtimeDelete.ID))
	if err != nil {
		return err
	}
	website, _ := websiteRepo.GetFirst(websiteRepo.WithRuntimeID(runtimeDelete.ID))
	if website.ID > 0 {
		return buserr.New(constant.ErrDelWithWebsite)
	}
	if runtime.Resource != constant.ResourceAppstore {
		return runtimeRepo.DeleteBy(commonRepo.WithByID(runtimeDelete.ID))
	}
	projectDir := runtime.GetPath()
	if out, err := compose.Down(runtime.GetComposePath()); err != nil && !runtimeDelete.ForceDelete {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	if runtime.Type == constant.RuntimePHP {
		client, err := docker.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()
		imageID, err := client.GetImageIDByName(runtime.Image)
		if err != nil {
			return err
		}
		if imageID != "" {
			if err := client.DeleteImage(imageID); err != nil {
				global.LOG.Errorf("delete image id [%s] error %v", imageID, err)
			}
		}
	}
	if err := files.NewFileOp().DeleteDir(projectDir); err != nil && !runtimeDelete.ForceDelete {
		return err
	}
	return runtimeRepo.DeleteBy(commonRepo.WithByID(runtimeDelete.ID))
}

func (r *RuntimeService) Get(id uint) (*response.RuntimeDTO, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}

	res := response.NewRuntimeDTO(*runtime)
	if runtime.Resource == constant.ResourceLocal {
		return &res, nil
	}
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(runtime.AppDetailID))
	if err != nil {
		return nil, err
	}
	res.AppID = appDetail.AppId
	switch runtime.Type {
	case constant.RuntimePHP:
		var (
			appForm   dto.AppForm
			appParams []response.AppParam
		)
		if err := json.Unmarshal([]byte(runtime.Params), &appForm); err != nil {
			return nil, err
		}
		envs, err := gotenv.Unmarshal(runtime.Env)
		if err != nil {
			return nil, err
		}
		if v, ok := envs["CONTAINER_PACKAGE_URL"]; ok {
			res.Source = v
		}
		res.Params = make(map[string]interface{})
		for k, v := range envs {
			if k == "PANEL_APP_PORT_HTTP" {
				port, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				res.Params[k] = port
				continue
			}
			res.Params[k] = v
		}

		for _, form := range appForm.FormFields {
			if v, ok := envs[form.EnvKey]; ok {
				appParam := response.AppParam{
					Edit:     false,
					Key:      form.EnvKey,
					Rule:     form.Rule,
					Type:     form.Type,
					Required: form.Required,
				}
				if form.Edit {
					appParam.Edit = true
				}
				appParam.LabelZh = form.LabelZh
				appParam.LabelEn = form.LabelEn
				appParam.Multiple = form.Multiple
				appParam.Value = v
				if form.Type == "select" {
					if form.Multiple {
						if v == "" {
							appParam.Value = []string{}
						} else {
							appParam.Value = strings.Split(v, ",")
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
				}
				appParams = append(appParams, appParam)
			}
		}
		res.AppParams = appParams
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython:
		res.Params = make(map[string]interface{})
		envs, err := gotenv.Unmarshal(runtime.Env)
		if err != nil {
			return nil, err
		}
		for k, v := range envs {
			switch k {
			case "NODE_APP_PORT", "PANEL_APP_PORT_HTTP", "JAVA_APP_PORT", "GO_APP_PORT", "APP_PORT", "port":
				port, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				res.Params[k] = port
			default:
				if strings.Contains(k, "CONTAINER_PORT") || strings.Contains(k, "HOST_PORT") {
					if strings.Contains(k, "CONTAINER_PORT") {
						r := regexp.MustCompile(`_(\d+)$`)
						matches := r.FindStringSubmatch(k)
						containerPort, err := strconv.Atoi(v)
						if err != nil {
							return nil, err
						}
						hostPort, err := strconv.Atoi(envs[fmt.Sprintf("HOST_PORT_%s", matches[1])])
						if err != nil {
							return nil, err
						}
						res.ExposedPorts = append(res.ExposedPorts, request.ExposedPort{
							ContainerPort: containerPort,
							HostPort:      hostPort,
						})
					}
				} else {
					res.Params[k] = v
				}
			}
		}
		if v, ok := envs["CONTAINER_PACKAGE_URL"]; ok {
			res.Source = v
		}
		composeByte, err := files.NewFileOp().GetContent(runtime.GetComposePath())
		if err != nil {
			return nil, err
		}
		res.Environments, err = getDockerComposeEnvironments(composeByte)
		if err != nil {
			return nil, err
		}
		volumes, err := getDockerComposeVolumes(composeByte)
		if err != nil {
			return nil, err
		}

		defaultVolumes := make(map[string]string)
		switch runtime.Type {
		case constant.RuntimeNode:
			defaultVolumes = constant.RuntimeDefaultVolumes
		case constant.RuntimeJava:
			defaultVolumes = constant.RuntimeDefaultVolumes
		case constant.RuntimeGo:
			defaultVolumes = constant.GoDefaultVolumes
		case constant.RuntimePython:
			defaultVolumes = constant.RuntimeDefaultVolumes
		}
		for _, volume := range volumes {
			exist := false
			for key, value := range defaultVolumes {
				if key == volume.Source && value == volume.Target {
					exist = true
					break
				}
			}
			if !exist {
				res.Volumes = append(res.Volumes, volume)
			}
		}
	}

	return &res, nil
}

func (r *RuntimeService) Update(req request.RuntimeUpdate) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if runtime.Resource == constant.ResourceLocal {
		runtime.Version = req.Version
		return runtimeRepo.Save(runtime)
	}
	oldImage := runtime.Image
	oldEnv := runtime.Env
	req.Port = int(req.Params["port"].(float64))
	switch runtime.Type {
	case constant.RuntimePHP:
		exist, _ := runtimeRepo.GetFirst(runtimeRepo.WithImage(req.Name), runtimeRepo.WithNotId(req.ID))
		if exist != nil {
			return buserr.New(constant.ErrImageExist)
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython:
		if runtime.Port != req.Port {
			if err = checkPortExist(req.Port); err != nil {
				return err
			}
			runtime.Port = req.Port
		}
		for _, export := range req.ExposedPorts {
			if err = checkPortExist(export.HostPort); err != nil {
				return err
			}
		}

		appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(runtime.AppDetailID))
		if err != nil {
			return err
		}
		app, err := appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
		if err != nil {
			return err
		}
		fileOp := files.NewFileOp()
		appVersionDir := path.Join(constant.AppResourceDir, app.Resource, app.Key, appDetail.Version)
		if !fileOp.Stat(appVersionDir) || appDetail.Update {
			if err := downloadApp(app, appDetail, nil, nil); err != nil {
				return err
			}
			_ = fileOp.Rename(path.Join(runtime.GetPath(), "run.sh"), path.Join(runtime.GetPath(), "run.sh.bak"))
			_ = fileOp.CopyFile(path.Join(appVersionDir, "run.sh"), runtime.GetPath())
		}
	}

	if containerName, ok := req.Params["CONTAINER_NAME"]; ok && containerName != getRuntimeEnv(runtime.Env, "CONTAINER_NAME") {
		if err := checkContainerName(containerName.(string)); err != nil {
			return err
		}
	}

	projectDir := path.Join(constant.RuntimeDir, runtime.Type, runtime.Name)
	create := request.RuntimeCreate{
		Image:   req.Image,
		Type:    runtime.Type,
		Source:  req.Source,
		Params:  req.Params,
		CodeDir: req.CodeDir,
		Version: req.Version,
		NodeConfig: request.NodeConfig{
			Port:         req.Port,
			Install:      true,
			ExposedPorts: req.ExposedPorts,
			Environments: req.Environments,
			Volumes:      req.Volumes,
		},
	}
	composeContent, envContent, _, err := handleParams(create, projectDir)
	if err != nil {
		return err
	}
	runtime.Env = string(envContent)
	runtime.DockerCompose = string(composeContent)

	switch runtime.Type {
	case constant.RuntimePHP:
		runtime.Image = req.Image
		runtime.Status = constant.RuntimeBuildIng
		_ = runtimeRepo.Save(runtime)
		client, err := docker.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()
		imageID, err := client.GetImageIDByName(oldImage)
		if err != nil {
			return err
		}
		go buildRuntime(runtime, imageID, oldEnv, req.Rebuild)
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython:
		runtime.Version = req.Version
		runtime.CodeDir = req.CodeDir
		runtime.Port = req.Port
		runtime.Status = constant.RuntimeReCreating
		_ = runtimeRepo.Save(runtime)
		go reCreateRuntime(runtime)
	}
	return nil
}

func (r *RuntimeService) GetNodePackageRunScript(req request.NodePackageReq) ([]response.PackageScripts, error) {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.CodeDir) {
		return nil, buserr.New(constant.ErrPathNotFound)
	}
	if !fileOp.Stat(path.Join(req.CodeDir, "package.json")) {
		return nil, buserr.New(constant.ErrPackageJsonNotFound)
	}
	content, err := fileOp.GetContent(path.Join(req.CodeDir, "package.json"))
	if err != nil {
		return nil, err
	}
	var packageMap map[string]interface{}
	err = json.Unmarshal(content, &packageMap)
	if err != nil {
		return nil, err
	}
	scripts, ok := packageMap["scripts"]
	if !ok {
		return nil, buserr.New(constant.ErrScriptsNotFound)
	}
	var packageScripts []response.PackageScripts
	for k, v := range scripts.(map[string]interface{}) {
		packageScripts = append(packageScripts, response.PackageScripts{
			Name:   k,
			Script: v.(string),
		})
	}
	return packageScripts, nil
}

func (r *RuntimeService) OperateRuntime(req request.RuntimeOperate) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			runtime.Status = constant.RuntimeError
			runtime.Message = err.Error()
			_ = runtimeRepo.Save(runtime)
		}
	}()
	switch req.Operate {
	case constant.RuntimeUp:
		if err = runComposeCmdWithLog(req.Operate, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
			return err
		}
		if err = SyncRuntimeContainerStatus(runtime); err != nil {
			return err
		}
	case constant.RuntimeDown:
		if err = runComposeCmdWithLog(req.Operate, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
			return err
		}
		runtime.Status = constant.RuntimeStopped
	case constant.RuntimeRestart:
		if err = restartRuntime(runtime); err != nil {
			return err
		}
		if err = SyncRuntimeContainerStatus(runtime); err != nil {
			return err
		}
	}
	return runtimeRepo.Save(runtime)
}

func (r *RuntimeService) GetNodeModules(req request.NodeModuleReq) ([]response.NodeModule, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return nil, err
	}
	var res []response.NodeModule
	nodeModulesPath := path.Join(runtime.CodeDir, "node_modules")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(nodeModulesPath) {
		return nil, buserr.New("ErrNodeModulesNotFound")
	}
	moduleDirs, err := os.ReadDir(nodeModulesPath)
	if err != nil {
		return nil, err
	}
	for _, moduleDir := range moduleDirs {
		packagePath := path.Join(nodeModulesPath, moduleDir.Name(), "package.json")
		if !fileOp.Stat(packagePath) {
			continue
		}
		content, err := fileOp.GetContent(packagePath)
		if err != nil {
			continue
		}
		module := response.NodeModule{}
		if err := json.Unmarshal(content, &module); err != nil {
			continue
		}
		res = append(res, module)
	}
	return res, nil
}

func (r *RuntimeService) OperateNodeModules(req request.NodeModuleOperateReq) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	containerName, err := env.GetEnvValueByKey(runtime.GetEnvPath(), "CONTAINER_NAME")
	if err != nil {
		return err
	}
	cmd := req.PkgManager
	switch req.Operate {
	case constant.RuntimeInstall:
		if req.PkgManager == constant.RuntimeNpm {
			cmd += " install"
		} else {
			cmd += " add"
		}
	case constant.RuntimeUninstall:
		if req.PkgManager == constant.RuntimeNpm {
			cmd += " uninstall"
		} else {
			cmd += " remove"
		}
	case constant.RuntimeUpdate:
		if req.PkgManager == constant.RuntimeNpm {
			cmd += " update"
		} else {
			cmd += " upgrade"
		}
	}
	cmd += " " + req.Module
	return cmd2.ExecContainerScript(containerName, cmd, 5*time.Minute)
}

func (r *RuntimeService) SyncForRestart() error {
	runtimes, err := runtimeRepo.List()
	if err != nil {
		return err
	}
	for _, runtime := range runtimes {
		if runtime.Status == constant.RuntimeBuildIng || runtime.Status == constant.RuntimeReCreating || runtime.Status == constant.RuntimeStarting || runtime.Status == constant.RuntimeCreating {
			runtime.Status = constant.SystemRestart
			runtime.Message = "System restart causing interrupt"
			_ = runtimeRepo.Save(&runtime)
		}
	}
	return nil
}

func (r *RuntimeService) SyncRuntimeStatus() error {
	runtimes, err := runtimeRepo.List()
	if err != nil {
		return err
	}
	for _, runtime := range runtimes {
		if runtime.Type == constant.RuntimeNode || runtime.Type == constant.RuntimeJava || runtime.Type == constant.RuntimeGo || runtime.Type == constant.RuntimePython {
			_ = SyncRuntimeContainerStatus(&runtime)
		}
	}
	return nil
}

func (r *RuntimeService) GetPHPExtensions(runtimeID uint) (response.PHPExtensionRes, error) {
	var res response.PHPExtensionRes
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(runtimeID))
	if err != nil {
		return res, err
	}
	phpCmd := fmt.Sprintf("docker exec -i %s %s", runtime.ContainerName, "php -m")
	out, err := cmd2.ExecWithTimeOut(phpCmd, 20*time.Second)
	if err != nil {
		if out != "" {
			return res, errors.New(out)
		}
		return res, err
	}
	extensions := strings.Split(out, "\n")
	exitExtensions := make(map[string]string)
	for _, ext := range extensions {
		extStr := strings.TrimSpace(ext)
		if extStr != "" && extStr != "[Zend Modules]" && extStr != "[PHP Modules]" {
			exitExtensions[strings.ToLower(extStr)] = extStr
		}
	}
	var phpExtensions []response.SupportExtension
	if err = json.Unmarshal(nginx_conf.PHPExtensionsJson, &phpExtensions); err != nil {
		return res, err
	}
	for _, ext := range phpExtensions {
		if _, ok := exitExtensions[strings.ToLower(ext.Check)]; ok {
			ext.Installed = true
		}
		res.SupportExtensions = append(res.SupportExtensions, ext)
	}
	for _, name := range exitExtensions {
		res.Extensions = append(res.Extensions, name)
	}
	sort.Slice(res.Extensions, func(i, j int) bool {
		return strings.ToLower(res.Extensions[i]) < strings.ToLower(res.Extensions[j])
	})
	return res, nil
}

func (r *RuntimeService) InstallPHPExtension(req request.PHPExtensionInstallReq) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	installTask, err := task.NewTaskWithOps(req.Name, task.TaskInstall, task.TaskScopeRuntimeExtension, req.TaskID, runtime.ID)
	if err != nil {
		return err
	}
	installTask.AddSubTask("", func(t *task.Task) error {
		installCmd := fmt.Sprintf("docker exec -i %s %s %s", runtime.ContainerName, "install-ext", req.Name)
		err = cmd2.ExecWithLogFile(installCmd, 15*time.Minute, t.Task.LogFile)
		if err != nil {
			return err
		}
		client, err := docker.NewClient()
		defer client.Close()
		if err == nil {
			oldImageID, err := client.GetImageIDByName(runtime.Image)
			commitCmd := fmt.Sprintf("docker commit %s %s", runtime.ContainerName, runtime.Image)
			err = cmd2.ExecWithLogFile(commitCmd, 15*time.Minute, t.Task.LogFile)
			if err != nil {
				return err
			}
			newImageID, err := client.GetImageIDByName(runtime.Image)
			if err == nil && newImageID != oldImageID {
				if err := client.DeleteImage(oldImageID); err != nil {
					t.Log(fmt.Sprintf("delete old image error %v", err))
				} else {
					t.Log("delete old image success")
				}
			}
		}
		if err = restartRuntime(runtime); err != nil {
			return err
		}
		return nil
	}, nil)
	go func() {
		err = installTask.Execute()
		if err == nil {
			envs, err := gotenv.Unmarshal(runtime.Env)
			if err != nil {
				global.LOG.Errorf("get runtime env error %v", err)
				return
			}
			extensions, ok := envs["PHP_EXTENSIONS"]
			exist := false
			var extensionArray []string
			if ok {
				extensionArray = strings.Split(extensions, ",")
				for _, ext := range extensionArray {
					if ext == req.Name {
						exist = true
						break
					}
				}
			}
			if !exist {
				extensionArray = append(extensionArray, req.Name)
				envs["PHP_EXTENSIONS"] = strings.Join(extensionArray, ",")
				if err = gotenv.Write(envs, runtime.GetEnvPath()); err != nil {
					global.LOG.Errorf("write runtime env error %v", err)
					return
				}
				envStr, err := gotenv.Marshal(envs)
				if err != nil {
					global.LOG.Errorf("marshal runtime env error %v", err)
					return
				}
				runtime.Env = envStr
				_ = runtimeRepo.Save(runtime)
			}
		}
	}()
	return nil
}

func (r *RuntimeService) UnInstallPHPExtension(req request.PHPExtensionInstallReq) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if err = unInstallPHPExtension(runtime, []string{req.Name}); err != nil {
		return err
	}
	if err = restartRuntime(runtime); err != nil {
		return err
	}
	return runtimeRepo.Save(runtime)
}

func (r *RuntimeService) GetPHPConfig(id uint) (*response.PHPConfig, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	phpConfigPath := path.Join(runtime.GetPath(), "conf", "php.ini")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(phpConfigPath) {
		return nil, buserr.WithName("ErrFileNotFound", "php.ini")
	}
	params := make(map[string]string)
	configFile, err := fileOp.OpenFile(phpConfigPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") {
			continue
		}
		matches := regexp.MustCompile(`^\s*([a-z_]+)\s*=\s*(.*)$`).FindStringSubmatch(line)
		if len(matches) == 3 {
			params[matches[1]] = matches[2]
		}
	}
	cfg, err := ini.Load(phpConfigPath)
	if err != nil {
		return nil, err
	}
	phpConfig, err := cfg.GetSection("PHP")
	if err != nil {
		return nil, err
	}
	disableFunctionStr := phpConfig.Key("disable_functions").Value()
	res := &response.PHPConfig{Params: params}
	if disableFunctionStr != "" {
		disableFunctions := strings.Split(disableFunctionStr, ",")
		if len(disableFunctions) > 0 {
			res.DisableFunctions = disableFunctions
		}
	}
	uploadMaxSize := phpConfig.Key("upload_max_filesize").Value()
	if uploadMaxSize != "" {
		res.UploadMaxSize = uploadMaxSize
	}
	return res, nil
}

func (r *RuntimeService) UpdatePHPConfig(req request.PHPConfigUpdate) (err error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	phpConfigPath := path.Join(runtime.GetPath(), "conf", "php.ini")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(phpConfigPath) {
		return buserr.WithName("ErrFileNotFound", "php.ini")
	}
	configFile, err := fileOp.OpenFile(phpConfigPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	contentBytes, err := fileOp.GetContent(phpConfigPath)
	if err != nil {
		return err
	}

	content := string(contentBytes)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, ";") {
			continue
		}
		switch req.Scope {
		case "params":
			for key, value := range req.Params {
				pattern := "^" + regexp.QuoteMeta(key) + "\\s*=\\s*.*$"
				if matched, _ := regexp.MatchString(pattern, line); matched {
					lines[i] = key + " = " + value
				}
			}
		case "disable_functions":
			pattern := "^" + regexp.QuoteMeta("disable_functions") + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(pattern, line); matched {
				lines[i] = "disable_functions" + " = " + strings.Join(req.DisableFunctions, ",")
				break
			}
		case "upload_max_filesize":
			pattern := "^" + regexp.QuoteMeta("post_max_size") + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(pattern, line); matched {
				lines[i] = "post_max_size" + " = " + req.UploadMaxSize
			}
			patternUpload := "^" + regexp.QuoteMeta("upload_max_filesize") + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(patternUpload, line); matched {
				lines[i] = "upload_max_filesize" + " = " + req.UploadMaxSize
			}
		}
	}
	updatedContent := strings.Join(lines, "\n")
	if err := fileOp.WriteFile(phpConfigPath, strings.NewReader(updatedContent), 0755); err != nil {
		return err
	}

	err = restartRuntime(runtime)
	if err != nil {
		_ = fileOp.WriteFile(phpConfigPath, strings.NewReader(string(contentBytes)), 0755)
		return err
	}
	return
}

func (r *RuntimeService) GetPHPConfigFile(req request.PHPFileReq) (*response.FileInfo, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return nil, err
	}
	configPath := ""
	switch req.Type {
	case constant.ConfigFPM:
		configPath = path.Join(runtime.GetPath(), "conf", "php-fpm.conf")
	case constant.ConfigPHP:
		configPath = path.Join(runtime.GetPath(), "conf", "php.ini")
	}
	info, err := files.NewFileInfo(files.FileOption{
		Path:   configPath,
		Expand: true,
	})
	if err != nil {
		return nil, err
	}
	return &response.FileInfo{FileInfo: *info}, nil
}

func (r *RuntimeService) UpdatePHPConfigFile(req request.PHPFileUpdate) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	configPath := ""
	if req.Type == constant.ConfigFPM {
		configPath = path.Join(runtime.GetPath(), "conf", "php-fpm.conf")
	} else {
		configPath = path.Join(runtime.GetPath(), "conf", "php.ini")
	}
	if err := files.NewFileOp().WriteFile(configPath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}
	if _, err := compose.Restart(runtime.GetComposePath()); err != nil {
		return err
	}
	return nil
}

func (r *RuntimeService) UpdateFPMConfig(req request.FPMConfig) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	cfg, err := ini.Load(runtime.GetFPMPath())
	if err != nil {
		return err
	}
	for k, v := range req.Params {
		var valueStr string
		switch v := v.(type) {
		case string:
			valueStr = v
		case int:
			valueStr = fmt.Sprintf("%d", v)
		case float64:
			valueStr = fmt.Sprintf("%.f", v)
		default:
			continue
		}
		cfg.Section("www").Key(k).SetValue(valueStr)
	}
	if err := cfg.SaveTo(runtime.GetFPMPath()); err != nil {
		return err
	}
	return nil
}

var PmKeys = map[string]struct {
}{
	"pm":                   {},
	"pm.max_children":      {},
	"pm.start_servers":     {},
	"pm.min_spare_servers": {},
	"pm.max_spare_servers": {},
}

func (r *RuntimeService) GetFPMConfig(id uint) (*request.FPMConfig, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	fileOp := files.NewFileOp()
	if !fileOp.Stat(runtime.GetFPMPath()) {
		return nil, buserr.WithName("ErrFileNotFound", "php-fpm.conf")
	}
	params := make(map[string]interface{})
	cfg, err := ini.Load(runtime.GetFPMPath())
	if err != nil {
		return nil, err
	}
	for _, key := range cfg.Section("www").Keys() {
		if _, ok := PmKeys[key.Name()]; ok {
			params[key.Name()] = key.Value()
		}
	}
	res := &request.FPMConfig{Params: params}
	return res, nil
}

func (r *RuntimeService) GetSupervisorProcess(id uint) ([]response.SupervisorProcessConfig, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	configDir := path.Join(constant.RuntimeDir, "php", runtime.Name, "supervisor", "supervisor.d")
	return handleProcessConfig(configDir, runtime.ContainerName)
}

func (r *RuntimeService) OperateSupervisorProcess(req request.PHPSupervisorProcessConfig) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	configDir := path.Join(constant.RuntimeDir, "php", runtime.Name, "supervisor")
	return handleProcess(configDir, req.SupervisorProcessConfig, runtime.ContainerName)
}

func (r *RuntimeService) OperateSupervisorProcessFile(req request.PHPSupervisorProcessFileReq) (string, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return "", err
	}
	supervisorDir := path.Join(constant.RuntimeDir, "php", runtime.Name, "supervisor")
	configDir := path.Join(supervisorDir, "supervisor.d")
	logFile := path.Join(supervisorDir, "log", fmt.Sprintf("%s.out.log", req.SupervisorProcessFileReq.Name))
	return handleSupervisorFile(req.SupervisorProcessFileReq, configDir, runtime.ContainerName, logFile)
}
