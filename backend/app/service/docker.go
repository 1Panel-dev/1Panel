package service

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
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
	IPTables    bool     `json:"iptables"`
	ExecOpts    []string `json:"exec-opts"`
}

func (u *DockerService) LoadDockerStatus() string {
	status := constant.StatusRunning
	stdout, err := cmd.Exec("systemctl is-active docker")
	if string(stdout) != "active\n" || err != nil {
		status = constant.Stopped
	}

	return status
}

func (u *DockerService) LoadDockerConf() *dto.DaemonJsonConf {
	var data dto.DaemonJsonConf
	data.IPTables = true
	data.Status = constant.StatusRunning
	stdout, err := cmd.Exec("systemctl is-active docker")
	if string(stdout) != "active\n" || err != nil {
		data.Status = constant.Stopped
	}
	data.IsSwarm = false
	stdout2, _ := cmd.Exec("docker info  | grep Swarm")
	if string(stdout2) == " Swarm: active\n" {
		data.IsSwarm = true
	}
	data.Version = "-"
	client, err := docker.NewDockerClient()
	if err == nil {
		ctx := context.Background()
		itemVersion, err := client.ServerVersion(ctx)
		if err == nil {
			data.Version = itemVersion.Version
		}
	}
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil {
		return &data
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return &data
	}
	var conf daemonJsonItem
	deamonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		return &data
	}
	arr, err := json.Marshal(deamonMap)
	if err != nil {
		return &data
	}
	if err := json.Unmarshal(arr, &conf); err != nil {
		return &data
	}
	if _, ok := deamonMap["iptables"]; !ok {
		conf.IPTables = true
	}
	data.CgroupDriver = "cgroupfs"
	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			data.CgroupDriver = strings.ReplaceAll(opt, "native.cgroupdriver=", "")
			break
		}
	}
	data.Mirrors = conf.Mirrors
	data.Registries = conf.Registries
	data.IPTables = conf.IPTables
	data.LiveRestore = conf.LiveRestore
	return &data
}

func (u *DockerService) UpdateConf(req dto.DaemonJsonConf) error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}

	file, err := os.ReadFile(constant.DaemonJsonPath)
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
	if req.IPTables {
		delete(deamonMap, "iptables")
	} else {
		deamonMap["iptables"] = false
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
	if len(deamonMap) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	newJson, err := json.MarshalIndent(deamonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	stdout, err := cmd.Exec("systemctl restart docker")
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *DockerService) UpdateConfByFile(req dto.DaemonJsonUpdateByFile) error {
	if len(req.File) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}
	file, err := os.OpenFile(constant.DaemonJsonPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	stdout, err := cmd.Exec("systemctl restart docker")
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func (u *DockerService) OperateDocker(req dto.DockerOperation) error {
	service := "docker"
	if req.Operation == "stop" {
		service = "docker.service"
		if req.StopSocket {
			service = "docker.socket"
		}
	}
	stdout, err := cmd.Execf("systemctl %s %s ", req.Operation, service)
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}
