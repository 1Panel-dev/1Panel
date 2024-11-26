package session

import (
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"path"
)

func Init() {
	global.SESSION = psession.NewPSession(path.Join(global.CONF.System.BaseDir, "1panel/db/session.db"))
	global.LOG.Info("init session successfully")
}
