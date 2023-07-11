package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	fileUtils "github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupService struct{}

type IBackupService interface {
	List() ([]dto.BackupInfo, error)
	SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error)
	LoadOneDriveInfo() (string, error)
	DownloadRecord(info dto.DownloadRecord) (string, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
	Update(ireq dto.BackupOperate) error
	Delete(id uint) error
	BatchDeleteRecord(ids []uint) error
	NewClient(backup *model.BackupAccount) (cloud_storage.CloudStorageClient, error)

	ListFiles(req dto.BackupSearchFile) ([]interface{}, error)

	MysqlBackup(db dto.CommonBackup) error
	MysqlRecover(db dto.CommonRecover) error
	MysqlRecoverByUpload(req dto.CommonRecover) error

	RedisBackup() error
	RedisRecover(db dto.CommonRecover) error

	WebsiteBackup(db dto.CommonBackup) error
	WebsiteRecover(req dto.CommonRecover) error

	AppBackup(db dto.CommonBackup) error
	AppRecover(req dto.CommonRecover) error
}

func NewIBackupService() IBackupService {
	return &BackupService{}
}

func (u *BackupService) List() ([]dto.BackupInfo, error) {
	ops, err := backupRepo.List(commonRepo.WithOrderBy("created_at desc"))
	var dtobas []dto.BackupInfo
	dtobas = append(dtobas, u.loadByType("LOCAL", ops))
	dtobas = append(dtobas, u.loadByType("OSS", ops))
	dtobas = append(dtobas, u.loadByType("S3", ops))
	dtobas = append(dtobas, u.loadByType("SFTP", ops))
	dtobas = append(dtobas, u.loadByType("MINIO", ops))
	dtobas = append(dtobas, u.loadByType("COS", ops))
	dtobas = append(dtobas, u.loadByType("KODO", ops))
	dtobas = append(dtobas, u.loadByType("OneDrive", ops))
	return dtobas, err
}

func (u *BackupService) SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error) {
	total, records, err := backupRepo.PageRecord(
		search.Page, search.PageSize,
		commonRepo.WithOrderBy("created_at desc"),
		commonRepo.WithByName(search.Name),
		commonRepo.WithByType(search.Type),
		backupRepo.WithByDetailName(search.DetailName),
	)
	var dtobas []dto.BackupRecords
	for _, group := range records {
		var item dto.BackupRecords
		if err := copier.Copy(&item, &group); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtobas = append(dtobas, item)
	}
	return total, dtobas, err
}

func (u *BackupService) LoadOneDriveInfo() (string, error) {
	OneDriveID, err := settingRepo.Get(settingRepo.WithByKey("OneDriveID"))
	if err != nil {
		return "", err
	}
	idItem, err := base64.StdEncoding.DecodeString(OneDriveID.Value)
	if err != nil {
		return "", err
	}
	return string(idItem), err
}

func (u *BackupService) DownloadRecord(info dto.DownloadRecord) (string, error) {
	if info.Source == "LOCAL" {
		return info.FileDir + "/" + info.FileName, nil
	}
	backup, _ := backupRepo.Get(commonRepo.WithByType(info.Source))
	if backup.ID == 0 {
		return "", constant.ErrRecordNotFound
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return "", err
	}
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp:
		varMap["username"] = backup.AccessKey
		varMap["password"] = backup.Credential
	case constant.OSS, constant.S3, constant.MinIo, constant.Cos, constant.Kodo:
		varMap["accessKey"] = backup.AccessKey
		varMap["secretKey"] = backup.Credential
	case constant.OneDrive:
		varMap["accessToken"] = backup.Credential
	}
	backClient, err := cloud_storage.NewCloudStorageClient(backup.Type, varMap)
	if err != nil {
		return "", fmt.Errorf("new cloud storage client failed, err: %v", err)
	}
	targetPath := fmt.Sprintf("%s/download/%s/%s", constant.DataDir, info.FileDir, info.FileName)
	if _, err := os.Stat(path.Dir(targetPath)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(targetPath), os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", path.Dir(targetPath), err)
		}
	}
	srcPath := fmt.Sprintf("%s/%s", info.FileDir, info.FileName)
	if len(backup.BackupPath) != 0 {
		itemPath := strings.TrimPrefix(backup.BackupPath, "/")
		itemPath = strings.TrimSuffix(itemPath, "/") + "/"
		srcPath = itemPath + srcPath
	}
	if exist, _ := backClient.Exist(srcPath); exist {
		isOK, err := backClient.Download(srcPath, targetPath)
		if !isOK {
			return "", fmt.Errorf("cloud storage download failed, err: %v", err)
		}
	}
	return targetPath, nil
}

func (u *BackupService) Create(backupDto dto.BackupOperate) error {
	backup, _ := backupRepo.Get(commonRepo.WithByType(backupDto.Type))
	if backup.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&backup, &backupDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	if backupDto.Type == constant.OneDrive {
		if err := u.loadAccessToken(&backup); err != nil {
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
	case constant.Sftp:
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
	cronjobs, _ := cronjobRepo.List(cronjobRepo.WithByBackupID(id))
	if len(cronjobs) != 0 {
		return buserr.New(constant.ErrBackupInUsed)
	}
	return backupRepo.Delete(commonRepo.WithByID(id))
}

func (u *BackupService) BatchDeleteRecord(ids []uint) error {
	records, err := backupRepo.ListRecord(commonRepo.WithIdsIn(ids))
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.Source == "LOCAL" {
			if err := os.Remove(record.FileDir + "/" + record.FileName); err != nil {
				global.LOG.Errorf("remove file %s failed, err: %v", record.FileDir+record.FileName, err)
			}
		} else {
			backupAccount, err := backupRepo.Get(commonRepo.WithByType(record.Source))
			if err != nil {
				return err
			}
			client, err := u.NewClient(&backupAccount)
			if err != nil {
				return err
			}
			if _, err = client.Delete(record.FileDir + record.FileName); err != nil {
				global.LOG.Errorf("remove file %s from %s failed, err: %v", record.FileDir+record.FileName, record.Source, err)
			}
		}
	}
	return backupRepo.DeleteRecord(context.Background(), commonRepo.WithIdsIn(ids))
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
	upMap["credential"] = req.Credential
	upMap["backup_path"] = req.BackupPath
	upMap["vars"] = req.Vars
	backup.Vars = req.Vars

	if req.Type == constant.OneDrive {
		if err := u.loadAccessToken(&backup); err != nil {
			return err
		}
		upMap["credential"] = backup.Credential
		upMap["vars"] = backup.Vars
	}
	if err := backupRepo.Update(req.ID, upMap); err != nil {
		return err
	}
	if backup.Type == "LOCAL" {
		if dir, ok := varMap["dir"]; ok {
			if dirStr, isStr := dir.(string); isStr {
				if strings.HasSuffix(dirStr, "/") {
					dirStr = dirStr[:strings.LastIndex(dirStr, "/")]
				}
				if err := copyDir(oldDir, dirStr); err != nil {
					_ = backupRepo.Update(req.ID, (map[string]interface{}{"vars": oldVars}))
					return err
				}
			}
		}
	}
	return nil
}

func (u *BackupService) ListFiles(req dto.BackupSearchFile) ([]interface{}, error) {
	backup, err := backupRepo.Get(backupRepo.WithByType(req.Type))
	if err != nil {
		return nil, err
	}
	client, err := u.NewClient(&backup)
	if err != nil {
		return nil, err
	}
	return client.ListObjects("system_snapshot/")
}

func (u *BackupService) NewClient(backup *model.BackupAccount) (cloud_storage.CloudStorageClient, error) {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return nil, err
	}
	if backup.Type == "LOCAL" {
		return nil, errors.New("not support")
	}
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp:
		varMap["username"] = backup.AccessKey
		varMap["password"] = backup.Credential
	case constant.OSS, constant.S3, constant.MinIo, constant.Cos, constant.Kodo:
		varMap["accessKey"] = backup.AccessKey
		varMap["secretKey"] = backup.Credential
	case constant.OneDrive:
		varMap["accessToken"] = backup.Credential
	}

	backClient, err := cloud_storage.NewCloudStorageClient(backup.Type, varMap)
	if err != nil {
		return nil, err
	}

	return backClient, nil
}

func (u *BackupService) loadByType(accountType string, accounts []model.BackupAccount) dto.BackupInfo {
	for _, account := range accounts {
		if account.Type == accountType {
			var item dto.BackupInfo
			if err := copier.Copy(&item, &account); err != nil {
				global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
			}
			return item
		}
	}
	return dto.BackupInfo{Type: accountType}
}

func (u *BackupService) loadAccessToken(backup *model.BackupAccount) error {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return fmt.Errorf("unmarshal backup vars failed, err: %v", err)
	}

	data := url.Values{}
	data.Set("client_id", global.CONF.System.OneDriveID)
	data.Set("client_secret", global.CONF.System.OneDriveSc)
	data.Set("grant_type", "authorization_code")
	data.Set("code", varMap["code"].(string))
	data.Set("redirect_uri", constant.OneDriveRedirectURI)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://login.microsoftonline.com/common/oauth2/v2.0/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("new http post client for access token failed, err: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request for access token failed, err: %v", err)
	}
	delete(varMap, "code")
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read data from response body failed, err: %v", err)
	}
	defer resp.Body.Close()

	token := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &token); err != nil {
		return fmt.Errorf("unmarshal data from response body failed, err: %v", err)
	}
	accessToken, ok := token["refresh_token"].(string)
	if !ok {
		return errors.New("no such access token in response")
	}

	itemVars, err := json.Marshal(varMap)
	if err != nil {
		return fmt.Errorf("json marshal var map failed, err: %v", err)
	}
	backup.Credential = accessToken
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

	fileOP := fileUtils.NewFileOp()
	for _, file := range files {
		srcPath := fmt.Sprintf("%s/%s", src, file.Name())
		dstPath := fmt.Sprintf("%s/%s", dst, file.Name())
		if file.IsDir() {
			if err = copyDir(srcPath, dstPath); err != nil {
				global.LOG.Errorf("copy dir %s to %s failed, err: %v", srcPath, dstPath, err)
			}
		} else {
			if err := fileOP.CopyFile(srcPath, dst); err != nil {
				global.LOG.Errorf("copy file %s to %s failed, err: %v", srcPath, dstPath, err)
			}
		}
	}

	return nil
}
