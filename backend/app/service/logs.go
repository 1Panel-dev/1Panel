package service

import (
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type LogService struct{}

const logs = "https://resource.fit2cloud.com/installation-log.sh"

type ILogService interface {
	CreateLoginLog(operation model.LoginLog) error
	PageLoginLog(search dto.SearchLgLogWithPage) (int64, interface{}, error)

	CreateOperationLog(operation model.OperationLog) error
	PageOperationLog(search dto.SearchOpLogWithPage) (int64, interface{}, error)

	LoadSystemLog() (string, error)

	CleanLogs(logtype string) error
}

func NewILogService() ILogService {
	return &LogService{}
}

func (u *LogService) CreateLoginLog(operation model.LoginLog) error {
	return logRepo.CreateLoginLog(&operation)
}

func (u *LogService) PageLoginLog(req dto.SearchLgLogWithPage) (int64, interface{}, error) {
	total, ops, err := logRepo.PageLoginLog(
		req.Page,
		req.PageSize,
		logRepo.WithByIP(req.IP),
		logRepo.WithByStatus(req.Status),
		commonRepo.WithOrderBy("created_at desc"),
	)
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

func (u *LogService) PageOperationLog(req dto.SearchOpLogWithPage) (int64, interface{}, error) {
	total, ops, err := logRepo.PageOperationLog(
		req.Page,
		req.PageSize,
		logRepo.WithByGroup(req.Source),
		logRepo.WithByLikeOperation(req.Operation),
		logRepo.WithByStatus(req.Status),
		commonRepo.WithOrderBy("created_at desc"),
	)
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

func (u *LogService) LoadSystemLog() (string, error) {
	filePath := path.Join(global.CONF.System.DataDir, "log/1Panel.log")
	if _, err := os.Stat(filePath); err != nil {
		return "", buserr.New("ErrHttpReqNotFound")
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (u *LogService) CleanLogs(logtype string) error {
	if logtype == "operation" {
		return logRepo.CleanOperation()
	}
	return logRepo.CleanLogin()
}

func writeLogs(version string) {
	_, _ = cmd.Execf("curl -sfL %s | sh -s 1p upgrade %s", logs, version)
}
