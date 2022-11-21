package app

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"path"
)

func Init() {
	constant.DefaultDataDir = "/opt/1Panel/data"
	constant.ResourceDir = path.Join(constant.DefaultDataDir, "resource")
	constant.AppResourceDir = path.Join(constant.ResourceDir, "apps")
	constant.AppInstallDir = path.Join(constant.DefaultDataDir, "apps")
	constant.BackupDir = path.Join(constant.DefaultDataDir, "backup")
	constant.AppBackupDir = path.Join(constant.BackupDir, "apps")

	dirs := []string{constant.DefaultDataDir, constant.ResourceDir, constant.AppResourceDir, constant.AppInstallDir, constant.BackupDir, constant.AppBackupDir}

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
	if !cli.NetworkExist("1panel") {
		if err := cli.CreateNetwork("1panel"); err != nil {
			fmt.Println("init docker client error", err.Error())
			return
		}
	}
}
