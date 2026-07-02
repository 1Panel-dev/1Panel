package constant

import "sync/atomic"

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

	// backup
	S3          = "S3"
	OSS         = "OSS"
	Sftp        = "SFTP"
	OneDrive    = "OneDrive"
	MinIo       = "MINIO"
	Cos         = "COS"
	Kodo        = "KODO"
	WebDAV      = "WebDAV"
	Local       = "LOCAL"
	UPYUN       = "UPYUN"
	ALIYUN      = "ALIYUN"
	GoogleDrive = "GoogleDrive"

	OneDriveRedirectURI = "http://localhost/login/authorized"
)

const (
	DirPerm  = 0755
	FilePerm = 0644
)

const (
	SyncSystemProxy                  = "SyncSystemProxy"
	SyncScripts                      = "SyncScripts"
	SyncBackupAccounts               = "SyncBackupAccounts"
	SyncAlertSetting                 = "SyncAlertSetting"
	SyncCustomApp                    = "SyncCustomApp"
	SyncLanguage                     = "SyncLanguage"
	SyncEdition                      = "SyncEdition"
	SyncSystemProxyWithRestartDocker = "SyncSystemProxyWithRestartDocker"
)

var WebUrlMap = map[string]struct{}{
	"/apps":           {},
	"/apps/all":       {},
	"/apps/installed": {},
	"/apps/upgrade":   {},
	"/apps/setting":   {},

	"/ai":                       {},
	"/ai/ai-proxy":              {},
	"/ai/ai-proxy/model-pool":   {},
	"/ai/ai-proxy/api-keys":     {},
	"/ai/ai-proxy/groups":       {},
	"/ai/ai-proxy/model-groups": {},
	"/ai/ai-proxy/usage":        {},
	"/ai/ai-proxy/content":      {},
	"/ai/skills-hub":            {},
	"/ai/skills-hub/targets":    {},
	"/ai/benchmark":             {},
	"/ai/model/account":         {},
	"/ai/model/local":           {},
	"/ai/gpu":                   {},
	"/ai/gpu/current":           {},
	"/ai/gpu/history":           {},
	"/ai/mcp":                   {},
	"/ai/agents/agent":          {},

	"/containers":                   {},
	"/containers/container/operate": {},
	"/containers/container":         {},
	"/containers/image":             {},
	"/containers/network":           {},
	"/containers/volume":            {},
	"/containers/repo":              {},
	"/containers/compose":           {},
	"/containers/template":          {},
	"/containers/setting":           {},
	"/containers/dashboard":         {},

	"/cronjobs":                 {},
	"/cronjobs/cronjob":         {},
	"/cronjobs/library":         {},
	"/cronjobs/cronjob/operate": {},

	"/databases":                   {},
	"/databases/mysql":             {},
	"/databases/mysql/remote":      {},
	"/databases/mongodb":           {},
	"/databases/mongodb/remote":    {},
	"/databases/postgresql":        {},
	"/databases/postgresql/remote": {},
	"/databases/redis":             {},
	"/databases/redis/remote":      {},

	"/hosts":                  {},
	"/hosts/files":            {},
	"/hosts/monitor/monitor":  {},
	"/hosts/monitor/setting":  {},
	"/hosts/firewall/port":    {},
	"/hosts/firewall/forward": {},
	"/hosts/firewall/ip":      {},
	"/hosts/firewall/advance": {},
	"/hosts/process/process":  {},
	"/hosts/process/network":  {},
	"/hosts/ssh/ssh":          {},
	"/hosts/ssh/log":          {},
	"/hosts/ssh/session":      {},
	"/hosts/disk":             {},

	"/terminal": {},

	"/logs":           {},
	"/logs/operation": {},
	"/logs/login":     {},
	"/logs/website":   {},
	"/logs/system":    {},
	"/logs/ssh":       {},
	"/logs/task":      {},

	"/settings":               {},
	"/settings/panel":         {},
	"/settings/backupaccount": {},
	"/settings/license":       {},
	"/settings/about":         {},
	"/settings/safe":          {},
	"/settings/alert":         {},
	"/settings/snapshot":      {},
	"/settings/expired":       {},

	"/toolbox":              {},
	"/toolbox/device":       {},
	"/toolbox/supervisor":   {},
	"/toolbox/clam":         {},
	"/toolbox/clam/setting": {},
	"/toolbox/ftp":          {},
	"/toolbox/fail2ban":     {},
	"/toolbox/clean":        {},

	"/websites":                 {},
	"/websites/ssl":             {},
	"/websites/runtimes/php":    {},
	"/websites/runtimes/node":   {},
	"/websites/runtimes/java":   {},
	"/websites/runtimes/go":     {},
	"/websites/runtimes/python": {},
	"/websites/runtimes/dotnet": {},

	"/login": {},

	"/xpack":                {},
	"/xpack/waf/dashboard":  {},
	"/xpack/waf/global":     {},
	"/xpack/waf/websites":   {},
	"/xpack/waf/log":        {},
	"/xpack/waf/block":      {},
	"/xpack/waf/blackwhite": {},
	"/xpack/waf/stat":       {},

	"/xpack/monitor/dashboard": {},
	"/xpack/monitor/setting":   {},
	"/xpack/monitor/rank":      {},
	"/xpack/monitor/log":       {},
	"/xpack/monitor/trend":     {},
	"/xpack/monitor/websites":  {},

	"/xpack/tamper":          {},
	"/xpack/gpu":             {},
	"/xpack/alert/dashboard": {},
	"/xpack/alert/log":       {},
	"/xpack/alert/setting":   {},
	"/xpack/setting":         {},
	"/xpack/node/dashboard":  {},
	"/xpack/node":            {},
	"/xpack/simple-node":     {},
	"/xpack/sync/file":       {},
	"/xpack/sync/image":      {},
	"/xpack/sync/ssl":        {},
	"/xpack/sync/app":        {},
	"/xpack/app":             {},
	"/xpack/app-upgrade":     {},

	"/xpack/cluster/mysql":    {},
	"/xpack/cluster/postgres": {},
	"/xpack/cluster/redis":    {},

	"/enterprise/users/list":          {},
	"/enterprise/users/roles":         {},
	"/enterprise/license":             {},
	"/enterprise/license-required":    {},
	"/enterprise/ops-report":          {},
	"/enterprise/ops-report/overview": {},
	"/enterprise/ops-report/system":   {},
	"/enterprise/ops-report/login":    {},
	"/enterprise/ops-report/website":  {},
	"/enterprise/ops-report/resource": {},
	"/enterprise/ops-report/cronjob":  {},
	"/enterprise/ops-report/alert":    {},
	"/enterprise/ops-report/history":  {},
	"/enterprise/ops-report/settings": {},
	"/enterprise/vm/list":             {},
	"/enterprise/vm/iso":              {},
	"/enterprise/vm/templates":        {},
	"/enterprise/vm/networks":         {},
	"/enterprise/vm/storage-pools":    {},
}

var DynamicRoutes = []string{
	`^/containers/composeDetail/[^/]+$`,
	`^/databases/mysql/setting/[^/]+/[^/]+$`,
	`^/databases/postgresql/setting/[^/]+/[^/]+$`,
	`^/websites/[^/]+/config/[^/]+$`,
	`^/s/[A-Za-z0-9]{10,16}$`,
}

var CertStore atomic.Value

var DaemonJsonPath = "/etc/docker/daemon.json"

const (
	RoleMaster = "master"
	RoleSlave  = "slave"
)
