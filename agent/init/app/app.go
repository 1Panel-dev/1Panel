package app

import (
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
)

func Init() {
	go func() {
		_ = docker.CreateDefaultDockerNetwork()
	}()
}
