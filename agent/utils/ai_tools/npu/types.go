package npu

type Info struct {
	Type          string   `json:"type"`
	DriverVersion string   `json:"driverVersion"`
	Devices       []Device `json:"npu"`
}

type Device struct {
	Type        string `json:"type"`
	Index       uint   `json:"index"`
	NPUIndex    uint   `json:"npuIndex"`
	ChipIndex   uint   `json:"chipIndex"`
	ProductName string `json:"productName"`
	BusID       string `json:"busID"`

	Health         string    `json:"health"`
	Temperature    string    `json:"temperature"`
	PowerDraw      string    `json:"powerDraw"`
	AICore         string    `json:"aiCore"`
	MemUsed        string    `json:"memUsed"`
	MemTotal       string    `json:"memTotal"`
	MemoryUsed     string    `json:"memoryUsed"`
	MemoryTotal    string    `json:"memoryTotal"`
	HBMUsed        string    `json:"hbmUsed"`
	HBMTotal       string    `json:"hbmTotal"`
	HugepagesUsed  string    `json:"hugepagesUsed"`
	HugepagesTotal string    `json:"hugepagesTotal"`
	Processes      []Process `json:"processes"`
}

type Process struct {
	PID         string `json:"pid"`
	ProcessName string `json:"processName"`
	UsedMemory  string `json:"usedMemory"`
}
