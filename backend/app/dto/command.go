package dto

type SearchCommandWithPage struct {
	SearchWithPage
	GroupID uint   `json:"groupID"`
	Info    string `json:"info"`
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
