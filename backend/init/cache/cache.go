package cache

import (
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/cache/badger_db"
	"github.com/dgraph-io/badger/v3"
)

func Init() {
	c := global.CONF.Cache

	cache, err := badger.Open(badger.DefaultOptions(c.Path))
	if err != nil {
		panic(err)
	}

	global.CACHE = badger_db.NewCacheDB(cache)
}
