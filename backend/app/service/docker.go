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
	LoadDockerConf() (*dto.DaemonJsonConf, error)
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

type daemonJsonItem struct {
	Status      string   `json:"status"`
	Mirrors     []string `json:"registry-mirrors"`
	Registries  []string `json:"insecure-registries"`
	Bip         string   `json:"bip"`
	LiveRestore bool     `json:"live-restore"`
	ExecOpts    []string `json:"exec-opts"`
}

func (u *DockerService) LoadDockerConf() (*dto.DaemonJsonConf, error) {
	file, err := ioutil.ReadFile(constant.DaemonJsonDir)
	if err != nil {
		return nil, err
	}
	var conf daemonJsonItem
	deamonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		return nil, err
	}
	arr, err := json.Marshal(deamonMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &conf); err != nil {
		return nil, err
	}
	driver := "cgroupfs"
	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			driver = strings.ReplaceAll(opt, "native.cgroupdriver=", "")
			break
		}
	}
	data := dto.DaemonJsonConf{
		Status:       conf.Status,
		Mirrors:      conf.Mirrors,
		Registries:   conf.Registries,
		Bip:          conf.Bip,
		LiveRestore:  conf.LiveRestore,
		CgroupDriver: driver,
	}

	return &data, nil
}

func (u *DockerService) UpdateConf(req dto.DaemonJsonConf) error {
	file, err := ioutil.ReadFile(constant.DaemonJsonDir)
	if err != nil {
		return err
	}
	deamonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		return err
	}
	if len(req.Registries) == 0 {
		delete(deamonMap, "insecure-registries")
	} else {
		deamonMap["insecure-registries"] = req.Registries
	}
	if len(req.Mirrors) == 0 {
		delete(deamonMap, "insecure-mirrors")
	} else {
		deamonMap["insecure-mirrors"] = req.Mirrors
	}
	if len(req.Bip) == 0 {
		delete(deamonMap, "bip")
	} else {
		deamonMap["bip"] = req.Bip
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
	if err := ioutil.WriteFile(constant.DaemonJsonDir, newJson, 0640); err != nil {
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
