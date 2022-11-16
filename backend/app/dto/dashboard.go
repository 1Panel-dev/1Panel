package dto

type DashboardBase struct {
	HaloEnabled        bool `json:"haloEnabled"`
	DateeaseEnabled    bool `json:"dateeaseEnabled"`
	JumpServerEnabled  bool `json:"jumpserverEnabled"`
	MeterSphereEnabled bool `json:"metersphereEnabled"`

	WebsiteNumber     int `json:"websiteNumber"`
	DatabaseNumber    int `json:"databaseNumber"`
	CronjobNumber     int `json:"cronjobNumber"`
	AppInstalldNumber int `json:"appInstalldNumber"`

	HostName             string `json:"hostname"`
	Os                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platformFamily"`
	PlatformVersion      string `json:"platformVersion"`
	KernelArch           string `json:"kernelArch"`
	VirtualizationSystem string `json:"virtualizationSystem"`

	CPUCores        int     `json:"cpuCores"`
	CPULogicalCores int     `json:"cpuLogicalCores"`
	CPUModelName    string  `json:"cpuModelName"`
	CPUPercent      float64 `json:"cpuPercent"`
}
