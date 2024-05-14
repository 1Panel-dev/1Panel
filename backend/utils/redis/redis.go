package redis

import (
	"fmt"

	"github.com/go-redis/redis"
)

type DBInfo struct {
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Password string `json:"password"`
}

func NewRedisClient(conn DBInfo) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%v", conn.Address, conn.Port),
		Password: conn.Password,
		DB:       0,
	})

	if _, err := client.Ping().Result(); err != nil {
		return client, err
	}
	return client, nil
}
