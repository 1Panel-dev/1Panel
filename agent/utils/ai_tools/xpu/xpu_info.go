package xpu

type XpuInfo struct {
	Type          string `json:"type"`
	DriverVersion string `json:"driverVersion"`

	Xpu []Xpu `json:"xpu"`
}

type Xpu struct {
	Basic     Basic     `json:"basic"`
	Stats     Stats     `json:"stats"`
	Processes []Process `json:"processes"`
}

type Basic struct {
	DeviceID      int    `json:"deviceID"`
	DeviceName    string `json:"deviceName"`
	VendorName    string `json:"vendorName"`
	DriverVersion string `json:"driverVersion"`
	Memory        string `json:"memory"`
	FreeMemory    string `json:"freeMemory"`
	PciBdfAddress string `json:"pciBdfAddress"`
}

type Stats struct {
	Power       string `json:"power"`
	Frequency   string `json:"frequency"`
	Temperature string `json:"temperature"`
	MemoryUsed  string `json:"memoryUsed"`
	MemoryUtil  string `json:"memoryUtil"`
}

type Process struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	SHR     string `json:"shr"`
	Memory  string `json:"memory"`
}

type XPUSimpleInfo struct {
	DeviceID    int    `json:"deviceID"`
	DeviceName  string `json:"deviceName"`
	Memory      string `json:"memory"`
	Temperature string `json:"temperature"`
	MemoryUsed  string `json:"memoryUsed"`
	Power       string `json:"power"`
	MemoryUtil  string `json:"memoryUtil"`
}
