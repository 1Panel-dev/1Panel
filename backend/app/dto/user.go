package dto

type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"`
	ImagePath string `json:"imagePath"`
}

type UserLoginInfo struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
