package configs

type ServerConfig struct {
	Sqlite    Sqlite    `mapstructure:"sqlite"`
	Mysql     Mysql     `mapstructure:"mysql"`
	System    System    `mapstructure:"system"`
	LogConfig LogConfig `mapstructure:"log"`
	JWT       JWT       `mapstructure:"jwt"`
	Session   Session   `mapstructure:"session"`
	CORS      CORS      `mapstructure:"cors"`
	Captcha   Captcha   `mapstructure:"captcha"`
	Encrypt   Encrypt   `mapstructure:"encrypt"`
	Csrf      Csrf      `mapstructure:"csrf"`
}
