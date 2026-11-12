package dto

type ModelChatMessage struct {
	Role    string `json:"role" validate:"required,oneof=user assistant"`
	Content string `json:"content" validate:"required,max=262144"`
}

type ModelChat struct {
	ID       uint               `json:"id" validate:"required"`
	Model    string             `json:"model" validate:"required,max=255"`
	Messages []ModelChatMessage `json:"messages" validate:"required,min=1,max=200,dive"`
}
