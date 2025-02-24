package model

type OllamaModel struct {
	BaseModel

	Name    string `json:"name"`
	Size    string `json:"size"`
	From    string `json:"from"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
