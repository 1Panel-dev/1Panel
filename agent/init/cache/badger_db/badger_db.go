package badger_db

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Cache struct {
	db *badger.DB
}

func NewCacheDB(db *badger.DB) *Cache {
	return &Cache{
		db: db,
	}
}

func (c *Cache) Set(key string, value interface{}) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		v := []byte(fmt.Sprintf("%v", value))
		return txn.Set([]byte(key), v)
	})
	return err
}

func (c *Cache) Del(key string) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	return err
}

func (c *Cache) Clean() error {
	return c.db.DropAll()
}

func (c *Cache) Get(key string) ([]byte, error) {
	var result []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			result = append([]byte{}, val...)
			return nil
		})
		return err
	})
	return result, err
}

func (c *Cache) SetWithTTL(key string, value interface{}, duration time.Duration) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		v := []byte(fmt.Sprintf("%v", value))
		e := badger.NewEntry([]byte(key), v).WithTTL(duration)
		return txn.SetEntry(e)
	})
	return err
}

func (c *Cache) PrefixScanKey(prefixStr string) ([]string, error) {
	var res []string
	err := c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefixStr)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			res = append(res, string(k))
			return nil
		}
		return nil
	})
	return res, err
}
