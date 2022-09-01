package dto

type CommandOperate struct {
	Name    string `json:"name" validate:"required"`
	Command string `json:"command" validate:"required"`
}

type CommandInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
}
