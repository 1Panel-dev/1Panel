package configs

type System struct {
	Port         string `mapstructure:"port"`
	DbFile       string `mapstructure:"db_file"`
	DbPath       string `mapstructure:"db_path"`
	LogPath      string `mapstructure:"log_path"`
	DataDir      string `mapstructure:"data_dir"`
	TmpDir       string `mapstructure:"tmp_dir"`
	Cache        string `mapstructure:"cache"`
	Backup       string `mapstructure:"backup"`
	AppRepoOwner string `mapstructure:"app_repo_owner"`
	AppRepoName  string `mapstructure:"app_repo_name"`
	EncryptKey   string `mapstructure:"encrypt_key"`
	BaseDir      string `mapstructure:"base_dir"`
	Mode         string `mapstructure:"mode"`
}
