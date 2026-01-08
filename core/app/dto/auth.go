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

type PasskeyBeginResponse struct {
	SessionID string      `json:"sessionId"`
	PublicKey interface{} `json:"publicKey"`
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
	Name      string `json:"name" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Captcha   string `json:"captcha"`
	CaptchaID string `json:"captchaID"`
	Language  string `json:"language" validate:"required,oneof=zh en 'zh-Hant' ko ja ru ms 'pt-BR' tr 'es-ES'"`
}

type MFALogin struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type SystemSetting struct {
	IsDemo   bool   `json:"isDemo"`
	Language string `json:"language"`
	IsIntl   bool   `json:"isIntl"`
}
