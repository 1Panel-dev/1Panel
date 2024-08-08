package service

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/core/utils/cloud_storage/client"
	fileUtils "github.com/1Panel-dev/1Panel/core/utils/files"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type BackupService struct{}

type IBackupService interface {
	SearchWithPage(search dto.SearchPageWithType) (int64, interface{}, error)
	LoadOneDriveInfo() (dto.OneDriveInfo, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
	Update(req dto.BackupOperate) error
	Delete(id uint) error
	NewClient(backup *model.BackupAccount) (cloud_storage.CloudStorageClient, error)

	Run()
}

func NewIBackupService() IBackupService {
	return &BackupService{}
}

func (u *BackupService) SearchWithPage(req dto.SearchPageWithType) (int64, interface{}, error) {
	count, accounts, err := backupRepo.Page(
		req.Page,
		req.PageSize,
		commonRepo.WithByType(req.Type),
		commonRepo.WithByName(req.Info),
		commonRepo.WithOrderBy("created_at desc"),
	)
	if err != nil {
		return 0, nil, err
	}
	var data []dto.BackupInfo
	for _, account := range accounts {
		var item dto.BackupInfo
		if err := copier.Copy(&item, &account); err != nil {
			global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
		}
		if !item.RememberAuth {
			item.AccessKey = ""
			item.Credential = ""
		} else {
			item.AccessKey = base64.StdEncoding.EncodeToString([]byte(item.AccessKey))
			item.Credential = base64.StdEncoding.EncodeToString([]byte(item.Credential))
		}

		if account.Type == constant.OneDrive {
			varMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(item.Vars), &varMap); err != nil {
				continue
			}
			delete(varMap, "refresh_token")
			itemVars, _ := json.Marshal(varMap)
			item.Vars = string(itemVars)
		}
		data = append(data, item)
	}
	return count, data, nil
}

func (u *BackupService) LoadOneDriveInfo() (dto.OneDriveInfo, error) {
	var data dto.OneDriveInfo
	data.RedirectUri = constant.OneDriveRedirectURI
	clientID, err := settingRepo.Get(settingRepo.WithByKey("OneDriveID"))
	if err != nil {
		return data, err
	}
	idItem, err := base64.StdEncoding.DecodeString(clientID.Value)
	if err != nil {
		return data, err
	}
	data.ClientID = string(idItem)
	clientSecret, err := settingRepo.Get(settingRepo.WithByKey("OneDriveSc"))
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
	backup, _ := backupRepo.Get(commonRepo.WithByName(req.Name))
	if backup.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&backup, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	if req.Type == constant.OneDrive {
		if err := u.loadAccessToken(&backup); err != nil {
			return err
		}
	}
	if req.Type != "LOCAL" {
		if _, err := u.checkBackupConn(&backup); err != nil {
			return buserr.WithMap("ErrBackupCheck", map[string]interface{}{"err": err.Error()}, err)
		}
	}
	if backup.Type == constant.OneDrive {
		if err := StartRefreshOneDriveToken(&backup); err != nil {
			return err
		}
	}
	if err := backupRepo.Create(&backup); err != nil {
		return err
	}
	return nil
}

func (u *BackupService) GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error) {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backupDto.Vars), &varMap); err != nil {
		return nil, err
	}
	switch backupDto.Type {
	case constant.Sftp, constant.WebDAV:
		varMap["username"] = backupDto.AccessKey
		varMap["password"] = backupDto.Credential
	case constant.OSS, constant.S3, constant.MinIo, constant.Cos, constant.Kodo:
		varMap["accessKey"] = backupDto.AccessKey
		varMap["secretKey"] = backupDto.Credential
	}
	client, err := cloud_storage.NewCloudStorageClient(backupDto.Type, varMap)
	if err != nil {
		return nil, err
	}
	return client.ListBuckets()
}

func (u *BackupService) Delete(id uint) error {
	backup, _ := backupRepo.Get(commonRepo.WithByID(id))
	if backup.ID == 0 {
		return constant.ErrRecordNotFound
	}
	if backup.Type == constant.Local {
		return buserr.New(constant.ErrBackupLocalDelete)
	}
	if backup.InUsed {
		return buserr.New(constant.ErrBackupInUsed)
	}
	if backup.Type == constant.OneDrive {
		global.Cron.Remove(cron.EntryID(backup.EntryID))
	}
	return backupRepo.Delete(commonRepo.WithByID(id))
}

func (u *BackupService) Update(req dto.BackupOperate) error {
	backup, err := backupRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return constant.ErrRecordNotFound
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(req.Vars), &varMap); err != nil {
		return err
	}

	oldVars := backup.Vars
	oldDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	upMap := make(map[string]interface{})
	upMap["bucket"] = req.Bucket
	upMap["access_key"] = req.AccessKey
	upMap["credential"] = req.Credential
	upMap["backup_path"] = req.BackupPath
	upMap["vars"] = req.Vars
	backup.Bucket = req.Bucket
	backup.Vars = req.Vars
	backup.Credential = req.Credential
	backup.AccessKey = req.AccessKey
	backup.BackupPath = req.BackupPath

	if req.Type == constant.OneDrive {
		global.Cron.Remove(cron.EntryID(backup.EntryID))
		if err := u.loadAccessToken(&backup); err != nil {
			return err
		}
		upMap["credential"] = backup.Credential
		upMap["vars"] = backup.Vars
		if err := StartRefreshOneDriveToken(&backup); err != nil {
			return err
		}
		upMap["entry_id"] = backup.EntryID
	}
	if backup.Type != "LOCAL" {
		isOk, err := u.checkBackupConn(&backup)
		if err != nil || !isOk {
			return buserr.WithMap("ErrBackupCheck", map[string]interface{}{"err": err.Error()}, err)
		}
	}

	if err := backupRepo.Update(req.ID, upMap); err != nil {
		return err
	}
	if backup.Type == "LOCAL" {
		if dir, ok := varMap["dir"]; ok {
			if dirStr, isStr := dir.(string); isStr {
				if strings.HasSuffix(dirStr, "/") && dirStr != "/" {
					dirStr = dirStr[:strings.LastIndex(dirStr, "/")]
				}
				if err := copyDir(oldDir, dirStr); err != nil {
					_ = backupRepo.Update(req.ID, map[string]interface{}{"vars": oldVars})
					return err
				}
			}
		}
	}
	return nil
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
	}

	backClient, err := cloud_storage.NewCloudStorageClient(backup.Type, varMap)
	if err != nil {
		return nil, err
	}

	return backClient, nil
}

func (u *BackupService) loadAccessToken(backup *model.BackupAccount) error {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return fmt.Errorf("unmarshal backup vars failed, err: %v", err)
	}
	refreshToken, err := client.RefreshToken("authorization_code", "refreshToken", varMap)
	if err != nil {
		return err
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

func loadLocalDir() (string, error) {
	backup, err := backupRepo.Get(commonRepo.WithByType("LOCAL"))
	if err != nil {
		return "", err
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return "", err
	}
	if _, ok := varMap["dir"]; !ok {
		return "", errors.New("load local backup dir failed")
	}
	baseDir, ok := varMap["dir"].(string)
	if ok {
		if _, err := os.Stat(baseDir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(baseDir, os.ModePerm); err != nil {
				return "", fmt.Errorf("mkdir %s failed, err: %v", baseDir, err)
			}
		}
		return baseDir, nil
	}
	return "", fmt.Errorf("error type dir: %T", varMap["dir"])
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, file := range files {
		srcPath := fmt.Sprintf("%s/%s", src, file.Name())
		dstPath := fmt.Sprintf("%s/%s", dst, file.Name())
		if file.IsDir() {
			if err = copyDir(srcPath, dstPath); err != nil {
				global.LOG.Errorf("copy dir %s to %s failed, err: %v", srcPath, dstPath, err)
			}
		} else {
			if err := fileUtils.CopyFile(srcPath, dst); err != nil {
				global.LOG.Errorf("copy file %s to %s failed, err: %v", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

func (u *BackupService) checkBackupConn(backup *model.BackupAccount) (bool, error) {
	client, err := u.NewClient(backup)
	if err != nil {
		return false, err
	}
	fileItem := path.Join(global.CONF.System.BaseDir, "1panel/tmp/test/1panel")
	if _, err := os.Stat(path.Dir(fileItem)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(fileItem), os.ModePerm); err != nil {
			return false, err
		}
	}
	file, err := os.OpenFile(fileItem, os.O_WRONLY|os.O_CREATE, 0666)
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

	targetPath := strings.TrimPrefix(path.Join(backup.BackupPath, "test/1panel"), "/")
	return client.Upload(fileItem, targetPath)
}

func StartRefreshOneDriveToken(backup *model.BackupAccount) error {
	service := NewIBackupService()
	oneDriveCronID, err := global.Cron.AddJob("0 3 */31 * *", service)
	if err != nil {
		global.LOG.Errorf("can not add OneDrive corn job: %s", err.Error())
		return err
	}
	backup.EntryID = uint(oneDriveCronID)
	return nil
}

func (u *BackupService) Run() {
	var backupItem model.BackupAccount
	_ = global.DB.Where("`type` = ?", "OneDrive").First(&backupItem)
	if backupItem.ID == 0 {
		return
	}
	global.LOG.Info("start to refresh token of OneDrive ...")
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backupItem.Vars), &varMap); err != nil {
		global.LOG.Errorf("Failed to refresh OneDrive token, please retry, err: %v", err)
		return
	}
	refreshToken, err := client.RefreshToken("refresh_token", "refreshToken", varMap)
	varMap["refresh_status"] = constant.StatusSuccess
	varMap["refresh_time"] = time.Now().Format(constant.DateTimeLayout)
	if err != nil {
		varMap["refresh_status"] = constant.StatusFailed
		varMap["refresh_msg"] = err.Error()
		global.LOG.Errorf("Failed to refresh OneDrive token, please retry, err: %v", err)
		return
	}
	varMap["refresh_token"] = refreshToken

	varsItem, _ := json.Marshal(varMap)
	_ = global.DB.Model(&model.BackupAccount{}).
		Where("id = ?", backupItem.ID).
		Updates(map[string]interface{}{
			"vars": varsItem,
		}).Error
	global.LOG.Info("Successfully refreshed OneDrive token.")
}
