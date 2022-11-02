package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

type RedisService struct{}

type IRedisService interface {
	UpdateConf(req dto.RedisConfUpdate) error

	LoadStatus() (*dto.RedisStatus, error)
	LoadConf() (*dto.RedisConf, error)
	LoadPersistenceConf() (*dto.RedisPersistence, error)

	// Backup(db dto.BackupDB) error
	// Recover(db dto.RecoverDB) error
}

func NewIRedisService() IRedisService {
	return &RedisService{}
}

func newRedisClient() (*redis.Client, error) {
	redisInfo, err := mysqlRepo.LoadRedisBaseInfo()
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: redisInfo.Password,
		DB:       0,
	})
	return client, nil
}

func (u *RedisService) UpdateConf(req dto.RedisConfUpdate) error {
	redisInfo, err := mysqlRepo.LoadRedisBaseInfo()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(fmt.Sprintf("/opt/1Panel/data/apps/redis/%s/conf/redis.conf", redisInfo.Name), os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	pos := int64(0)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		if bytes := updateConfFile(line, "timeout", req.Timeout); len(bytes) != 0 {
			_, _ = file.WriteAt(bytes, pos)
		}
		if bytes := updateConfFile(line, "maxclients", req.Maxclients); len(bytes) != 0 {
			_, _ = file.WriteAt(bytes, pos)
		}
		if bytes := updateConfFile(line, "databases", req.Databases); len(bytes) != 0 {
			_, _ = file.WriteAt(bytes, pos)
		}
		if bytes := updateConfFile(line, "requirepass", req.Requirepass); len(bytes) != 0 {
			_, _ = file.WriteAt(bytes, pos)
		}
		if bytes := updateConfFile(line, "maxmemory", req.Maxmemory); len(bytes) != 0 {
			_, _ = file.WriteAt(bytes, pos)
		}
		pos += int64(len(line))
	}
	return nil
}

func (u *RedisService) LoadStatus() (*dto.RedisStatus, error) {
	client, err := newRedisClient()
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

func (u *RedisService) LoadConf() (*dto.RedisConf, error) {
	redisInfo, err := mysqlRepo.LoadRedisBaseInfo()
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: redisInfo.Password,
		DB:       0,
	})
	var item dto.RedisConf
	item.ContainerName = redisInfo.ContainerName
	item.Name = redisInfo.Name
	item.Timeout = configGetStr(client, "timeout")
	item.Maxclients = configGetStr(client, "maxclients")
	item.Databases = configGetStr(client, "databases")
	item.Requirepass = configGetStr(client, "requirepass")
	item.Maxmemory = configGetStr(client, "maxmemory")
	return &item, nil
}

func (u *RedisService) LoadPersistenceConf() (*dto.RedisPersistence, error) {
	redisInfo, err := mysqlRepo.LoadRedisBaseInfo()
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", redisInfo.Port),
		Password: redisInfo.Password,
		DB:       0,
	})
	var item dto.RedisPersistence
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

func updateConfFile(line, param string, value string) []byte {
	var bytes []byte
	if strings.HasPrefix(line, param) || strings.HasPrefix(line, "# "+param) {
		if len(value) == 0 || value == "0" {
			bytes = []byte(fmt.Sprintf("# %s", param))
		} else {
			bytes = []byte(fmt.Sprintf("%s %v", param, value))
		}
		return bytes
	}
	return bytes
}
