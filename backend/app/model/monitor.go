package model

type MonitorBase struct {
	BaseModel
	Cpu float64 `gorm:"type:float" json:"cpu"`

	LoadUsage float64 `gorm:"type:float" json:"loadUsage"`
	CpuLoad1  float64 `gorm:"type:float" json:"cpuLoad1"`
	CpuLoad5  float64 `gorm:"type:float" json:"cpuLoad5"`
	CpuLoad15 float64 `gorm:"type:float" json:"cpuLoad15"`

	Memory float64 `gorm:"type:float" json:"memory"`
}

type MonitorIO struct {
	BaseModel
	Name       string `json:"name"`
	ReadCount  uint64 `gorm:"type:decimal" json:"readCount"`
	WriteCount uint64 `gorm:"type:decimal" json:"writeCount"`
	ReadTime   uint64 `gorm:"type:decimal" json:"readTime"`
	WriteTime  uint64 `gorm:"type:decimal" json:"writeTime"`
	ReadByte   uint64 `gorm:"type:decimal(32)" json:"readByte"`
	WriteByte  uint64 `gorm:"type:decimal(32)" json:"writeByte"`

	Read  uint64 `gorm:"type:decimal" json:"read"`
	Write uint64 `gorm:"type:decimal" json:"write"`
	Count uint64 `gorm:"type:decimal" json:"count"`
	Time  uint64 `gorm:"type:decimal" json:"time"`
}

type MonitorNetwork struct {
	BaseModel
	Name      string  `json:"name"`
	BytesSent uint64  `gorm:"type:decimal(32)" json:"bytesSent"`
	BytesRecv uint64  `gorm:"type:decimal(32)" json:"bytesRecv"`
	Up        float64 `gorm:"type:float" json:"up"`
	Down      float64 `gorm:"type:float" json:"down"`
}
