package model

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/global"
)

type Runtime struct {
	BaseModel
	Name          string `gorm:"not null" json:"name"`
	AppDetailID   uint   `json:"appDetailID"`
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
	return path.Join(global.Dir.RuntimeDir, r.Type, r.Name)
}

func (r *Runtime) GetFPMPath() string {
	return path.Join(global.Dir.RuntimeDir, r.Type, r.Name, "conf", "php-fpm.conf")
}

func (r *Runtime) GetPHPPath() string {
	return path.Join(global.Dir.RuntimeDir, r.Type, r.Name, "conf", "php.ini")
}

func (r *Runtime) GetLogPath() string {
	return path.Join(r.GetPath(), "build.log")
}
