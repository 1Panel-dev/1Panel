package configs

type JWT struct {
	HeaderName  string `mapstructure:"header_name"`
	SigningKey  string `mapstructure:"signing_key"`
	ExpiresTime int64  `mapstructure:"expires_time"`
	BufferTime  int64  `mapstructure:"buffer_time"`
	Issuer      string `mapstructure:"issuer"`
}
