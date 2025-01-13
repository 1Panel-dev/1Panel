package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cloud_storage"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupRecordService struct{}

type IBackupRecordService interface {
	SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error)
	SearchRecordsByCronjobWithPage(search dto.RecordSearchByCronjob) (int64, []dto.BackupRecords, error)
	DownloadRecord(info dto.DownloadRecord) (string, error)
	DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error
	BatchDeleteRecord(ids []uint) error
	ListAppRecords(name, detailName, fileName string) ([]model.BackupRecord, error)

	ListFiles(req dto.OperateByID) []string
}

func NewIBackupRecordService() IBackupRecordService {
	return &BackupRecordService{}
}

func (u *BackupRecordService) SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error) {
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

func (u *BackupRecordService) SearchRecordsByCronjobWithPage(search dto.RecordSearchByCronjob) (int64, []dto.BackupRecords, error) {
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

func (u *BackupRecordService) DownloadRecord(info dto.DownloadRecord) (string, error) {
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

func (u *BackupRecordService) DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error {
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

func (u *BackupRecordService) BatchDeleteRecord(ids []uint) error {
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

func (u *BackupRecordService) ListAppRecords(name, detailName, fileName string) ([]model.BackupRecord, error) {
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

func (u *BackupRecordService) ListFiles(req dto.OperateByID) []string {
	var datas []string
	_, client, err := NewBackupClientWithID(req.ID)
	if err != nil {
		return datas
	}
	prefix := "system_snapshot"
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

func (u *BackupRecordService) loadRecordSize(records []model.BackupRecord) ([]dto.BackupRecords, error) {
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

type loadSizeHelper struct {
	isOk       bool
	backupName string
	backupPath string
	client     cloud_storage.CloudStorageClient
}
