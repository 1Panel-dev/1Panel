package dto

type DeviceBaseInfo struct {
	DNS       []string     `json:"dns"`
	Hosts     []HostHelper `json:"hosts"`
	Hostname  string       `json:"hostname"`
	TimeZone  string       `json:"timeZone"`
	LocalTime string       `json:"localTime"`
	Ntp       string       `json:"ntp"`
	User      string       `json:"user"`
}

type HostHelper struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
}

type TimeZoneOptions struct {
	From  string   `json:"from"`
	Zones []string `json:"zones"`
}

type ChangePasswd struct {
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}
