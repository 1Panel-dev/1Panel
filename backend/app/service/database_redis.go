package service

import (
	"encoding/json"
	"fmt"
	"strings"
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

	UpdateConf(req dto.RedisConfUpdate) error

	CleanAll(req dto.RedisBaseReq) error
	LoadState(req dto.RedisBaseReq) (*dto.RedisStatus, error)
	LoadConf(req dto.RedisBaseReq) (*dto.RedisConf, error)
	LoadRedisRunningVersion() ([]string, error)

	// Backup(db dto.BackupDB) error
	// Recover(db dto.RecoverDB) error
}

func NewIRedisService() IRedisService {
	return &RedisService{}
}

func newRedisClient(name string, db int) (*redis.Client, error) {
	redisInfo, err := mysqlRepo.LoadRedisBaseInfoByName(name)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81",
		DB:       db,
	})
	return client, nil
}

func (u *RedisService) SearchWithPage(search dto.SearchRedisWithPage) (int64, interface{}, error) {
	client, err := newRedisClient(search.RedisName, search.DB)
	if err != nil {
		return 0, nil, err
	}
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

func (u *RedisService) LoadRedisRunningVersion() ([]string, error) {
	return mysqlRepo.LoadRunningVersion([]string{"redis"})
}

func (u *RedisService) UpdateConf(req dto.RedisConfUpdate) error {
	client, err := newRedisClient(req.RedisName, 0)
	if err != nil {
		return err
	}
	if _, err := client.ConfigSet(req.ParamName, req.Value).Result(); err != nil {
		return err
	}
	if _, err := client.ConfigRewrite().Result(); err != nil {
		return err
	}

	return nil
}

func (u *RedisService) Set(setData dto.RedisDataSet) error {
	client, err := newRedisClient(setData.RedisName, setData.DB)
	if err != nil {
		return err
	}
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
	client, err := newRedisClient(req.RedisName, req.DB)
	if err != nil {
		return err
	}
	if _, err := client.Del(req.Names...).Result(); err != nil {
		return err
	}
	return nil
}

func (u *RedisService) CleanAll(req dto.RedisBaseReq) error {
	client, err := newRedisClient(req.RedisName, req.DB)
	if err != nil {
		return err
	}
	if _, err := client.FlushAll().Result(); err != nil {
		return err
	}
	return nil
}

func (u *RedisService) LoadState(req dto.RedisBaseReq) (*dto.RedisStatus, error) {
	client, err := newRedisClient(req.RedisName, req.DB)
	if err != nil {
		return nil, err
	}
	stdStr, err := client.Info().Result()
	if err != nil {
		return nil, err
	}
	rows := strings.Split(stdStr, "\r\n")
	rowMap := make(map[string]string)
	for _, v := range rows {
		itemRow := strings.Split(v, ":")
		if len(itemRow) == 2 {
			rowMap[itemRow[0]] = itemRow[1]
		}
	}
	var info dto.RedisStatus
	arr, err := json.Marshal(rowMap)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(arr, &info)
	return &info, nil
}

func (u *RedisService) LoadConf(req dto.RedisBaseReq) (*dto.RedisConf, error) {
	client, err := newRedisClient(req.RedisName, req.DB)
	if err != nil {
		return nil, err
	}
	var item dto.RedisConf
	item.Timeout = configGetStr(client, "timeout")
	item.Maxclients = configGetStr(client, "maxclients")
	item.Databases = configGetStr(client, "databases")
	item.Requirepass = configGetStr(client, "requirepass")
	item.Maxmemory = configGetStr(client, "maxmemory")

	item.Dir = configGetStr(client, "dir")
	item.Appendonly = configGetStr(client, "appendonly")
	item.Appendfsync = configGetStr(client, "appendfsync")
	item.Save = configGetStr(client, "save")
	return &item, nil
}

func configGetStr(client *redis.Client, param string) string {
	item, _ := client.ConfigGet(param).Result()
	if len(item) == 2 {
		if value, ok := item[1].(string); ok {
			return value
		}
	}
	return ""
}
