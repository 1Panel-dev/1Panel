package service

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/toolbox"
	"github.com/jinzhu/copier"
)

type FtpService struct{}

type IFtpService interface {
	LoadBaseInfo() (dto.FtpBaseInfo, error)
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Operate(operation string) error
	Create(req dto.FtpCreate) (uint, error)
	CreateWebsite(req dto.FtpCreate) (uint, error)
	Delete(req dto.BatchDeleteReq) error
	Update(req dto.FtpUpdate) error
	Sync() error
	LoadLog(req dto.FtpLogSearch) (int64, interface{}, error)
}

func NewIFtpService() IFtpService {
	return &FtpService{}
}

func (f *FtpService) LoadBaseInfo() (dto.FtpBaseInfo, error) {
	var baseInfo dto.FtpBaseInfo
	baseInfo.IsActive, baseInfo.IsExist = toolbox.FtpStatus()
	return baseInfo, nil
}

func (f *FtpService) LoadLog(req dto.FtpLogSearch) (int64, interface{}, error) {
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return 0, nil, err
	}
	logItem, err := client.LoadLogs(req.User, req.Operation)
	if err != nil {
		return 0, nil, err
	}
	sort.Slice(logItem, func(i, j int) bool {
		return logItem[i].Time > logItem[j].Time
	})
	var logs []toolbox.FtpLog
	total, start, end := len(logItem), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		logs = make([]toolbox.FtpLog, 0)
	} else {
		if end >= total {
			end = total
		}
		logs = logItem[start:end]
	}
	return int64(total), logs, nil
}

func (u *FtpService) Operate(operation string) error {
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return err
	}
	return client.Operate(operation)
}

func (f *FtpService) SearchWithPage(req dto.SearchWithPage) (int64, interface{}, error) {
	total, lists, err := ftpRepo.Page(req.Page, req.PageSize, ftpRepo.WithLikeUser(req.Info), repo.WithOrderDesc("created_at"))
	if err != nil {
		return 0, nil, err
	}
	var users []dto.FtpInfo
	for _, user := range lists {
		var item dto.FtpInfo
		if err := copier.Copy(&item, &user); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		item.Password, _ = encrypt.StringDecrypt(item.Password)
		users = append(users, item)
	}
	return total, users, err
}

func (f *FtpService) Sync() error {
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return err
	}
	lists, err := client.LoadList()
	if err != nil {
		return err
	}
	listsInDB, err := ftpRepo.GetList()
	if err != nil {
		return err
	}
	currentData := make(map[string]model.Ftp)
	for _, item := range listsInDB {
		currentData[item.User] = item
	}
	sameData := make(map[string]struct{})
	for _, item := range lists {
		if itemInDB, ok := currentData[item.User]; ok {
			sameData[item.User] = struct{}{}
			if item.Path != itemInDB.Path || item.Status != itemInDB.Status || item.UID != itemInDB.UID || item.GID != itemInDB.GID {
				if err := ftpRepo.Update(itemInDB.ID, map[string]interface{}{
					"path":   item.Path,
					"status": item.Status,
					"uid":    item.UID,
					"gid":    item.GID,
				}); err != nil {
					return err
				}
			}
		} else {
			if err := ftpRepo.Create(&model.Ftp{
				User:   item.User,
				Path:   item.Path,
				Status: item.Status,
				UID:    item.UID,
				GID:    item.GID,
			}); err != nil {
				return err
			}
		}
	}
	for _, item := range listsInDB {
		if _, ok := sameData[item.User]; !ok {
			_ = ftpRepo.Update(item.ID, map[string]interface{}{"status": constant.StatusDeleted})
		}
	}
	return nil
}

func (f *FtpService) Create(req dto.FtpCreate) (uint, error) {
	return f.create(req, false)
}

func (f *FtpService) CreateWebsite(req dto.FtpCreate) (uint, error) {
	return f.create(req, true)
}

func (f *FtpService) create(req dto.FtpCreate, website bool) (uint, error) {
	if err := toolbox.ValidateFtpRootPath(req.Path); err != nil {
		return 0, err
	}
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return 0, err
	}
	if _, err := os.Stat(req.Path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(req.Path, os.ModePerm); err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}
	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return 0, err
	}
	userInDB, _ := ftpRepo.Get(ftpRepo.WithByUser(req.User))
	if userInDB.ID != 0 {
		return 0, buserr.New("ErrRecordExist")
	}
	var ftp model.Ftp
	if err := copier.Copy(&ftp, &req); err != nil {
		return 0, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	uid, gid := uint(constant.WebsiteUID), uint(constant.WebsiteGID)
	if !website {
		uid, gid, err = toolbox.EnsureStandaloneFtpIdentity()
		if err != nil {
			return 0, err
		}
	}
	if err := client.UserAdd(req.User, req.Password, req.Path, uid, gid); err != nil {
		return 0, err
	}
	ftp.Status = constant.StatusEnable
	ftp.Password = pass
	ftp.UID = uid
	ftp.GID = gid
	if err := ftpRepo.Create(&ftp); err != nil {
		if rollbackErr := client.UserDel(req.User); rollbackErr != nil {
			return 0, errors.Join(err, fmt.Errorf("rollback FTP user %s failed: %w", req.User, rollbackErr))
		}
		return 0, err
	}
	return ftp.ID, nil
}

func (f *FtpService) Delete(req dto.BatchDeleteReq) error {
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return err
	}
	for _, id := range req.Ids {
		ftpItem, err := ftpRepo.Get(repo.WithByID(id))
		if err != nil {
			return err
		}
		_ = client.UserDel(ftpItem.User)
		_ = ftpRepo.Delete(repo.WithByID(id))
	}
	return nil
}

func (f *FtpService) Update(req dto.FtpUpdate) error {
	if err := toolbox.ValidateFtpRootPath(req.Path); err != nil {
		return err
	}
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return err
	}
	if _, err := os.Stat(req.Path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(req.Path, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	pass, err := encrypt.StringEncrypt(req.Password)
	if err != nil {
		return err
	}
	ftpItem, _ := ftpRepo.Get(repo.WithByID(req.ID))
	if ftpItem.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	passItem, err := encrypt.StringDecrypt(ftpItem.Password)
	if err != nil {
		return err
	}

	needReload := false
	updates := make(map[string]interface{})
	if req.Password != passItem {
		if err := client.SetPasswd(ftpItem.User, req.Password); err != nil {
			return err
		}
		updates["password"] = pass
		needReload = true
	}
	if req.Status != ftpItem.Status {
		if err := client.SetStatus(ftpItem.User, req.Status); err != nil {
			return err
		}
		updates["status"] = req.Status
		needReload = true
	}
	if req.Path != ftpItem.Path {
		uid, gid := ftpItem.UID, ftpItem.GID
		if uid == 0 || gid == 0 {
			uid, gid = uint(constant.WebsiteUID), uint(constant.WebsiteGID)
		}
		if err := client.SetPath(ftpItem.User, req.Path, uid, gid); err != nil {
			return err
		}
		updates["path"] = req.Path
		needReload = true
	}
	if req.Description != ftpItem.Description {
		updates["description"] = req.Description
	}
	if needReload {
		_ = client.Reload()
	}
	if len(updates) != 0 {
		return ftpRepo.Update(ftpItem.ID, updates)
	}
	return nil
}
