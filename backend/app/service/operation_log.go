package service

import "github.com/1Panel-dev/1Panel/app/model"

type OperationService struct{}

type IOperationService interface {
	Create(operation model.OperationLog) error
}

func NewIOperationService() IOperationService {
	return &OperationService{}
}

func (u *OperationService) Create(operation model.OperationLog) error {
	return operationRepo.Create(&operation)
}
