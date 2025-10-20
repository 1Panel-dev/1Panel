package global

type ServerConfig struct {
	Base      Base      `mapstructure:"base"`
	Conn      Conn      `mapstructure:"conn"`
	RemoteURL RemoteURL `mapstructure:"remote_url"`
	LogConfig LogConfig `mapstructure:"log"`
}

type Base struct {
	Mode           string `mapstructure:"mode"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Language       string `mapstructure:"language"`
	IsDemo         bool   `mapstructure:"is_demo"`
	IsIntl         bool   `mapstructure:"is_intl"`
	IsOffLine      bool   `mapstructure:"is_offline"`
	IsFxplay       bool   `mapstructure:"is_fxplay"`
	Version        string `mapstructure:"version"`
	InstallDir     string `mapstructure:"install_dir"`
	ChangeUserInfo string `mapstructure:"change_user_info"`
	EncryptKey     string `mapstructure:"encrypt_key"`
}

type Conn struct {
	Port        string `mapstructure:"port"`
	BindAddress string `mapstructure:"bindAddress"`
	Ipv6        string `mapstructure:"ipv6"`
	SSL         string `mapstructure:"ssl"`
	Entrance    string `mapstructure:"entrance"`
}

type ApiInterface struct {
	ApiKey             string `mapstructure:"api_key"`
	ApiInterfaceStatus string `mapstructure:"api_interface_status"`
	IpWhiteList        string `mapstructure:"ip_white_list"`
	ApiKeyValidityTime string `mapstructure:"api_key_validity_time"`
}

type RemoteURL struct {
	RepoUrl     string `mapstructure:"repo_url"`
	ResourceURL string `mapstructure:"resource_url"`
}

type LogConfig struct {
	Level     string `mapstructure:"level"`
	TimeZone  string `mapstructure:"timeZone"`
	LogName   string `mapstructure:"log_name"`
	LogSuffix string `mapstructure:"log_suffix"`
	MaxBackup int    `mapstructure:"max_backup"`
}
