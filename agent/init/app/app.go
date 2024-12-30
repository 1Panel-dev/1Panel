package app

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func Init() {
	constant.DataDir = global.CONF.System.DataDir
	constant.ResourceDir = path.Join(constant.DataDir, "resource")
	constant.AppResourceDir = path.Join(constant.ResourceDir, "apps")
	constant.AppInstallDir = path.Join(constant.DataDir, "apps")
	constant.RuntimeDir = path.Join(constant.DataDir, "runtime")

	constant.LocalAppResourceDir = path.Join(constant.AppResourceDir, "local")
	constant.LocalAppInstallDir = path.Join(constant.AppInstallDir, "local")
	constant.RemoteAppResourceDir = path.Join(constant.AppResourceDir, "remote")
	constant.CustomAppResourceDir = path.Join(constant.AppResourceDir, "custom")

	constant.LogDir = path.Join(global.CONF.System.DataDir, "log")
	constant.SSLLogDir = path.Join(constant.LogDir, "ssl")

	dirs := []string{
		constant.DataDir, constant.ResourceDir, constant.AppResourceDir, constant.AppInstallDir,
		global.CONF.System.Backup, constant.RuntimeDir, constant.LocalAppResourceDir,
		constant.RemoteAppResourceDir, constant.SSLLogDir,
		constant.CustomAppResourceDir,
	}

	fileOp := files.NewFileOp()
	for _, dir := range dirs {
		createDir(fileOp, dir)
	}

	go func() {
		_ = docker.CreateDefaultDockerNetwork()

		if f, err := firewall.NewFirewallClient(); err == nil {
			_ = f.EnableForward()
		}
	}()
}

func createDir(fileOp files.FileOp, dirPath string) {
	if !fileOp.Stat(dirPath) {
		_ = fileOp.CreateDir(dirPath, constant.DirPerm)
	}
}
