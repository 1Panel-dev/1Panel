package configs

type ServerConfig struct {
	Sqlite    Sqlite    `mapstructure:"sqlite"`
	Mysql     Mysql     `mapstructure:"mysql"`
	System    System    `mapstructure:"system"`
	LogConfig LogConfig `mapstructure:"log"`
	JWT       JWT       `mapstructure:"jwt"`
	CORS      CORS      `mapstructure:"cors"`
}
