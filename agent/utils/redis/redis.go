package redis

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/go-redis/redis"
)

type DBInfo struct {
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Password string `json:"password"`
	Timeout  uint   `json:"timeout"`
}

func NewRedisClient(conn DBInfo) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%v", conn.Address, conn.Port),
		Password:    conn.Password,
		DB:          0,
		DialTimeout: time.Duration(conn.Timeout) * time.Second,
	})

	if _, err := client.Ping().Result(); err != nil {
		global.LOG.Errorf("check redis conn failed, err: %v", err)
		return client, err
	}
	return client, nil
}
