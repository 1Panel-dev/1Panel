package constant

type DBContext string

const (
	TimeOut5s  = 5
	TimeOut20s = 20
	TimeOut5m  = 300

	DateLayout         = "2006-01-02" // or use time.DateOnly while go version >= 1.20
	DefaultDate        = "1970-01-01"
	DateTimeLayout     = "2006-01-02 15:04:05" // or use time.DateTime while go version >= 1.20
	DateTimeSlimLayout = "20060102150405"

	OrderDesc = "descending"
	OrderAsc  = "ascending"

	StatusEnable  = "Enable"
	StatusDisable = "Disable"

	// backup
	S3                  = "S3"
	OSS                 = "OSS"
	Sftp                = "SFTP"
	OneDrive            = "OneDrive"
	MinIo               = "MINIO"
	Cos                 = "COS"
	Kodo                = "KODO"
	WebDAV              = "WebDAV"
	Local               = "LOCAL"
	UPYUN               = "UPYUN"
	ALIYUN              = "ALIYUN"
	OneDriveRedirectURI = "http://localhost/login/authorized"
)
