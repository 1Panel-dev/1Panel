package dto

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type SearchTaskLogReq struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	TaskID string `json:"taskID"`
	PageInfo
}

type TaskDTO struct {
	model.Task
}

type SystemLogReq struct {
	PageSize  int       `json:"pageSize" validate:"omitempty,min=1,max=500"`
	Cursor    string    `json:"cursor"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Keyword   string    `json:"keyword"`
	Priority  string    `json:"priority"`
	Service   string    `json:"service"`
}

type SystemLogRes struct {
	Source     string          `json:"source"`
	Items      []SystemLogItem `json:"items"`
	HasMore    bool            `json:"hasMore"`
	NextCursor string          `json:"nextCursor"`
}

type SystemLogItem struct {
	Timestamp int64  `json:"-"`
	Time      string `json:"time"`
	Priority  string `json:"priority"`
	Service   string `json:"service"`
	Message   string `json:"message"`
	Raw       string `json:"raw"`
}
