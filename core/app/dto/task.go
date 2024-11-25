package dto

import "github.com/1Panel-dev/1Panel/core/app/model"

type SearchTaskLogReq struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	PageInfo
}

type TaskDTO struct {
	model.Task
}
