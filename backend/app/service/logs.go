package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type LogService struct{}

type ILogService interface {
	CreateLoginLog(operation model.LoginLog) error
	PageLoginLog(search dto.PageInfo) (int64, interface{}, error)

	CreateOperationLog(operation model.OperationLog) error
	PageOperationLog(search dto.PageInfo) (int64, interface{}, error)

	CleanLogs(logtype string) error
}

func NewILogService() ILogService {
	return &LogService{}
}

func (u *LogService) CreateLoginLog(operation model.LoginLog) error {
	return logRepo.CreateLoginLog(&operation)
}

func (u *LogService) PageLoginLog(search dto.PageInfo) (int64, interface{}, error) {
	total, ops, err := logRepo.PageLoginLog(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.LoginLog
	for _, op := range ops {
		var item dto.LoginLog
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoOps = append(dtoOps, item)
	}
	return total, dtoOps, err
}

func (u *LogService) CreateOperationLog(operation model.OperationLog) error {
	return logRepo.CreateOperationLog(&operation)
}

func (u *LogService) PageOperationLog(search dto.PageInfo) (int64, interface{}, error) {
	total, ops, err := logRepo.PageOperationLog(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.OperationLog
	for _, op := range ops {
		var item dto.OperationLog
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoOps = append(dtoOps, item)
	}
	return total, dtoOps, err
}

func (u *LogService) CleanLogs(logtype string) error {
	if logtype == "operation" {
		return logRepo.CleanOperation()
	}
	return logRepo.CleanLogin()
}
