package constant

import (
	"path"
)

var (
	DefaultDataDir = "/opt/1Panel/data"
	ResourceDir    = path.Join(DefaultDataDir, "resource")
	AppResourceDir = path.Join(ResourceDir, "apps")
	AppInstallDir  = path.Join(DefaultDataDir, "apps")
)

const (
	TmpDir      = "/opt/1Panel/data/tmp"
	TaskDir     = "/opt/1Panel/data/task"
	DownloadDir = "/opt/1Panel/data/download"
	UploadDir   = "/opt/1Panel/data/uploads"
)
