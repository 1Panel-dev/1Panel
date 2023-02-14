package app

import (
	"fmt"
	"path"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

func Init() {
	constant.DataDir = global.CONF.System.DataDir
	constant.ResourceDir = path.Join(constant.DataDir, "resource")
	constant.AppResourceDir = path.Join(constant.ResourceDir, "apps")
	constant.AppInstallDir = path.Join(constant.DataDir, "apps")

	dirs := []string{constant.DataDir, constant.ResourceDir, constant.AppResourceDir, constant.AppInstallDir, global.CONF.System.Backup}

	fileOp := files.NewFileOp()
	for _, dir := range dirs {
		createDir(fileOp, dir)
	}

	createDefaultDockerNetwork()
}

func createDir(fileOp files.FileOp, dirPath string) {
	if !fileOp.Stat(dirPath) {
		_ = fileOp.CreateDir(dirPath, 0755)
	}
}

func createDefaultDockerNetwork() {
	cli, err := docker.NewClient()
	if err != nil {
		fmt.Println("init docker client error", err.Error())
		return
	}
	if !cli.NetworkExist("1panel-network") {
		if err := cli.CreateNetwork("1panel-network"); err != nil {
			fmt.Println("init docker client error", err.Error())
			return
		}
	}
}
