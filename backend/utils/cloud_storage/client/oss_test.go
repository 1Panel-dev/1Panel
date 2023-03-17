package client

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/db"
	"github.com/1Panel-dev/1Panel/backend/init/log"
	"github.com/1Panel-dev/1Panel/backend/init/viper"
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

	name := "directory/directory-test1/20220928104331.tar.gz"
	targetPath := constant.DataDir + "/download/directory/directory-test1/"
	if _, err := os.Stat(targetPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetPath, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}
	if err := bucket.GetObjectToFile(name, targetPath+"20220928104231.tar.gz"); err != nil {
		fmt.Println(err)
	}
}

func getObjectsFormResponse(lor oss.ListObjectsResult) string {
	var output string
	for _, object := range lor.Objects {
		output += object.Key + "  "
	}
	return output
}
