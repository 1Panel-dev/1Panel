package global

type ServerConfig struct {
	Base      Base      `mapstructure:"base"`
	RemoteURL RemoteURL `mapstructure:"remote_url"`
	Log       LogConfig `mapstructure:"log"`
}

type Base struct {
	Port       string `mapstructure:"port"`
	Version    string `mapstructure:"version"`
	EncryptKey string `mapstructure:"encrypt_key"`
	Mode       string `mapstructure:"mode"` // xpack [ Enable / Disable ]
	IsDemo     bool   `mapstructure:"is_demo"`
	InstallDir string `mapstructure:"install_dir"`
}

type RemoteURL struct {
	AppRepo     string `mapstructure:"app_repo"`
	RepoUrl     string `mapstructure:"repo_url"`
	ResourceUrl string `mapstructure:"resource_url"`
}

type SystemDir struct {
	BaseDir        string
	DbDir          string
	LogDir         string
	TaskDir        string
	DataDir        string
	TmpDir         string
	LocalBackupDir string

	AppDir               string
	ResourceDir          string
	AppResourceDir       string
	AppInstallDir        string
	LocalAppResourceDir  string
	LocalAppInstallDir   string
	RemoteAppResourceDir string
	CustomAppResourceDir string
	RuntimeDir           string
	RecycleBinDir        string
	SSLLogDir            string
	McpDir               string
}

type LogConfig struct {
	Level     string `mapstructure:"level"`
	TimeZone  string `mapstructure:"timeZone"`
	LogName   string `mapstructure:"log_name"`
	LogSuffix string `mapstructure:"log_suffix"`
	MaxBackup int    `mapstructure:"max_backup"`
}
