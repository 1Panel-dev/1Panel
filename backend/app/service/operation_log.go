package service

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type OperationService struct{}

type IOperationService interface {
	Page(page, size int) (int64, interface{}, error)
	Create(operation model.OperationLog) error
	BatchDelete(ids []uint) error
}

func NewIOperationService() IOperationService {
	return &OperationService{}
}

func (u *OperationService) Create(operation model.OperationLog) error {
	return operationRepo.Create(&operation)
}

func (u *OperationService) Page(page, size int) (int64, interface{}, error) {
	total, ops, err := operationRepo.Page(page, size, commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.OperationLogBack
	for _, op := range ops {
		var item dto.OperationLogBack
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		item.Body = filterSensitive(item.Body)
		var res dto.Response
		if err := json.Unmarshal([]byte(item.Resp), &res); err != nil {
			global.LOG.Errorf("unmarshal failed, err: %+v", err)
			dtoOps = append(dtoOps, item)
			continue
		}
		item.Status = res.Code
		if item.Status != 200 {
			item.ErrorMessage = res.Msg
		}
		dtoOps = append(dtoOps, item)
	}
	return total, dtoOps, err
}

func (u *OperationService) BatchDelete(ids []uint) error {
	return operationRepo.Delete(commonRepo.WithIdsIn(ids))
}

func filterSensitive(vars string) string {
	var Sensitives = []string{"password", "Password"}
	ops := make(map[string]string)
	if err := json.Unmarshal([]byte(vars), &ops); err != nil {
		return vars
	}
	for k := range ops {
		for _, sen := range Sensitives {
			if k == sen {
				delete(ops, k)
				continue
			}
		}
	}
	backStr, err := json.Marshal(ops)
	if err != nil {
		return ""
	}
	return string(backStr)
}
