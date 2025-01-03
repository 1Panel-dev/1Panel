package configs

type System struct {
	Mode string `mapstructure:"mode"`

	Port       string `mapstructure:"port"`
	Version    string `mapstructure:"version"`
	BaseDir    string `mapstructure:"base_dir"`
	EncryptKey string `mapstructure:"encrypt_key"`

	DbFile  string `mapstructure:"db_agent_file"`
	DbPath  string `mapstructure:"db_path"`
	LogPath string `mapstructure:"log_path"`
	DataDir string `mapstructure:"data_dir"`
	TmpDir  string `mapstructure:"tmp_dir"`
	Cache   string `mapstructure:"cache"`
	Backup  string `mapstructure:"backup"`

	RepoUrl     string `mapstructure:"repo_url"`
	ResourceUrl string `mapstructure:"resource_url"`
	IsDemo      bool   `mapstructure:"is_demo"`
	AppRepo     string `mapstructure:"app_repo"`
}
