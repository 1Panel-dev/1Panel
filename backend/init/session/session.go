package session

import (
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/session/psession"
)

func Init() {
	global.SESSION = psession.NewPSession(global.CACHE)
}
