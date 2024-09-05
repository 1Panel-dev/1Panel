package cache

import (
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/cache/badger_db"
	"github.com/dgraph-io/badger/v4"
)

func Init() {
	c := global.CONF.System.Cache
	_ = os.RemoveAll(c)
	_ = os.Mkdir(c, 0755)
	options := badger.Options{
		Dir:                c,
		ValueDir:           c,
		ValueLogFileSize:   64 << 20,
		ValueLogMaxEntries: 10 << 20,
		VLogPercentile:     0.1,

		MemTableSize:                  32 << 20,
		BaseTableSize:                 2 << 20,
		BaseLevelSize:                 10 << 20,
		TableSizeMultiplier:           2,
		LevelSizeMultiplier:           10,
		MaxLevels:                     7,
		NumGoroutines:                 4,
		MetricsEnabled:                true,
		NumCompactors:                 2,
		NumLevelZeroTables:            5,
		NumLevelZeroTablesStall:       15,
		NumMemtables:                  1,
		BloomFalsePositive:            0.01,
		BlockSize:                     2 * 1024,
		SyncWrites:                    false,
		NumVersionsToKeep:             1,
		CompactL0OnClose:              false,
		VerifyValueChecksum:           false,
		BlockCacheSize:                32 << 20,
		IndexCacheSize:                0,
		ZSTDCompressionLevel:          1,
		EncryptionKey:                 []byte{},
		EncryptionKeyRotationDuration: 10 * 24 * time.Hour, // Default 10 days.
		DetectConflicts:               true,
		NamespaceOffset:               -1,
	}

	cache, err := badger.Open(options)
	if err != nil {
		panic(err)
	}
	_ = cache.DropAll()
	global.CacheDb = cache
	global.CACHE = badger_db.NewCacheDB(cache)
	global.LOG.Info("init cache successfully")
}
