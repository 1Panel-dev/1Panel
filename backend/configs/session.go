package configs

type Session struct {
	SessionName string `mapstructure:"session_name"`
	ExpiresTime int    `mapstructure:"expires_time"`
}
