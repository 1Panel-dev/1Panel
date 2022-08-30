package dto

type GroupCreate struct {
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type GroupUpdate struct {
	Name string `json:"name" validate:"required"`
}
