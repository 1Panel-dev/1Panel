package constant

import (
	"path"
)

var (
	DefaultDataDir = "/opt/1Panel/data"
	LogDir         = "opt/1Panel/log"
	ResourceDir    = path.Join(DefaultDataDir, "resource")
	AppResourceDir = path.Join(ResourceDir, "apps")
	AppInstallDir  = path.Join(DefaultDataDir, "apps")
	BackupDir      = path.Join(DefaultDataDir, "backup")
	AppBackupDir   = path.Join(BackupDir, "apps")
)
