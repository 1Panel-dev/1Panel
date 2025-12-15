package constant

const (
	WebRunning = "Running"
	WebStopped = "Stopped"

	ProtocolHTTP   = "HTTP"
	ProtocolHTTPS  = "HTTPS"
	ProtocolStream = "TCP/UDP"

	NewApp       = "new"
	InstalledApp = "installed"

	Deployment = "deployment"
	Static     = "static"
	Proxy      = "proxy"
	Runtime    = "runtime"
	Subsite    = "subsite"
	Stream     = "stream"

	SSLExisted = "existed"
	SSLAuto    = "auto"
	SSLManual  = "manual"

	DNSAccount = "dnsAccount"
	DnsManual  = "dnsManual"
	Http       = "http"
	Manual     = "manual"
	SelfSigned = "selfSigned"
	FromMaster = "fromMaster"

	StartWeb = "start"
	StopWeb  = "stop"
	SetHttps = "setHttps"

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
