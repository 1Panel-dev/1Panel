package model

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

type Runtime struct {
	BaseModel
	Name          string `gorm:"not null" json:"name"`
	AppDetailID   uint   `json:"appDetailId"`
	Image         string `json:"image"`
	WorkDir       string `json:"workDir"`
	DockerCompose string `json:"dockerCompose"`
	Env           string `json:"env"`
	Params        string `json:"params"`
	Version       string `gorm:"not null" json:"version"`
	Type          string `gorm:"not null" json:"type"`
	Status        string `gorm:"not null" json:"status"`
	Resource      string `gorm:"not null" json:"resource"`
	Port          int    `json:"port"`
	Message       string `json:"message"`
	CodeDir       string `json:"codeDir"`
	ContainerName string `json:"containerName"`
}

func (r *Runtime) GetComposePath() string {
	return path.Join(r.GetPath(), "docker-compose.yml")
}

func (r *Runtime) GetEnvPath() string {
	return path.Join(r.GetPath(), ".env")
}

func (r *Runtime) GetPath() string {
	return path.Join(constant.RuntimeDir, r.Type, r.Name)
}

func (r *Runtime) GetLogPath() string {
	return path.Join(r.GetPath(), "build.log")
}
