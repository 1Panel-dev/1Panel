package configs

type System struct {
	Port    int    `mapstructure:"port"`
	DbType  string `mapstructure:"db_type"`
	DataDir string `mapstructure:"data_dir"`
	Cache   string `mapstructure:"cache"`
	Backup  string `mapstructure:"backup"`
	AppOss  string `mapstructure:"app_oss"`
}
