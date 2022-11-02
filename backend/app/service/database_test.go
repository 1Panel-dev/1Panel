package service

import (
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/go-redis/redis"
)

func TestMysql(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "172.16.10.143:6379",
		Password: "",
		DB:       0,
	})
	fmt.Println(client.Ping().Result())

	var item dto.RedisPersistence
	dir, _ := client.ConfigGet("dir").Result()
	if len(dir) == 2 {
		if value, ok := dir[1].(string); ok {
			item.Dir = value
		}
	}
	appendonly, _ := client.ConfigGet("appendonly").Result()
	if len(appendonly) == 2 {
		if value, ok := appendonly[1].(string); ok {
			item.Appendonly = value
		}
	}
	appendfsync, _ := client.ConfigGet("appendfsync").Result()
	if len(appendfsync) == 2 {
		if value, ok := appendfsync[1].(string); ok {
			item.Appendfsync = value
		}
	}
	save, _ := client.ConfigGet("save").Result()
	if len(save) == 2 {
		if value, ok := save[1].(string); ok {
			item.Save = value
		}
	}
	fmt.Println(item)
}
