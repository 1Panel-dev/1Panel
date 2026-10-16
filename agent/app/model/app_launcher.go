package model

type AppLauncher struct {
	BaseModel
	Key string `json:"key"`
}

type QuickDirectory struct {
	BaseModel
	Path   string `json:"path"`
	Alias  string `json:"alias"`
	Legacy bool   `json:"legacy" gorm:"-"`
}

type QuickJump struct {
	BaseModel
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Recommend int    `json:"recommend"`
	IsShow    bool   `json:"isShow"`
	Router    string `json:"router"`
}
