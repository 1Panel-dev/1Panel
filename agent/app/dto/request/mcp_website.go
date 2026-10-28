package request

type MCPWebsiteBinding struct {
	Operation string `json:"operation"`
	RequestID string `json:"requestId"`
	BindingID string `json:"bindingId"`
	WebsiteID uint   `json:"websiteId"`
	Domain    string `json:"domain"`
	EnableSSL bool   `json:"enableSSL"`
	SSLID     uint   `json:"sslID"`
	Upstream  string `json:"upstream"`
	CoreHost  string `json:"coreHost"`
	ProxyKey  string `json:"proxyKey"`
}
