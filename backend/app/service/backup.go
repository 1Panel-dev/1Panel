package service

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage/client"
	fileUtils "github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupService struct{}

type IBackupService interface {
	List() ([]dto.BackupInfo, error)
	SearchRecordsWithPage(search dto.RecordSearch) (int64, []dto.BackupRecords, error)
	LoadSize(req dto.RecordSearch) ([]dto.BackupFile, error)
	SearchRecordsByCronjobWithPage(search dto.RecordSearchByCronjob) (int64, []dto.BackupRecords, error)
	LoadSizeByCronjob(req dto.RecordSearchByCronjob) ([]dto.BackupFile, error)
	LoadOneDriveInfo() (dto.OneDriveInfo, error)
	DownloadRecord(info dto.DownloadRecord) (string, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
	Update(ireq dto.BackupOperate) error
	Delete(id uint) error
	DeleteRecordByName(backupType, name, detailName string, withDeleteFile bool) error
	BatchDeleteRecord(ids []uint) error
	NewClient(backup *model.BackupAccount) (cloud_storage.CloudStorageClient, error)
	ListAppRecords(name, detailName, fileName string) ([]model.BackupRecord, error)

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

	Run()
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
	dtobas = append(dtobas, u.loadByType("WebDAV", ops))
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
	if err != nil {
		return 0, nil, err
	}

	var list []dto.BackupRecords
	for _, item := range records {
		var itemRecord dto.BackupRecords
		if err := copier.Copy(&itemRecord, &item); err != nil {
			continue
		}
		list = append(list, itemRecord)
	}
	return total, list, err
}

func (u *BackupService) LoadSize(req dto.RecordSearch) ([]dto.BackupFile, error) {
	_, records, err := backupRepo.PageRecord(
		req.Page, req.PageSize,
		commonRepo.WithOrderBy("created_at desc"),
		commonRepo.WithByName(req.Name),
		commonRepo.WithByType(req.Type),
		backupRepo.WithByDetailName(req.DetailName),
	)
	if err != nil {
		return nil, err
	}
	return u.loadRecordSize(records)
}

func (u *BackupService) ListAppRecords(name, detailName, fileName string) ([]model.BackupRecord, error) {
	records, err := backupRepo.ListRecord(
		commonRepo.WithOrderBy("created_at asc"),
		commonRepo.WithByName(name),
		commonRepo.WithByType("app"),
		backupRepo.WithFileNameStartWith(fileName),
		backupRepo.WithByDetailName(detailName),
	)
	if err != nil {
		return nil, err
	}
	return records, err
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

	var list []dto.BackupRecords
	for _, item := range records {
		var itemRecord dto.BackupRecords
		if err := copier.Copy(&itemRecord, &item); err != nil {
			continue
		}
		list = append(list, itemRecord)
	}
	return total, list, err
}

func (u *BackupService) LoadSizeByCronjob(req dto.RecordSearchByCronjob) ([]dto.BackupFile, error) {
	_, records, err := backupRepo.PageRecord(
		req.Page, req.PageSize,
		commonRepo.WithOrderBy("created_at desc"),
		backupRepo.WithByCronID(req.CronjobID),
	)
	if err != nil {
		return nil, err
	}
	return u.loadRecordSize(records)
}

type loadSizeHelper struct {
	isOk       bool
	backupPath string
	client     cloud_storage.CloudStorageClient
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

func (u *BackupService) DownloadRecord(info dto.DownloadRecord) (string, error) {
	if info.Source == "LOCAL" {
		localDir, err := loadLocalDir()
		if err != nil {
			return "", err
		}
		return path.Join(localDir, info.FileDir, info.FileName), nil
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

func (u *BackupService) Create(req dto.BackupOperate) error {
	backup, _ := backupRepo.Get(commonRepo.WithByType(req.Type))
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
		StartRefreshOneDriveToken()
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
	if backup.Type == constant.OneDrive {
		global.Cron.Remove(global.OneDriveCronID)
	}
	cronjobs, _ := cronjobRepo.List(cronjobRepo.WithByDefaultDownload(backup.Type))
	if len(cronjobs) != 0 {
		return buserr.New(constant.ErrBackupInUsed)
	}
	return backupRepo.Delete(commonRepo.WithByID(id))
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
		if err := u.loadAccessToken(&backup); err != nil {
			return err
		}
		upMap["credential"] = backup.Credential
		upMap["vars"] = backup.Vars
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
				global.CONF.System.Backup = dirStr
			}
		}
	}
	return nil
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

func (u *BackupService) loadByType(accountType string, accounts []model.BackupAccount) dto.BackupInfo {
	for _, account := range accounts {
		if account.Type == accountType {
			var item dto.BackupInfo
			if err := copier.Copy(&item, &account); err != nil {
				global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
			}
			if account.Type == constant.OneDrive {
				varMap := make(map[string]interface{})
				if err := json.Unmarshal([]byte(item.Vars), &varMap); err != nil {
					return dto.BackupInfo{Type: accountType}
				}
				delete(varMap, "refresh_token")
				itemVars, _ := json.Marshal(varMap)
				item.Vars = string(itemVars)
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

func (u *BackupService) loadRecordSize(records []model.BackupRecord) ([]dto.BackupFile, error) {
	var datas []dto.BackupFile
	clientMap := make(map[string]loadSizeHelper)
	var wg sync.WaitGroup
	for i := 0; i < len(records); i++ {
		var item dto.BackupFile
		item.ID = records[i].ID
		item.Name = records[i].FileName
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

func (u *BackupService) checkBackupConn(backup *model.BackupAccount) (bool, error) {
	client, err := u.NewClient(backup)
	if err != nil {
		return false, err
	}
	fileItem := path.Join(global.CONF.System.TmpDir, "test", "1panel")
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

func StartRefreshOneDriveToken() {
	service := NewIBackupService()
	oneDriveCronID, err := global.Cron.AddJob("0 3 */31 * *", service)
	if err != nil {
		global.LOG.Errorf("can not add OneDrive corn job: %s", err.Error())
		return
	}
	global.OneDriveCronID = oneDriveCronID
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
