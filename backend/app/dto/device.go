package dto

type DeviceBaseInfo struct {
	DNS       []string     `json:"dns"`
	Hosts     []HostHelper `json:"hosts"`
	Hostname  string       `json:"hostname"`
	TimeZone  string       `json:"timeZone"`
	LocalTime string       `json:"localTime"`
	Ntp       string       `json:"ntp"`
	User      string       `json:"user"`

	SwapMemoryTotal     uint64 `json:"swapMemoryTotal"`
	SwapMemoryAvailable uint64 `json:"swapMemoryAvailable"`
	SwapMemoryUsed      uint64 `json:"swapMemoryUsed"`
	MaxSize             uint64 `json:"maxSize"`

	SwapDetails []SwapHelper `json:"swapDetails"`
}

type HostHelper struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
}

type SwapHelper struct {
	Path string `json:"path" validate:"required"`
	Size uint64 `json:"size"`
	Used string `json:"used"`

	IsNew bool `json:"isNew"`
}

type TimeZoneOptions struct {
	From  string   `json:"from"`
	Zones []string `json:"zones"`
}

type ChangePasswd struct {
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}
