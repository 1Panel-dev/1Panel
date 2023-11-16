package dto

type DeviceBaseInfo struct {
	TimeZone    string       `json:"timeZone"`
	LocalTime   string       `json:"localTime"`
	NameServers []string     `json:"nameServer"`
	Hosts       []HostHelper `json:"hosts"`
}

type HostHelper struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
}

type TimeZoneOptions struct {
	From  string   `json:"from"`
	Zones []string `json:"zones"`
}
