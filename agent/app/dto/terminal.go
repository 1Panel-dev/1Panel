package dto

type TerminalSessionClose struct {
	ID string `json:"id" validate:"required"`
}

type TerminalSessionRevoke struct {
	Scope         string `json:"scope" validate:"required,oneof=auth_session user all"`
	UserID        string `json:"userId"`
	AuthSessionID string `json:"authSessionId"`
}
