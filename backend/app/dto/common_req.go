package dto

type PageInfo struct {
	Page     int `json:"page" validate:"required,number"`
	PageSize int `json:"pageSize" validate:"required,number"`
	Limit    int `json:"limit" validate:"required,number"`
}

type OperationWithName struct {
	Name string `json:"name" validate:"required"`
}

type OperationWithNameAndType struct {
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"`
}

type Login struct {
	Name       string `json:"name" validate:"name,required"`
	Password   string `json:"password" validate:"required"`
	Captcha    string `json:"captcha"`
	CaptchaID  string `json:"captchaID"`
	AuthMethod string `json:"authMethod"`
}
