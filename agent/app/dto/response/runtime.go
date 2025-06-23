package response

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type RuntimeDTO struct {
	ID              uint                   `json:"id"`
	Name            string                 `json:"name"`
	Resource        string                 `json:"resource"`
	AppDetailID     uint                   `json:"appDetailID"`
	AppID           uint                   `json:"appID"`
	Source          string                 `json:"source"`
	Status          string                 `json:"status"`
	Type            string                 `json:"type"`
	Image           string                 `json:"image"`
	Params          map[string]interface{} `json:"params"`
	Message         string                 `json:"message"`
	Version         string                 `json:"version"`
	CreatedAt       time.Time              `json:"createdAt"`
	CodeDir         string                 `json:"codeDir"`
	AppParams       []AppParam             `json:"appParams"`
	Port            string                 `json:"port"`
	Path            string                 `json:"path"`
	ExposedPorts    []request.ExposedPort  `json:"exposedPorts"`
	Environments    []request.Environment  `json:"environments"`
	Volumes         []request.Volume       `json:"volumes"`
	ContainerStatus string                 `json:"containerStatus"`
	Container       string                 `json:"container"`
	Remark          string                 `json:"remark"`
}

type PackageScripts struct {
	Name   string `json:"name"`
	Script string `json:"script"`
}

func NewRuntimeDTO(runtime model.Runtime) RuntimeDTO {
	return RuntimeDTO{
		ID:          runtime.ID,
		Name:        runtime.Name,
		Resource:    runtime.Resource,
		AppDetailID: runtime.AppDetailID,
		Status:      runtime.Status,
		Type:        runtime.Type,
		Image:       runtime.Image,
		Message:     runtime.Message,
		CreatedAt:   runtime.CreatedAt,
		CodeDir:     runtime.CodeDir,
		Version:     runtime.Version,
		Port:        runtime.Port,
		Path:        runtime.GetPath(),
		Container:   runtime.ContainerName,
		Remark:      runtime.Remark,
	}
}

type NodeModule struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	License     string `json:"license"`
	Description string `json:"description"`
}

type SupportExtension struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Installed   bool     `json:"installed"`
	Check       string   `json:"check"`
	Versions    []string `json:"versions"`
	File        string   `json:"file"`
}

type PHPExtensionRes struct {
	Extensions        []string           `json:"extensions"`
	SupportExtensions []SupportExtension `json:"supportExtensions"`
}
