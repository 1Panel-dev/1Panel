package dto

type HostNetworkAddress struct {
	Editable bool   `json:"editable"`
	Reason   string `json:"reason"`
	Address  string `json:"address"`
	Family   string `json:"family"`
	Scope    string `json:"scope"`
	Managed  bool   `json:"managed"`
	Applied  bool   `json:"applied"`
}

type HostNetworkInterface struct {
	Backend   string               `json:"backend"`
	Reason    string               `json:"reason"`
	Index     int                  `json:"index"`
	Name      string               `json:"name"`
	State     string               `json:"state"`
	LinkState string               `json:"linkState"`
	MAC       string               `json:"mac"`
	MTU       int                  `json:"mtu"`
	Flags     []string             `json:"flags"`
	Addresses []HostNetworkAddress `json:"addresses"`
}

type HostNetworkAddressRequest struct {
	Interface string `json:"interface" validate:"required,max=15"`
	Address   string `json:"address" validate:"required,cidr"`
}

type HostNetworkRoute struct {
	Editable    bool                      `json:"editable"`
	Reason      string                    `json:"reason"`
	Source      string                    `json:"source"`
	Destination string                    `json:"destination" validate:"required"`
	Gateway     string                    `json:"gateway" validate:"omitempty,ipv4|ipv6"`
	Interface   string                    `json:"interface" validate:"omitempty,max=15"`
	Metric      uint32                    `json:"metric"`
	Table       string                    `json:"table"`
	Protocol    string                    `json:"protocol"`
	Scope       string                    `json:"scope"`
	Family      string                    `json:"family" validate:"required,oneof=IPv4 IPv6"`
	ReadOnly    bool                      `json:"readOnly"`
	NextHops    []HostNetworkRouteNextHop `json:"nextHops"`
	Managed     bool                      `json:"managed"`
	Applied     bool                      `json:"applied"`
}

type HostNetworkRouteNextHop struct {
	Gateway   string `json:"gateway"`
	Interface string `json:"interface"`
	Weight    uint32 `json:"weight"`
}

type HostNetworkRouteBatchCheck struct {
	Routes []HostNetworkRoute `json:"routes" validate:"required,min=1,max=16,dive"`
}

type HostNetworkRouteCheck struct {
	Skipped   bool   `json:"skipped,omitempty"`
	Failed    bool   `json:"failed,omitempty"`
	Reachable bool   `json:"reachable"`
	Message   string `json:"message"`
}

type HostNetworkPersistence struct {
	Supported bool   `json:"supported"`
	OS        string `json:"os"`
	Reason    string `json:"reason"`
}

type HostNetworkAddressChange struct {
	HostNetworkAddressRequest
	PreviousAddress string `json:"previousAddress"`
	Confirmation    string `json:"confirmation" validate:"required"`
}

type HostNetworkRouteChange struct {
	HostNetworkRoute
	Previous     *HostNetworkRoute `json:"previous,omitempty"`
	Confirmation string            `json:"confirmation" validate:"required"`
}

type HostNetworkOverview struct {
	Persistence HostNetworkPersistence `json:"persistence"`
	Interfaces  []HostNetworkInterface `json:"interfaces"`
	Routes      []HostNetworkRoute     `json:"routes"`
}
