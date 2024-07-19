package session

import (
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
)

func Init() {
	global.SESSION = psession.NewPSession(global.CACHE)
	global.LOG.Info("init session successfully")
}
