package dto

import "time"

type ComposeTemplateCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type ComposeTemplateUpdate struct {
	ID          uint   `json:"id"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type ComposeTemplateInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
}
