package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/db"
	"github.com/1Panel-dev/1Panel/init/log"
	"github.com/1Panel-dev/1Panel/init/viper"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func TestCron(t *testing.T) {
	viper.Init()
	log.Init()
	db.Init()

	var backup model.BackupAccount
	if err := global.DB.Where("id = ?", 2).First(&backup).Error; err != nil {
		fmt.Println(err)
	}

	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		fmt.Println(err)
	}
	varMap["type"] = backup.Type
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp:
		varMap["password"] = backup.Credential
	case constant.OSS, constant.S3, constant.MinIo:
		varMap["secretKey"] = backup.Credential
	}
	endpoint := varMap["endpoint"].(string)
	accessKey := varMap["accessKey"].(string)
	secretKey := varMap["secretKey"].(string)
	client, err := oss.New(endpoint, accessKey, secretKey)
	if err != nil {
		fmt.Println(err)
	}
	bucket, err := client.Bucket(backup.Bucket)
	if err != nil {
		fmt.Println(err)
	}
	lor, err := bucket.ListObjects(oss.Prefix("directory/directory-test1/"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("my objects:", getObjectsFormResponse(lor))
}

func getObjectsFormResponse(lor oss.ListObjectsResult) string {
	var output string
	for _, object := range lor.Objects {
		output += object.Key + "  "
	}
	return output
}
