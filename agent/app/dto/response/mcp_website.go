package response

type MCPWebsiteBinding struct {
	State         string `json:"state"`
	WebsiteID     uint   `json:"websiteId"`
	BindingID     string `json:"bindingId"`
	RequestID     string `json:"requestId"`
	Domain        string `json:"domain"`
	Endpoint      string `json:"endpoint"`
	EnableSSL     bool   `json:"enableSSL"`
	SSLID         uint   `json:"sslID"`
	WebsiteStatus string `json:"websiteStatus"`
	Reason        string `json:"reason"`
	Code          string `json:"code,omitempty"`
	Error         string `json:"error,omitempty"`
	Definitive    bool   `json:"definitive,omitempty"`
}
