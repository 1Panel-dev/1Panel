package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type MysqlService struct{}

type IMysqlService interface {
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Create(mysqlDto dto.MysqlDBCreate) error
	Delete(ids []uint) error
}

func NewIMysqlService() IMysqlService {
	return &MysqlService{}
}

func (u *MysqlService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, mysqls, err := mysqlRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Info))
	var dtoMysqls []dto.MysqlDBInfo
	for _, mysql := range mysqls {
		var item dto.MysqlDBInfo
		if err := copier.Copy(&item, &mysql); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoMysqls = append(dtoMysqls, item)
	}
	return total, dtoMysqls, err
}

func (u *MysqlService) Create(mysqlDto dto.MysqlDBCreate) error {
	mysql, _ := mysqlRepo.Get(commonRepo.WithByName(mysqlDto.Name))
	if mysql.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&mysql, &mysqlDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := mysqlRepo.Create(&mysql); err != nil {
		return err
	}
	return nil
}

func (u *MysqlService) Delete(ids []uint) error {
	if len(ids) == 1 {
		mysql, _ := mysqlRepo.Get(commonRepo.WithByID(ids[0]))
		if mysql.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return mysqlRepo.Delete(commonRepo.WithByID(ids[0]))
	}
	return mysqlRepo.Delete(commonRepo.WithIdsIn(ids))
}
