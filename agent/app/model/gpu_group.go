package model

type GPUGroup struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageType   string `json:"imageType"`
	Config      string `json:"config"`
	Revision    uint   `json:"revision"`
}
