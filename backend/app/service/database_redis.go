package service

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

type RedisService struct{}

type IRedisService interface {
	SearchWithPage(search dto.SearchRedisWithPage) (int64, interface{}, error)
	Set(setData dto.RedisDataSet) error
	Delete(info dto.RedisDelBatch) error
	CleanAll(db int) error

	// Backup(db dto.BackupDB) error
	// Recover(db dto.RecoverDB) error
}

func NewIRedisService() IRedisService {
	return &RedisService{}
}

func (u *RedisService) SearchWithPage(search dto.SearchRedisWithPage) (int64, interface{}, error) {
	redisInfo, err := mysqlRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return 0, nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       search.DB,
	})
	total, err := client.DbSize().Result()
	if err != nil {
		return 0, nil, err
	}
	keys, _, err := client.Scan(uint64((search.Page-1)*search.PageSize), "*", int64(search.PageSize)).Result()
	if err != nil {
		return 0, nil, err
	}
	var data []dto.RedisData
	for _, key := range keys {
		var dataItem dto.RedisData
		dataItem.Key = key
		value, err := client.Get(key).Result()
		if err != nil {
			return 0, nil, err
		}
		dataItem.Value = value
		typeVal, err := client.Type(key).Result()
		if err != nil {
			return 0, nil, err
		}
		dataItem.Type = typeVal
		length, err := client.StrLen(key).Result()
		if err != nil {
			return 0, nil, err
		}
		dataItem.Length = length
		ttl, err := client.TTL(key).Result()
		if err != nil {
			return 0, nil, err
		}
		dataItem.Expiration = int64(ttl / 1000000000)
		data = append(data, dataItem)
	}
	return total, data, nil
}

func (u *RedisService) Set(setData dto.RedisDataSet) error {
	redisInfo, err := mysqlRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       setData.DB,
	})
	value, _ := client.Get(setData.Key).Result()
	if err != nil {
		return err
	}
	if len(value) != 0 {
		if _, err := client.Del(setData.Key).Result(); err != nil {
			return err
		}
	}
	if _, err := client.Set(setData.Key, setData.Value, time.Duration(setData.Expiration*int64(time.Second))).Result(); err != nil {
		return err
	}
	return nil
}

func (u *RedisService) Delete(req dto.RedisDelBatch) error {
	redisInfo, err := mysqlRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       req.DB,
	})
	if _, err := client.Del(req.Names...).Result(); err != nil {
		return err
	}
	return nil
}

func (u *RedisService) CleanAll(db int) error {
	redisInfo, err := mysqlRepo.LoadBaseInfoByKey("redis")
	if err != nil {
		return err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       db,
	})
	if _, err := client.FlushAll().Result(); err != nil {
		return err
	}
	return nil
}
