package dto

type AppDatabase struct {
	ServiceName string `json:"PANEL_DB_HOST"`
	DbName      string `json:"PANEL_DB_NAME"`
	DbUser      string `json:"PANEL_DB_USER"`
	Password    string `json:"PANEL_DB_USER_PASSWORD"`
}

type AuthParam struct {
	RootPassword string `json:"PANEL_DB_ROOT_PASSWORD"`
	RootUser     string `json:"PANEL_DB_ROOT_USER"`
}

type RedisAuthParam struct {
	RootPassword string `json:"PANEL_REDIS_ROOT_PASSWORD"`
}

type MinioAuthParam struct {
	RootPassword string `json:"PANEL_MINIO_ROOT_PASSWORD"`
	RootUser     string `json:"PANEL_MINIO_ROOT_USER"`
}

type ContainerExec struct {
	ContainerName string      `json:"containerName"`
	DbParam       AppDatabase `json:"dbParam"`
	Auth          AuthParam   `json:"auth"`
}

type AppOssConfig struct {
	Version string `json:"version"`
	Package string `json:"package"`
}

type AppVersion struct {
	Version       string `json:"version"`
	DetailId      uint   `json:"detailId"`
	DockerCompose string `json:"dockerCompose"`
}

type AppList struct {
	Valid        bool     `json:"valid"`
	Violations   []string `json:"violations"`
	LastModified int      `json:"lastModified"`

	Apps  []AppDefine     `json:"apps"`
	Extra ExtraProperties `json:"additionalProperties"`
}

type AppDefine struct {
	Icon         string `json:"icon"`
	Name         string `json:"name"`
	ReadMe       string `json:"readMe"`
	LastModified int    `json:"lastModified"`

	AppProperty AppProperty        `json:"additionalProperties"`
	Versions    []AppConfigVersion `json:"versions"`
}

type LocalAppAppDefine struct {
	AppProperty AppProperty `json:"additionalProperties" yaml:"additionalProperties"`
}

type LocalAppParam struct {
	AppParams LocalAppInstallDefine `json:"additionalProperties" yaml:"additionalProperties"`
}

type LocalAppInstallDefine struct {
	FormFields interface{} `json:"formFields" yaml:"formFields"`
}

type ExtraProperties struct {
	Tags    []Tag  `json:"tags"`
	Version string `json:"version"`
}

type AppProperty struct {
	Name               string   `json:"name"`
	Type               string   `json:"type"`
	Tags               []string `json:"tags"`
	ShortDescZh        string   `json:"shortDescZh"`
	ShortDescEn        string   `json:"shortDescEn"`
	Description        Locale   `json:"description"`
	Key                string   `json:"key"`
	Required           []string `json:"Required"`
	CrossVersionUpdate bool     `json:"crossVersionUpdate"`
	Limit              int      `json:"limit"`
	Recommend          int      `json:"recommend"`
	Website            string   `json:"website"`
	Github             string   `json:"github"`
	Document           string   `json:"document"`
	Version            float64  `json:"version"`
	GpuSupport         bool     `json:"gpuSupport"`
}

type AppConfigVersion struct {
	Name                string      `json:"name"`
	LastModified        int         `json:"lastModified"`
	DownloadUrl         string      `json:"downloadUrl"`
	DownloadCallBackUrl string      `json:"downloadCallBackUrl"`
	AppForm             interface{} `json:"additionalProperties"`
}

type Tag struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Sort    int    `json:"sort"`
	Locales Locale `json:"locales"`
}

type Locale struct {
	En     string `json:"en"`
	Ja     string `json:"ja"`
	Ms     string `json:"ms"`
	PtBr   string `json:"pt-br" yaml:"pt-br"`
	Ru     string `json:"ru"`
	ZhHant string `json:"zh-hant" yaml:"zh-hant"`
	Zh     string `json:"zh"`
	Ko     string `json:"ko"`
}

type AppForm struct {
	FormFields     []AppFormFields `json:"formFields"`
	SupportVersion float64         `json:"supportVersion"`
}

type AppFormFields struct {
	Type     string         `json:"type"`
	LabelZh  string         `json:"labelZh"`
	LabelEn  string         `json:"labelEn"`
	Label    Locale         `json:"label"`
	Required bool           `json:"required"`
	Default  interface{}    `json:"default"`
	EnvKey   string         `json:"envKey"`
	Disabled bool           `json:"disabled"`
	Edit     bool           `json:"edit"`
	Rule     string         `json:"rule"`
	Multiple bool           `json:"multiple"`
	Child    interface{}    `json:"child"`
	Values   []AppFormValue `json:"values"`
}

type AppFormValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type AppResource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

var AppToolMap = map[string]string{
	"mysql": "phpmyadmin",
	"redis": "redis-commander",
}

type AppInstallInfo struct {
	ID   uint   `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}
