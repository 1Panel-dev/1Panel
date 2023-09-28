package response

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"time"
)

type RuntimeDTO struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Resource    string                 `json:"resource"`
	AppDetailID uint                   `json:"appDetailID"`
	AppID       uint                   `json:"appID"`
	Source      string                 `json:"source"`
	Status      string                 `json:"status"`
	Type        string                 `json:"type"`
	Image       string                 `json:"image"`
	Params      map[string]interface{} `json:"params"`
	Message     string                 `json:"message"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"createdAt"`
	CodeDir     string                 `json:"codeDir"`
	AppParams   []AppParam             `json:"appParams"`
	Port        int                    `json:"port"`
	Path        string                 `json:"path"`
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
	}
}
