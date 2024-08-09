package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	fileUtils "github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupService struct{}

type IBackupService interface {
	Operate(req dto.BackupOperate) error

	SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error)
	SearchRecordsByCronjobWithPage(search dto.RecordSearchByCronjob) (int64, []dto.BackupRecords, error)
	DownloadRecord(info dto.DownloadRecord) (string, error)
	DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error
	BatchDeleteRecord(ids []uint) error
	NewClient(backup *model.BackupAccount) (cloud_storage.CloudStorageClient, error)

	ListFiles(req dto.BackupSearchFile) []string

	MysqlBackup(db dto.CommonBackup) error
	PostgresqlBackup(db dto.CommonBackup) error
	MysqlRecover(db dto.CommonRecover) error
	PostgresqlRecover(db dto.CommonRecover) error
	MysqlRecoverByUpload(req dto.CommonRecover) error
	PostgresqlRecoverByUpload(req dto.CommonRecover) error

	RedisBackup(db dto.CommonBackup) error
	RedisRecover(db dto.CommonRecover) error

	WebsiteBackup(db dto.CommonBackup) error
	WebsiteRecover(req dto.CommonRecover) error

	AppBackup(db dto.CommonBackup) (*model.BackupRecord, error)
	AppRecover(req dto.CommonRecover) error
}

func NewIBackupService() IBackupService {
	return &BackupService{}
}

func (u *BackupService) SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error) {
	total, records, err := backupRepo.PageRecord(
		search.Page, search.PageSize,
		commonRepo.WithOrderBy("created_at desc"),
		commonRepo.WithByName(search.Name),
		commonRepo.WithByType(search.Type),
		backupRepo.WithByDetailName(search.DetailName),
	)
	if err != nil {
		return 0, nil, err
	}

	datas, err := u.loadRecordSize(records)
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].CreatedAt.After(datas[j].CreatedAt)
	})
	return total, datas, err
}

func (u *BackupService) SearchRecordsByCronjobWithPage(search dto.RecordSearchByCronjob) (int64, []dto.BackupRecords, error) {
	total, records, err := backupRepo.PageRecord(
		search.Page, search.PageSize,
		commonRepo.WithOrderBy("created_at desc"),
		backupRepo.WithByCronID(search.CronjobID),
	)
	if err != nil {
		return 0, nil, err
	}

	datas, err := u.loadRecordSize(records)
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].CreatedAt.After(datas[j].CreatedAt)
	})
	return total, datas, err
}

type loadSizeHelper struct {
	isOk       bool
	backupPath string
	client     cloud_storage.CloudStorageClient
}

func (u *BackupService) DownloadRecord(info dto.DownloadRecord) (string, error) {
	backup, _ := backupRepo.Get(commonRepo.WithByType(info.Source))
	if backup.ID == 0 {
		return "", constant.ErrRecordNotFound
	}
	if info.Source == "LOCAL" {
		localDir, err := loadLocalDir()
		if err != nil {
			return "", err
		}
		return path.Join(localDir, info.FileDir, info.FileName), nil
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return "", err
	}
	varMap["bucket"] = backup.Bucket
	switch backup.Type {
	case constant.Sftp, constant.WebDAV:
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
		srcPath = path.Join(strings.TrimPrefix(backup.BackupPath, "/"), srcPath)
	}
	if exist, _ := backClient.Exist(srcPath); exist {
		isOK, err := backClient.Download(srcPath, targetPath)
		if !isOK {
			return "", fmt.Errorf("cloud storage download failed, err: %v", err)
		}
	}
	return targetPath, nil
}

func (u *BackupService) Operate(req dto.BackupOperate) error {
	for i := 0; i < len(req.Data); i++ {
		encryptKeyItem, err := encrypt.StringEncryptWithBase64(req.Data[i].AccessKey)
		if err != nil {
			return err
		}
		req.Data[i].AccessKey = encryptKeyItem
		encryptCredentialItem, err := encrypt.StringEncryptWithBase64(req.Data[i].Credential)
		if err != nil {
			return err
		}
		req.Data[i].Credential = encryptCredentialItem
	}
	if req.Operate == "add" {
		return backupRepo.Create(req.Data)
	}
	if req.Operate == "remove" {
		var names []string
		for _, item := range req.Data {
			names = append(names, item.Name)
		}
		return backupRepo.Delete(commonRepo.WithNamesIn(names))
	}
	global.LOG.Debug("走到了这里")
	for _, item := range req.Data {
		local, _ := backupRepo.Get(commonRepo.WithByName(item.Name))
		if local.ID == 0 {
			if err := backupRepo.Create([]model.BackupAccount{item}); err != nil {
				return err
			}
			continue
		}
		if item.Type == constant.Local {
			if local.ID != 0 && item.Vars != local.Vars {
				oldPath, err := loadLocalDirByStr(local.Vars)
				if err != nil {
					return err
				}
				newPath, err := loadLocalDirByStr(item.Vars)
				if err != nil {
					return err
				}
				if strings.HasSuffix(newPath, "/") && newPath != "/" {
					newPath = newPath[:strings.LastIndex(newPath, "/")]
				}
				if err := copyDir(oldPath, newPath); err != nil {
					return err
				}
				global.CONF.System.Backup = newPath
			}
		}
		item.ID = local.ID

		global.LOG.Debug("走到了这里111")
		if err := backupRepo.Save(&item); err != nil {
			return err
		}
	}
	return nil
}

func (u *BackupService) DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error {
	if !withDeleteFile {
		return backupRepo.DeleteRecord(context.Background(), commonRepo.WithByType(backupType), commonRepo.WithByName(name), backupRepo.WithByDetailName(detailName))
	}

	records, err := backupRepo.ListRecord(commonRepo.WithByType(backupType), commonRepo.WithByName(name), backupRepo.WithByDetailName(detailName))
	if err != nil {
		return err
	}

	for _, record := range records {
		backupAccount, err := backupRepo.Get(commonRepo.WithByType(record.Source))
		if err != nil {
			global.LOG.Errorf("load backup account %s info from db failed, err: %v", record.Source, err)
			continue
		}
		client, err := u.NewClient(&backupAccount)
		if err != nil {
			global.LOG.Errorf("new client for backup account %s failed, err: %v", record.Source, err)
			continue
		}
		if _, err = client.Delete(path.Join(record.FileDir, record.FileName)); err != nil {
			global.LOG.Errorf("remove file %s from %s failed, err: %v", path.Join(record.FileDir, record.FileName), record.Source, err)
		}
		_ = backupRepo.DeleteRecord(context.Background(), commonRepo.WithByID(record.ID))
	}
	return nil
}

func (u *BackupService) BatchDeleteRecord(ids []uint) error {
	records, err := backupRepo.ListRecord(commonRepo.WithIdsIn(ids))
	if err != nil {
		return err
	}
	for _, record := range records {
		backupAccount, err := backupRepo.Get(commonRepo.WithByType(record.Source))
		if err != nil {
			global.LOG.Errorf("load backup account %s info from db failed, err: %v", record.Source, err)
			continue
		}
		client, err := u.NewClient(&backupAccount)
		if err != nil {
			global.LOG.Errorf("new client for backup account %s failed, err: %v", record.Source, err)
			continue
		}
		if _, err = client.Delete(path.Join(record.FileDir, record.FileName)); err != nil {
			global.LOG.Errorf("remove file %s from %s failed, err: %v", path.Join(record.FileDir, record.FileName), record.Source, err)
		}
	}
	return backupRepo.DeleteRecord(context.Background(), commonRepo.WithIdsIn(ids))
}

func (u *BackupService) ListFiles(req dto.BackupSearchFile) []string {
	var datas []string
	backup, err := backupRepo.Get(backupRepo.WithByType(req.Type))
	if err != nil {
		return datas
	}
	client, err := u.NewClient(&backup)
	if err != nil {
		return datas
	}
	prefix := "system_snapshot"
	if len(backup.BackupPath) != 0 {
		prefix = path.Join(strings.TrimPrefix(backup.BackupPath, "/"), prefix)
	}
	files, err := client.ListObjects(prefix)
	if err != nil {
		global.LOG.Debugf("load files from %s failed, err: %v", req.Type, err)
		return datas
	}
	for _, file := range files {
		if len(file) != 0 {
			datas = append(datas, path.Base(file))
		}
	}
	return datas
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

func (u *BackupService) loadRecordSize(records []model.BackupRecord) ([]dto.BackupRecords, error) {
	var datas []dto.BackupRecords
	clientMap := make(map[string]loadSizeHelper)
	var wg sync.WaitGroup
	for i := 0; i < len(records); i++ {
		var item dto.BackupRecords
		if err := copier.Copy(&item, &records[i]); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		itemPath := path.Join(records[i].FileDir, records[i].FileName)
		if _, ok := clientMap[records[i].Source]; !ok {
			backup, err := backupRepo.Get(commonRepo.WithByType(records[i].Source))
			if err != nil {
				global.LOG.Errorf("load backup model %s from db failed, err: %v", records[i].Source, err)
				clientMap[records[i].Source] = loadSizeHelper{}
				datas = append(datas, item)
				continue
			}
			client, err := u.NewClient(&backup)
			if err != nil {
				global.LOG.Errorf("load backup client %s from db failed, err: %v", records[i].Source, err)
				clientMap[records[i].Source] = loadSizeHelper{}
				datas = append(datas, item)
				continue
			}
			item.Size, _ = client.Size(path.Join(strings.TrimLeft(backup.BackupPath, "/"), itemPath))
			datas = append(datas, item)
			clientMap[records[i].Source] = loadSizeHelper{backupPath: strings.TrimLeft(backup.BackupPath, "/"), client: client, isOk: true}
			continue
		}
		if clientMap[records[i].Source].isOk {
			wg.Add(1)
			go func(index int) {
				item.Size, _ = clientMap[records[index].Source].client.Size(path.Join(clientMap[records[index].Source].backupPath, itemPath))
				datas = append(datas, item)
				wg.Done()
			}(i)
		} else {
			datas = append(datas, item)
		}
	}
	wg.Wait()
	return datas, nil
}

func loadLocalDir() (string, error) {
	backup, err := backupRepo.Get(commonRepo.WithByType("LOCAL"))
	if err != nil {
		return "", err
	}
	return loadLocalDirByStr(backup.Vars)
}

func loadLocalDirByStr(vars string) (string, error) {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(vars), &varMap); err != nil {
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
