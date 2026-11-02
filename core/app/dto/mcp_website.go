package dto

type MCPWebsiteBind struct {
	RequestID string `json:"requestId"`
	Revision  uint64 `json:"revision"`
	Domain    string `json:"domain"`
	EnableSSL bool   `json:"enableSSL"`
	SSLID     uint   `json:"sslID"`
}

type MCPWebsiteUpdate struct {
	RequestID string `json:"requestId"`
	Revision  uint64 `json:"revision"`
	EnableSSL bool   `json:"enableSSL"`
	SSLID     uint   `json:"sslID"`
}

type MCPWebsiteUnbind struct {
	RequestID     string `json:"requestId"`
	Revision      uint64 `json:"revision"`
	DeleteWebsite bool   `json:"deleteWebsite,omitempty"`
}

type MCPWebsitePendingRequest struct {
	Operation     string `json:"operation"`
	RequestID     string `json:"requestId"`
	Domain        string `json:"domain"`
	EnableSSL     bool   `json:"enableSSL"`
	SSLID         uint   `json:"sslID"`
	DeleteWebsite bool   `json:"deleteWebsite,omitempty"`
}

type MCPWebsiteStatus struct {
	State          string                    `json:"state"`
	WebsiteID      uint                      `json:"websiteId"`
	Domain         string                    `json:"domain"`
	Endpoint       string                    `json:"endpoint"`
	EnableSSL      bool                      `json:"enableSSL"`
	SSLID          uint                      `json:"sslID"`
	WebsiteStatus  string                    `json:"websiteStatus"`
	RequestID      string                    `json:"requestId"`
	Revision       uint64                    `json:"revision"`
	Reason         string                    `json:"reason,omitempty"`
	PendingRequest *MCPWebsitePendingRequest `json:"pendingRequest,omitempty"`
	WebsiteDeleted bool                      `json:"websiteDeleted,omitempty"`
}
