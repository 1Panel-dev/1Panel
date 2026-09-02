package model

type AlertConfig struct {
	BaseModel
	UID          string `json:"uid"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Config       string `json:"config"`
	SecretConfig string `json:"-"`
}
