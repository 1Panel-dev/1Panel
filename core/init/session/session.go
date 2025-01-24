package session

import (
	"path"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
)

func Init() {
	global.SESSION = psession.NewPSession(path.Join(global.CONF.Base.InstallDir, "1panel/db/session.db"))
	global.LOG.Info("init session successfully")
}
