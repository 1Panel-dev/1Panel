package dto

type SearchCommandWithPage struct {
	PageInfo
	OrderBy string `json:"orderBy" validate:"required,oneof=name command createdAt"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
	GroupID uint   `json:"groupID"`
	Type    string `json:"type" validate:"required,oneof=redis command"`
	Info    string `json:"info"`
}

type CommandOperate struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	GroupID     uint   `json:"groupID"`
	GroupBelong string `json:"groupBelong"`
	Name        string `json:"name" validate:"required"`
	Command     string `json:"command" validate:"required"`
}

type CommandInfo struct {
	ID          uint   `json:"id"`
	GroupID     uint   `json:"groupID"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Command     string `json:"command"`
	GroupBelong string `json:"groupBelong"`
}

type CommandTree struct {
	Label    string        `json:"label"`
	Value    string        `json:"value"`
	Children []CommandTree `json:"children"`
}

type CommandDelete struct {
	Type string `json:"type" validate:"required,oneof=redis command"`
	IDs  []uint `json:"ids"`
}
