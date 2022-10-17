package cache

import (
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/cache/badger_db"
	"github.com/dgraph-io/badger/v3"
	"time"
)

func Init() {
	c := global.CONF.Cache

	options := badger.Options{
		Dir:                c.Path,
		ValueDir:           c.Path,
		ValueLogFileSize:   102400000,
		ValueLogMaxEntries: 100000,
		VLogPercentile:     0.1,

		MemTableSize:                  64 << 20,
		BaseTableSize:                 2 << 20,
		BaseLevelSize:                 10 << 20,
		TableSizeMultiplier:           2,
		LevelSizeMultiplier:           10,
		MaxLevels:                     7,
		NumGoroutines:                 8,
		MetricsEnabled:                true,
		NumCompactors:                 4,
		NumLevelZeroTables:            5,
		NumLevelZeroTablesStall:       15,
		NumMemtables:                  5,
		BloomFalsePositive:            0.01,
		BlockSize:                     4 * 1024,
		SyncWrites:                    false,
		NumVersionsToKeep:             1,
		CompactL0OnClose:              false,
		VerifyValueChecksum:           false,
		BlockCacheSize:                256 << 20,
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

	global.CACHE = badger_db.NewCacheDB(cache)
}
