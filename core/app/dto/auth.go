package dto

type CaptchaResponse struct {
	CaptchaID string `json:"captchaID"`
	ImagePath string `json:"imagePath"`
}

type UserLoginInfo struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Token      string `json:"token"`
	MfaStatus  string `json:"mfaStatus"`
	MfaSession string `json:"mfaSession"`
}

type PasskeyBeginResponse struct {
	SessionID string      `json:"sessionId"`
	PublicKey interface{} `json:"publicKey"`
}

type Login struct {
	Name      string `json:"name" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Captcha   string `json:"captcha"`
	CaptchaID string `json:"captchaID"`
	Language  string `json:"language" validate:"required,oneof=zh en 'zh-Hant' ko ja ru ms 'pt-BR' tr 'es-ES' fa"`
}

type SystemSetting struct {
	IsDemo   bool   `json:"isDemo"`
	Language string `json:"language"`
	IsIntl   bool   `json:"isIntl"`
}

type PasskeyID struct {
	ID string `json:"id" validate:"required"`
}

// mfa
type MFALogin struct {
	SessionID string `json:"sessionId" validate:"required"`
	Code      string `json:"code" validate:"required"`
}

type MfaRequest struct {
	Title    string `json:"title" validate:"required"`
	Interval int    `json:"interval" validate:"required"`
}

type MfaCredential struct {
	Secret   string `json:"secret" validate:"required"`
	Code     string `json:"code" validate:"required"`
	Interval int    `json:"interval" validate:"required"`
}

type ApiInterfaceConfig struct {
	ApiInterfaceStatus string `json:"apiInterfaceStatus"`
	ApiKey             string `json:"apiKey"`
	IpWhiteList        string `json:"ipWhiteList"`
	ApiKeyValidityTime int    `json:"apiKeyValidityTime"`
}

type CurrentUserInfo struct {
	Name              string `json:"name"`
	MFAStatus         string `json:"mfaStatus"`
	MFAInterval       int    `json:"mfaInterval"`
	ComplexitySetting string `json:"complexitySetting"`

	ApiInterfaceStatus string `json:"apiInterfaceStatus"`
	ApiKey             string `json:"apiKey"`
	IpWhiteList        string `json:"ipWhiteList"`
	ApiKeyValidityTime int    `json:"apiKeyValidityTime"`

	Role        string                `json:"role"`
	Permissions []string              `json:"permissions"`
	NodeRoles   []CurrentUserNodeRole `json:"nodeRoles"`
}

type CurrentUserNodeRole struct {
	NodeID   uint   `json:"nodeId"`
	NodeName string `json:"nodeName"`
	RoleID   uint   `json:"roleId"`
	RoleName string `json:"roleName"`
}

type CurrentUserUpdate struct {
	Name        string `json:"name" validate:"required"`
	Password    string `json:"password"`
	OldPassword string `json:"oldPassword"`
}
