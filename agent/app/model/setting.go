package model

type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"not null;"`
	Value string `json:"value"`
	About string `json:"about"`
}

type CommonDescription struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	DetailType  string `json:"detailType"`
	IsPinned    bool   `json:"isPinned"`
	Description string `json:"description"`
}

type NodeInfo struct {
	Scope     string `json:"scope"`
	BaseDir   string `json:"baseDir"`
	NodePort  uint   `json:"nodePort"`
	Version   string `json:"version"`
	ServerCrt string `json:"serverCrt"`
	ServerKey string `json:"serverKey"`
}

type LocalConnInfo struct {
	Addr       string `json:"addr"`
	Port       uint   `json:"port"`
	User       string `json:"user"`
	AuthMode   string `json:"authMode"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	PassPhrase string `json:"passPhrase"`
}
