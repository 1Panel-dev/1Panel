package configs

type System struct {
	Port        int    `mapstructure:"port"`
	DbType      string `mapstructure:"db_type"`
	Level       string `mapstructure:"level"`
	DataDir     string `mapstructure:"data_dir"`
	ResourceDir string `mapstructure:"resource_dir"`
	AppDir      string `mapstructure:"app_dir"`
	AppOss      string `mapstructure:"app_oss"`
}
