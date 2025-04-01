package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/jinzhu/copier"
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
	LoadRecordSize(req dto.SearchForSize) ([]dto.RecordFileSize, error)
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
	accounts, _ := backupRepo.List()
	var data []dto.BackupRecords
	for _, record := range records {
		var item dto.BackupRecords
		if err := copier.Copy(&item, &record); err != nil {
			global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
		}
		for _, account := range accounts {
			if account.ID == record.DownloadAccountID {
				item.DownloadAccountID = account.ID
				item.AccountName = account.Name
				item.AccountType = account.Type
				break
			}
		}
		data = append(data, item)
	}
	return total, data, err
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
	accounts, _ := backupRepo.List()
	var data []dto.BackupRecords
	for _, record := range records {
		var item dto.BackupRecords
		if err := copier.Copy(&item, &record); err != nil {
			global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
		}
		for _, account := range accounts {
			if account.ID == record.DownloadAccountID {
				item.DownloadAccountID = account.ID
				item.AccountName = account.Name
				item.AccountType = account.Type
				break
			}
		}
		data = append(data, item)
	}
	return total, data, err
}

func (u *BackupRecordService) DownloadRecord(info dto.DownloadRecord) (string, error) {
	account, client, err := NewBackupClientWithID(info.DownloadAccountID)
	if err != nil {
		return "", fmt.Errorf("new cloud storage client failed, err: %v", err)
	}
	if account.Type == "LOCAL" {
		return path.Join(global.Dir.LocalBackupDir, info.FileDir, info.FileName), nil
	}
	targetPath := fmt.Sprintf("%s/download/%s/%s", global.Dir.DataDir, info.FileDir, info.FileName)
	if _, err := os.Stat(path.Dir(targetPath)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(targetPath), os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", path.Dir(targetPath), err)
		}
	}
	srcPath := fmt.Sprintf("%s/%s", info.FileDir, info.FileName)
	if len(account.BackupPath) != 0 {
		srcPath = path.Join(account.BackupPath, srcPath)
	}
	if exist, _ := client.Exist(srcPath); exist {
		isOK, err := client.Download(srcPath, targetPath)
		if !isOK || err != nil {
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
	backupItem, client, err := NewBackupClientWithID(req.ID)
	if err != nil {
		return []string{}
	}
	prefix := "system_snapshot"
	if len(backupItem.BackupPath) != 0 {
		prefix = path.Join(backupItem.BackupPath, prefix)
	}
	files, err := client.ListObjects(prefix)
	if err != nil {
		global.LOG.Debugf("load files failed, err: %v", err)
		return []string{}
	}
	return files
}

type backupSizeHelper struct {
	ID         uint   `json:"id"`
	DownloadID uint   `json:"downloadID"`
	FilePath   string `json:"filePath"`
	Size       uint   `json:"size"`
}

func (u *BackupRecordService) LoadRecordSize(req dto.SearchForSize) ([]dto.RecordFileSize, error) {
	var list []backupSizeHelper
	switch req.Type {
	case "snapshot":
		_, records, err := snapshotRepo.Page(req.Page, req.PageSize, repo.WithByLikeName(req.Info), repo.WithOrderRuleBy(req.OrderBy, req.Order))
		if err != nil {
			return nil, err
		}
		for _, item := range records {
			list = append(list, backupSizeHelper{ID: item.ID, DownloadID: item.DownloadAccountID, FilePath: fmt.Sprintf("system_snapshot/%s.tar.gz", item.Name)})
		}
	case "cronjob":
		_, records, err := backupRepo.PageRecord(req.Page, req.PageSize, repo.WithOrderBy("created_at desc"), backupRepo.WithByCronID(req.CronjobID))
		if err != nil {
			return nil, err
		}
		for _, item := range records {
			list = append(list, backupSizeHelper{ID: item.ID, DownloadID: item.DownloadAccountID, FilePath: path.Join(item.FileDir, item.FileName)})
		}
	default:
		_, records, err := backupRepo.PageRecord(
			req.Page, req.PageSize,
			repo.WithOrderBy("created_at desc"),
			repo.WithByName(req.Name),
			repo.WithByType(req.Type),
			repo.WithByDetailName(req.DetailName),
		)
		if err != nil {
			return nil, err
		}
		for _, item := range records {
			list = append(list, backupSizeHelper{ID: item.ID, DownloadID: item.DownloadAccountID, FilePath: path.Join(item.FileDir, item.FileName)})
		}
	}
	recordMap := make(map[uint]struct{})
	var recordIds []string
	for _, record := range list {
		if _, ok := recordMap[record.DownloadID]; !ok {
			recordMap[record.DownloadID] = struct{}{}
			recordIds = append(recordIds, fmt.Sprintf("%v", record.DownloadID))
		}
	}
	clientMap, err := NewBackupClientMap(recordIds)
	if err != nil {
		return nil, err
	}

	var datas []dto.RecordFileSize
	var wg sync.WaitGroup
	for i := 0; i < len(list); i++ {
		datas = append(datas, dto.RecordFileSize{ID: list[i].ID})
		if val, ok := clientMap[fmt.Sprintf("%v", list[i].DownloadID)]; ok {
			wg.Add(1)
			go func(index int) {
				datas[index].Size, _ = val.client.Size(path.Join(val.backupPath, list[i].FilePath))
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	return datas, nil
}
