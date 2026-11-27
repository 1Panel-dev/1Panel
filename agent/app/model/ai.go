package model

type OllamaModel struct {
	BaseModel

	Description string `json:"description"`
	Name        string `json:"name"`
	Size        string `json:"size"`
	From        string `json:"from"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}
