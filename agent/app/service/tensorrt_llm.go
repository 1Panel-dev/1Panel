package service

import (
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
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
	"path"
	"strconv"
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
		env, _ := gotenv.Unmarshal(item.Env)
		serverDTO.Version = env["VERSION"]
		serverDTO.Model = env["MODEL_NAME"]
		serverDTO.ModelDir = env["MODEL_PATH"]
		serverDTO.Dir = path.Join(global.Dir.TensorRTLLMDir, item.Name)
		items = append(items, serverDTO)
	}
	res.Total = total
	res.Items = items
	return res
}

func handleLLMParams(llm *model.TensorRTLLM) error {
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

	services[llm.Name] = serviceValue
	composeByte, err := yaml.Marshal(composeMap)
	if err != nil {
		return err
	}
	llm.DockerCompose = string(composeByte)
	return nil
}

func handleLLMEnv(llm *model.TensorRTLLM, create request.TensorRTLLMCreate) gotenv.Env {
	env := make(gotenv.Env)
	env["CONTAINER_NAME"] = create.ContainerName
	env["PANEL_APP_PORT_HTTP"] = strconv.Itoa(llm.Port)
	env["MODEL_PATH"] = create.ModelDir
	env["MODEL_NAME"] = create.Model
	env["VERSION"] = create.Version
	if create.HostIP != "" {
		env["HOST_IP"] = create.HostIP
	} else {
		env["HOST_IP"] = ""
	}
	envStr, _ := gotenv.Marshal(env)
	llm.Env = envStr
	return env
}

func (t TensorRTLLMService) Create(create request.TensorRTLLMCreate) error {
	servers, _ := tensorrtLLMRepo.List()
	for _, server := range servers {
		if server.Port == create.Port {
			return buserr.New("ErrPortInUsed")
		}
		if server.ContainerName == create.ContainerName {
			return buserr.New("ErrContainerName")
		}
		if server.Name == create.Name {
			return buserr.New("ErrNameIsExist")
		}
	}
	if err := checkPortExist(create.Port); err != nil {
		return err
	}
	if err := checkContainerName(create.ContainerName); err != nil {
		return err
	}

	tensorrtLLMDir := path.Join(global.Dir.TensorRTLLMDir, create.Name)
	filesOP := files.NewFileOp()
	if !filesOP.Stat(tensorrtLLMDir) {
		_ = filesOP.CreateDir(tensorrtLLMDir, 0644)
	}
	tensorrtLLM := &model.TensorRTLLM{
		Name:          create.Name,
		ContainerName: create.ContainerName,
		Port:          create.Port,
		Status:        constant.StatusStarting,
	}

	if err := handleLLMParams(tensorrtLLM); err != nil {
		return err
	}
	env := handleLLMEnv(tensorrtLLM, create)
	llmDir := path.Join(global.Dir.TensorRTLLMDir, create.Name)
	envPath := path.Join(llmDir, ".env")
	if err := gotenv.Write(env, envPath); err != nil {
		return err
	}
	dockerComposePath := path.Join(llmDir, "docker-compose.yml")
	if err := filesOP.SaveFile(dockerComposePath, tensorrtLLM.DockerCompose, 0644); err != nil {
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
	if tensorrtLLM.Port != req.Port {
		if err := checkPortExist(req.Port); err != nil {
			return err
		}
	}
	if tensorrtLLM.ContainerName != req.ContainerName {
		if err := checkContainerName(req.ContainerName); err != nil {
			return err
		}
	}

	tensorrtLLM.ContainerName = req.ContainerName
	tensorrtLLM.Port = req.Port
	if err := handleLLMParams(tensorrtLLM); err != nil {
		return err
	}

	newEnv, err := gotenv.Unmarshal(tensorrtLLM.Env)
	if err != nil {
		return err
	}
	newEnv["CONTAINER_NAME"] = req.ContainerName
	newEnv["PANEL_APP_PORT_HTTP"] = strconv.Itoa(tensorrtLLM.Port)
	newEnv["MODEL_PATH"] = req.ModelDir
	newEnv["MODEL_NAME"] = req.Model
	newEnv["VERSION"] = req.Version
	if req.HostIP != "" {
		newEnv["HOST_IP"] = req.HostIP
	} else {
		newEnv["HOST_IP"] = ""
	}
	envStr, _ := gotenv.Marshal(newEnv)
	tensorrtLLM.Env = envStr
	llmDir := path.Join(global.Dir.TensorRTLLMDir, tensorrtLLM.Name)
	envPath := path.Join(llmDir, ".env")
	if err := gotenv.Write(newEnv, envPath); err != nil {
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
