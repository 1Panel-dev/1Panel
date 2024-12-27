package client

type FireInfo struct {
	Family   string `json:"family"`  // ipv4 ipv6
	Address  string `json:"address"` // Anywhere
	Port     string `json:"port"`
	Protocol string `json:"protocol"` // tcp udp tcp/udp
	Strategy string `json:"strategy"` // accept drop

	Num        string `json:"num"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort"`

	UsedStatus  string `json:"usedStatus"`
	Description string `json:"description"`
}

type Forward struct {
	Num        string `json:"num"`
	Protocol   string `json:"protocol"`
	Port       string `json:"port"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort"`
}

type IptablesNatInfo struct {
	Num         string `json:"num"`
	Target      string `json:"target"`
	Protocol    string `json:"protocol"`
	Opt         string `json:"opt"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	SrcPort     string `json:"srcPort"`
	DestPort    string `json:"destPort"`
}

type IptablesFilterInfo struct {
	Num         string `json:"num"`
	Target      string `json:"target"`
	Protocol    string `json:"protocol"`
	Opt         string `json:"opt"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}
