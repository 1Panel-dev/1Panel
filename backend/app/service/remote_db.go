package service

import (
	"context"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql"
	"github.com/1Panel-dev/1Panel/backend/utils/mysql/client"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type RemoteDBService struct{}

type IRemoteDBService interface {
	Get(name string) (dto.RemoteDBInfo, error)
	SearchWithPage(search dto.RemoteDBSearch) (int64, interface{}, error)
	Create(req dto.RemoteDBCreate) error
	Update(req dto.RemoteDBUpdate) error
	Delete(id uint) error
	List(dbType string) ([]dto.RemoteDBOption, error)
}

func NewIRemoteDBService() IRemoteDBService {
	return &RemoteDBService{}
}

func (u *RemoteDBService) SearchWithPage(search dto.RemoteDBSearch) (int64, interface{}, error) {
	total, dbs, err := remoteDBRepo.Page(search.Page, search.PageSize,
		commonRepo.WithByType(search.Type),
		commonRepo.WithLikeName(search.Info),
		remoteDBRepo.WithoutByFrom("local"),
	)
	var datas []dto.RemoteDBInfo
	for _, db := range dbs {
		var item dto.RemoteDBInfo
		if err := copier.Copy(&item, &db); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		datas = append(datas, item)
	}
	return total, datas, err
}

func (u *RemoteDBService) Get(name string) (dto.RemoteDBInfo, error) {
	var data dto.RemoteDBInfo
	remote, err := remoteDBRepo.Get(commonRepo.WithByName(name))
	if err != nil {
		return data, err
	}
	if err := copier.Copy(&data, &remote); err != nil {
		return data, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return data, nil
}

func (u *RemoteDBService) List(dbType string) ([]dto.RemoteDBOption, error) {
	dbs, err := remoteDBRepo.GetList(commonRepo.WithByType(dbType))
	var datas []dto.RemoteDBOption
	for _, db := range dbs {
		var item dto.RemoteDBOption
		if err := copier.Copy(&item, &db); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		datas = append(datas, item)
	}
	return datas, err
}

func (u *RemoteDBService) Create(req dto.RemoteDBCreate) error {
	db, _ := remoteDBRepo.Get(commonRepo.WithByName(req.Name))
	if db.ID != 0 {
		return constant.ErrRecordExist
	}
	if _, err := mysql.NewMysqlClient(client.DBInfo{
		From:     "remote",
		Address:  req.Address,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Timeout:  300,
	}); err != nil {
		return err
	}
	if err := copier.Copy(&db, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := remoteDBRepo.Create(&db); err != nil {
		return err
	}
	return nil
}

func (u *RemoteDBService) Delete(id uint) error {
	db, _ := remoteDBRepo.Get(commonRepo.WithByID(id))
	if db.ID == 0 {
		return constant.ErrRecordNotFound
	}
	if err := remoteDBRepo.Delete(commonRepo.WithByID(id)); err != nil {
		return err
	}
	if db.From != "local" {
		if err := mysqlRepo.Delete(context.Background(), remoteDBRepo.WithByFrom(db.Name)); err != nil {
			return err
		}
	}
	return nil
}

func (u *RemoteDBService) Update(req dto.RemoteDBUpdate) error {
	if _, err := mysql.NewMysqlClient(client.DBInfo{
		From:     "remote",
		Address:  req.Address,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Timeout:  300,
	}); err != nil {
		return err
	}

	upMap := make(map[string]interface{})
	upMap["version"] = req.Version
	upMap["address"] = req.Address
	upMap["port"] = req.Port
	upMap["username"] = req.Username
	upMap["password"] = req.Password
	upMap["description"] = req.Description
	return remoteDBRepo.Update(req.ID, upMap)
}
