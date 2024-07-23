package dto

type SearchCommandWithPage struct {
	PageInfo
	OrderBy string `json:"orderBy" validate:"required,oneof=name command created_at"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
	GroupID uint   `json:"groupID"`
	Info    string `json:"info"`
	Name    string `json:"name"`
}

type CommandOperate struct {
	ID          uint   `json:"id"`
	GroupID     uint   `json:"groupID"`
	GroupBelong string `json:"groupBelong"`
	Name        string `json:"name" validate:"required"`
	Command     string `json:"command" validate:"required"`
}

type CommandInfo struct {
	ID          uint   `json:"id"`
	GroupID     uint   `json:"groupID"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	GroupBelong string `json:"groupBelong"`
}

type CommandTree struct {
	ID       uint          `json:"id"`
	Label    string        `json:"label"`
	Children []CommandInfo `json:"children"`
}

type RedisCommand struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
}
