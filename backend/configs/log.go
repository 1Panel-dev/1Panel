package configs

type LogConfig struct {
	Level     string `mapstructure:"level"`
	TimeZone  string `mapstructure:"timeZone"`
	Path      string `mapstructure:"path"`
	LogName   string `mapstructure:"log_name"`
	LogSuffix string `mapstructure:"log_suffix"`
	LogSize   int    `mapstructure:"log_size"`
	LogBackup int    `mapstructure:"log_backup"`
	LogData   int    `mapstructure:"log_data"`
}
