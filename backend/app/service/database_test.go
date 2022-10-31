package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/go-redis/redis"
)

func TestMysql(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       0,
	})
	// fmt.Println(rdb.Get("dqwas"))

	client.Set("omg", "111", 10*time.Minute)
	client.Set("omg1", "111", 10*time.Minute)
	client.Set("omg2", "111", 10*time.Minute)
	client.Set("omg3", "111", 10*time.Minute)
	client.Set("omg4", "111", 10*time.Minute)
	client.Set("omg5", "111", 10*time.Minute)
	client.Set("omg6", "111", 10*time.Minute)
	client.Set("omg7", "111", 10*time.Minute)
	client.Set("omg8", "111", 10*time.Minute)
	client.Set("omg9", "111", 10*time.Minute)
	keys, _, err := client.Scan(0, "*", 5).Result()
	if err != nil {
		panic(err)
	}

	var data []dto.RedisData
	for _, key := range keys {
		var dataItem dto.RedisData
		dataItem.Key = key
		value, err := client.Get(key).Result()
		if err != nil {
			fmt.Println(err)
		}
		dataItem.Value = value
		typeVal, err := client.Type(key).Result()
		if err != nil {
			fmt.Println(err)
		}
		dataItem.Type = typeVal
		length, err := client.StrLen(key).Result()
		if err != nil {
			fmt.Println(err)
		}
		dataItem.Length = length
		ttl, err := client.TTL(key).Result()
		if err != nil {
			fmt.Println(err)
		}
		dataItem.Expiration = int64(ttl / 1000000000)
		data = append(data, dataItem)
	}
	fmt.Println(data)
}
