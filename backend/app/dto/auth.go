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

type MfaCredential struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

type Login struct {
	Name       string `json:"name"`
	Password   string `json:"password"`
	Captcha    string `json:"captcha"`
	CaptchaID  string `json:"captchaID"`
	AuthMethod string `json:"authMethod"`
}

type MFALogin struct {
	Name       string `json:"name"`
	Password   string `json:"password"`
	Code       string `json:"code"`
	AuthMethod string `json:"authMethod"`
}

type InitUser struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}
