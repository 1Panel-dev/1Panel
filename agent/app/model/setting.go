package model

type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"not null;"`
	Value string `json:"value"`
	About string `json:"about"`
}

type NodeInfo struct {
	Scope     string `json:"scope"`
	BaseDir   string `json:"baseDir"`
	NodePort  uint   `json:"nodePort"`
	Version   string `json:"version"`
	ServerCrt string `json:"serverCrt"`
	ServerKey string `json:"serverKey"`
}
