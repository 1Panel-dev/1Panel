package client

type FireInfo struct {
	ID       uint   `json:"id"`
	Chain    string `json:"chain"`
	Family   string `json:"family"`  // ipv4 ipv6
	Address  string `json:"address"` // Anywhere
	Port     string `json:"port"`
	Protocol string `json:"protocol"` // tcp udp tcp/udp
	Strategy string `json:"strategy"` // accept drop

	Num        string `json:"num"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort"`
	Interface  string `json:"interface"`

	UsedStatus  string `json:"usedStatus"`
	Description string `json:"description"`
}

type Forward struct {
	Num        string `json:"num"`
	Protocol   string `json:"protocol"`
	Port       string `json:"port"`
	TargetIP   string `json:"targetIP"`
	TargetPort string `json:"targetPort"`
	Interface  string `json:"interface"`
}
