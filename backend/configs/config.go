package configs

type ServerConfig struct {
	Sqlite    Sqlite    `mapstructure:"sqlite"`
	Mysql     Mysql     `mapstructure:"mysql"`
	System    System    `mapstructure:"system"`
	LogConfig LogConfig `mapstructure:"log"`
	CORS      CORS      `mapstructure:"cors"`
	Encrypt   Encrypt   `mapstructure:"encrypt"`
	Csrf      Csrf      `mapstructure:"csrf"`
	Cache     Cache     `mapstructure:"cache"`
}
