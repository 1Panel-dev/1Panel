package configs

type System struct {
	Port         string `mapstructure:"port"`
	DbFile       string `mapstructure:"db_file"`
	DbPath       string `mapstructure:"db_path"`
	LogPath      string `mapstructure:"log_path"`
	DataDir      string `mapstructure:"data_dir"`
	Cache        string `mapstructure:"cache"`
	Backup       string `mapstructure:"backup"`
	AppOss       string `mapstructure:"app_oss"`
	AppRepoOwner string `mapstructure:"app_repo_owner"`
	AppRepoName  string `mapstructure:"app_repo_name"`
}
