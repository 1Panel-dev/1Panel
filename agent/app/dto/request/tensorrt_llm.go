package request

import "github.com/1Panel-dev/1Panel/agent/app/dto"

type TensorRTLLMSearch struct {
	dto.PageInfo
	Name string `json:"name"`
}

type TensorRTLLMCreate struct {
	Name          string `json:"name" validate:"required"`
	ContainerName string `json:"containerName"  validate:"required"`
	Version       string `json:"version"  validate:"required"`
	ModelDir      string `json:"modelDir" validate:"required"`
	Image         string `json:"image"  validate:"required"`
	Command       string `json:"command" validate:"required"`
	DockerConfig
}

type TensorRTLLMUpdate struct {
	ID uint `json:"id" validate:"required"`
	TensorRTLLMCreate
}

type TensorRTLLMDelete struct {
	ID uint `json:"id" validate:"required"`
}

type TensorRTLLMOperate struct {
	ID      uint   `json:"id" validate:"required"`
	Operate string `json:"operate" validate:"required"`
}

type DockerConfig struct {
	ExposedPorts []ExposedPort `json:"exposedPorts"`
	Environments []Environment `json:"environments"`
	Volumes      []Volume      `json:"volumes"`
}
