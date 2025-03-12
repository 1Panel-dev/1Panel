package service

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/agent/utils/cloud_storage/client"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/jinzhu/copier"
)

type BackupService struct{}

type IBackupService interface {
	CheckUsed(name string, isPublic bool) error
	Sync(req dto.SyncFromMaster) error

	LoadBackupOptions() ([]dto.BackupOption, error)
	SearchWithPage(search dto.SearchPageWithType) (int64, interface{}, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
	Update(req dto.BackupOperate) error
	Delete(id uint) error
	RefreshToken(req dto.OperateByID) error
	GetLocalDir() (string, error)

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

func (u *BackupService) GetLocalDir() (string, error) {
	account, err := backupRepo.Get(repo.WithByType(constant.Local))
	if err != nil {
		return "", err
	}
	return account.BackupPath, nil
}

func (u *BackupService) SearchWithPage(req dto.SearchPageWithType) (int64, interface{}, error) {
	options := []repo.DBOption{repo.WithOrderBy("created_at desc")}
	if len(req.Type) != 0 {
		options = append(options, repo.WithByType(req.Type))
	}
	if len(req.Info) != 0 {
		options = append(options, repo.WithByType(req.Info))
	}
	count, accounts, err := backupRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	var data []dto.BackupInfo
	for _, account := range accounts {
		var item dto.BackupInfo
		if err := copier.Copy(&item, &account); err != nil {
			global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
		}
		if item.Type != constant.Sftp && item.Type != constant.Local {
			item.BackupPath = path.Join("/", strings.TrimPrefix(item.BackupPath, "/"))
		}
		if !item.RememberAuth {
			item.AccessKey = ""
			item.Credential = ""
			if account.Type == constant.Sftp {
				varMap := make(map[string]interface{})
				if err := json.Unmarshal([]byte(item.Vars), &varMap); err != nil {
					continue
				}
				delete(varMap, "passPhrase")
				itemVars, _ := json.Marshal(varMap)
				item.Vars = string(itemVars)
			}
		} else {
			item.AccessKey, _ = encrypt.StringDecryptWithBase64(item.AccessKey)
			item.Credential, _ = encrypt.StringDecryptWithBase64(item.Credential)
		}

		if account.Type == constant.OneDrive || account.Type == constant.ALIYUN || account.Type == constant.GoogleDrive {
			varMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(item.Vars), &varMap); err != nil {
				continue
			}
			delete(varMap, "refresh_token")
			delete(varMap, "drive_id")
			itemVars, _ := json.Marshal(varMap)
			item.Vars = string(itemVars)
		}
		data = append(data, item)
	}
	return count, data, nil
}

func (u *BackupService) Create(req dto.BackupOperate) error {
	if req.Type == constant.Local {
		return buserr.New("ErrBackupLocalCreate")
	}
	if req.Type != constant.Sftp && req.BackupPath != "/" {
		req.BackupPath = strings.TrimPrefix(req.BackupPath, "/")
	}
	backup, _ := backupRepo.Get(repo.WithByName(req.Name))
	if backup.ID != 0 {
		return buserr.New("ErrRecordExist")
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
		if err := loadRefreshTokenByCode(&backup); err != nil {
			return err
		}
	}
	if req.Type != constant.Local {
		isOk, err := u.checkBackupConn(&backup)
		if err != nil || !isOk {
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

func (u *BackupService) Delete(id uint) error {
	backup, _ := backupRepo.Get(repo.WithByID(id))
	if backup.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if backup.Type == constant.Local {
		return buserr.New("ErrBackupLocalDelete")
	}
	if err := u.CheckUsed(backup.Name, false); err != nil {
		return err
	}
	return backupRepo.Delete(repo.WithByID(id))
}

func (u *BackupService) Update(req dto.BackupOperate) error {
	backup, _ := backupRepo.Get(repo.WithByID(req.ID))
	if backup.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if req.Type != constant.Sftp && req.Type != constant.Local && req.BackupPath != "/" {
		req.BackupPath = strings.TrimPrefix(req.BackupPath, "/")
	}
	var newBackup model.BackupAccount
	if err := copier.Copy(&newBackup, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
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
	if backup.Type == constant.Local {
		if err := changeLocalBackup(backup.BackupPath, newBackup.BackupPath); err != nil {
			return err
		}
	}

	if newBackup.Type == constant.OneDrive || newBackup.Type == constant.GoogleDrive {
		if err := loadRefreshTokenByCode(&newBackup); err != nil {
			return err
		}
	}
	if backup.Type != constant.Local {
		isOk, err := u.checkBackupConn(&newBackup)
		if err != nil || !isOk {
			return buserr.WithMap("ErrBackupCheck", map[string]interface{}{"err": err.Error()}, err)
		}
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
	return nil
}

func (u *BackupService) RefreshToken(req dto.OperateByID) error {
	backup, _ := backupRepo.Get(repo.WithByID(req.ID))
	if backup.ID == 0 {
		return buserr.New("ErrRecordNotFound")
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

func (u *BackupService) checkBackupConn(backup *model.BackupAccount) (bool, error) {
	client, err := newClient(backup, false)
	if err != nil {
		return false, err
	}
	fileItem := path.Join(global.Dir.BaseDir, "1panel/tmp/test/1panel")
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
			return buserr.New("ErrRecordNotFound")
		}
		return backupRepo.Delete(repo.WithByID(account.ID))
	case "update":
		if account.ID == 0 {
			return buserr.New("ErrRecordNotFound")
		}
		accountItem.ID = account.ID
		accountItem.CreatedAt = account.CreatedAt
		accountItem.UpdatedAt = account.UpdatedAt
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

func (u *BackupService) CheckUsed(name string, isPublic bool) error {
	account, _ := backupRepo.Get(repo.WithByName(name), backupRepo.WithByPublic(isPublic))
	if account.ID == 0 {
		return nil
	}
	cronjobs, _ := cronjobRepo.List()
	for _, job := range cronjobs {
		if job.DownloadAccountID == account.ID {
			return buserr.New("ErrBackupInUsed")
		}
		ids := strings.Split(job.SourceAccountIDs, ",")
		for _, idItem := range ids {
			if idItem == fmt.Sprintf("%v", account.ID) {
				return buserr.New("ErrBackupInUsed")
			}
		}
	}
	return nil
}

func NewBackupClientWithID(id uint) (*model.BackupAccount, cloud_storage.CloudStorageClient, error) {
	account, _ := backupRepo.Get(repo.WithByID(id))
	backClient, err := newClient(&account, true)
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
	var idItems []uint
	for i := 0; i < len(ids); i++ {
		item, _ := strconv.Atoi(ids[i])
		idItems = append(idItems, uint(item))
	}
	accounts, _ = backupRepo.List(repo.WithByIDs(idItems))
	clientMap := make(map[string]backupClientHelper)
	for _, item := range accounts {
		backClient, err := newClient(&item, true)
		if err != nil {
			return nil, err
		}
		clientMap[fmt.Sprintf("%v", item.ID)] = backupClientHelper{
			client:      backClient,
			name:        item.Name,
			backupPath:  item.BackupPath,
			accountType: item.Type,
			id:          item.ID,
		}
	}
	return clientMap, nil
}

func newClient(account *model.BackupAccount, isEncrypt bool) (cloud_storage.CloudStorageClient, error) {
	varMap := make(map[string]interface{})
	if len(account.Vars) != 0 {
		if err := json.Unmarshal([]byte(account.Vars), &varMap); err != nil {
			return nil, err
		}
	}
	varMap["bucket"] = account.Bucket
	varMap["backupPath"] = account.BackupPath
	if isEncrypt {
		account.AccessKey, _ = encrypt.StringDecrypt(account.AccessKey)
		account.Credential, _ = encrypt.StringDecrypt(account.Credential)
	}
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

func loadRefreshTokenByCode(backup *model.BackupAccount) error {
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

func loadBackupNamesByID(accountIDs string, downloadID uint) ([]string, string, error) {
	accountIDList := strings.Split(accountIDs, ",")
	var ids []uint
	for _, item := range accountIDList {
		if len(item) != 0 {
			itemID, _ := strconv.Atoi(item)
			ids = append(ids, uint(itemID))
		}
	}
	list, err := backupRepo.List(repo.WithByIDs(ids))
	if err != nil {
		return nil, "", err
	}
	var accounts []string
	var downloadAccount string
	for _, item := range list {
		itemName := fmt.Sprintf("%s - %s", item.Type, item.Name)
		accounts = append(accounts, itemName)
		if item.ID == downloadID {
			downloadAccount = itemName
		}
	}
	return accounts, downloadAccount, nil
}

func changeLocalBackup(oldPath, newPath string) error {
	fileOp := files.NewFileOp()
	if fileOp.Stat(path.Join(oldPath, "app")) {
		if err := fileOp.Mv(path.Join(oldPath, "app"), newPath); err != nil {
			return err
		}
	}
	if fileOp.Stat(path.Join(oldPath, "database")) {
		if err := fileOp.Mv(path.Join(oldPath, "database"), newPath); err != nil {
			return err
		}
	}
	if fileOp.Stat(path.Join(oldPath, "directory")) {
		if err := fileOp.Mv(path.Join(oldPath, "directory"), newPath); err != nil {
			return err
		}
	}
	if fileOp.Stat(path.Join(oldPath, "system_snapshot")) {
		if err := fileOp.Mv(path.Join(oldPath, "system_snapshot"), newPath); err != nil {
			return err
		}
	}
	if fileOp.Stat(path.Join(oldPath, "website")) {
		if err := fileOp.CopyDir(path.Join(oldPath, "website"), newPath); err != nil {
			return err
		}
	}
	if fileOp.Stat(path.Join(oldPath, "log")) {
		if err := fileOp.Mv(path.Join(oldPath, "log"), newPath); err != nil {
			return err
		}
	}
	return nil
}
