package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/ai"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/env"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

type TensorRTLLMService struct{}

type ITensorRTLLMService interface {
	Page(req request.TensorRTLLMSearch) response.TensorRTLLMsRes
	Create(create request.TensorRTLLMCreate) error
	Update(req request.TensorRTLLMUpdate) error
	Delete(id uint) error
	Operate(req request.TensorRTLLMOperate) error
}

func NewITensorRTLLMService() ITensorRTLLMService {
	return &TensorRTLLMService{}
}

func (t TensorRTLLMService) Page(req request.TensorRTLLMSearch) response.TensorRTLLMsRes {
	var (
		res   response.TensorRTLLMsRes
		items []response.TensorRTLLMDTO
	)

	total, data, _ := tensorrtLLMRepo.Page(req.PageInfo.Page, req.PageInfo.PageSize)
	for _, item := range data {
		_ = syncTensorRTLLMContainerStatus(&item)
		serverDTO := response.TensorRTLLMDTO{
			TensorRTLLM: item,
		}
		envs, _ := gotenv.Unmarshal(item.Env)
		serverDTO.Version = envs["VERSION"]
		serverDTO.ModelDir = envs["MODEL_PATH"]
		serverDTO.Dir = path.Join(global.Dir.TensorRTLLMDir, item.Name)
		serverDTO.Image = envs["IMAGE"]
		serverDTO.Command = getCommand(item.Env)

		for k, v := range envs {
			if strings.Contains(k, "CONTAINER_PORT") || strings.Contains(k, "HOST_PORT") {
				if strings.Contains(k, "CONTAINER_PORT") {
					matches := re.GetRegex(re.TrailingDigitsPattern).FindStringSubmatch(k)
					if len(matches) < 2 {
						continue
					}
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
					serverDTO.ExposedPorts = append(serverDTO.ExposedPorts, request.ExposedPort{
						ContainerPort: containerPort,
						HostPort:      hostPort,
						HostIP:        hostIP,
					})
				}
			}
		}

		composeByte, err := files.NewFileOp().GetContent(path.Join(global.Dir.TensorRTLLMDir, item.Name, "docker-compose.yml"))
		if err == nil {
			serverDTO.Environments, _ = getDockerComposeEnvironments(composeByte)
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
	return res
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
		for i := range create.ExposedPorts {
			containerPortStr := fmt.Sprintf("CONTAINER_PORT_%d", i)
			hostPortStr := fmt.Sprintf("HOST_PORT_%d", i)
			hostIPStr := fmt.Sprintf("HOST_IP_%d", i)
			ports = append(ports, fmt.Sprintf("${%s}:${%s}:${%s}", hostIPStr, hostPortStr, containerPortStr))
		}
		serviceValue["ports"] = ports
	}

	delete(serviceValue, "environment")
	var environments []interface{}
	environments = append(environments, fmt.Sprintf("MODEL_PATH=%s", create.ModelDir))
	for _, e := range create.Environments {
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
		volumes = append(volumes, fmt.Sprintf("%s:%s", volume.Source, volume.Target))
	}
	serviceValue["volumes"] = volumes

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
		containerPortStr := fmt.Sprintf("CONTAINER_PORT_%d", i)
		hostPortStr := fmt.Sprintf("HOST_PORT_%d", i)
		hostIPStr := fmt.Sprintf("HOST_IP_%d", i)
		envMap[containerPortStr] = strconv.Itoa(port.ContainerPort)
		envMap[hostPortStr] = strconv.Itoa(port.HostPort)
		envMap[hostIPStr] = port.HostIP
	}
	orders := []string{"MODEL_PATH", "COMMAND"}
	envStr, _ := env.MarshalWithOrder(envMap, orders)
	llm.Env = envStr
	return envMap
}

func (t TensorRTLLMService) Create(create request.TensorRTLLMCreate) error {
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
		if err := checkPortExist(export.HostPort); err != nil {
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
	if create.ModelSpeedup {
		if err := handleModelArchive(create.ModelType, create.ModelDir); err != nil {
			return err
		}
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
	go startTensorRTLLM(tensorrtLLM)
	return nil
}

func (t TensorRTLLMService) Update(req request.TensorRTLLMUpdate) error {
	tensorrtLLM, err := tensorrtLLMRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
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
	go startTensorRTLLM(tensorrtLLM)
	return nil
}

func (t TensorRTLLMService) Delete(id uint) error {
	tensorrtLLM, err := tensorrtLLMRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return err
	}
	composePath := path.Join(global.Dir.TensorRTLLMDir, tensorrtLLM.Name, "docker-compose.yml")
	_, _ = compose.Down(composePath)
	_ = files.NewFileOp().DeleteDir(path.Join(global.Dir.TensorRTLLMDir, tensorrtLLM.Name))
	return tensorrtLLMRepo.DeleteBy(repo.WithByID(id))
}

func (t TensorRTLLMService) Operate(req request.TensorRTLLMOperate) error {
	tensorrtLLM, err := tensorrtLLMRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	composePath := path.Join(global.Dir.TensorRTLLMDir, tensorrtLLM.Name, "docker-compose.yml")
	var out string
	switch req.Operate {
	case "start":
		out, err = compose.Up(composePath)
		tensorrtLLM.Status = constant.StatusRunning
	case "stop":
		out, err = compose.Down(composePath)
		tensorrtLLM.Status = constant.StatusStopped
	case "restart":
		out, err = compose.Restart(composePath)
		tensorrtLLM.Status = constant.StatusRunning
	}
	if err != nil {
		tensorrtLLM.Status = constant.StatusError
		tensorrtLLM.Message = out
	}
	return tensorrtLLMRepo.Save(tensorrtLLM)
}

func startTensorRTLLM(tensorrtLLM *model.TensorRTLLM) {
	composePath := path.Join(global.Dir.TensorRTLLMDir, tensorrtLLM.Name, "docker-compose.yml")
	if tensorrtLLM.Status != constant.StatusNormal {
		_, _ = compose.Down(composePath)
	}
	if out, err := compose.Up(composePath); err != nil {
		tensorrtLLM.Status = constant.StatusError
		tensorrtLLM.Message = out
	} else {
		tensorrtLLM.Status = constant.StatusRunning
		tensorrtLLM.Message = ""
	}
	_ = syncTensorRTLLMContainerStatus(tensorrtLLM)
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
		if tensorrtLLM.Status == constant.StatusStarting {
			return nil
		}
		tensorrtLLM.Status = constant.StatusStopped
		return tensorrtLLMRepo.Save(tensorrtLLM)
	}
	container := containers[0]
	switch container.State {
	case "exited":
		tensorrtLLM.Status = constant.StatusError
	case "running":
		tensorrtLLM.Status = constant.StatusRunning
	case "paused":
		tensorrtLLM.Status = constant.StatusStopped
	case "restarting":
		tensorrtLLM.Status = constant.StatusRestarting
	default:
		if tensorrtLLM.Status != constant.StatusStarting {
			tensorrtLLM.Status = constant.StatusStopped
		}
	}
	return tensorrtLLMRepo.Save(tensorrtLLM)
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
