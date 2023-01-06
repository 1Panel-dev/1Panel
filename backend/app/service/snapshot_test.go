package service

import (
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/db"
	"github.com/1Panel-dev/1Panel/backend/init/viper"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

func TestSnaa(t *testing.T) {
	fileOp := files.NewFileOp()

	fmt.Println(fileOp.CopyFile("/Users/slooop/Documents/编码规范.pdf", "/Users/slooop/Downloads"))
	// fmt.Println(fileOp.Compress([]string{"/Users/slooop/Documents/编码规范.pdf", "/Users/slooop/Downloads/1Panel.db"}, "/Users/slooop/Downloads/", "test.tar.gz", files.TarGz))
}

func TestOss(t *testing.T) {
	viper.Init()
	db.Init()

	var backup model.BackupAccount
	if err := global.DB.Where("id = ?", 6).First(&backup).Error; err != nil {
		fmt.Println(err)
	}
	backupAccont, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(backupAccont.Upload("/Users/slooop/Downloads/1Panel.db", "database/1Panel.db"))
}
