package service

import (
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

type TaskLogService struct{}

type ITaskLogService interface {
	Page(req dto.SearchTaskLogReq) (int64, []dto.TaskDTO, error)
	ReadByLine(req request.TaskLogReadReq) (*response.FileLineContent, error)
	SyncForRestart() error
	CountExecutingTask() (int64, error)
}

func NewITaskService() ITaskLogService {
	return &TaskLogService{}
}

func (u *TaskLogService) Page(req dto.SearchTaskLogReq) (int64, []dto.TaskDTO, error) {
	opts := []repo.DBOption{
		repo.WithOrderDesc("created_at"),
	}
	if req.TaskID != "" {
		opts = append(opts, taskRepo.WithByID(req.TaskID))
	}
	if req.Status != "" {
		opts = append(opts, repo.WithByStatus(req.Status))
	}
	if req.Type != "" {
		opts = append(opts, repo.WithByType(req.Type))
	}

	total, tasks, err := taskRepo.Page(
		req.Page,
		req.PageSize,
		opts...,
	)
	var items []dto.TaskDTO
	for _, t := range tasks {
		item := dto.TaskDTO{
			Task: t,
		}
		items = append(items, item)
	}
	return total, items, err
}

func (u *TaskLogService) ReadByLine(req request.TaskLogReadReq) (*response.FileLineContent, error) {
	opts := []repo.DBOption{}
	if req.TaskID != "" {
		opts = append(opts, taskRepo.WithByID(req.TaskID))
	} else {
		opts = append(opts, repo.WithOrderRuleBy("created_at", "desc"), repo.WithByType(req.TaskType), taskRepo.WithOperate(req.TaskOperate), taskRepo.WithResourceID(req.ResourceID))
	}
	taskModel, err := taskRepo.GetFirst(opts...)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(taskModel.LogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var (
		lines       []string
		isEndOfFile bool
		scope       string
		logFileRes  *dto.LogFileRes
	)
	if stat.Size() > files.MaxReadFileSize {
		lines, _ = files.TailFromEnd(taskModel.LogFile, req.PageSize)
		isEndOfFile = true
		scope = "tail"
	} else {
		logFileRes, err = files.ReadFileByLine(taskModel.LogFile, req.Page, req.PageSize, req.Latest)
		if err != nil {
			return nil, err
		}
		scope = "page"
		lines = logFileRes.Lines
	}

	res := &response.FileLineContent{
		End:        isEndOfFile,
		Path:       taskModel.LogFile,
		TaskStatus: taskModel.Status,
		Lines:      lines,
		Scope:      scope,
	}
	if logFileRes != nil {
		res.TotalLines = logFileRes.TotalLines
		res.Total = logFileRes.TotalPages
		res.End = logFileRes.IsEndOfFile
	}
	return res, nil
}

func (u *TaskLogService) SyncForRestart() error {
	return taskRepo.UpdateRunningTaskToFailed()
}

func (u *TaskLogService) CountExecutingTask() (int64, error) {
	return taskRepo.CountExecutingTask()
}
