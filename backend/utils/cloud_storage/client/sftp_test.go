package client

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
)

func TestCronS(t *testing.T) {
	viper.Init()
	log.Init()
	db.Init()

	var backup model.BackupAccount
	if err := global.DB.Where("id = ?", 5).First(&backup).Error; err != nil {
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
	client, err := NewS3Client(varMap)
	if err != nil {
		fmt.Println(err)
	}

	_, _ = client.ListObjects("directory/directory-test-s3")
}
