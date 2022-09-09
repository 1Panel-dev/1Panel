package badger_db

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"time"
)

type Cache struct {
	db *badger.DB
}

func NewCacheDB(db *badger.DB) *Cache {
	return &Cache{
		db: db,
	}
}

func (c *Cache) SetNX(key string, value interface{}) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			v := []byte(fmt.Sprintf("%v", value))
			return txn.Set([]byte(key), v)
		}
		return err
	})
	return err
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
