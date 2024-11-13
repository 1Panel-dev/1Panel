package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
	"github.com/pkg/errors"
)

type DockerService struct{}

type IDockerService interface {
	UpdateConf(req dto.SettingUpdate) error
	UpdateLogOption(req dto.LogOption) error
	UpdateIpv6Option(req dto.Ipv6Option) error
	UpdateConfByFile(info dto.DaemonJsonUpdateByFile) error
	LoadDockerStatus() string
	LoadDockerConf() *dto.DaemonJsonConf
	OperateDocker(req dto.DockerOperation) error
}

func NewIDockerService() IDockerService {
	return &DockerService{}
}

type daemonJsonItem struct {
	Status       string    `json:"status"`
	Mirrors      []string  `json:"registry-mirrors"`
	Registries   []string  `json:"insecure-registries"`
	LiveRestore  bool      `json:"live-restore"`
	Ipv6         bool      `json:"ipv6"`
	FixedCidrV6  string    `json:"fixed-cidr-v6"`
	Ip6Tables    bool      `json:"ip6tables"`
	Experimental bool      `json:"experimental"`
	IPTables     bool      `json:"iptables"`
	ExecOpts     []string  `json:"exec-opts"`
	LogOption    logOption `json:"log-opts"`
}
type logOption struct {
	LogMaxSize string `json:"max-size"`
	LogMaxFile string `json:"max-file"`
}

func (u *DockerService) LoadDockerStatus() string {
	client, err := docker.NewDockerClient()
	if err != nil {
		return constant.Stopped
	}
	defer client.Close()
	if _, err := client.Ping(context.Background()); err != nil {
		return constant.Stopped
	}

	return constant.StatusRunning
}

func (u *DockerService) LoadDockerConf() *dto.DaemonJsonConf {
	ctx := context.Background()
	var data dto.DaemonJsonConf
	data.IPTables = true
	data.Status = constant.StatusRunning
	data.Version = "-"
	client, err := docker.NewDockerClient()
	if err != nil {
		data.Status = constant.Stopped
	} else {
		defer client.Close()
		if _, err := client.Ping(ctx); err != nil {
			data.Status = constant.Stopped
		}
		itemVersion, err := client.ServerVersion(ctx)
		if err == nil {
			data.Version = itemVersion.Version
		}
	}
	data.IsSwarm = false
	stdout2, _ := cmd.Exec("docker info  | grep Swarm")
	if string(stdout2) == " Swarm: active\n" {
		data.IsSwarm = true
	}
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil {
		return &data
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return &data
	}
	var conf daemonJsonItem
	daemonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &daemonMap); err != nil {
		return &data
	}
	arr, err := json.Marshal(daemonMap)
	if err != nil {
		return &data
	}
	if err := json.Unmarshal(arr, &conf); err != nil {
		return &data
	}
	if _, ok := daemonMap["iptables"]; !ok {
		conf.IPTables = true
	}
	data.CgroupDriver = "cgroupfs"
	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			data.CgroupDriver = strings.ReplaceAll(opt, "native.cgroupdriver=", "")
			break
		}
	}
	data.Ipv6 = conf.Ipv6
	data.FixedCidrV6 = conf.FixedCidrV6
	data.Ip6Tables = conf.Ip6Tables
	data.Experimental = conf.Experimental
	data.LogMaxSize = conf.LogOption.LogMaxSize
	data.LogMaxFile = conf.LogOption.LogMaxFile
	data.Mirrors = conf.Mirrors
	data.Registries = conf.Registries
	data.IPTables = conf.IPTables
	data.LiveRestore = conf.LiveRestore
	return &data
}

func (u *DockerService) UpdateConf(req dto.SettingUpdate) error {
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	switch req.Key {
	case "Registries":
		req.Value = strings.TrimSuffix(req.Value, ",")
		if len(req.Value) == 0 {
			delete(daemonMap, "insecure-registries")
		} else {
			daemonMap["insecure-registries"] = strings.Split(req.Value, ",")
		}
	case "Mirrors":
		req.Value = strings.TrimSuffix(req.Value, ",")
		if len(req.Value) == 0 {
			delete(daemonMap, "registry-mirrors")
		} else {
			daemonMap["registry-mirrors"] = strings.Split(req.Value, ",")
		}
	case "Ipv6":
		if req.Value == "disable" {
			delete(daemonMap, "ipv6")
			delete(daemonMap, "fixed-cidr-v6")
			delete(daemonMap, "ip6tables")
			delete(daemonMap, "experimental")
		}
	case "LogOption":
		if req.Value == "disable" {
			delete(daemonMap, "log-opts")
		}
	case "LiveRestore":
		if req.Value == "disable" {
			delete(daemonMap, "live-restore")
		} else {
			daemonMap["live-restore"] = true
		}
	case "IPtables":
		if req.Value == "enable" {
			delete(daemonMap, "iptables")
		} else {
			daemonMap["iptables"] = false
		}
	case "Driver":
		if opts, ok := daemonMap["exec-opts"]; ok {
			if optsValue, isArray := opts.([]interface{}); isArray {
				for i := 0; i < len(optsValue); i++ {
					if opt, isStr := optsValue[i].(string); isStr {
						if strings.HasPrefix(opt, "native.cgroupdriver=") {
							optsValue[i] = "native.cgroupdriver=" + req.Value
							break
						}
					}
				}
			}
		} else {
			if req.Value == "systemd" {
				daemonMap["exec-opts"] = []string{"native.cgroupdriver=systemd"}
			}
		}
	case "http-proxy", "https-proxy":
		delete(daemonMap, "proxies")
		if len(req.Value) > 0 {
			proxies := map[string]interface{}{
				req.Key: req.Value,
			}
			daemonMap["proxies"] = proxies
		}
	case "socks5-proxy", "close-proxy":
		delete(daemonMap, "proxies")
		if len(req.Value) > 0 {
			proxies := map[string]interface{}{
				"http-proxy":  req.Value,
				"https-proxy": req.Value,
			}
			daemonMap["proxies"] = proxies
		}
	}
	if len(daemonMap) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		if err := restartDocker(); err != nil {
			return err
		}
		return nil
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}
	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := restartDocker(); err != nil {
		return err
	}
	return nil
}
func createIfNotExistDaemonJsonFile() error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		var daemonFile *os.File
		daemonFile, err = os.Create(constant.DaemonJsonPath)
		if err != nil {
			return err
		}
		defer daemonFile.Close()
	}
	return nil
}

func (u *DockerService) UpdateLogOption(req dto.LogOption) error {
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	changeLogOption(daemonMap, req.LogMaxFile, req.LogMaxSize)
	if len(daemonMap) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := restartDocker(); err != nil {
		return err
	}
	return nil
}

func (u *DockerService) UpdateIpv6Option(req dto.Ipv6Option) error {
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}

	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	daemonMap := make(map[string]interface{})
	_ = json.Unmarshal(file, &daemonMap)

	daemonMap["ipv6"] = true
	daemonMap["fixed-cidr-v6"] = req.FixedCidrV6
	if req.Ip6Tables {
		daemonMap["ip6tables"] = req.Ip6Tables
	}
	if req.Experimental {
		daemonMap["experimental"] = req.Experimental
	}
	if len(daemonMap) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		return nil
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}

	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := restartDocker(); err != nil {
		return err
	}
	return nil
}

func (u *DockerService) UpdateConfByFile(req dto.DaemonJsonUpdateByFile) error {
	if len(req.File) == 0 {
		_ = os.Remove(constant.DaemonJsonPath)
		if err := restartDocker(); err != nil {
			return err
		}
		return nil
	}
	err := createIfNotExistDaemonJsonFile()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(constant.DaemonJsonPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	if err := validateDockerConfig(); err != nil {
		return err
	}

	if err := restartDocker(); err != nil {
		return err
	}
	return nil
}

func (u *DockerService) OperateDocker(req dto.DockerOperation) error {
	service := "docker"
	sudo := cmd.SudoHandleCmd()
	dockerCmd, err := getDockerRestartCommand()
	if err != nil {
		return err
	}
	if req.Operation == "stop" {
		isSocketActive, _ := systemctl.IsActive("docker.socket")
		if isSocketActive {
			std, err := cmd.Execf("%s systemctl stop docker.socket", sudo)
			if err != nil {
				global.LOG.Errorf("handle systemctl stop docker.socket failed, err: %v", std)
			}
		}
	}

	if req.Operation == "restart" {
		if err := validateDockerConfig(); err != nil {
			return err
		}
	}

	stdout, err := cmd.Execf("%s %s %s", dockerCmd, req.Operation, service)
	if err != nil {
		return errors.New(stdout)
	}
	return nil
}

func changeLogOption(daemonMap map[string]interface{}, logMaxFile, logMaxSize string) {
	if opts, ok := daemonMap["log-opts"]; ok {
		if len(logMaxFile) != 0 || len(logMaxSize) != 0 {
			daemonMap["log-driver"] = "json-file"
		}
		optsMap, isMap := opts.(map[string]interface{})
		if isMap {
			if len(logMaxFile) != 0 {
				optsMap["max-file"] = logMaxFile
			} else {
				delete(optsMap, "max-file")
			}
			if len(logMaxSize) != 0 {
				optsMap["max-size"] = logMaxSize
			} else {
				delete(optsMap, "max-size")
			}
			if len(optsMap) == 0 {
				delete(daemonMap, "log-opts")
			}
		} else {
			optsMap := make(map[string]interface{})
			if len(logMaxFile) != 0 {
				optsMap["max-file"] = logMaxFile
			}
			if len(logMaxSize) != 0 {
				optsMap["max-size"] = logMaxSize
			}
			if len(optsMap) != 0 {
				daemonMap["log-opts"] = optsMap
			}
		}
	} else {
		if len(logMaxFile) != 0 || len(logMaxSize) != 0 {
			daemonMap["log-driver"] = "json-file"
		}
		optsMap := make(map[string]interface{})
		if len(logMaxFile) != 0 {
			optsMap["max-file"] = logMaxFile
		}
		if len(logMaxSize) != 0 {
			optsMap["max-size"] = logMaxSize
		}
		if len(optsMap) != 0 {
			daemonMap["log-opts"] = optsMap
		}
	}
}

func validateDockerConfig() error {
	if !cmd.Which("dockerd") {
		return nil
	}
	stdout, err := cmd.Exec("dockerd --validate")
	if strings.Contains(stdout, "unknown flag: --validate") {
		return nil
	}
	if err != nil || (stdout != "" && strings.TrimSpace(stdout) != "configuration OK") {
		return fmt.Errorf("Docker configuration validation failed, err: %v", stdout)
	}
	return nil
}

func getDockerRestartCommand() (string, error) {
	stdout, err := cmd.Exec("which docker")
	if err != nil {
		return "", fmt.Errorf("failed to find docker: %v", err)
	}
	dockerPath := stdout
	if strings.Contains(dockerPath, "snap") {
		return "snap", nil
	}
	return "systemctl", nil
}

func restartDocker() error {
	restartCmd, err := getDockerRestartCommand()
	if err != nil {
		return err
	}
	stdout, err := cmd.Execf("%s restart docker", restartCmd)
	if err != nil {
		return fmt.Errorf("failed to restart Docker: %s", stdout)
	}
	return nil
}
