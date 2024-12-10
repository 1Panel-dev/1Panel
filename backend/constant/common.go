package constant

import "sync/atomic"

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

	DateLayout         = "2006-01-02" // or use time.DateOnly while go version >= 1.20
	DefaultDate        = "1970-01-01"
	DateTimeLayout     = "2006-01-02 15:04:05" // or use time.DateTime while go version >= 1.20
	DateTimeSlimLayout = "20060102150405"
)

var WebUrlMap = map[string]struct{}{
	"/apps":           {},
	"/apps/all":       {},
	"/apps/installed": {},
	"/apps/upgrade":   {},

	"/containers":           {},
	"/containers/container": {},
	"/containers/image":     {},
	"/containers/network":   {},
	"/containers/volume":    {},
	"/containers/repo":      {},
	"/containers/compose":   {},
	"/containers/template":  {},
	"/containers/setting":   {},

	"/cronjobs": {},

	"/databases":                   {},
	"/databases/mysql":             {},
	"/databases/mysql/remote":      {},
	"/databases/postgresql":        {},
	"/databases/postgresql/remote": {},
	"/databases/redis":             {},
	"/databases/redis/remote":      {},

	"/hosts":                  {},
	"/hosts/files":            {},
	"/hosts/monitor/monitor":  {},
	"/hosts/monitor/setting":  {},
	"/hosts/terminal":         {},
	"/hosts/firewall/port":    {},
	"/hosts/firewall/forward": {},
	"/hosts/firewall/ip":      {},
	"/hosts/process/process":  {},
	"/hosts/process/network":  {},
	"/hosts/ssh/ssh":          {},
	"/hosts/ssh/log":          {},
	"/hosts/ssh/session":      {},

	"/logs":           {},
	"/logs/operation": {},
	"/logs/login":     {},
	"/logs/website":   {},
	"/logs/system":    {},
	"/logs/ssh":       {},

	"/settings":               {},
	"/settings/panel":         {},
	"/settings/backupaccount": {},
	"/settings/license":       {},
	"/settings/about":         {},
	"/settings/safe":          {},
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

	"/xpack":                   {},
	"/xpack/waf/dashboard":     {},
	"/xpack/waf/global":        {},
	"/xpack/waf/websites":      {},
	"/xpack/waf/log":           {},
	"/xpack/waf/block":         {},
	"/xpack/monitor/dashboard": {},
	"/xpack/monitor/setting":   {},
	"/xpack/monitor/rank":      {},
	"/xpack/monitor/log":       {},
	"/xpack/tamper":            {},
	"/xpack/gpu":               {},
	"/xpack/alert/dashboard":   {},
	"/xpack/alert/log":         {},
	"/xpack/alert/setting":     {},
	"/xpack/setting":           {},
}

var DynamicRoutes = []string{
	`^/containers/composeDetail/[^/]+$`,
	`^/databases/mysql/setting/[^/]+/[^/]+$`,
	`^/databases/postgresql/setting/[^/]+/[^/]+$`,
	`^/websites/[^/]+/config/[^/]+$`,
}

var CertStore atomic.Value
