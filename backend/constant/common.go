package constant

type DBContext string

const (
	DB DBContext = "db"

	SystemRestart = "systemRestart"

	TypeWebsite = "website"
	TypePhp     = "php"
	TypeSSL     = "ssl"
	TypeSystem  = "system"
)

const (
	TimeOut5s  = 5
	TimeOut20s = 20
	TimeOut5m  = 300
)
