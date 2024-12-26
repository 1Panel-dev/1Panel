package model

type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"not null;"`
	Value string `json:"value"`
	About string `json:"about"`
}

type NodeInfo struct {
	Scope      string `json:"scope"`
	BaseDir    string `json:"baseDir"`
	Version    string `json:"version"`
	EncryptKey string `json:"encryptKey"`
	ServerCrt  string `json:"serverCrt"`
	ServerKey  string `json:"serverKey"`
}
