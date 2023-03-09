package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupService struct{}

type IBackupService interface {
	List() ([]dto.BackupInfo, error)
	SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error)
	DownloadRecord(info dto.DownloadRecord) (string, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
	Update(ireq dto.BackupOperate) error
	BatchDelete(ids []uint) error
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
	ossExist, s3Exist, sftpExist, minioExist := false, false, false, false
	for _, group := range ops {
		switch group.Type {
		case "OSS":
			ossExist = true
		case "S3":
			s3Exist = true
		case "SFTP":
			sftpExist = true
		case "MINIO":
			minioExist = true
		}
		var item dto.BackupInfo
		if err := copier.Copy(&item, &group); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtobas = append(dtobas, item)
	}
	if !ossExist {
		dtobas = append(dtobas, dto.BackupInfo{Type: "OSS"})
	}
	if !s3Exist {
		dtobas = append(dtobas, dto.BackupInfo{Type: "S3"})
	}
	if !sftpExist {
		dtobas = append(dtobas, dto.BackupInfo{Type: "SFTP"})
	}
	if !minioExist {
		dtobas = append(dtobas, dto.BackupInfo{Type: "MINIO"})
	}
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
	varMap["type"] = backup.Type
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp:
		varMap["username"] = backup.AccessKey
		varMap["password"] = backup.Credential
	case constant.OSS, constant.S3, constant.MinIo:
		varMap["accessKey"] = backup.AccessKey
		varMap["secretKey"] = backup.Credential
	}
	backClient, err := cloud_storage.NewCloudStorageClient(varMap)
	if err != nil {
		return "", fmt.Errorf("new cloud storage client failed, err: %v", err)
	}
	tempPath := fmt.Sprintf("%sdownload%s", constant.DataDir, info.FileDir)
	if _, err := os.Stat(tempPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(tempPath, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}
	targetPath := tempPath + info.FileName
	if _, err = os.Stat(targetPath); err != nil && os.IsNotExist(err) {
		isOK, err := backClient.Download(info.FileName, targetPath)
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
	varMap["type"] = backupDto.Type
	switch backupDto.Type {
	case constant.Sftp:
		varMap["username"] = backupDto.AccessKey
		varMap["password"] = backupDto.Credential
	case constant.OSS, constant.S3, constant.MinIo:
		varMap["accessKey"] = backupDto.AccessKey
		varMap["secretKey"] = backupDto.Credential
	}
	client, err := cloud_storage.NewCloudStorageClient(varMap)
	if err != nil {
		return nil, err
	}
	return client.ListBuckets()
}

func (u *BackupService) BatchDelete(ids []uint) error {
	return backupRepo.Delete(commonRepo.WithIdsIn(ids))
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
			backupAccount, err := backupRepo.Get(commonRepo.WithByName(record.Source))
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
	upMap := make(map[string]interface{})
	upMap["bucket"] = req.Bucket
	upMap["credential"] = req.Credential
	upMap["vars"] = req.Vars
	if err := backupRepo.Update(req.ID, upMap); err != nil {
		return err
	}
	if backup.Type == "LOCAL" {
		if dir, ok := varMap["dir"]; ok {
			if dirStr, isStr := dir.(string); isStr {
				if strings.HasSuffix(dirStr, "/") {
					dirStr = dirStr[:strings.LastIndex(dirStr, "/")]
				}
				if err := updateBackupDir(dirStr); err != nil {
					upMap["vars"] = backup.Vars
					_ = backupRepo.Update(req.ID, upMap)
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
	varMap["type"] = backup.Type
	if backup.Type == "LOCAL" {
		return nil, errors.New("not support")
	}
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp:
		varMap["username"] = backup.AccessKey
		varMap["password"] = backup.Credential
	case constant.OSS, constant.S3, constant.MinIo:
		varMap["accessKey"] = backup.AccessKey
		varMap["secretKey"] = backup.Credential
	}

	backClient, err := cloud_storage.NewCloudStorageClient(varMap)
	if err != nil {
		return nil, err
	}

	return backClient, nil
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

func updateBackupDir(dir string) error {
	oldDir := global.CONF.System.Backup
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	if strings.HasSuffix(oldDir, "/") {
		oldDir = oldDir[:strings.LastIndex(oldDir, "/")]
	}
	stdout, err := cmd.Execf("cp -r %s/* %s", oldDir, dir)
	if err != nil {
		return errors.New(string(stdout))
	}
	global.CONF.System.Backup = dir
	return nil
}
