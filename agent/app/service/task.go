package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
)

type TaskLogService struct{}

type ITaskLogService interface {
	Page(req dto.SearchTaskLogReq) (int64, []dto.TaskDTO, error)
	SyncForRestart() error
	CountExecutingTask() (int64, error)
}

func NewITaskService() ITaskLogService {
	return &TaskLogService{}
}

func (u *TaskLogService) Page(req dto.SearchTaskLogReq) (int64, []dto.TaskDTO, error) {
	opts := []repo.DBOption{
		repo.WithOrderBy("created_at desc"),
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

func (u *TaskLogService) SyncForRestart() error {
	return taskRepo.UpdateRunningTaskToFailed()
}

func (u *TaskLogService) CountExecutingTask() (int64, error) {
	return taskRepo.CountExecutingTask()
}
