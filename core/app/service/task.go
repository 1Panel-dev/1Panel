package service

import (
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
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
		opts = append(opts, commonRepo.WithByStatus(req.Status))
	}
	if req.Type != "" {
		opts = append(opts, commonRepo.WithByType(req.Type))
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
