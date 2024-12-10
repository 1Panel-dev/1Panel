package service

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type LogService struct{}

const logs = "https://resource.fit2cloud.com/installation-log.sh"

type ILogService interface {
	ListSystemLogFile() ([]string, error)
	CreateLoginLog(operation model.LoginLog) error
	PageLoginLog(search dto.SearchLgLogWithPage) (int64, interface{}, error)

	CreateOperationLog(operation *model.OperationLog) error
	PageOperationLog(search dto.SearchOpLogWithPage) (int64, interface{}, error)

	CleanLogs(logtype string) error
}

func NewILogService() ILogService {
	return &LogService{}
}

func (u *LogService) CreateLoginLog(operation model.LoginLog) error {
	return logRepo.CreateLoginLog(&operation)
}

func (u *LogService) ListSystemLogFile() ([]string, error) {
	logDir := path.Join(global.CONF.System.BaseDir, "1panel/log")
	var files []string
	if err := filepath.Walk(logDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "1Panel") {
			if info.Name() == "1Panel.log" {
				files = append(files, time.Now().Format("2006-01-02"))
				return nil
			}
			itemFileName := strings.TrimPrefix(info.Name(), "1Panel-")
			itemFileName = strings.TrimSuffix(itemFileName, ".gz")
			itemFileName = strings.TrimSuffix(itemFileName, ".log")
			files = append(files, itemFileName)
			return nil
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(files) < 2 {
		return files, nil
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	return files, nil
}

func (u *LogService) PageLoginLog(req dto.SearchLgLogWithPage) (int64, interface{}, error) {
	options := []global.DBOption{
		repo.WithOrderBy("created_at desc"),
	}
	if len(req.IP) != 0 {
		options = append(options, logRepo.WithByIP(req.IP))
	}
	if len(req.Status) != 0 {
		options = append(options, repo.WithByStatus(req.Status))
	}
	total, ops, err := logRepo.PageLoginLog(
		req.Page,
		req.PageSize,
		options...,
	)
	var dtoOps []dto.LoginLog
	for _, op := range ops {
		var item dto.LoginLog
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrTransform, err.Error())
		}
		dtoOps = append(dtoOps, item)
	}
	return total, dtoOps, err
}

func (u *LogService) CreateOperationLog(operation *model.OperationLog) error {
	return logRepo.CreateOperationLog(operation)
}

func (u *LogService) PageOperationLog(req dto.SearchOpLogWithPage) (int64, interface{}, error) {
	options := []global.DBOption{
		repo.WithOrderBy("created_at desc"),
		logRepo.WithByLikeOperation(req.Operation),
	}
	if len(req.Source) != 0 {
		options = append(options, logRepo.WithBySource(req.Source))
	}
	if len(req.Status) != 0 {
		options = append(options, repo.WithByStatus(req.Status))
	}

	total, ops, err := logRepo.PageOperationLog(
		req.Page,
		req.PageSize,
		options...,
	)
	var dtoOps []dto.OperationLog
	for _, op := range ops {
		var item dto.OperationLog
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrTransform, err.Error())
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

func writeLogs(version string) {
	_, _ = cmd.Execf("curl -sfL %s | sh -s 1p upgrade %s", logs, version)
}
