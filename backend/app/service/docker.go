package service

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/pkg/errors"
)

type DockerService struct{}

type IDockerService interface {
	UpdateConf(req dto.DaemonJsonConf) error
	UpdateConfByFile(info dto.DaemonJsonUpdateByFile) error
	LoadDockerStatus() string
	LoadDockerConf() *dto.DaemonJsonConf
	OperateDocker(req dto.DockerOperation) error
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

type daemonJsonItem struct {
	Status      string   `json:"status"`
	Mirrors     []string `json:"registry-mirrors"`
	Registries  []string `json:"insecure-registries"`
	LiveRestore bool     `json:"live-restore"`
	ExecOpts    []string `json:"exec-opts"`
}

func (u *DockerService) LoadDockerStatus() string {
	status := constant.StatusRunning
	// cmd := exec.Command("systemctl", "is-active", "docker")
	// stdout, err := cmd.CombinedOutput()
	// if string(stdout) != "active\n" || err != nil {
	// 	status = constant.Stopped
	// }

	return status
}

func (u *DockerService) LoadDockerConf() *dto.DaemonJsonConf {
	status := constant.StatusRunning
	cmd := exec.Command("systemctl", "is-active", "docker")
	stdout, err := cmd.CombinedOutput()
	if string(stdout) != "active\n" || err != nil {
		status = constant.Stopped
	}
	fileSetting, err := settingRepo.Get(settingRepo.WithByKey("DaemonJsonPath"))
	if err != nil {
		return &dto.DaemonJsonConf{Status: status}
	}
	if len(fileSetting.Value) == 0 {
		return &dto.DaemonJsonConf{Status: status}
	}
	if _, err := os.Stat(fileSetting.Value); err != nil {
		return &dto.DaemonJsonConf{Status: status}
	}
	file, err := ioutil.ReadFile(fileSetting.Value)
	if err != nil {
		return &dto.DaemonJsonConf{Status: status}
	}
	var conf daemonJsonItem
	deamonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		return &dto.DaemonJsonConf{Status: status}
	}
	arr, err := json.Marshal(deamonMap)
	if err != nil {
		return &dto.DaemonJsonConf{Status: status}
	}
	if err := json.Unmarshal(arr, &conf); err != nil {
		return &dto.DaemonJsonConf{Status: status}
	}
	driver := "cgroupfs"
	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			driver = strings.ReplaceAll(opt, "native.cgroupdriver=", "")
			break
		}
	}
	data := dto.DaemonJsonConf{
		Status:       status,
		Mirrors:      conf.Mirrors,
		Registries:   conf.Registries,
		LiveRestore:  conf.LiveRestore,
		CgroupDriver: driver,
	}

	return &data
}

func (u *DockerService) UpdateConf(req dto.DaemonJsonConf) error {
	fileSetting, err := settingRepo.Get(settingRepo.WithByKey("DaemonJsonPath"))
	if err != nil {
		return err
	}
	if len(fileSetting.Value) == 0 {
		return errors.New("error daemon.json path in request")
	}
	if _, err := os.Stat(fileSetting.Value); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(fileSetting.Value, os.ModePerm); err != nil {
			if err != nil {
				return err
			}
		}
	}

	file, err := ioutil.ReadFile(fileSetting.Value)
	if err != nil {
		return err
	}
	deamonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &deamonMap)

	if len(req.Registries) == 0 {
		delete(deamonMap, "insecure-registries")
	} else {
		deamonMap["insecure-registries"] = req.Registries
	}
	if len(req.Mirrors) == 0 {
		delete(deamonMap, "registry-mirrors")
	} else {
		deamonMap["registry-mirrors"] = req.Mirrors
	}
	if !req.LiveRestore {
		delete(deamonMap, "live-restore")
	} else {
		deamonMap["live-restore"] = req.LiveRestore
	}
	if opts, ok := deamonMap["exec-opts"]; ok {
		if optsValue, isArray := opts.([]interface{}); isArray {
			for i := 0; i < len(optsValue); i++ {
				if opt, isStr := optsValue[i].(string); isStr {
					if strings.HasPrefix(opt, "native.cgroupdriver=") {
						optsValue[i] = "native.cgroupdriver=" + req.CgroupDriver
						break
					}
				}
			}
		}
	} else {
		if req.CgroupDriver == "systemd" {
			deamonMap["exec-opts"] = []string{"native.cgroupdriver=systemd"}
		}
	}
	newJson, err := json.MarshalIndent(deamonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(fileSetting.Value, newJson, 0640); err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "restart", "docker")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *DockerService) UpdateConfByFile(req dto.DaemonJsonUpdateByFile) error {
	file, err := os.OpenFile(req.Path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	cmd := exec.Command("systemctl", "restart", "docker")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *DockerService) OperateDocker(req dto.DockerOperation) error {
	cmd := exec.Command("systemctl", req.Operation, "docker")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}
