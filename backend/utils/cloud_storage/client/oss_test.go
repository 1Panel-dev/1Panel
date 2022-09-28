package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

	name := "directory/directory-test1/20220928104331.tar.gz"
	targetPath := constant.DownloadDir + "directory/directory-test1/"
	if _, err := os.Stat(targetPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetPath, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}
	if err := bucket.GetObjectToFile(name, targetPath+"20220928104231.tar.gz"); err != nil {
		fmt.Println(err)
	}
}

func TestDir(t *testing.T) {
	files, err := ioutil.ReadDir("/opt/1Panel/task/directory/directory-test1-3")
	if len(files) <= 10 {
		return
	}
	for i := 0; i < len(files)-10; i++ {
		os.Remove("/opt/1Panel/task/directory/directory-test1-3/" + files[i].Name())
	}
	fmt.Println(files, err)
}

func getObjectsFormResponse(lor oss.ListObjectsResult) string {
	var output string
	for _, object := range lor.Objects {
		output += object.Key + "  "
	}
	return output
}
