package configs

type ServerConfig struct {
	BaseDir   string    `mapstructure:"base_dir"`
	System    System    `mapstructure:"system"`
	LogConfig LogConfig `mapstructure:"log"`
}
