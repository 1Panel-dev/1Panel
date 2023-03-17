package client

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/db"
	"github.com/1Panel-dev/1Panel/backend/init/log"
	"github.com/1Panel-dev/1Panel/backend/init/viper"
)

func TestMinio(t *testing.T) {
	viper.Init()
	log.Init()
	db.Init()

	var backup model.BackupAccount
	if err := global.DB.Where("id = ?", 3).First(&backup).Error; err != nil {
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
	client, _ := NewMinIoClient(varMap)
	_, _ = client.ListObjects("directory/directory-test-minio/")
}
