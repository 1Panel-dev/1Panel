package constant

import (
	"path"

	"github.com/1Panel-dev/1Panel/backend/global"
)

var (
	DataDir              = global.CONF.System.DataDir
	ResourceDir          = path.Join(DataDir, "resource")
	AppResourceDir       = path.Join(ResourceDir, "apps")
	AppInstallDir        = path.Join(DataDir, "apps")
	LocalAppResourceDir  = path.Join(AppResourceDir, "local")
	LocalAppInstallDir   = path.Join(AppInstallDir, "local")
	RemoteAppResourceDir = path.Join(AppResourceDir, "remote")
	RuntimeDir           = path.Join(DataDir, "runtime")
	RecycleBinDir        = "/.1panel_clash"
	SSLLogDir            = path.Join(global.CONF.System.DataDir, "log", "ssl")
)
