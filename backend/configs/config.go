package configs

type ServerConfig struct {
	BaseDir   string    `mapstructure:"base_dir"`
	System    System    `mapstructure:"system"`
	Sqlite    Sqlite    `mapstructure:"sqlite"`
	LogConfig LogConfig `mapstructure:"log"`
	CORS      CORS      `mapstructure:"cors"`
	Encrypt   Encrypt   `mapstructure:"encrypt"`
}
