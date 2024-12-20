package service

import (
	"os"
	"sort"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/toolbox"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type FtpService struct{}

type IFtpService interface {
	LoadBaseInfo() (dto.FtpBaseInfo, error)
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Operate(operation string) error
	Create(req dto.FtpCreate) (uint, error)
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
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return baseInfo, err
	}
	baseInfo.IsActive, baseInfo.IsExist = client.Status()
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
	total, lists, err := ftpRepo.Page(req.Page, req.PageSize, ftpRepo.WithByUser(req.Info), commonRepo.WithOrderBy("created_at desc"))
	if err != nil {
		return 0, nil, err
	}
	var users []dto.FtpInfo
	for _, user := range lists {
		var item dto.FtpInfo
		if err := copier.Copy(&item, &user); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
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
		return nil
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
			if item.Path != itemInDB.Path || item.Status != itemInDB.Status {
				if err := ftpRepo.Update(itemInDB.ID, map[string]interface{}{"path": item.Path, "status": item.Status}); err != nil {
					return err
				}
			}
		} else {
			if err := ftpRepo.Create(&model.Ftp{User: item.User, Path: item.Path, Status: item.Status}); err != nil {
				return err
			}
		}
	}

	for _, item := range listsInDB {
		if _, ok := sameData[item.User]; !ok {
			_ = ftpRepo.Update(item.ID, map[string]interface{}{"status": "deleted"})
		}
	}
	return nil
}

func (f *FtpService) Create(req dto.FtpCreate) (uint, error) {
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
	userInDB, _ := ftpRepo.Get(hostRepo.WithByUser(req.User))
	if userInDB.ID != 0 {
		return 0, constant.ErrRecordExist
	}
	client, err := toolbox.NewFtpClient()
	if err != nil {
		return 0, err
	}
	if err := client.UserAdd(req.User, req.Password, req.Path); err != nil {
		return 0, err
	}
	var ftp model.Ftp
	if err := copier.Copy(&ftp, &req); err != nil {
		return 0, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	ftp.Status = constant.StatusEnable
	ftp.Password = pass
	if err := ftpRepo.Create(&ftp); err != nil {
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
		ftpItem, err := ftpRepo.Get(commonRepo.WithByID(id))
		if err != nil {
			return err
		}
		_ = client.UserDel(ftpItem.User)
		_ = ftpRepo.Delete(commonRepo.WithByID(id))
	}
	return nil
}

func (f *FtpService) Update(req dto.FtpUpdate) error {
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
	ftpItem, _ := ftpRepo.Get(commonRepo.WithByID(req.ID))
	if ftpItem.ID == 0 {
		return constant.ErrRecordNotFound
	}
	passItem, err := encrypt.StringDecrypt(ftpItem.Password)
	if err != nil {
		return err
	}

	client, err := toolbox.NewFtpClient()
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
		if err := client.SetPath(ftpItem.User, req.Path); err != nil {
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
