package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/req_helper/proxy_local"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/jinzhu/copier"
)

type BackupService struct{}

type IBackupService interface {
	LoadBackupClientInfo(clientType string) (dto.BackupClientInfo, error)
	Create(backupDto dto.BackupOperate) error
	Update(req dto.BackupOperate) error
	Delete(name string) error
	RefreshToken(req dto.OperateByName) error
}

func NewIBackupService() IBackupService {
	return &BackupService{}
}

func (u *BackupService) LoadBackupClientInfo(clientType string) (dto.BackupClientInfo, error) {
	var data dto.BackupClientInfo
	clientIDKey := "OneDriveID"
	clientIDSc := "OneDriveSc"
	data.RedirectUri = constant.OneDriveRedirectURI
	clientID, err := settingRepo.Get(repo.WithByKey(clientIDKey))
	if err != nil {
		return data, err
	}
	idItem, err := base64.StdEncoding.DecodeString(clientID.Value)
	if err != nil {
		return data, err
	}
	data.ClientID = string(idItem)
	clientSecret, err := settingRepo.Get(repo.WithByKey(clientIDSc))
	if err != nil {
		return data, err
	}
	secretItem, err := base64.StdEncoding.DecodeString(clientSecret.Value)
	if err != nil {
		return data, err
	}
	data.ClientSecret = string(secretItem)

	return data, err
}

func (u *BackupService) Create(req dto.BackupOperate) error {
	if !req.IsPublic {
		return buserr.New("ErrBackupPublic")
	}
	backup, _ := backupRepo.Get(repo.WithByName(req.Name))
	if backup.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if req.Type != constant.Sftp && req.BackupPath != "/" {
		req.BackupPath = strings.TrimPrefix(req.BackupPath, "/")
	}
	if err := copier.Copy(&backup, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	itemAccessKey, err := base64.StdEncoding.DecodeString(backup.AccessKey)
	if err != nil {
		return err
	}
	backup.AccessKey = string(itemAccessKey)
	itemCredential, err := base64.StdEncoding.DecodeString(backup.Credential)
	if err != nil {
		return err
	}
	backup.Credential = string(itemCredential)

	backup.AccessKey, err = encrypt.StringEncrypt(backup.AccessKey)
	if err != nil {
		return err
	}
	backup.Credential, err = encrypt.StringEncrypt(backup.Credential)
	if err != nil {
		return err
	}
	if err := backupRepo.Create(&backup); err != nil {
		return err
	}
	if err := xpack.Sync(constant.SyncBackupAccounts); err != nil {
		global.LOG.Errorf("sync backup account to node failed, err: %v", err)
	}
	return nil
}

func (u *BackupService) Delete(name string) error {
	backup, _ := backupRepo.Get(repo.WithByName(name))
	if backup.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if !backup.IsPublic {
		return buserr.New("ErrBackupPublic")
	}
	if backup.Type == constant.Local {
		return buserr.New("ErrBackupLocal")
	}
	if _, err := proxy_local.NewLocalClient(fmt.Sprintf("/api/v2/backups/check/%s", name), http.MethodGet, nil, nil); err != nil {
		global.LOG.Errorf("check used of local cronjob failed, err: %v", err)
		return buserr.New("ErrBackupInUsed")
	}
	if err := xpack.CheckBackupUsed(name); err != nil {
		global.LOG.Errorf("check used of node cronjob failed, err: %v", err)
		return buserr.New("ErrBackupInUsed")
	}

	if err := backupRepo.Delete(repo.WithByName(name)); err != nil {
		return err
	}
	if err := xpack.Sync(constant.SyncBackupAccounts); err != nil {
		global.LOG.Errorf("sync backup account to node failed, err: %v", err)
	}
	return nil
}

func (u *BackupService) Update(req dto.BackupOperate) error {
	backup, _ := backupRepo.Get(repo.WithByName(req.Name))
	if backup.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if !backup.IsPublic {
		return buserr.New("ErrBackupPublic")
	}
	if backup.Type == constant.Local {
		return buserr.New("ErrBackupLocal")
	}
	if req.Type != constant.Sftp && req.BackupPath != "/" {
		req.BackupPath = strings.TrimPrefix(req.BackupPath, "/")
	}
	var newBackup model.BackupAccount
	if err := copier.Copy(&newBackup, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	newBackup.ID = backup.ID
	itemAccessKey, err := base64.StdEncoding.DecodeString(newBackup.AccessKey)
	if err != nil {
		return err
	}
	newBackup.AccessKey = string(itemAccessKey)
	itemCredential, err := base64.StdEncoding.DecodeString(newBackup.Credential)
	if err != nil {
		return err
	}
	newBackup.Credential = string(itemCredential)
	newBackup.AccessKey, err = encrypt.StringEncrypt(newBackup.AccessKey)
	if err != nil {
		return err
	}
	newBackup.Credential, err = encrypt.StringEncrypt(newBackup.Credential)
	if err != nil {
		return err
	}
	newBackup.ID = backup.ID
	newBackup.CreatedAt = backup.CreatedAt
	newBackup.UpdatedAt = backup.UpdatedAt
	if err := backupRepo.Save(&newBackup); err != nil {
		return err
	}
	if err := xpack.Sync(constant.SyncBackupAccounts); err != nil {
		global.LOG.Errorf("sync backup account to node failed, err: %v", err)
	}
	return nil
}

func (u *BackupService) RefreshToken(req dto.OperateByName) error {
	backup, _ := backupRepo.Get(repo.WithByName(req.Name))
	if backup.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if !backup.IsPublic {
		return buserr.New("ErrBackupPublic")
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return fmt.Errorf("failed to refresh %s - %s token, please retry, err: %v", backup.Type, backup.Name, err)
	}
	var (
		refreshToken string
		err          error
	)
	switch backup.Type {
	case constant.OneDrive:
		refreshToken, err = cloud_storage.RefreshToken("refresh_token", "refreshToken", varMap)
	case constant.ALIYUN:
		refreshToken, err = cloud_storage.RefreshALIToken(varMap)
	}
	if err != nil {
		varMap["refresh_status"] = constant.StatusFailed
		varMap["refresh_msg"] = err.Error()
		return fmt.Errorf("failed to refresh %s-%s token, please retry, err: %v", backup.Type, backup.Name, err)
	}
	varMap["refresh_status"] = constant.StatusSuccess
	varMap["refresh_time"] = time.Now().Format(constant.DateTimeLayout)
	varMap["refresh_token"] = refreshToken

	varsItem, _ := json.Marshal(varMap)
	backup.Vars = string(varsItem)
	if err := backupRepo.Save(&backup); err != nil {
		return err
	}
	if err := xpack.Sync(constant.SyncBackupAccounts); err != nil {
		global.LOG.Errorf("sync backup account to node failed, err: %v", err)
	}
	return nil
}
