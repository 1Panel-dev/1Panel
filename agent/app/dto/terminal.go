package dto

type TerminalSessionClose struct {
	ID string `json:"id" validate:"required"`
}
