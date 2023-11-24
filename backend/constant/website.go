package constant

const (
	WebRunning = "Running"
	WebStopped = "Stopped"

	DateLayout     = "2006-01-02"
	DateTimeLayout = "2006-01-02 15:04:05"
	DefaultDate    = "1970-01-01"

	ProtocolHTTP  = "HTTP"
	ProtocolHTTPS = "HTTPS"

	NewApp       = "new"
	InstalledApp = "installed"

	Deployment = "deployment"
	Static     = "static"
	Proxy      = "proxy"
	Runtime    = "runtime"

	SSLExisted = "existed"
	SSLAuto    = "auto"
	SSLManual  = "manual"

	DNSAccount = "dnsAccount"
	DnsManual  = "dnsManual"
	Http       = "http"
	Manual     = "manual"
	SelfSigned = "selfSigned"

	StartWeb = "start"
	StopWeb  = "stop"

	HTTPSOnly   = "HTTPSOnly"
	HTTPAlso    = "HTTPAlso"
	HTTPToHTTPS = "HTTPToHTTPS"

	GetLog     = "get"
	DisableLog = "disable"
	EnableLog  = "enable"
	DeleteLog  = "delete"

	AccessLog = "access.log"
	ErrorLog  = "error.log"

	ConfigPHP = "php"
	ConfigFPM = "fpm"

	SSLInit       = "init"
	SSLError      = "error"
	SSLReady      = "ready"
	SSLApply      = "applying"
	SSLApplyError = "applyError"
)
