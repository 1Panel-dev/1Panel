package client

type FireInfo struct {
	Family   string `json:"family"`  // ipv4 ipv6
	Address  string `json:"address"` // Anywhere
	Port     string `json:"port"`
	Protocol string `json:"protocol"` // tcp udp tcp/udp
	Strategy string `json:"strategy"` // accept drop

	UsedStatus  string `json:"usedStatus"`
	Description string `json:"description"`
}

type Forward struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Target   string `json:"target"`
}
