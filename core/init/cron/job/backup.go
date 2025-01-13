package job

import (
	"encoding/json"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage/client"
)

type backup struct{}

func NewBackupJob() *backup {
	return &backup{}
}

func (b *backup) Run() {
	var backups []model.BackupAccount
	_ = global.DB.Where("`type` in (?) AND is_public = 0", []string{constant.OneDrive, constant.ALIYUN, constant.GoogleDrive}).Find(&backups)
	if len(backups) == 0 {
		return
	}
	for _, backupItem := range backups {
		if backupItem.ID == 0 {
			continue
		}
		global.LOG.Infof("Start to refresh %s-%s access_token ...", backupItem.Type, backupItem.Name)
		varMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(backupItem.Vars), &varMap); err != nil {
			global.LOG.Errorf("Failed to refresh %s - %s token, please retry, err: %v", backupItem.Type, backupItem.Name, err)
			continue
		}
		var (
			refreshToken string
			err          error
		)
		switch backupItem.Type {
		case constant.OneDrive:
			refreshToken, err = client.RefreshToken("refresh_token", "refreshToken", varMap)
		case constant.GoogleDrive:
			refreshToken, err = client.RefreshGoogleToken("refresh_token", "refreshToken", varMap)
		case constant.ALIYUN:
			refreshToken, err = client.RefreshALIToken(varMap)
		}
		if err != nil {
			varMap["refresh_status"] = constant.StatusFailed
			varMap["refresh_msg"] = err.Error()
			global.LOG.Errorf("Failed to refresh %s-%s token, please retry, err: %v", backupItem.Type, backupItem.Name, err)
			continue
		}
		varMap["refresh_status"] = constant.StatusSuccess
		varMap["refresh_time"] = time.Now().Format(constant.DateTimeLayout)
		varMap["refresh_token"] = refreshToken

		varsItem, _ := json.Marshal(varMap)
		_ = global.DB.Model(&model.BackupAccount{}).Where("id = ?", backupItem.ID).Updates(map[string]interface{}{"vars": string(varsItem)}).Error
		global.LOG.Infof("Refresh %s-%s access_token successful!", backupItem.Type, backupItem.Name)
	}
}
