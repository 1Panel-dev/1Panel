package dto

type GroupCreate struct {
	ID   uint   `json:"id"`
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type GroupSearch struct {
	Type string `json:"type" validate:"required"`
}

type GroupUpdate struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type" validate:"required"`
	IsDefault bool   `json:"isDefault"`
}

type GroupInfo struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"isDefault"`
}
