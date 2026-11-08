package service

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"gopkg.in/yaml.v3"
)

const modelGPUGroupKey = "x-1panel-gpu-group"

var gpuGroupMu sync.Mutex

type GPUGroupService struct{}

func (s *GPUGroupService) List() ([]dto.GPUGroupInfo, error) {
	groups, err := gpuGroupRepo.List()
	if err != nil {
		return nil, err
	}
	result := make([]dto.GPUGroupInfo, 0, len(groups))
	for _, group := range groups {
		item := dto.GPUGroupInfo{GPUGroupSave: dto.GPUGroupSave{ID: group.ID, Name: group.Name, Description: group.Description, ImageType: group.ImageType}, Revision: group.Revision, Deployments: []string{}}
		if err := json.Unmarshal([]byte(group.Config), &item.GPU); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if len(result) == 0 {
		return result, nil
	}
	installs, err := gpuGroupRepo.Deployments()
	if err != nil {
		return nil, err
	}
	for _, install := range installs {
		var project struct {
			Services map[string]map[string]interface{} `yaml:"services"`
		}
		if err := yaml.Unmarshal([]byte(install.DockerCompose), &project); err != nil {
			return nil, fmt.Errorf("read GPU group for %s: %w", install.Name, err)
		}
		for _, service := range project.Services {
			value, ok := service[modelGPUGroupKey].(string)
			if !ok {
				continue
			}
			var source dto.GPUGroupSource
			if err := json.Unmarshal([]byte(value), &source); err != nil {
				return nil, err
			}
			for i := range result {
				if result[i].ID == source.ID && !slices.Contains(result[i].Deployments, install.Name) {
					result[i].Deployments = append(result[i].Deployments, install.Name)
				}
			}
		}
	}
	return result, nil
}

func (s *GPUGroupService) Save(req dto.GPUGroupSave) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fmt.Errorf("GPU group name is required")
	}
	req.GPU.Group = nil
	if !req.GPU.Enabled || req.GPU.Mode != "devices" {
		return fmt.Errorf("GPU groups require specific GPU devices")
	}
	if err := validateModelGPU(&req.GPU, req.ImageType); err != nil {
		return err
	}
	topology := LoadModelGPUDevices()
	if err := validateGPUGroupDevices(&req.GPU, topology, req.ImageType); err != nil {
		return err
	}
	if req.GPU.NUMA != nil && req.GPU.NUMA.Mode != "off" {
		if _, err := resolveModelNUMANodes(&req.GPU, topology, req.ImageType, ""); err != nil {
			return err
		}
	}
	config, err := json.Marshal(req.GPU)
	if err != nil {
		return err
	}
	gpuGroupMu.Lock()
	defer gpuGroupMu.Unlock()
	groups, err := gpuGroupRepo.List()
	if err != nil {
		return err
	}
	group := model.GPUGroup{}
	for _, existing := range groups {
		if existing.ID != req.ID && existing.Name == req.Name {
			return buserr.New("ErrRecordExist")
		}
		if existing.ID == req.ID {
			group = existing
		}
	}
	if req.ID != 0 && group.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	group.Name, group.ImageType, group.Config = req.Name, req.ImageType, string(config)
	group.Description = strings.TrimSpace(req.Description)
	group.Revision++
	return gpuGroupRepo.Save(&group)
}

func (s *GPUGroupService) Delete(id uint) error {
	gpuGroupMu.Lock()
	defer gpuGroupMu.Unlock()
	return gpuGroupRepo.Delete(id)
}

func validateGPUGroupDevices(config *dto.ModelGPUConfig, topology dto.ModelGPUDevices, imageType string) error {
	seen := map[string]bool{}
	for _, id := range config.DeviceIDs {
		if seen[id] {
			return fmt.Errorf("duplicate GPU device: %s", id)
		}
		seen[id] = true
		index := slices.IndexFunc(topology.Devices, func(device dto.ModelGPUDevice) bool {
			return device.ID == id && device.Type == imageType && device.Available
		})
		if index < 0 {
			return fmt.Errorf("GPU device is unavailable: %s", id)
		}
	}
	return nil
}
