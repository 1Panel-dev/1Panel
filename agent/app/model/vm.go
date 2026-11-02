package model

type VirtualMachine struct {
	BaseModel

	Name         string   `json:"name"`
	UUID         string   `json:"uuid"`
	CPU          int      `json:"cpu"`
	Memory       int64    `json:"memory"`
	DiskSize     int64    `json:"diskSize"`
	DiskBus      string   `json:"diskBus"`
	DiskPath     string   `json:"diskPath"`
	StoragePool  string   `json:"storagePool"`
	ISOPaths     []string `json:"isoPaths" gorm:"serializer:json"`
	ImagePath    string   `json:"imagePath"`
	Network      string   `json:"network"`
	BootOrder    string   `json:"bootOrder"`
	AutoStart    bool     `json:"autoStart"`
	Status       string   `json:"status"`
	TaskID       string   `json:"taskID"`
	CreateTaskID string   `json:"createTaskID"`
	Description  string   `json:"description"`
}

type VMTemplate struct {
	BaseModel

	Name        string `json:"name"`
	SourceVM    string `json:"sourceVM"`
	CPU         int    `json:"cpu"`
	Memory      int64  `json:"memory"`
	DiskSize    int64  `json:"diskSize"`
	DiskBus     string `json:"diskBus"`
	DiskPath    string `json:"diskPath"`
	StoragePool string `json:"storagePool"`
	Network     string `json:"network"`
	BootOrder   string `json:"bootOrder"`
	Description string `json:"description"`
}

type VMIso struct {
	BaseModel

	Type         string `json:"type"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Status       string `json:"status"`
	Error        string `json:"error"`
	TempFileID   string `json:"tempFileID"`
	TempFileType string `json:"tempFileType"`
	Description  string `json:"description"`
}

type VMNetwork struct {
	BaseModel

	Name            string `json:"name"`
	Type            string `json:"type"`
	BridgeName      string `json:"bridgeName"`
	SourceInterface string `json:"sourceInterface"`
	CIDR            string `json:"cidr" gorm:"column:cidr"`
	Gateway         string `json:"gateway"`
	DHCPStart       string `json:"dhcpStart"`
	DHCPEnd         string `json:"dhcpEnd"`
	AutoStart       bool   `json:"autoStart"`
	Status          string `json:"status"`
	Description     string `json:"description"`
	Error           string `json:"error"`
}

type VMStorage struct {
	BaseModel

	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	AutoStart   bool   `json:"autoStart"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Error       string `json:"error"`
}
