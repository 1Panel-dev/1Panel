package request

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

type FileHistorySearchReq struct {
	dto.PageInfo
	Path      string `json:"path"`
	Scope     string `json:"scope" validate:"required,oneof=current all"`
	Operation string `json:"operation"`
}

type FileHistoryContentReq struct {
	ID uint `json:"id" validate:"required"`
}

type FileHistoryIDs []uint

func (ids *FileHistoryIDs) UnmarshalJSON(data []byte) error {
	type alias []uint
	var direct alias
	if err := json.Unmarshal(data, &direct); err == nil {
		*ids = FileHistoryIDs(direct)
		return nil
	}

	var wrapped struct {
		IDs []uint `json:"ids"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	*ids = FileHistoryIDs(wrapped.IDs)
	return nil
}

type FileHistoryDeleteReq struct {
	IDs FileHistoryIDs `json:"ids" validate:"required"`
}

type FileHistoryRestoreReq struct {
	ID uint `json:"id" validate:"required"`
}

type FileHistorySettingUpdate struct {
	Enable      string `json:"enable" validate:"required,oneof=Enable Disable"`
	MaxPerPath  int    `json:"maxPerPath" validate:"min=0,max=1000"`
	DiskQuotaMB int    `json:"diskQuotaMB" validate:"min=0,max=1048576"`
}
