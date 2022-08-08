package configs

type Session struct {
	SessionKey     string `mapstructure:"session_key"`
	SessionUserKey string `mapstructure:"session_user_key"`
	SessionName    string `mapstructure:"session_name"`
	ExpiresTime    int    `mapstructure:"expires_time"`
}
