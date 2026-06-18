package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	fcgiclient "github.com/tomasen/fcgi_client"

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
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/env"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
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

var pullRuntimeComposeImages = compose.PullComposeImages
var downRuntimeCompose = compose.Down

type runtimeTaskMeta struct {
	resourceName string
	operate      string
	scope        string
	taskID       string
	resourceID   uint
}

func buildRuntimeCreateTaskMeta(create request.RuntimeCreate, runtimeID uint) runtimeTaskMeta {
	return runtimeTaskMeta{
		resourceName: create.Name,
		operate:      task.TaskCreate,
		scope:        task.TaskScopeRuntime,
		taskID:       create.TaskID,
		resourceID:   runtimeID,
	}
}

func buildRuntimeDeleteTaskMeta(deleteReq request.RuntimeDelete, runtime *model.Runtime) runtimeTaskMeta {
	return runtimeTaskMeta{
		resourceName: runtime.Name,
		operate:      task.TaskDelete,
		scope:        task.TaskScopeRuntime,
		taskID:       deleteReq.TaskID,
		resourceID:   deleteReq.ID,
	}
}

func runtimeCreateInitialStatus(create request.RuntimeCreate) string {
	if create.Type == constant.RuntimePHP {
		if create.Resource == constant.ResourceLocal {
			return constant.StatusNormal
		}
		return constant.StatusBuilding
	}
	return constant.StatusCreating
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
	if create.Type == constant.RuntimePHP && create.Resource == constant.ResourceLocal {
		runtime := &model.Runtime{
			Name:     create.Name,
			Resource: create.Resource,
			Type:     create.Type,
			Version:  create.Version,
			Status:   runtimeCreateInitialStatus(create),
			Remark:   create.Remark,
		}
		if err := runtimeRepo.Create(context.Background(), runtime); err != nil {
			return nil, err
		}
		if err := startRuntimeCreateTask(create, runtime, model.App{}, model.AppDetail{}, ""); err != nil {
			runtime.Status = constant.StatusError
			runtime.Message = err.Error()
			_ = runtimeRepo.Save(runtime)
			return nil, err
		}
		return runtime, nil
	}
	var hostPorts []string
	switch create.Type {
	case constant.RuntimePHP:
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
			if err := checkPortExistWithProtocol(export.HostPort, export.Protocol); err != nil {
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
		Status:        runtimeCreateInitialStatus(create),
	}
	if err := runtimeRepo.Create(context.Background(), runtime); err != nil {
		return nil, err
	}
	if err := startRuntimeCreateTask(create, runtime, app, appDetail, appVersionDir); err != nil {
		runtime.Status = constant.StatusError
		runtime.Message = err.Error()
		_ = runtimeRepo.Save(runtime)
		return nil, err
	}
	return runtime, nil
}

func startRuntimeCreateTask(create request.RuntimeCreate, runtime *model.Runtime, app model.App, appDetail model.AppDetail, appVersionDir string) error {
	meta := buildRuntimeCreateTaskMeta(create, runtime.ID)
	createTask, err := task.NewTaskWithOps(meta.resourceName, meta.operate, meta.scope, meta.taskID, meta.resourceID)
	if err != nil {
		return err
	}
	createTask.AddSubTask(task.GetTaskName(create.Name, task.TaskCreate, task.TaskScopeRuntime), func(t *task.Task) error {
		return executeRuntimeCreateTask(create, runtime, app, appDetail, appVersionDir, t)
	}, nil)
	go func() {
		_ = createTask.Execute()
	}()
	return nil
}

func executeRuntimeCreateTask(create request.RuntimeCreate, runtime *model.Runtime, app model.App, appDetail model.AppDetail, appVersionDir string, taskItem *task.Task) error {
	if create.Type == constant.RuntimePHP && create.Resource == constant.ResourceLocal {
		runtime.Status = constant.StatusNormal
		runtime.Message = ""
		return runtimeRepo.Save(runtime)
	}

	if err := downloadApp(app, appDetail, nil, taskItem.Logger); err != nil {
		return markRuntimeCreateFailed(runtime, err)
	}
	go func() {
		RequestDownloadCallBack(appDetail.DownloadCallBackUrl)
	}()

	fileOp := files.NewFileOp()
	var err error
	switch create.Type {
	case constant.RuntimePHP:
		err = handlePHP(create, runtime, fileOp, appVersionDir)
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		err = handleRuntime(create, runtime, fileOp, appVersionDir)
	}
	if err != nil {
		return markRuntimeCreateFailed(runtime, err)
	}
	if err = runtimeRepo.Save(runtime); err != nil {
		return err
	}

	switch create.Type {
	case constant.RuntimePHP:
		return buildRuntimeWithResult(runtime, "", "", false)
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		if err := pullRuntimeImagesBeforeStart(create, runtime, taskItem); err != nil {
			return markRuntimeCreateFailed(runtime, err)
		}
		return startRuntimeWithResult(runtime)
	}
	return nil
}

func pullRuntimeImagesBeforeStart(create request.RuntimeCreate, runtime *model.Runtime, taskItem *task.Task) error {
	switch create.Type {
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
		return pullRuntimeComposeImages(runtime.GetComposePath(), false, taskItem)
	default:
		return nil
	}
}

func markRuntimeCreateFailed(runtime *model.Runtime, err error) error {
	runtime.Status = constant.StatusError
	runtime.Message = err.Error()
	_ = runtimeRepo.Save(runtime)
	return err
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
		detail, _ := appDetailRepo.GetFirst(repo.WithByID(runtime.AppDetailID))
		if detail.AppId == 0 {
			appID, appDetailID := handleRuntimeDetailID(runtime)
			runtimeDTO.AppDetailID = appDetailID
			runtimeDTO.AppID = appID
		} else {
			runtimeDTO.AppID = detail.AppId
		}
		for k, v := range envs {
			if !isComposePortEnvKey(k) {
				runtimeDTO.Params[k] = v
			}
		}
		runtimeDTO.ExposedPorts, _ = loadComposeExposedPortsFromEnv(envs, "", false)
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
	return startRuntimeDeleteTask(runtimeDelete, runtime)
}

func startRuntimeDeleteTask(runtimeDelete request.RuntimeDelete, runtime *model.Runtime) error {
	meta := buildRuntimeDeleteTaskMeta(runtimeDelete, runtime)
	deleteTask, err := task.NewTaskWithOps(meta.resourceName, meta.operate, meta.scope, meta.taskID, meta.resourceID)
	if err != nil {
		return err
	}
	deleteTask.AddSubTask(task.GetTaskName(runtime.Name, task.TaskDelete, task.TaskScopeRuntime), func(t *task.Task) error {
		return executeRuntimeDeleteTask(runtimeDelete, runtime, t)
	}, nil)
	go func() {
		_ = deleteTask.Execute()
	}()
	return nil
}

func executeRuntimeDeleteTask(runtimeDelete request.RuntimeDelete, runtime *model.Runtime, taskItem *task.Task) error {
	if runtime.Resource != constant.ResourceAppstore {
		if runtimeDelete.DeleteImage {
			deleteRuntimeImages(runtime, taskItem)
		}
		return runtimeRepo.DeleteBy(repo.WithByID(runtimeDelete.ID))
	}
	projectDir := runtime.GetPath()
	if err := stopRuntimeBeforeDelete(runtime, runtimeDelete.ForceDelete, taskItem); err != nil {
		return err
	}
	if runtimeDelete.DeleteImage {
		deleteRuntimeImages(runtime, taskItem)
	}
	if err := files.NewFileOp().DeleteDir(projectDir); err != nil && !runtimeDelete.ForceDelete {
		return err
	}
	return runtimeRepo.DeleteBy(repo.WithByID(runtimeDelete.ID))
}

func getRuntimeStopLog() string {
	logStr := i18n.GetMsgByKey("Stop") + i18n.GetMsgByKey("Runtime")
	if strings.TrimSpace(logStr) == "" {
		return "StopRuntime"
	}
	return logStr
}

func stopRuntimeBeforeDelete(runtime *model.Runtime, forceDelete bool, taskItem *task.Task) error {
	logStr := getRuntimeStopLog()
	if taskItem != nil {
		taskItem.Log(logStr)
	}
	out, err := downRuntimeCompose(runtime.GetComposePath())
	if err != nil && !forceDelete {
		if out != "" {
			err = errors.New(out)
		}
		if taskItem != nil {
			taskItem.LogFailedWithErr(logStr, err)
		}
		return err
	}
	if taskItem != nil {
		taskItem.LogSuccess(logStr)
	}
	return nil
}

func getRuntimeDeleteImages(runtime *model.Runtime) ([]string, error) {
	var images []string
	imageMap := make(map[string]struct{})
	appendImage := func(image string) {
		image = strings.TrimSpace(image)
		if image == "" {
			return
		}
		if _, ok := imageMap[image]; ok {
			return
		}
		imageMap[image] = struct{}{}
		images = append(images, image)
	}

	appendImage(runtime.Image)
	if runtime.DockerCompose != "" {
		composeImages, err := docker.GetImagesFromDockerCompose([]byte(runtime.Env), []byte(runtime.DockerCompose))
		if err != nil {
			return nil, err
		}
		for _, image := range composeImages {
			appendImage(image)
		}
	}
	return images, nil
}

func deleteRuntimeImages(runtime *model.Runtime, taskItem *task.Task) {
	logPrefix := i18n.GetMsgByKey("TaskDelete") + i18n.GetMsgByKey("Image")
	images, err := getRuntimeDeleteImages(runtime)
	if err != nil {
		global.LOG.Errorf("get runtime [%s] delete images error %v", runtime.Name, err)
		if taskItem != nil {
			taskItem.LogFailedWithErr(logPrefix, err)
		}
		return
	}
	if len(images) == 0 {
		if taskItem != nil {
			taskItem.LogFailedWithErr(logPrefix, errors.New(i18n.GetWithName("ErrImageNotExist", runtime.Name)))
		}
		return
	}
	client, err := docker.NewClient()
	if err != nil {
		global.LOG.Errorf("delete runtime images [%s] error %v", runtime.Name, err)
		if taskItem != nil {
			taskItem.LogFailedWithErr(logPrefix, err)
		}
		return
	}
	defer client.Close()

	for _, imageName := range images {
		logStr := logPrefix + imageName
		if taskItem != nil {
			taskItem.Log(logStr)
		}
		imageID, err := client.GetImageIDByName(imageName)
		if err != nil {
			global.LOG.Errorf("get runtime image [%s] error %v", imageName, err)
			if taskItem != nil {
				taskItem.LogFailedWithErr(logStr, err)
			}
			continue
		}
		if imageID == "" {
			if taskItem != nil {
				taskItem.LogFailedWithErr(logStr, errors.New(i18n.GetWithName("ErrImageNotExist", imageName)))
			}
			continue
		}
		if err := client.DeleteImage(imageID); err != nil {
			global.LOG.Errorf("delete image id [%s] error %v", imageID, err)
			if taskItem != nil {
				taskItem.LogFailedWithErr(logStr, err)
			}
			continue
		}
		if taskItem != nil {
			taskItem.LogSuccess(logStr)
		}
	}
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
							if strSlice, ok := appParam.Value.([]string); ok && len(strSlice) > 0 && strSlice[0] == "" {
								appParam.Value = strSlice[1:]
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
		ownedPortKeys := runtimeOwnedPortKeys(runtime.Env)
		for _, export := range req.ExposedPorts {
			hostPorts = append(hostPorts, strconv.Itoa(export.HostPort))
			_, owned := ownedPortKeys[composePortCheckKey(export.HostPort, export.Protocol)]
			if err = checkRuntimePortExistWithProtocol(export.HostPort, export.Protocol, !owned, runtime.ID); err != nil {
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
			ExtraHosts:   req.ExtraHosts,
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
		return res, nil
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
	operation := getOperation(req.Operate, req.PkgManager)
	execArgs := []string{"exec", "-i", containerName, req.PkgManager, operation}
	if strings.TrimSpace(req.Module) != "" {
		execArgs = append(execArgs, req.Module)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	installCmd := exec.CommandContext(ctx, "docker", execArgs...)
	output, err := installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %s, error: %w", string(output), err)
	}
	return nil
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
	out, err := cmdMgr.RunWithStdout("docker", "exec", "-i", runtime.ContainerName, "php", "-m")
	if err != nil {
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
	installTask.AddSubTask("", func(t *task.Task) error {
		err = cmd.NewCommandMgr(cmd.WithTask(*installTask), cmd.WithTimeout(20*time.Minute)).
			Run("docker", "exec", "-i", runtime.ContainerName, "install-ext", req.Name)
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
		err = cmd.NewCommandMgr(cmd.WithTask(*installTask), cmd.WithTimeout(15*time.Minute)).
			Run("docker", "commit", runtime.ContainerName, runtime.Image)
		if err != nil {
			return err
		}
		handlePHPDir(*runtime)
		if err = restartRuntime(runtime); err != nil {
			return err
		}
		newImageID, err := client.GetImageIDByName(runtime.Image)
		if err == nil && newImageID != oldImageID {
			if err := client.DeleteImage(oldImageID); err != nil {
				t.Log(fmt.Sprintf("delete old image %s failed %v", oldImageID, err))
			} else {
				t.Log("delete old image success")
			}
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
		matches := re.GetRegex(re.PhpAssignmentPattern).FindStringSubmatch(line)
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
	timeout := phpConfig.Key("max_execution_time").Value()
	if timeout != "" {
		res.MaxExecutionTime = timeout
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
				if phpConfigLineMatchesKey(line, key) {
					lines[i] = fmt.Sprintf("%s = %s", key, value)
				}
			}
		case "disable_functions":
			if phpConfigLineMatchesKey(line, "disable_functions") {
				lines[i] = fmt.Sprintf("disable_functions = %s", strings.Join(req.DisableFunctions, ","))
			}
		case "upload_max_filesize":
			if phpConfigLineMatchesKey(line, "post_max_size") {
				lines[i] = fmt.Sprintf("post_max_size = %s", req.UploadMaxSize)
			}
			if phpConfigLineMatchesKey(line, "upload_max_filesize") {
				lines[i] = fmt.Sprintf("upload_max_filesize = %s", req.UploadMaxSize)
			}
		case "max_execution_time":
			if phpConfigLineMatchesKey(line, "max_execution_time") {
				lines[i] = fmt.Sprintf("max_execution_time = %s", req.MaxExecutionTime)
			}
			if phpConfigLineMatchesKey(line, "max_input_time") {
				lines[i] = fmt.Sprintf("max_input_time = %s", req.MaxExecutionTime)
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

func phpConfigLineMatchesKey(line, key string) bool {
	trim := strings.TrimSpace(line)

	idx := strings.Index(trim, "=")
	if idx == -1 {
		return false
	}
	currentKey := strings.TrimSpace(trim[:idx])
	return currentKey == key
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
	ownedPortKeys := runtimeOwnedPortKeys(runtime.Env)
	for _, export := range req.ExposedPorts {
		if strconv.Itoa(export.HostPort) == runtime.Port {
			return buserr.WithName("ErrPHPRuntimePortFailed", strconv.Itoa(export.HostPort))
		}
		if export.ContainerPort == 9000 {
			return buserr.New("ErrPHPPortIsDefault")
		}
		_, owned := ownedPortKeys[composePortCheckKey(export.HostPort, export.Protocol)]
		if err = checkRuntimePortExistWithProtocol(export.HostPort, export.Protocol, !owned, runtime.ID); err != nil {
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
		if isComposePortEnvKey(k) || strings.Contains(k, "APP_PORT") {
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
			ExtraHosts:   req.ExtraHosts,
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
		ExtraHosts:    runtimeDTO.ExtraHosts,
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
