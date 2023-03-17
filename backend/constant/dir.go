package constant

import (
	"path"

	"github.com/1Panel-dev/1Panel/backend/global"
)

var (
	DataDir        = global.CONF.System.DataDir
	ResourceDir    = path.Join(DataDir, "resource")
	AppResourceDir = path.Join(ResourceDir, "apps")
	AppInstallDir  = path.Join(DataDir, "apps")
)
