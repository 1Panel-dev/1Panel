package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupService struct{}

type IBackupService interface {
	CheckUsed(id uint) error
	Sync(req dto.SyncFromMaster) error

	LoadBackupOptions() ([]dto.BackupOption, error)

	SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error)
	SearchRecordsByCronjobWithPage(search dto.RecordSearchByCronjob) (int64, []dto.BackupRecords, error)
	DownloadRecord(info dto.DownloadRecord) (string, error)
	DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error
	BatchDeleteRecord(ids []uint) error
	ListAppRecords(name, detailName, fileName string) ([]model.BackupRecord, error)

	ListFiles(req dto.OperateByID) []string

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

func (u *BackupService) Sync(req dto.SyncFromMaster) error {
	var accountItem model.BackupAccount
	if err := json.Unmarshal([]byte(req.Data), &accountItem); err != nil {
		return err
	}
	accountItem.AccessKey, _ = encrypt.StringEncryptWithBase64(accountItem.AccessKey)
	accountItem.Credential, _ = encrypt.StringEncryptWithBase64(accountItem.Credential)
	account, _ := backupRepo.Get(repo.WithByName(req.Name))
	switch req.Operation {
	case "create":
		if account.ID != 0 {
			accountItem.ID = account.ID
			return backupRepo.Save(&accountItem)
		}
		return backupRepo.Create(&accountItem)
	case "delete":
		if account.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return backupRepo.Delete(repo.WithByID(account.ID))
	case "update":
		if account.ID == 0 {
			return constant.ErrRecordNotFound
		}
		accountItem.ID = account.ID
		return backupRepo.Save(&accountItem)
	default:
		return fmt.Errorf("not support such operation %s", req.Operation)
	}
}

func (u *BackupService) LoadBackupOptions() ([]dto.BackupOption, error) {
	accounts, err := backupRepo.List(repo.WithOrderBy("created_at desc"))
	if err != nil {
		return nil, err
	}
	var data []dto.BackupOption
	for _, account := range accounts {
		var item dto.BackupOption
		if err := copier.Copy(&item, &account); err != nil {
			global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
		}
		data = append(data, item)
	}
	return data, nil
}

func (u *BackupService) SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error) {
	total, records, err := backupRepo.PageRecord(
		search.Page, search.PageSize,
		repo.WithOrderBy("created_at desc"),
		repo.WithByName(search.Name),
		repo.WithByType(search.Type),
		repo.WithByDetailName(search.DetailName),
	)
	if err != nil {
		return 0, nil, err
	}

	if total == 0 {
		return 0, nil, nil
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
		repo.WithOrderBy("created_at desc"),
		backupRepo.WithByCronID(search.CronjobID),
	)
	if err != nil {
		return 0, nil, err
	}

	if total == 0 {
		return 0, nil, nil
	}
	datas, err := u.loadRecordSize(records)
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].CreatedAt.After(datas[j].CreatedAt)
	})
	return total, datas, err
}

func (u *BackupService) CheckUsed(id uint) error {
	cronjobs, _ := cronjobRepo.List()
	for _, job := range cronjobs {
		if job.DownloadAccountID == id {
			return buserr.New(constant.ErrBackupInUsed)
		}
		ids := strings.Split(job.SourceAccountIDs, ",")
		for _, idItem := range ids {
			if idItem == fmt.Sprintf("%v", id) {
				return buserr.New(constant.ErrBackupInUsed)
			}
		}
	}
	return nil
}

type loadSizeHelper struct {
	isOk       bool
	backupName string
	backupPath string
	client     cloud_storage.CloudStorageClient
}

func (u *BackupService) DownloadRecord(info dto.DownloadRecord) (string, error) {
	account, client, err := NewBackupClientWithID(info.DownloadAccountID)
	if err != nil {
		return "", fmt.Errorf("new cloud storage client failed, err: %v", err)
	}
	if account.Type == "LOCAL" {
		return path.Join(global.CONF.System.Backup, info.FileDir, info.FileName), nil
	}
	targetPath := fmt.Sprintf("%s/download/%s/%s", constant.DataDir, info.FileDir, info.FileName)
	if _, err := os.Stat(path.Dir(targetPath)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(targetPath), os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", path.Dir(targetPath), err)
		}
	}
	srcPath := fmt.Sprintf("%s/%s", info.FileDir, info.FileName)
	if len(account.BackupPath) != 0 {
		srcPath = path.Join(strings.TrimPrefix(account.BackupPath, "/"), srcPath)
	}
	if exist, _ := client.Exist(srcPath); exist {
		isOK, err := client.Download(srcPath, targetPath)
		if !isOK {
			return "", fmt.Errorf("cloud storage download failed, err: %v", err)
		}
	}
	return targetPath, nil
}

func (u *BackupService) DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error {
	if !withDeleteFile {
		return backupRepo.DeleteRecord(context.Background(), repo.WithByType(backupType), repo.WithByName(name), repo.WithByDetailName(detailName))
	}

	records, err := backupRepo.ListRecord(repo.WithByType(backupType), repo.WithByName(name), repo.WithByDetailName(detailName))
	if err != nil {
		return err
	}

	for _, record := range records {
		_, client, err := NewBackupClientWithID(record.DownloadAccountID)
		if err != nil {
			global.LOG.Errorf("new client for backup account failed, err: %v", err)
			continue
		}
		if _, err = client.Delete(path.Join(record.FileDir, record.FileName)); err != nil {
			global.LOG.Errorf("remove file %s failed, err: %v", path.Join(record.FileDir, record.FileName), err)
		}
		_ = backupRepo.DeleteRecord(context.Background(), repo.WithByID(record.ID))
	}
	return nil
}

func (u *BackupService) BatchDeleteRecord(ids []uint) error {
	records, err := backupRepo.ListRecord(repo.WithByIDs(ids))
	if err != nil {
		return err
	}
	for _, record := range records {
		_, client, err := NewBackupClientWithID(record.DownloadAccountID)
		if err != nil {
			global.LOG.Errorf("new client for backup account failed, err: %v", err)
			continue
		}
		if _, err = client.Delete(path.Join(record.FileDir, record.FileName)); err != nil {
			global.LOG.Errorf("remove file %s failed, err: %v", path.Join(record.FileDir, record.FileName), err)
		}
	}
	return backupRepo.DeleteRecord(context.Background(), repo.WithByIDs(ids))
}

func (u *BackupService) ListAppRecords(name, detailName, fileName string) ([]model.BackupRecord, error) {
	records, err := backupRepo.ListRecord(
		repo.WithOrderBy("created_at asc"),
		repo.WithByName(name),
		repo.WithByType("app"),
		backupRepo.WithFileNameStartWith(fileName),
		backupRepo.WithByDetailName(detailName),
	)
	if err != nil {
		return nil, err
	}
	return records, err
}

func (u *BackupService) ListFiles(req dto.OperateByID) []string {
	var datas []string
	account, client, err := NewBackupClientWithID(req.ID)
	if err != nil {
		return datas
	}
	prefix := "system_snapshot"
	if len(account.BackupPath) != 0 {
		prefix = path.Join(strings.TrimPrefix(account.BackupPath, "/"), prefix)
	}
	files, err := client.ListObjects(prefix)
	if err != nil {
		global.LOG.Debugf("load files failed, err: %v", err)
		return datas
	}
	for _, file := range files {
		if len(file) != 0 {
			datas = append(datas, path.Base(file))
		}
	}
	return datas
}

func (u *BackupService) loadRecordSize(records []model.BackupRecord) ([]dto.BackupRecords, error) {
	recordMap := make(map[uint]struct{})
	var recordIds []string
	for _, record := range records {
		if _, ok := recordMap[record.DownloadAccountID]; !ok {
			recordMap[record.DownloadAccountID] = struct{}{}
			recordIds = append(recordIds, fmt.Sprintf("%v", record.DownloadAccountID))
		}
	}
	clientMap, err := NewBackupClientMap(recordIds)
	if err != nil {
		return nil, err
	}

	var datas []dto.BackupRecords
	var wg sync.WaitGroup
	for i := 0; i < len(records); i++ {
		var item dto.BackupRecords
		if err := copier.Copy(&item, &records[i]); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}

		itemPath := path.Join(records[i].FileDir, records[i].FileName)
		if val, ok := clientMap[fmt.Sprintf("%v", records[i].DownloadAccountID)]; ok {
			item.AccountName = val.name
			item.AccountType = val.accountType
			item.DownloadAccountID = val.id
			wg.Add(1)
			go func(index int) {
				item.Size, _ = val.client.Size(path.Join(strings.TrimLeft(val.backupPath, "/"), itemPath))
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

func NewBackupClientWithID(id uint) (*model.BackupAccount, cloud_storage.CloudStorageClient, error) {
	var account model.BackupAccount
	if global.IsMaster {
		var setting model.Setting
		if err := global.CoreDB.Where("key = ?", "EncryptKey").First(&setting).Error; err != nil {
			return nil, nil, err
		}
		if err := global.CoreDB.Where("id = ?", id).First(&account).Error; err != nil {
			return nil, nil, err
		}
		if account.ID == 0 {
			return nil, nil, constant.ErrRecordNotFound
		}
		account.AccessKey, _ = encrypt.StringDecryptWithKey(account.AccessKey, setting.Value)
		account.Credential, _ = encrypt.StringDecryptWithKey(account.Credential, setting.Value)
	} else {
		account, _ = backupRepo.Get(repo.WithByID(id))
	}
	backClient, err := newClient(&account)
	if err != nil {
		return nil, nil, err
	}
	return &account, backClient, nil
}

type backupClientHelper struct {
	id          uint
	accountType string
	name        string
	backupPath  string
	client      cloud_storage.CloudStorageClient
}

func NewBackupClientMap(ids []string) (map[string]backupClientHelper, error) {
	var accounts []model.BackupAccount
	if global.IsMaster {
		var setting model.Setting
		if err := global.CoreDB.Where("key = ?", "EncryptKey").First(&setting).Error; err != nil {
			return nil, err
		}
		if err := global.CoreDB.Where("id in (?)", ids).Find(&accounts).Error; err != nil {
			return nil, err
		}
		if len(accounts) == 0 {
			return nil, constant.ErrRecordNotFound
		}
		for i := 0; i < len(accounts); i++ {
			accounts[i].AccessKey, _ = encrypt.StringDecryptWithKey(accounts[i].AccessKey, setting.Value)
			accounts[i].Credential, _ = encrypt.StringDecryptWithKey(accounts[i].Credential, setting.Value)
		}
	} else {
		var idItems []uint
		for i := 0; i < len(ids); i++ {
			item, _ := strconv.Atoi(ids[i])
			idItems = append(idItems, uint(item))
		}
		accounts, _ = backupRepo.List(repo.WithByIDs(idItems))
	}
	clientMap := make(map[string]backupClientHelper)
	for _, item := range accounts {
		if !global.IsMaster {
			accessItem, err := base64.StdEncoding.DecodeString(item.AccessKey)
			if err != nil {
				return nil, err
			}
			item.AccessKey = string(accessItem)
			secretItem, err := base64.StdEncoding.DecodeString(item.Credential)
			if err != nil {
				return nil, err
			}
			item.Credential = string(secretItem)
		}
		backClient, err := newClient(&item)
		if err != nil {
			return nil, err
		}
		pathItem := item.BackupPath
		if item.BackupPath != "/" {
			pathItem = strings.TrimPrefix(item.BackupPath, "/")
		}
		clientMap[fmt.Sprintf("%v", item.ID)] = backupClientHelper{
			client:      backClient,
			backupPath:  pathItem,
			name:        item.Name,
			accountType: item.Type,
			id:          item.ID,
		}
	}
	return clientMap, nil
}

func newClient(account *model.BackupAccount) (cloud_storage.CloudStorageClient, error) {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(account.Vars), &varMap); err != nil {
		return nil, err
	}
	varMap["bucket"] = account.Bucket
	switch account.Type {
	case constant.Sftp, constant.WebDAV:
		varMap["username"] = account.AccessKey
		varMap["password"] = account.Credential
	case constant.OSS, constant.S3, constant.MinIo, constant.Cos, constant.Kodo:
		varMap["accessKey"] = account.AccessKey
		varMap["secretKey"] = account.Credential
	case constant.UPYUN:
		varMap["operator"] = account.AccessKey
		varMap["password"] = account.Credential
	}

	client, err := cloud_storage.NewCloudStorageClient(account.Type, varMap)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func LoadLocalDirByStr(vars string) (string, error) {
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
			return baseDir, nil
		}
		return baseDir, nil
	}
	return "", fmt.Errorf("error type dir: %T", varMap["dir"])
}
