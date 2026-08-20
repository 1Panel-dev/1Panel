package xpu

type DeviceUtilByProc struct {
	DeviceID      int     `json:"device_id"`
	MemSize       float64 `json:"mem_size"`
	ProcessID     int     `json:"process_id"`
	ProcessName   string  `json:"process_name"`
	SharedMemSize float64 `json:"shared_mem_size"`
}

type DeviceUtilByProcList struct {
	DeviceUtilByProcList []DeviceUtilByProc `json:"device_util_by_proc_list"`
}

type discoveryDevice struct {
	DeviceFunctionType string `json:"device_function_type"`
	DeviceID           int    `json:"device_id"`
	DeviceName         string `json:"device_name"`
	DeviceType         string `json:"device_type"`
	DrmDevice          string `json:"drm_device"`
	PciBdfAddress      string `json:"pci_bdf_address"`
	PciDeviceID        string `json:"pci_device_id"`
	UUID               string `json:"uuid"`
	VendorName         string `json:"vendor_name"`

	MemoryPhysicalSizeByte string `json:"memory_physical_size_byte"`
	MemoryFreeSizeByte     string `json:"memory_free_size_byte"`
	DriverVersion          string `json:"driver_version"`
}

type discoveryInfo struct {
	DeviceList []discoveryDevice `json:"device_list"`
}

type DeviceLevelMetric struct {
	MetricsType string  `json:"metrics_type"`
	Value       float64 `json:"value"`
}

type DeviceStats struct {
	DeviceID    int                 `json:"device_id"`
	DeviceLevel []DeviceLevelMetric `json:"device_level"`
}

type Info struct {
	Type          string `json:"type"`
	DriverVersion string `json:"driverVersion"`

	Devices []Device `json:"xpu"`
}

type Device struct {
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
	GPUUtil     string `json:"gpuUtil"`
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
