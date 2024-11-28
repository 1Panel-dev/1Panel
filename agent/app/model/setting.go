package model

type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"not null;"`
	Value string `json:"value"`
	About string `json:"about"`
}

type NodeInfo struct {
	BaseDir     string `json:"baseDir"`
	Version     string `json:"version"`
	MasterAddr  string `json:"masterAddr"`
	EncryptKey  string `json:"encryptKey"`
	ServerCrt   string `json:"serverCrt"`
	ServerKey   string `json:"serverKey"`
	CurrentNode string `json:"currentNode"`
}
