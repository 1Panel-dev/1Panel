package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type AIToolService struct{}

type IAIToolService interface {
	Search(search dto.SearchWithPage) (int64, []dto.OllamaModelInfo, error)
	Create(name string) error
	Delete(name string) error
	LoadDetail(name string) (string, error)
}

func NewIAIToolService() IAIToolService {
	return &AIToolService{}
}

func (u *AIToolService) Search(req dto.SearchWithPage) (int64, []dto.OllamaModelInfo, error) {
	ollamaBaseInfo, err := appInstallRepo.LoadBaseInfo("ollama", "")
	if err != nil {
		return 0, nil, err
	}
	if ollamaBaseInfo.Status != constant.Running {
		return 0, nil, nil
	}
	stdout, err := cmd.Execf("docker exec %s ollama list", ollamaBaseInfo.ContainerName)
	if err != nil {
		return 0, nil, err
	}
	var list []dto.OllamaModelInfo
	modelMaps := make(map[string]struct{})
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		if parts[0] == "NAME" {
			continue
		}
		modelMaps[parts[0]] = struct{}{}
		list = append(list, dto.OllamaModelInfo{Name: parts[0], Size: parts[2] + " " + parts[3], Modified: strings.Join(parts[4:], " ")})
	}
	entries, _ := os.ReadDir(path.Join(global.CONF.System.DataDir, "log", "AITools"))
	for _, item := range entries {
		if _, ok := modelMaps[item.Name()]; ok {
			continue
		}
		if _, ok := modelMaps[item.Name()+":latest"]; ok {
			continue
		}
		list = append(list, dto.OllamaModelInfo{Name: item.Name(), Size: "-", Modified: "-"})
	}
	if len(req.Info) != 0 {
		length, count := len(list), 0
		for count < length {
			if !strings.Contains(list[count].Name, req.Info) {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}

	var records []dto.OllamaModelInfo
	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]dto.OllamaModelInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list[start:end]
	}
	return int64(total), records, err
}

func (u *AIToolService) LoadDetail(name string) (string, error) {
	if cmd.CheckIllegal(name) {
		return "", buserr.New(constant.ErrCmdIllegal)
	}
	ollamaBaseInfo, err := appInstallRepo.LoadBaseInfo("ollama", "")
	if err != nil {
		return "", err
	}
	if ollamaBaseInfo.Status != constant.Running {
		return "", nil
	}
	stdout, err := cmd.Execf("docker exec %s ollama show %s", ollamaBaseInfo.ContainerName, name)
	if err != nil {
		return "", err
	}
	return stdout, err
}

func (u *AIToolService) Create(name string) error {
	if cmd.CheckIllegal(name) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	ollamaBaseInfo, err := appInstallRepo.LoadBaseInfo("ollama", "")
	if err != nil {
		return err
	}
	if ollamaBaseInfo.Status != constant.Running {
		return nil
	}
	fileName := strings.ReplaceAll(name, ":", "-")
	logItem := path.Join(global.CONF.System.DataDir, "log", "AITools", fileName)
	if _, err := os.Stat(path.Dir(logItem)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(logItem), os.ModePerm); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	go func() {
		defer file.Close()
		cmd := exec.Command("docker", "exec", ollamaBaseInfo.ContainerName, "ollama", "run", name)
		multiWriter := io.MultiWriter(os.Stdout, file)
		cmd.Stdout = multiWriter
		cmd.Stderr = multiWriter
		if err := cmd.Run(); err != nil {
			global.LOG.Errorf("ollama pull %s failed, err: %v", name, err)
			_, _ = file.WriteString("ollama pull failed!")
			return
		}
		global.LOG.Infof("ollama pull %s successful!", name)
		_, _ = file.WriteString("ollama pull successful!")
	}()

	return nil
}

func (u *AIToolService) Delete(name string) error {
	if cmd.CheckIllegal(name) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	ollamaBaseInfo, err := appInstallRepo.LoadBaseInfo("ollama", "")
	if err != nil {
		return err
	}
	if ollamaBaseInfo.Status != constant.Running {
		return nil
	}
	stdout, err := cmd.Execf("docker exec %s ollama list", ollamaBaseInfo.ContainerName)
	if err != nil {
		return err
	}
	isExist := false
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		if parts[0] == "NAME" {
			continue
		}
		if parts[0] == name {
			isExist = true
			break
		}
	}

	if isExist {
		stdout, err := cmd.Execf("docker exec %s ollama rm %s", ollamaBaseInfo.ContainerName, name)
		if err != nil {
			return fmt.Errorf("handle ollama rm %s failed, stdout: %s, err: %v", name, stdout, err)
		}
	}
	logItem := path.Join(global.CONF.System.DataDir, "log", "AITools", name)
	_ = os.Remove(logItem)
	logItem2 := path.Join(global.CONF.System.DataDir, "log", "AITools", strings.TrimSuffix(name, ":latest"))
	if logItem2 != logItem {
		_ = os.Remove(logItem2)
	}
	return nil
}
