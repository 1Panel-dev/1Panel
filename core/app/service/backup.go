package service

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage/client"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/req_helper/proxy_local"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/jinzhu/copier"
)

type BackupService struct{}

type IBackupService interface {
	LoadBackupClientInfo(clientType string) (dto.BackupClientInfo, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
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
	if clientType == constant.GoogleDrive {
		clientIDKey = "GoogleID"
		clientIDSc = "GoogleSc"
		data.RedirectUri = constant.GoogleRedirectURI
	} else {
		data.RedirectUri = constant.OneDriveRedirectURI
	}
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

	if req.Type == constant.OneDrive || req.Type == constant.GoogleDrive {
		if err := u.loadRefreshTokenByCode(&backup); err != nil {
			return err
		}
	}
	if req.Type != "LOCAL" {
		if _, err := u.checkBackupConn(&backup); err != nil {
			return buserr.WithMap("ErrBackupCheck", map[string]interface{}{"err": err.Error()}, err)
		}
	}

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
	go syncAccountToAgent(backup, "create")
	return nil
}

func (u *BackupService) GetBuckets(req dto.ForBuckets) ([]interface{}, error) {
	itemAccessKey, err := base64.StdEncoding.DecodeString(req.AccessKey)
	if err != nil {
		return nil, err
	}
	req.AccessKey = string(itemAccessKey)
	itemCredential, err := base64.StdEncoding.DecodeString(req.Credential)
	if err != nil {
		return nil, err
	}
	req.Credential = string(itemCredential)

	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(req.Vars), &varMap); err != nil {
		return nil, err
	}
	switch req.Type {
	case constant.Sftp, constant.WebDAV:
		varMap["username"] = req.AccessKey
		varMap["password"] = req.Credential
	case constant.OSS, constant.S3, constant.MinIo, constant.Cos, constant.Kodo:
		varMap["accessKey"] = req.AccessKey
		varMap["secretKey"] = req.Credential
	}
	client, err := cloud_storage.NewCloudStorageClient(req.Type, varMap)
	if err != nil {
		return nil, err
	}
	return client.ListBuckets()
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
	if _, err := proxy_local.NewLocalClient(fmt.Sprintf("/api/v2/backups/check/%s", name), http.MethodGet, nil); err != nil {
		global.LOG.Errorf("check used of local cronjob failed, err: %v", err)
		return buserr.New("ErrBackupInUsed")
	}
	if err := xpack.CheckBackupUsed(name); err != nil {
		global.LOG.Errorf("check used of node cronjob failed, err: %v", err)
		return buserr.New("ErrBackupInUsed")
	}

	go syncAccountToAgent(backup, "delete")
	return backupRepo.Delete(repo.WithByName(name))
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

	if newBackup.Type == constant.OneDrive || newBackup.Type == constant.GoogleDrive {
		if err := u.loadRefreshTokenByCode(&backup); err != nil {
			return err
		}
	}
	isOk, err := u.checkBackupConn(&newBackup)
	if err != nil || !isOk {
		return buserr.WithMap("ErrBackupCheck", map[string]interface{}{"err": err.Error()}, err)
	}

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
	go syncAccountToAgent(newBackup, "update")
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
		return fmt.Errorf("Failed to refresh %s - %s token, please retry, err: %v", backup.Type, backup.Name, err)
	}
	var (
		refreshToken string
		err          error
	)
	switch backup.Type {
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
		return fmt.Errorf("Failed to refresh %s-%s token, please retry, err: %v", backup.Type, backup.Name, err)
	}
	varMap["refresh_status"] = constant.StatusSuccess
	varMap["refresh_time"] = time.Now().Format(constant.DateTimeLayout)
	varMap["refresh_token"] = refreshToken

	varsItem, _ := json.Marshal(varMap)
	backup.Vars = string(varsItem)
	return backupRepo.Save(&backup)
}

func (u *BackupService) NewClient(backup *model.BackupAccount) (cloud_storage.CloudStorageClient, error) {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return nil, err
	}
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp, constant.WebDAV:
		varMap["username"] = backup.AccessKey
		varMap["password"] = backup.Credential
	case constant.OSS, constant.S3, constant.MinIo, constant.Cos, constant.Kodo:
		varMap["accessKey"] = backup.AccessKey
		varMap["secretKey"] = backup.Credential
	case constant.UPYUN:
		varMap["operator"] = backup.AccessKey
		varMap["password"] = backup.Credential
	}

	backClient, err := cloud_storage.NewCloudStorageClient(backup.Type, varMap)
	if err != nil {
		return nil, err
	}

	return backClient, nil
}

func (u *BackupService) loadRefreshTokenByCode(backup *model.BackupAccount) error {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return fmt.Errorf("unmarshal backup vars failed, err: %v", err)
	}
	refreshToken := ""
	var err error
	if backup.Type == constant.GoogleDrive {
		refreshToken, err = client.RefreshGoogleToken("authorization_code", "refreshToken", varMap)
		if err != nil {
			return err
		}
	} else {
		refreshToken, err = client.RefreshToken("authorization_code", "refreshToken", varMap)
		if err != nil {
			return err
		}
	}
	delete(varMap, "code")
	varMap["refresh_status"] = constant.StatusSuccess
	varMap["refresh_time"] = time.Now().Format(constant.DateTimeLayout)
	varMap["refresh_token"] = refreshToken
	itemVars, err := json.Marshal(varMap)
	if err != nil {
		return fmt.Errorf("json marshal var map failed, err: %v", err)
	}
	backup.Vars = string(itemVars)
	return nil
}

func (u *BackupService) checkBackupConn(backup *model.BackupAccount) (bool, error) {
	client, err := u.NewClient(backup)
	if err != nil {
		return false, err
	}
	fileItem := path.Join(global.CONF.Base.InstallDir, "1panel/tmp/test/1panel")
	if _, err := os.Stat(path.Dir(fileItem)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(fileItem), os.ModePerm); err != nil {
			return false, err
		}
	}
	file, err := os.OpenFile(fileItem, os.O_WRONLY|os.O_CREATE, constant.FilePerm)
	if err != nil {
		return false, err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString("1Panel 备份账号测试文件。\n")
	_, _ = write.WriteString("1Panel 備份賬號測試文件。\n")
	_, _ = write.WriteString("1Panel Backs up account test files.\n")
	_, _ = write.WriteString("1Panelアカウントのテストファイルをバックアップします。\n")
	write.Flush()

	targetPath := path.Join(backup.BackupPath, "test/1panel")
	if backup.Type != constant.Sftp && backup.Type != constant.Local && targetPath != "/" {
		targetPath = strings.TrimPrefix(targetPath, "/")
	}
	if _, err := client.Upload(fileItem, targetPath); err != nil {
		return false, err
	}
	_, _ = client.Delete(path.Join(backup.BackupPath, "test"))
	return true, nil
}

func syncAccountToAgent(backup model.BackupAccount, operation string) {
	if !backup.IsPublic {
		return
	}
	backup.AccessKey, _ = encrypt.StringDecryptWithBase64(backup.AccessKey)
	backup.Credential, _ = encrypt.StringDecryptWithBase64(backup.Credential)
	itemData, _ := json.Marshal(backup)
	itemJson := dto.SyncToAgent{Name: backup.Name, Operation: operation, Data: string(itemData)}
	bodyItem, _ := json.Marshal(itemJson)
	_ = xpack.RequestToAllAgent("/api/v2/backups/sync", http.MethodPost, bytes.NewReader((bodyItem)))
	_, _ = proxy_local.NewLocalClient("/api/v2/backups/sync", http.MethodPost, bytes.NewReader((bodyItem)))
}
