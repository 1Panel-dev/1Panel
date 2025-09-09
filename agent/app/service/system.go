package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"os/exec"
)

type SystemService struct{}

type ISystemService interface {
	IsComponentExist(name string) response.ComponentInfo
}

func NewISystemService() ISystemService {
	return &SystemService{}
}

func (s *SystemService) IsComponentExist(name string) response.ComponentInfo {
	info := response.ComponentInfo{}
	path, err := exec.LookPath(name)
	if err != nil {
		info.Exists = false
		info.Error = err.Error()
		return info
	}
	info.Exists = true
	info.Path = path
	return info
}
