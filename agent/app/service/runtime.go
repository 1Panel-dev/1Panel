package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	fcgiclient "github.com/tomasen/fcgi_client"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"gopkg.in/ini.v1"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
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
	UpdateRemark(req request.RuntimeRemark) error

	GetPHPExtensions(runtimeID uint) (response.PHPExtensionRes, error)
	InstallPHPExtension(req request.PHPExtensionInstallReq) error
	UnInstallPHPExtension(req request.PHPExtensionInstallReq) error

	GetPHPConfig(id uint) (*response.PHPConfig, error)
	UpdatePHPConfig(req request.PHPConfigUpdate) (err error)
	UpdatePHPConfigFile(req request.PHPFileUpdate) error
	GetPHPConfigFile(req request.PHPFileReq) (*response.FileInfo, error)
	UpdateFPMConfig(req request.FPMConfig) error
	GetFPMConfig(id uint) (*request.FPMConfig, error)

	UpdatePHPContainer(req request.PHPContainerConfig) error
	GetPHPContainerConfig(id uint) (*request.PHPContainerConfig, error)

	GetSupervisorProcess(id uint) ([]response.SupervisorProcessConfig, error)
	OperateSupervisorProcess(req request.PHPSupervisorProcessConfig) error
	OperateSupervisorProcessFile(req request.PHPSupervisorProcessFileReq) (string, error)

	GetFPMStatus(runtimeID uint) ([]response.FpmStatusItem, error)
}

func NewRuntimeService() IRuntimeService {
	return &RuntimeService{}
}

func (r *RuntimeService) Create(create request.RuntimeCreate) (*model.Runtime, error) {
	var (
		opts []repo.DBOption
	)
	if create.Name != "" {
		opts = append(opts, repo.WithByName(create.Name))
	}
	if create.Type != "" {
		opts = append(opts, repo.WithByType(create.Type))
	}
	exist, _ := runtimeRepo.GetFirst(context.Background(), opts...)
	if exist != nil {
		return nil, buserr.New("ErrNameIsExist")
	}
	fileOp := files.NewFileOp()

	runtimeDir := path.Join(global.Dir.RuntimeDir, create.Type)
	if !fileOp.Stat(runtimeDir) {
		if err := fileOp.CreateDir(runtimeDir, constant.DirPerm); err != nil {
			return nil, err
		}
	}
	var hostPorts []string
	switch create.Type {
	case constant.RuntimePHP:
		if create.Resource == constant.ResourceLocal {
			runtime := &model.Runtime{
				Name:     create.Name,
				Resource: create.Resource,
				Type:     create.Type,
				Version:  create.Version,
				Status:   constant.StatusNormal,
				Remark:   create.Remark,
			}
			return nil, runtimeRepo.Create(context.Background(), runtime)
		}
		exist, _ = runtimeRepo.GetFirst(context.Background(), runtimeRepo.WithImage(create.Image))
		if exist != nil {
			return nil, buserr.New("ErrImageExist")
		}
		fpmPort, ok := create.Params["PANEL_APP_PORT_HTTP"]
		if !ok {
			return nil, buserr.New("ErrPortNotFound")
		}
		hostPorts = append(hostPorts, fmt.Sprintf("%.0f", fpmPort.(float64)))
		if err := checkPortExist(int(fpmPort.(float64))); err != nil {
			return nil, err
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		if !fileOp.Stat(create.CodeDir) {
			return nil, buserr.New("ErrPathNotFound")
		}
		create.Install = true
		for _, export := range create.ExposedPorts {
			hostPorts = append(hostPorts, strconv.Itoa(export.HostPort))
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

	appDetail, err := appDetailRepo.GetFirst(repo.WithByID(create.AppDetailID))
	if err != nil {
		return nil, err
	}
	app, err := appRepo.GetFirst(repo.WithByID(appDetail.AppId))
	if err != nil {
		return nil, err
	}

	appVersionDir := filepath.Join(app.GetAppResourcePath(), appDetail.Version)
	if !fileOp.Stat(appVersionDir) {
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
		Port:          strings.Join(hostPorts, ","),
		Remark:        create.Remark,
	}

	switch create.Type {
	case constant.RuntimePHP:
		if err = handlePHP(create, runtime, fileOp, appVersionDir); err != nil {
			return nil, err
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		if err = handleRuntime(create, runtime, fileOp, appVersionDir); err != nil {
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
		opts = append(opts, repo.WithByLikeName(req.Name))
	}
	if req.Status != "" {
		if req.Type == constant.TypePhp {
			opts = append(opts, runtimeRepo.WithNormalStatus(req.Status))
		} else {
			opts = append(opts, runtimeRepo.WithStatus(req.Status))
		}
	}
	if req.Type != "" {
		opts = append(opts, repo.WithByType(req.Type))
	}
	total, runtimes, err := runtimeRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	if len(runtimes) == 0 {
		return 0, res, nil
	}
	if err = SyncRuntimesStatus(runtimes); err != nil {
		return 0, nil, err
	}
	for _, runtime := range runtimes {
		if runtime.Resource == constant.ResourceLocal {
			runtime.Status = constant.StatusNormal
		}
		runtimeDTO := response.NewRuntimeDTO(runtime)
		runtimeDTO.Params = make(map[string]interface{})
		envs, err := gotenv.Unmarshal(runtime.Env)
		if err != nil {
			return 0, nil, err
		}
		for k, v := range envs {
			runtimeDTO.Params[k] = v
			if strings.Contains(k, "CONTAINER_PORT") || strings.Contains(k, "HOST_PORT") {
				if strings.Contains(k, "CONTAINER_PORT") {
					r := regexp.MustCompile(`_(\d+)$`)
					matches := r.FindStringSubmatch(k)
					containerPort, err := strconv.Atoi(v)
					if err != nil {
						continue
					}
					hostPort, err := strconv.Atoi(envs[fmt.Sprintf("HOST_PORT_%s", matches[1])])
					if err != nil {
						continue
					}
					hostIP := envs[fmt.Sprintf("HOST_IP_%s", matches[1])]
					if hostIP == "" {
						hostIP = "0.0.0.0"
					}
					runtimeDTO.ExposedPorts = append(runtimeDTO.ExposedPorts, request.ExposedPort{
						ContainerPort: containerPort,
						HostPort:      hostPort,
						HostIP:        hostIP,
					})
				}
			}
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(runtimeDelete.ID))
	if err != nil {
		return err
	}
	website, _ := websiteRepo.GetFirst(websiteRepo.WithRuntimeID(runtimeDelete.ID))
	if website.ID > 0 {
		return buserr.New("ErrDelWithWebsite")
	}
	if runtime.Resource != constant.ResourceAppstore {
		return runtimeRepo.DeleteBy(repo.WithByID(runtimeDelete.ID))
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
	return runtimeRepo.DeleteBy(repo.WithByID(runtimeDelete.ID))
}

func (r *RuntimeService) Get(id uint) (*response.RuntimeDTO, error) {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(id))
	if err != nil {
		return nil, err
	}

	res := response.NewRuntimeDTO(*runtime)
	if runtime.Resource == constant.ResourceLocal {
		return &res, nil
	}
	appDetail, err := appDetailRepo.GetFirst(repo.WithByID(runtime.AppDetailID))
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
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		if err := handleRuntimeDTO(&res, *runtime); err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func (r *RuntimeService) Update(req request.RuntimeUpdate) error {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if runtime.Resource == constant.ResourceLocal {
		runtime.Version = req.Version
		return runtimeRepo.Save(runtime)
	}
	oldImage := runtime.Image
	oldEnv := runtime.Env
	var hostPorts []string
	switch runtime.Type {
	case constant.RuntimePHP:
		exist, _ := runtimeRepo.GetFirst(context.Background(), runtimeRepo.WithImage(req.Name), runtimeRepo.WithNotId(req.ID))
		if exist != nil {
			return buserr.New("ErrImageExist")
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		for _, export := range req.ExposedPorts {
			hostPorts = append(hostPorts, strconv.Itoa(export.HostPort))
			if err = checkRuntimePortExist(export.HostPort, false, runtime.ID); err != nil {
				return err
			}
		}

		appDetail, err := appDetailRepo.GetFirst(repo.WithByID(runtime.AppDetailID))
		if err != nil {
			return err
		}
		app, err := appRepo.GetFirst(repo.WithByID(appDetail.AppId))
		if err != nil {
			return err
		}
		fileOp := files.NewFileOp()
		appVersionDir := path.Join(global.Dir.AppResourceDir, app.Resource, app.Key, appDetail.Version)
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
		runtime.ContainerName = containerName.(string)
	}

	projectDir := path.Join(global.Dir.RuntimeDir, runtime.Type, runtime.Name)
	create := request.RuntimeCreate{
		Image:   req.Image,
		Type:    runtime.Type,
		Source:  req.Source,
		Params:  req.Params,
		CodeDir: req.CodeDir,
		Version: req.Version,
		Remark:  req.Remark,
		NodeConfig: request.NodeConfig{
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
	runtime.Remark = req.Remark
	runtime.Env = string(envContent)
	runtime.DockerCompose = string(composeContent)

	switch runtime.Type {
	case constant.RuntimePHP:
		runtime.Image = req.Image
		runtime.Status = constant.StatusBuilding
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
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		runtime.Version = req.Version
		runtime.CodeDir = req.CodeDir
		runtime.Port = strings.Join(hostPorts, ",")
		runtime.Status = constant.StatusReCreating
		runtime.ContainerName = req.Params["CONTAINER_NAME"].(string)
		_ = runtimeRepo.Save(runtime)
		go reCreateRuntime(runtime)
	}
	return nil
}

func (r *RuntimeService) GetNodePackageRunScript(req request.NodePackageReq) ([]response.PackageScripts, error) {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.CodeDir) {
		return nil, buserr.New("ErrPathNotFound")
	}
	if !fileOp.Stat(path.Join(req.CodeDir, "package.json")) {
		return nil, buserr.New("ErrPackageJsonNotFound")
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
		return nil, buserr.New("ErrScriptsNotFound")
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			runtime.Status = constant.StatusError
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
		runtime.Status = constant.StatusStopped
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
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

	cmdMgr := cmd2.NewCommandMgr(cmd2.WithTimeout(5 * time.Minute))
	return cmdMgr.Run("docker", "exec", "-i", containerName, "bash", "-c", fmt.Sprintf("'%s'", cmd))
}

func (r *RuntimeService) SyncForRestart() error {
	runtimes, err := runtimeRepo.List()
	if err != nil {
		return err
	}
	for _, runtime := range runtimes {
		if runtime.Status == constant.StatusBuilding || runtime.Status == constant.StatusReCreating || runtime.Status == constant.StatusStarting || runtime.Status == constant.StatusCreating {
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
		if runtime.Type == constant.RuntimeNode || runtime.Type == constant.RuntimeJava || runtime.Type == constant.RuntimeGo || runtime.Type == constant.RuntimePython || runtime.Type == constant.RuntimeDotNet {
			_ = SyncRuntimeContainerStatus(&runtime)
		}
	}
	return nil
}

func (r *RuntimeService) GetPHPExtensions(runtimeID uint) (response.PHPExtensionRes, error) {
	var res response.PHPExtensionRes
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(runtimeID))
	if err != nil {
		return res, err
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Second))
	out, err := cmdMgr.RunWithStdoutBashCf("docker exec -i %s php -m", runtime.ContainerName)
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
	if err = json.Unmarshal(nginx_conf.GetWebsiteFile("php_extensions.json"), &phpExtensions); err != nil {
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if task.CheckResourceTaskIsExecuting(task.TaskInstall, task.TaskScopeRuntimeExtension, runtime.ID) {
		return buserr.New("ErrInstallExtension")
	}
	installTask, err := task.NewTaskWithOps(req.Name, task.TaskInstall, task.TaskScopeRuntimeExtension, req.TaskID, runtime.ID)
	if err != nil {
		return err
	}
	cmdMgr := cmd2.NewCommandMgr(cmd.WithTask(*installTask), cmd.WithTimeout(15*time.Minute))
	installTask.AddSubTask("", func(t *task.Task) error {
		err = cmdMgr.RunBashCf("docker exec -i %s install-ext %s", runtime.ContainerName, req.Name)
		if err != nil {
			return err
		}
		client, err := docker.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()
		oldImageID, err := client.GetImageIDByName(runtime.Image)
		if err != nil {
			return err
		}
		err = cmdMgr.RunBashCf("docker commit %s %s", runtime.ContainerName, runtime.Image)
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
		handlePHPDir(*runtime)
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
				extensions = strings.TrimPrefix(extensions, ",")
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(id))
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
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
	if err := fileOp.WriteFile(phpConfigPath, strings.NewReader(updatedContent), constant.DirPerm); err != nil {
		return err
	}

	err = restartRuntime(runtime)
	if err != nil {
		_ = fileOp.WriteFile(phpConfigPath, strings.NewReader(string(contentBytes)), constant.DirPerm)
		return err
	}
	return
}

func (r *RuntimeService) GetPHPConfigFile(req request.PHPFileReq) (*response.FileInfo, error) {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	configPath := ""
	if req.Type == constant.ConfigFPM {
		configPath = path.Join(runtime.GetPath(), "conf", "php-fpm.conf")
	} else {
		configPath = path.Join(runtime.GetPath(), "conf", "php.ini")
	}
	if err := files.NewFileOp().WriteFile(configPath, strings.NewReader(req.Content), constant.DirPerm); err != nil {
		return err
	}
	if _, err := compose.Restart(runtime.GetComposePath()); err != nil {
		return err
	}
	return nil
}

func (r *RuntimeService) UpdateFPMConfig(req request.FPMConfig) error {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
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
	if _, err := compose.Restart(runtime.GetComposePath()); err != nil {
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
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(id))
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

func (r *RuntimeService) UpdatePHPContainer(req request.PHPContainerConfig) error {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	var (
		composeContent []byte
	)
	for _, export := range req.ExposedPorts {
		if strconv.Itoa(export.HostPort) == runtime.Port {
			return buserr.WithName("ErrPHPRuntimePortFailed", strconv.Itoa(export.HostPort))
		}
		if export.ContainerPort == 9000 {
			return buserr.New("ErrPHPPortIsDefault")
		}
		if err = checkRuntimePortExist(export.HostPort, false, runtime.ID); err != nil {
			return err
		}
	}
	if req.ContainerName != "" && req.ContainerName != getRuntimeEnv(runtime.Env, "CONTAINER_NAME") {
		if err := checkContainerName(req.ContainerName); err != nil {
			return err
		}
		runtime.ContainerName = req.ContainerName
	}
	fileOp := files.NewFileOp()
	projectDir := path.Join(global.Dir.RuntimeDir, runtime.Type, runtime.Name)
	composeContent, err = fileOp.GetContent(path.Join(projectDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	envPath := path.Join(projectDir, ".env")
	if !fileOp.Stat(envPath) {
		_ = fileOp.CreateFile(envPath)
	}
	envs, err := gotenv.Read(envPath)
	if err != nil {
		return err
	}
	for k := range envs {
		if strings.HasPrefix(k, "CONTAINER_PORT_") || strings.HasPrefix(k, "HOST_PORT_") || strings.HasPrefix(k, "HOST_IP_") || strings.Contains(k, "APP_PORT") {
			delete(envs, k)
		}
	}
	create := request.RuntimeCreate{
		Image:  runtime.Image,
		Type:   runtime.Type,
		Params: make(map[string]interface{}),
		NodeConfig: request.NodeConfig{
			ExposedPorts: req.ExposedPorts,
			Environments: req.Environments,
			Volumes:      req.Volumes,
		},
	}
	composeContent, err = handleCompose(envs, composeContent, create, projectDir)
	if err != nil {
		return err
	}
	newMap := make(map[string]string)
	handleMap(create.Params, newMap)
	maps.Copy(envs, newMap)
	envs["PANEL_APP_PORT_HTTP"] = runtime.Port
	envStr, err := gotenv.Marshal(envs)
	if err != nil {
		return err
	}
	if err = gotenv.Write(envs, envPath); err != nil {
		return err
	}
	envContent := []byte(envStr)
	runtime.Env = string(envContent)
	runtime.DockerCompose = string(composeContent)
	runtime.Status = constant.StatusReCreating
	_ = runtimeRepo.Save(runtime)
	go reCreateRuntime(runtime)
	return nil
}

func (r *RuntimeService) GetPHPContainerConfig(id uint) (*request.PHPContainerConfig, error) {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	runtimeDTO := response.NewRuntimeDTO(*runtime)
	if err := handleRuntimeDTO(&runtimeDTO, *runtime); err != nil {
		return nil, err
	}
	res := &request.PHPContainerConfig{
		ID:            runtime.ID,
		ContainerName: runtime.ContainerName,
		ExposedPorts:  runtimeDTO.ExposedPorts,
		Environments:  runtimeDTO.Environments,
		Volumes:       runtimeDTO.Volumes,
	}
	return res, nil
}

func (r *RuntimeService) GetSupervisorProcess(id uint) ([]response.SupervisorProcessConfig, error) {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	configDir := path.Join(global.Dir.RuntimeDir, "php", runtime.Name, "supervisor", "supervisor.d")
	return handleProcessConfig(configDir, runtime.ContainerName)
}

func (r *RuntimeService) OperateSupervisorProcess(req request.PHPSupervisorProcessConfig) error {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	configDir := path.Join(global.Dir.RuntimeDir, "php", runtime.Name, "supervisor")
	return handleProcess(configDir, req.SupervisorProcessConfig, runtime.ContainerName)
}

func (r *RuntimeService) OperateSupervisorProcessFile(req request.PHPSupervisorProcessFileReq) (string, error) {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return "", err
	}
	supervisorDir := path.Join(global.Dir.RuntimeDir, "php", runtime.Name, "supervisor")
	configDir := path.Join(supervisorDir, "supervisor.d")
	logFile := path.Join(supervisorDir, "log", fmt.Sprintf("%s.out.log", req.SupervisorProcessFileReq.Name))
	return handleSupervisorFile(req.SupervisorProcessFileReq, configDir, runtime.ContainerName, logFile)
}

func (r *RuntimeService) UpdateRemark(req request.RuntimeRemark) error {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	runtime.Remark = req.Remark
	return runtimeRepo.Save(runtime)
}

func (r *RuntimeService) GetFPMStatus(runtimeID uint) ([]response.FpmStatusItem, error) {
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(runtimeID))
	if err != nil {
		return nil, err
	}
	fcgiClient, err := fcgiclient.DialTimeout("tcp", "127.0.0.1:"+runtime.Port, 10*time.Second)
	if err != nil {
		return nil, errors.New("<UNK> FastCGI <UNK>: " + err.Error())
	}
	defer fcgiClient.Close()

	reqEnv := map[string]string{
		"REQUEST_METHOD":    "GET",
		"REQUEST_URI":       "/status",
		"SCRIPT_FILENAME":   "/status",
		"SCRIPT_NAME":       "/status",
		"QUERY_STRING":      "",
		"CONTENT_TYPE":      "",
		"CONTENT_LENGTH":    "0",
		"SERVER_SOFTWARE":   "go-fcgi-client",
		"SERVER_NAME":       "localhost",
		"SERVER_PORT":       runtime.Port,
		"REMOTE_ADDR":       "127.0.0.1",
		"GATEWAY_INTERFACE": "CGI/1.1",
	}

	resp, err := fcgiClient.Get(reqEnv)
	if err != nil {
		return nil, errors.New("<UNK> FastCGI <UNK>: " + err.Error())
	}
	defer resp.Body.Close()

	var status []response.FpmStatusItem
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		status = append(status, response.FpmStatusItem{
			Key:   key,
			Value: value,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("<UNK> FastCGI <UNK>: %v", err))
	}
	return status, nil
}
