package request

import "github.com/1Panel-dev/1Panel/agent/app/dto"

type TensorRTLLMSearch struct {
	dto.PageInfo
	Name string `json:"name"`
}

type TensorRTLLMCreate struct {
	Name          string `json:"name" validate:"required"`
	ContainerName string `json:"containerName"`
	Port          int    `json:"port" validate:"required"`
	Version       string `json:"version"  validate:"required"`
	ModelDir      string `json:"modelDir" validate:"required"`
	Model         string `json:"model" validate:"required"`
	HostIP        string `json:"hostIP"`
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
