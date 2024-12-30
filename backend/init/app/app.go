package app

import (
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/firewall"
	"path"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
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

	constant.SSLLogDir = path.Join(global.CONF.System.DataDir, "log", "ssl")

	dirs := []string{constant.DataDir, constant.ResourceDir, constant.AppResourceDir, constant.AppInstallDir,
		global.CONF.System.Backup, constant.RuntimeDir, constant.LocalAppResourceDir, constant.RemoteAppResourceDir, constant.SSLLogDir}

	fileOp := files.NewFileOp()
	for _, dir := range dirs {
		createDir(fileOp, dir)
	}

	go func() {
		_ = docker.CreateDefaultDockerNetwork()

		if f, err := firewall.NewFirewallClient(); err == nil {
			if err = f.EnableForward(); err != nil {
				global.LOG.Errorf("init port forward failed, err: %v", err)
			}
		}
	}()
}

func createDir(fileOp files.FileOp, dirPath string) {
	if !fileOp.Stat(dirPath) {
		_ = fileOp.CreateDir(dirPath, 0755)
	}
}
