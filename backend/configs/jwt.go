package configs

type JWT struct {
	SigningKey  string `mapstructure:"signing_key"`  // jwt签名
	ExpiresTime int64  `mapstructure:"expires_time"` // 过期时间
	BufferTime  int64  `mapstructure:"buffer_time"`  // 缓冲时间
	Issuer      string `mapstructure:"issuer"`
}
