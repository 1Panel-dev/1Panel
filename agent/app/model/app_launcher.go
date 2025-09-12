package model

type AppLauncher struct {
	BaseModel
	Key string `json:"key"`
}

type QuickJump struct {
	BaseModel
	Name      string `json:"name"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Recommend int    `json:"recommend"`
	IsShow    bool   `json:"isShow"`
	Router    string `json:"router"`
}
