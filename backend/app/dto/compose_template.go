package dto

import "time"

type ComposeTemplateCreate struct {
	Name        string `json:"name" validate:"required"`
	From        string `json:"from" validate:"required,oneof=edit path"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Content     string `json:"content"`
}

type ComposeTemplateUpdate struct {
	From        string `json:"from" validate:"required,oneof=edit path"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Content     string `json:"content"`
}

type ComposeTemplateInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	From        string    `json:"from"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Content     string    `json:"content"`
}
