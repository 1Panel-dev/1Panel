package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/ai"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/env"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

var tensorRTLLMMu sync.Mutex

type TensorRTLLMService struct{}

type ITensorRTLLMService interface {
	Page(req request.TensorRTLLMSearch) (response.TensorRTLLMsRes, error)
	Create(create request.TensorRTLLMCreate) error
	Update(req request.TensorRTLLMUpdate) error
	Delete(req request.TensorRTLLMDelete) error
	Operate(req request.TensorRTLLMOperate) error
}

func NewITensorRTLLMService() ITensorRTLLMService {
	return &TensorRTLLMService{}
}

func (t TensorRTLLMService) Page(req request.TensorRTLLMSearch) (response.TensorRTLLMsRes, error) {
	var (
		res   response.TensorRTLLMsRes
		items []response.TensorRTLLMDTO
	)

	total, data, err := tensorrtLLMRepo.Page(req.PageInfo.Page, req.PageInfo.PageSize)
	if err != nil {
		return res, err
	}
	for _, item := range data {
		operationTask, err := taskRepo.GetExecutingTask(task.TaskScopeTensorRT, item.ID)
		if err != nil {
			return res, err
		}
		if operationTask.ID != "" {
			item.Status = localEngineTaskStatus(operationTask.Operate, item.Status)
		} else if err := syncTensorRTLLMContainerStatus(&item); err != nil {
			return res, err
		}
		serverDTO := response.TensorRTLLMDTO{
			TensorRTLLM: item,
			TaskID:      operationTask.ID,
		}
		envs, _ := gotenv.Unmarshal(item.Env)
		serverDTO.Version = envs["VERSION"]
		serverDTO.ModelDir = envs["MODEL_PATH"]
		serverDTO.Dir = path.Join(global.Dir.TensorRTLLMDir, item.Name)
		serverDTO.Image = envs["IMAGE"]
		serverDTO.Command = getCommand(item.Env)
		var project map[string]interface{}
		if yaml.Unmarshal([]byte(item.DockerCompose), &project) == nil {
			services, _ := project["services"].(map[string]interface{})
			if service, ok := services[item.Name].(map[string]interface{}); ok {
				params := make(map[string]interface{}, len(envs))
				for key, value := range envs {
					params[key] = value
				}
				serverDTO.GPU = ReadModelGPU(service, params, "nvidia")
			}
		}

		serverDTO.ExposedPorts, _ = loadComposeExposedPortsFromEnv(envs, "0.0.0.0", false)

		composeByte, err := files.NewFileOp().GetContent(path.Join(global.Dir.TensorRTLLMDir, item.Name, "docker-compose.yml"))
		if err == nil {
			serverDTO.Environments, _ = getDockerComposeEnvironments(composeByte)
			environments := make([]request.Environment, 0, len(serverDTO.Environments))
			for _, item := range serverDTO.Environments {
				if item.Key == "MODEL_PATH" || (serverDTO.GPU != nil && (item.Key == "NVIDIA_VISIBLE_DEVICES" || item.Key == "NVIDIA_DRIVER_CAPABILITIES")) {
					continue
				}
				environments = append(environments, item)
			}
			serverDTO.Environments = environments
		}
		volumes, err := getDockerComposeVolumes(composeByte)
		if err == nil {
			var defaultVolumes = map[string]string{
				"${MODEL_PATH}": "${MODEL_PATH}",
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
					serverDTO.Volumes = append(serverDTO.Volumes, volume)
				}
			}
		}
		items = append(items, serverDTO)
	}
	res.Total = total
	res.Items = items
	return res, nil
}

func handleLLMParams(llm *model.TensorRTLLM, create request.TensorRTLLMCreate) error {
	var composeContent []byte
	if llm.ID == 0 {
		composeContent = ai.DefaultTensorrtLLMCompose
	} else {
		composeContent = []byte(llm.DockerCompose)
	}
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return err
	}
	services, serviceValid := composeMap["services"].(map[string]interface{})
	if !serviceValid {
		return buserr.New("ErrFileParse")
	}
	serviceName := ""
	serviceValue := make(map[string]interface{})

	if llm.ID > 0 {
		serviceName = llm.Name
		serviceValue = services[serviceName].(map[string]interface{})
	} else {
		for name, service := range services {
			serviceName = name
			serviceValue = service.(map[string]interface{})
			break
		}
		delete(services, serviceName)
	}

	delete(serviceValue, "ports")
	if len(create.ExposedPorts) > 0 {
		var ports []interface{}
		for i, port := range create.ExposedPorts {
			containerPortStr, hostPortStr, hostIPStr, _ := composePortEnvKeys(i)
			ports = append(ports, formatComposePortMapping(hostIPStr, hostPortStr, containerPortStr, port.Protocol))
		}
		serviceValue["ports"] = ports
	}

	delete(serviceValue, "environment")
	var environments []interface{}
	environments = append(environments, fmt.Sprintf("MODEL_PATH=%s", create.ModelDir))
	for _, e := range create.Environments {
		if e.Key == "MODEL_PATH" {
			continue
		}
		environments = append(environments, fmt.Sprintf("%s=%s", e.Key, e.Value))
	}
	serviceValue["environment"] = environments

	var volumes []interface{}
	var defaultVolumes = map[string]string{
		"${MODEL_PATH}": "${MODEL_PATH}",
	}
	for k, v := range defaultVolumes {
		volumes = append(volumes, fmt.Sprintf("%s:%s", k, v))
	}
	for _, volume := range create.Volumes {
		volumes = append(volumes, formatComposeVolume(volume.Source, volume.Target, volume.Mode))
	}
	serviceValue["volumes"] = volumes
	gpuParams := map[string]interface{}{}
	existingEnv, err := gotenv.Unmarshal(llm.Env)
	if err != nil {
		return err
	}
	for key, value := range existingEnv {
		gpuParams[key] = value
	}
	gpuParams["MODEL_PATH"] = create.ModelDir
	if err := ApplyModelGPU(serviceValue, create.GPU, gpuParams, "nvidia"); err != nil {
		return err
	}

	services[llm.Name] = serviceValue
	composeByte, err := yaml.Marshal(composeMap)
	if err != nil {
		return err
	}
	llm.DockerCompose = string(composeByte)
	return nil
}

func handleLLMEnv(llm *model.TensorRTLLM, create request.TensorRTLLMCreate) gotenv.Env {
	envMap := make(gotenv.Env)
	envMap["CONTAINER_NAME"] = create.ContainerName
	envMap["MODEL_PATH"] = create.ModelDir
	envMap["VERSION"] = create.Version
	envMap["IMAGE"] = create.Image
	envMap["COMMAND"] = create.Command
	for i, port := range create.ExposedPorts {
		containerPortStr, hostPortStr, hostIPStr, protocolStr := composePortEnvKeys(i)
		envMap[containerPortStr] = strconv.Itoa(port.ContainerPort)
		envMap[hostPortStr] = strconv.Itoa(port.HostPort)
		envMap[hostIPStr] = port.HostIP
		envMap[protocolStr] = normalizeComposeProtocol(port.Protocol)
	}
	orders := []string{"MODEL_PATH", "COMMAND"}
	envStr, _ := env.MarshalWithOrder(envMap, orders)
	llm.Env = envStr
	return envMap
}

func (t TensorRTLLMService) Create(create request.TensorRTLLMCreate) error {
	tensorRTLLMMu.Lock()
	defer tensorRTLLMMu.Unlock()
	servers, _ := tensorrtLLMRepo.List()
	for _, server := range servers {
		if server.ContainerName == create.ContainerName {
			return buserr.New("ErrContainerName")
		}
		if server.Name == create.Name {
			return buserr.New("ErrNameIsExist")
		}
	}
	for _, export := range create.ExposedPorts {
		if err := checkPortExistWithProtocol(export.HostPort, export.Protocol); err != nil {
			return err
		}
	}
	if err := checkContainerName(create.ContainerName); err != nil {
		return err
	}

	tensorrtLLMDir := path.Join(global.Dir.TensorRTLLMDir, create.Name)
	filesOp := files.NewFileOp()
	if !filesOp.Stat(tensorrtLLMDir) {
		_ = filesOp.CreateDir(tensorrtLLMDir, 0644)
	}
	tensorrtLLM := &model.TensorRTLLM{
		Name:          create.Name,
		ContainerName: create.ContainerName,
		Status:        constant.StatusStarting,
		ModelType:     create.ModelType,
		ModelSpeedup:  create.ModelSpeedup,
	}

	if err := handleLLMParams(tensorrtLLM, create); err != nil {
		return err
	}
	envMap := handleLLMEnv(tensorrtLLM, create)
	llmDir := path.Join(global.Dir.TensorRTLLMDir, create.Name)
	envPath := path.Join(llmDir, ".env")
	if err := env.WriteWithOrder(envMap, envPath, []string{"MODEL_PATH", "COMMAND"}); err != nil {
		return err
	}
	dockerComposePath := path.Join(llmDir, "docker-compose.yml")
	if err := filesOp.SaveFile(dockerComposePath, tensorrtLLM.DockerCompose, 0644); err != nil {
		return err
	}
	tensorrtLLM.Status = constant.StatusStarting

	if err := tensorrtLLMRepo.Create(tensorrtLLM); err != nil {
		return err
	}
	return runTensorRTLLMTask(tensorrtLLM, task.TaskCreate, "create", create.TaskID)
}

func (t TensorRTLLMService) Update(req request.TensorRTLLMUpdate) error {
	tensorRTLLMMu.Lock()
	defer tensorRTLLMMu.Unlock()
	tensorrtLLM, err := tensorrtLLMRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	operationTask, err := taskRepo.GetExecutingTask(task.TaskScopeTensorRT, tensorrtLLM.ID)
	if err != nil {
		return err
	}
	if operationTask.ID != "" {
		return buserr.New("TaskIsExecuting")
	}
	if tensorrtLLM.ContainerName != req.ContainerName {
		if err := checkContainerName(req.ContainerName); err != nil {
			return err
		}
	}
	tensorrtLLM.ModelType = req.ModelType
	tensorrtLLM.ModelSpeedup = req.ModelSpeedup
	tensorrtLLM.ContainerName = req.ContainerName

	if err := handleLLMParams(tensorrtLLM, req.TensorRTLLMCreate); err != nil {
		return err
	}

	envMap := handleLLMEnv(tensorrtLLM, req.TensorRTLLMCreate)
	envStr, _ := gotenv.Marshal(envMap)
	tensorrtLLM.Env = envStr
	llmDir := path.Join(global.Dir.TensorRTLLMDir, tensorrtLLM.Name)
	envPath := path.Join(llmDir, ".env")
	if err := env.WriteWithOrder(envMap, envPath, []string{"MODEL_PATH", "COMMAND"}); err != nil {
		return err
	}
	dockerComposePath := path.Join(llmDir, "docker-compose.yml")
	if err := files.NewFileOp().SaveFile(dockerComposePath, tensorrtLLM.DockerCompose, 0644); err != nil {
		return err
	}
	tensorrtLLM.Status = constant.StatusStarting
	if err := tensorrtLLMRepo.Save(tensorrtLLM); err != nil {
		return err
	}
	return runTensorRTLLMTask(tensorrtLLM, task.TaskUpdate, "update", req.TaskID)
}

func (t TensorRTLLMService) Delete(req request.TensorRTLLMDelete) error {
	tensorRTLLMMu.Lock()
	defer tensorRTLLMMu.Unlock()
	tensorrtLLM, err := tensorrtLLMRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	return runTensorRTLLMTask(tensorrtLLM, task.TaskDelete, "delete", req.TaskID)
}

func (t TensorRTLLMService) Operate(req request.TensorRTLLMOperate) error {
	tensorRTLLMMu.Lock()
	defer tensorRTLLMMu.Unlock()
	tensorrtLLM, err := tensorrtLLMRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	var operation string
	switch req.Operate {
	case "start":
		operation = "Run"
	case "stop":
		operation = "Stop"
	case "restart":
		operation = task.TaskRestart
	default:
		return fmt.Errorf("unsupported TensorRT operation: %s", req.Operate)
	}
	return runTensorRTLLMTask(tensorrtLLM, operation, req.Operate, req.TaskID)
}

func runTensorRTLLMTask(llm *model.TensorRTLLM, operation, action, taskID string) error {
	currentTask, err := taskRepo.GetExecutingTask(task.TaskScopeTensorRT, llm.ID)
	if err != nil {
		return err
	}
	if currentTask.ID != "" {
		return buserr.New("TaskIsExecuting")
	}
	operationTask, err := task.NewTaskWithOps(llm.Name, operation, task.TaskScopeTensorRT, taskID, llm.ID)
	if err != nil {
		return err
	}
	operationTask.AddSubTaskWithOps(operationTask.Name, func(t *task.Task) error {
		if action == "stop" {
			llm.Status = constant.StatusWaiting
		} else {
			llm.Status = localEngineTaskStatus(operation, constant.StatusStarting)
		}
		if err := tensorrtLLMRepo.Save(llm); err != nil {
			return err
		}
		composePath := path.Join(global.Dir.TensorRTLLMDir, llm.Name, "docker-compose.yml")
		var out string
		var err error
		switch action {
		case "create":
			if llm.ModelSpeedup {
				envs, err := gotenv.Unmarshal(llm.Env)
				if err != nil {
					return err
				}
				if err := handleModelArchive(llm.ModelType, envs["MODEL_PATH"]); err != nil {
					return err
				}
			}
			out, err = compose.Up(composePath)
		case "start":
			out, err = compose.Up(composePath)
		case "update":
			out, err = compose.Down(composePath)
			if strings.TrimSpace(out) != "" {
				t.Log(out)
			}
			if err != nil {
				return err
			}
			out, err = compose.Up(composePath)
		case "stop":
			out, err = compose.Stop(composePath)
		case "restart":
			out, err = compose.Restart(composePath)
		case "delete":
			out, err = compose.Down(composePath)
		default:
			return fmt.Errorf("unsupported TensorRT operation: %s", action)
		}
		if strings.TrimSpace(out) != "" {
			t.Log(out)
		}
		if err != nil {
			return err
		}
		if action == "delete" {
			if err := files.NewFileOp().DeleteDir(path.Join(global.Dir.TensorRTLLMDir, llm.Name)); err != nil {
				return err
			}
			return tensorrtLLMRepo.DeleteBy(repo.WithByID(llm.ID))
		}
		if err := syncTensorRTLLMContainerStatus(llm); err != nil {
			return err
		}
		llm.Message = ""
		return tensorrtLLMRepo.Save(llm)
	}, nil, 0, 0)
	go func() {
		_ = operationTask.Execute()
	}()
	return nil
}

func syncTensorRTLLMContainerStatus(tensorrtLLM *model.TensorRTLLM) error {
	containerNames := []string{tensorrtLLM.ContainerName}
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	containers, err := cli.ListContainersByName(containerNames)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		tensorrtLLM.Status = constant.StatusError
		tensorrtLLM.Message = buserr.WithName("ErrContainerNotFound", tensorrtLLM.ContainerName).Error()
		return nil
	}
	container := containers[0]
	switch container.State {
	case "exited":
		tensorrtLLM.Status = constant.StatusStopped
	case "running":
		tensorrtLLM.Status = constant.StatusRunning
	case "paused":
		tensorrtLLM.Status = constant.StatusPaused
	case "restarting":
		tensorrtLLM.Status = constant.StatusRestarting
	default:
		tensorrtLLM.Status = constant.StatusUnHealthy
	}
	return nil
}

func findModelArchive(modelType string) (string, error) {
	const baseDir = "/home/models"
	prefix := fmt.Sprintf("FusionXplay_%s_Accelerator", modelType)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", baseDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".tar.gz") {
			return filepath.Join(baseDir, name), nil
		}
	}

	return "", fmt.Errorf("no FusionXplay_%s_Accelerator*.tar.gz found in /home/models", modelType)
}

func handleModelArchive(modelType string, modelDir string) error {
	filePath, err := findModelArchive(modelType)
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	if err = fileOp.TarGzExtractPro(filePath, modelDir, ""); err != nil {
		return err
	}
	if err = fileOp.ChmodR(path.Join(modelDir, "fusionxpark_accelerator"), 0755, false); err != nil {
		return err
	}
	return nil
}

func getCommand(envStr string) string {
	lines := strings.Split(envStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "COMMAND=") {
			return strings.TrimPrefix(line, "COMMAND=")
		}
	}
	return ""
}
