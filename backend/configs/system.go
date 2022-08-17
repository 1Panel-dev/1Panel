package configs

type System struct {
	Port   int    `mapstructure:"port"`
	DbType string `mapstructure:"db_type"`
	Level  string `mapstructure:"level"`
}
