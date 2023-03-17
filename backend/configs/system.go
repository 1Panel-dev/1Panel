package configs

type System struct {
	Port       string `mapstructure:"port"`
	DbFile     string `mapstructure:"db_file"`
	DbPath     string `mapstructure:"db_path"`
	LogPath    string `mapstructure:"log_path"`
	DataDir    string `mapstructure:"data_dir"`
	TmpDir     string `mapstructure:"tmp_dir"`
	Cache      string `mapstructure:"cache"`
	Backup     string `mapstructure:"backup"`
	EncryptKey string `mapstructure:"encrypt_key"`
	BaseDir    string `mapstructure:"base_dir"`
	Mode       string `mapstructure:"mode"`
	RepoUrl    string `mapstructure:"repo_url"`
	Version    string `mapstructure:"version"`
	IsDemo     bool   `mapstructure:"is_demo"`
}
