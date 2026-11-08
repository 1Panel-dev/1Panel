package service

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func LoadModelGPUUsage(ctx context.Context, req dto.ModelGPUUsageRequest) (dto.ModelGPUUsage, error) {
	topology := LoadModelGPUDevices()
	result := dto.ModelGPUUsage{Devices: []dto.ModelGPUUsageDevice{}, Warnings: topology.Warnings}
	values, err := godotenv.Unmarshal(req.Env)
	if err != nil {
		return result, err
	}
	params := make(map[string]interface{}, len(values))
	for key, value := range values {
		params[key] = value
	}
	var project struct {
		Services map[string]map[string]interface{} `yaml:"services"`
	}
	if err := yaml.Unmarshal([]byte(req.DockerCompose), &project); err != nil {
		return result, err
	}
	selected, err := modelGPUUsageSelection(project.Services, params, req.ImageType, topology.Devices)
	if err != nil {
		return result, err
	}
	for _, device := range topology.Devices {
		if possible, ok := selected[device.ID]; ok {
			result.Devices = append(result.Devices, dto.ModelGPUUsageDevice{ModelGPUDevice: device, Possible: possible, Deployments: []dto.ModelGPUDeployment{}})
		}
	}
	if len(result.Devices) == 0 {
		return result, nil
	}
	installs, err := appInstallRepo.ListBy(ctx)
	if err != nil {
		return result, err
	}
	client, err := docker.NewDockerClient()
	if err != nil {
		return result, err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	containers, err := client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return result, err
	}
	states := make(map[string]string)
	for _, item := range containers {
		for _, name := range item.Names {
			states[strings.TrimPrefix(name, "/")] = item.State
		}
	}
	for _, install := range installs {
		if install.ID == req.ExcludeInstallID {
			continue
		}
		var project struct {
			Services map[string]map[string]interface{} `yaml:"services"`
		}
		if err := yaml.Unmarshal([]byte(install.DockerCompose), &project); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", install.Name, err))
			continue
		}
		var env map[string]interface{}
		if err := json.Unmarshal([]byte(install.Env), &env); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", install.Name, err))
			continue
		}
		for serviceName, service := range project.Services {
			devices, err := modelGPUUsageSelection(map[string]map[string]interface{}{serviceName: service}, env, req.ImageType, topology.Devices)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s/%s: %v", install.Name, serviceName, err))
				continue
			}
			name := install.ContainerName
			if raw, ok := service["container_name"].(string); ok {
				name, err = resolveModelGPUValue(raw, env)
				if err != nil {
					result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", install.Name, err))
					continue
				}
			} else if len(project.Services) != 1 {
				name = ""
			}
			status := states[name]
			if status == "" {
				status = "unknown"
			}
			for i := range result.Devices {
				if possible, ok := devices[result.Devices[i].ID]; ok {
					result.Devices[i].Deployments = append(result.Devices[i].Deployments, dto.ModelGPUDeployment{Name: install.Name + "/" + serviceName, Status: status, Possible: possible})
				}
			}
		}
	}
	return result, nil
}

func modelGPUUsageSelection(services map[string]map[string]interface{}, params map[string]interface{}, imageType string, devices []dto.ModelGPUDevice) (map[string]bool, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("Docker Compose must contain services")
	}
	selected := map[string]bool{}
	for _, service := range services {
		config := ReadModelGPU(service, params, imageType)
		if config.Error != "" {
			return nil, fmt.Errorf("%s", config.Error)
		}
		if !config.Enabled {
			continue
		}
		matched := false
		for _, device := range devices {
			if device.Type != imageType {
				continue
			}
			matches := config.Mode == "all" || config.Mode == "count" || slices.Contains(config.DeviceIDs, device.ID)
			if imageType == "nvidia" {
				matches = matches || slices.Contains(config.DeviceIDs, "nvidia:"+device.UUID) || slices.Contains(config.DeviceIDs, "nvidia:"+strconv.Itoa(device.Index))
			}
			if matches {
				matched = true
				possible := config.Mode == "count"
				if previous, exists := selected[device.ID]; !exists || previous {
					selected[device.ID] = possible
				}
			}
		}
		if !matched {
			return nil, fmt.Errorf("no detected %s GPU matches the deployment", imageType)
		}
	}
	return selected, nil
}
