package constant

type DBContext string

const (
	DB DBContext = "db"

	SystemRestart = "systemRestart"

	TypeWebsite       = "website"
	TypePhp           = "php"
	TypeSSL           = "ssl"
	TypeSystem        = "system"
	TypeTask          = "task"
	TypeImagePull     = "image-pull"
	TypeImagePush     = "image-push"
	TypeImageBuild    = "image-build"
	TypeComposeCreate = "compose-create"
)

const (
	TimeOut5s  = 5
	TimeOut20s = 20
	TimeOut5m  = 300

	DateLayout               = "2006-01-02"          // or use time.DateOnly while go version >= 1.20
	DateTimeLayout           = "2006-01-02 15:04:05" // or use time.DateTime while go version >= 1.20
	DateTimeSlimLayout       = "20060102150405"
	WebsiteDefaultExpireDate = "9999-12-31"
)

const (
	DirPerm  = 0755
	FilePerm = 0644
)
