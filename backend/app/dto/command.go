package dto

type CommandCreate struct {
	Name    string `json:"name" validate:"required"`
	Command string `json:"command" validate:"required"`
}

type CommandUpdate struct {
	Name string `json:"name" validate:"required"`
}

type CommandInfo struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}
