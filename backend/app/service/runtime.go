package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	cmd2 "github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/env"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
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
}

func NewRuntimeService() IRuntimeService {
	return &RuntimeService{}
}

func (r *RuntimeService) Create(create request.RuntimeCreate) (*model.Runtime, error) {
	var (
		opts []repo.DBOption
	)
	if create.Name != "" {
		opts = append(opts, commonRepo.WithByName(create.Name))
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
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
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
		if containerName, ok := create.Params["CONTAINER_NAME"]; ok {
			if err := checkContainerName(containerName.(string)); err != nil {
				return nil, err
			}
		}
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
	if !fileOp.Stat(appVersionDir) {
		if err = downloadApp(app, appDetail, nil); err != nil {
			return nil, err
		}
	}

	runtime := &model.Runtime{
		Name:        create.Name,
		AppDetailID: create.AppDetailID,
		Type:        create.Type,
		Image:       create.Image,
		Resource:    create.Resource,
		Version:     create.Version,
	}

	switch create.Type {
	case constant.RuntimePHP:
		if err = handlePHP(create, runtime, fileOp, appVersionDir); err != nil {
			return nil, err
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		runtime.Port = create.Port
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
		opts = append(opts, commonRepo.WithLikeName(req.Name))
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
	if runtime.Resource == constant.ResourceAppstore {
		projectDir := runtime.GetPath()
		switch runtime.Type {
		case constant.RuntimePHP:
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
		case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
			if out, err := compose.Down(runtime.GetComposePath()); err != nil && !runtimeDelete.ForceDelete {
				if out != "" {
					return errors.New(out)
				}
				return err
			}
		}
		if err := files.NewFileOp().DeleteDir(projectDir); err != nil && !runtimeDelete.ForceDelete {
			return err
		}
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
		res.Params = make(map[string]interface{})
		envs, err := gotenv.Unmarshal(runtime.Env)
		if err != nil {
			return nil, err
		}
		for k, v := range envs {
			switch k {
			case "NODE_APP_PORT", "PANEL_APP_PORT_HTTP", "JAVA_APP_PORT", "GO_APP_PORT", "APP_PORT":
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
	switch runtime.Type {
	case constant.RuntimePHP:
		exist, _ := runtimeRepo.GetFirst(runtimeRepo.WithImage(req.Name), runtimeRepo.WithNotId(req.ID))
		if exist != nil {
			return buserr.New(constant.ErrImageExist)
		}
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
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
		if containerName, ok := req.Params["CONTAINER_NAME"]; ok {
			envs, err := gotenv.Unmarshal(runtime.Env)
			if err != nil {
				return err
			}
			oldContainerName := envs["CONTAINER_NAME"]
			if containerName != oldContainerName {
				if err := checkContainerName(containerName.(string)); err != nil {
					return err
				}
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
			if err := downloadApp(app, appDetail, nil); err != nil {
				return err
			}
			_ = fileOp.Rename(path.Join(runtime.GetPath(), "run.sh"), path.Join(runtime.GetPath(), "run.sh.bak"))
			_ = fileOp.CopyFile(path.Join(appVersionDir, "run.sh"), runtime.GetPath())
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
		go buildRuntime(runtime, imageID, req.Rebuild)
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
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
	go func() {
		_ = docker.CreateDefaultDockerNetwork()
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
		if err = runComposeCmdWithLog(constant.RuntimeDown, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
			return err
		}
		if err = runComposeCmdWithLog(constant.RuntimeUp, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
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
		if runtime.Type != constant.RuntimePHP {
			_ = SyncRuntimeContainerStatus(&runtime)
		}
	}
	return nil
}
