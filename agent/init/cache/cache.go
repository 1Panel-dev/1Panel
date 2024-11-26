package cache

import (
	"github.com/1Panel-dev/1Panel/agent/global"
	cachedb "github.com/1Panel-dev/1Panel/agent/init/cache/db"
)

func Init() {
	global.CACHE = cachedb.NewCacheDB()
}
