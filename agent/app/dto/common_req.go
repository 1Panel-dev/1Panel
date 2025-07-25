package dto

type SearchWithPage struct {
	PageInfo
	Info string `json:"info"`
}

type SearchPageWithType struct {
	PageInfo
	Info string `json:"info"`
	Type string `json:"type"`
}

type PageInfo struct {
	Page     int `json:"page" validate:"required,number"`
	PageSize int `json:"pageSize" validate:"required,number"`
}

type UpdateDescription struct {
	ID          uint   `json:"id" validate:"required"`
	Description string `json:"description" validate:"max=256"`
}

type OperationWithName struct {
	Name string `json:"name" validate:"required"`
}

type OperateByID struct {
	ID uint `json:"id" validate:"required"`
}
type OperateByIDs struct {
	IDs []uint `json:"ids"`
}

type Operate struct {
	Operation string `json:"operation" validate:"required"`
}

type SearchByFilter struct {
	Filter string `json:"filter"`
}

type BatchDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

type FilePath struct {
	Path string `json:"path" validate:"required"`
}

type DeleteByName struct {
	Name string `json:"name" validate:"required"`
}

type UpdateByFile struct {
	File string `json:"file"`
}

type UpdateByNameAndFile struct {
	Name string `json:"name"`
	File string `json:"file"`
}

type OperationWithNameAndType struct {
	Name string `json:"name"`
	Type string `json:"type" validate:"required"`
}

type UpdateGroup struct {
	Group    uint `json:"group"`
	NewGroup uint `json:"newGroup"`
}

type OperateWithTask struct {
	TaskID string `json:"taskID"`
}

type ForceDelete struct {
	IDs         []uint `json:"ids"`
	ForceDelete bool   `json:"forceDelete"`
}
