package dto

import "time"

// TerminalSessionInfo describes one live web terminal session.
type TerminalSessionInfo struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	HostID       uint      `json:"hostId"`
	Title        string    `json:"title"`
	Pinned       bool      `json:"pinned"`
	Attached     bool      `json:"attached"`
	CreatedAt    time.Time `json:"createdAt"`
	LastActiveAt time.Time `json:"lastActiveAt"`
	// set only while no websocket is bound
	DetachedAt *time.Time `json:"detachedAt,omitempty"`
	// set only for a pinned session that is currently detached
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type TerminalSessionPin struct {
	ID     string `json:"id" validate:"required"`
	Pinned bool   `json:"pinned"`
}

type TerminalSessionClose struct {
	ID string `json:"id" validate:"required"`
}
