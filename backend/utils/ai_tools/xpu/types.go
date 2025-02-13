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

type Device struct {
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

type DeviceInfo struct {
	DeviceList []Device `json:"device_list"`
}

type DeviceLevelMetric struct {
	MetricsType string  `json:"metrics_type"`
	Value       float64 `json:"value"`
}

type DeviceStats struct {
	DeviceID    int                 `json:"device_id"`
	DeviceLevel []DeviceLevelMetric `json:"device_level"`
}
