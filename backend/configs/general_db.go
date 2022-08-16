package configs

type GeneralDB struct {
	Path         string `mapstructure:"path"`           // 服务器地址
	Port         string `mapstructure:"port"`           //:端口
	Config       string `mapstructure:"config"`         // 高级配置
	Dbname       string `mapstructure:"db_name"`        // 数据库名
	Username     string `mapstructure:"username"`       // 数据库用户名
	Password     string `mapstructure:"password"`       // 数据库密码
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 空闲中的最大连接数
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 打开到数据库的最大连接数
	DbFile       string `mapstructure:"db_file"`
}
