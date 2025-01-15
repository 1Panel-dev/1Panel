package dto

type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"`
	ImagePath string `json:"imagePath"`
}

type UserLoginInfo struct {
	Name      string `json:"name"`
	Token     string `json:"token"`
	MfaStatus string `json:"mfaStatus"`
}

type MfaRequest struct {
	Title    string `json:"title" validate:"required"`
	Interval int    `json:"interval" validate:"required"`
}

type MfaCredential struct {
	Secret   string `json:"secret" validate:"required"`
	Code     string `json:"code" validate:"required"`
	Interval string `json:"interval" validate:"required"`
}

type Login struct {
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required"`
	IgnoreCaptcha bool   `json:"ignoreCaptcha"`
	Captcha       string `json:"captcha"`
	CaptchaID     string `json:"captchaID"`
	AuthMethod    string `json:"authMethod" validate:"required,oneof=jwt session"`
	Language      string `json:"language" validate:"required,oneof=zh en tw ja ko ru ms 'pt-BR'"`
}

type MFALogin struct {
	Name       string `json:"name" validate:"required"`
	Password   string `json:"password" validate:"required"`
	Code       string `json:"code" validate:"required"`
	AuthMethod string `json:"authMethod"`
}
