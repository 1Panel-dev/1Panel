package dir

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func Init() {
	fileOp := files.NewFileOp()
	baseDir := global.CONF.Base.InstallDir
	_, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/docker/compose/"))

	global.Dir.BaseDir, _ = fileOp.CreateDirWithPath(true, baseDir)
	global.Dir.DataDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel"))
	global.Dir.DbDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/db"))
	global.Dir.LogDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/log"))
	global.Dir.TaskDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/log/task"))
	global.Dir.TmpDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/tmp"))

	global.Dir.AppDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/apps"))
	global.Dir.ResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/resource"))
	global.Dir.AppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/resource/apps"))
	global.Dir.AppInstallDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/apps"))
	global.Dir.LocalAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/resource/apps/local"))
	global.Dir.LocalAppInstallDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/apps/local"))
	global.Dir.RemoteAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/resource/apps/remote"))
	global.Dir.CustomAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/resource/apps/custom"))
	global.Dir.RuntimeDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/runtime"))
	global.Dir.RecycleBinDir, _ = fileOp.CreateDirWithPath(true, "/.1panel_clash")
	global.Dir.SSLLogDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/log/ssl"))
	global.Dir.McpDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/mcp"))
	global.Dir.ConvertLogDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "1panel/log/convert"))
}
