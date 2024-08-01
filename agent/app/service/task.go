package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
)

type TaskLogService struct{}

type ITaskLogService interface {
	Page(req dto.SearchTaskLogReq) (int64, []dto.TaskDTO, error)
}

func NewITaskService() ITaskLogService {
	return &TaskLogService{}
}

func (u *TaskLogService) Page(req dto.SearchTaskLogReq) (int64, []dto.TaskDTO, error) {
	opts := []repo.DBOption{
		commonRepo.WithOrderBy("created_at desc"),
	}
	if req.Status != "" {
		opts = append(opts, taskRepo.WithStatus(req.Status))
	}
	if req.Type != "" {
		opts = append(opts, taskRepo.WithType(req.Type))
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
