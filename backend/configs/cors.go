package configs

type CORS struct {
	Mode      string          `mapstructure:"mode"`
	WhiteList []CORSWhiteList `mapstructure:"whitelist"`
}

type CORSWhiteList struct {
	AllowOrigin      string `mapstructure:"allow-origin"`
	AllowMethods     string `mapstructure:"allow-methods"`
	AllowHeaders     string `mapstructure:"allow-headers"`
	ExposeHeaders    string `mapstructure:"expose-headers"`
	AllowCredentials bool   `mapstructure:"allow-credentials"`
}
