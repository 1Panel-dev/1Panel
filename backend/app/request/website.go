package request

type WebsiteWafReq struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Key       string `json:"key" validate:"required"`
	Rule      string `json:"rule" validate:"required"`
}

type WebsiteWafUpdate struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Key       string `json:"key" validate:"required"`
	Enable    bool   `json:"enable" validate:"required"`
}
