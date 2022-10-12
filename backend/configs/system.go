package configs

type System struct {
	Port    int    `mapstructure:"port"`
	DbType  string `mapstructure:"db_type"`
	Level   string `mapstructure:"level"`
	DataDir string `mapstructure:"data_dir"`
	AppOss  string `mapstructure:"app_oss"`
}
