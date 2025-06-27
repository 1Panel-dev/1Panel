package constant

const (
	ResourceLocal    = "local"
	ResourceAppstore = "appstore"

	RuntimePHP    = "php"
	RuntimeNode   = "node"
	RuntimeJava   = "java"
	RuntimeGo     = "go"
	RuntimePython = "python"
	RuntimeDotNet = "dotnet"

	RuntimeProxyUnix = "unix"
	RuntimeProxyTcp  = "tcp"

	RuntimeUp      = "up"
	RuntimeDown    = "down"
	RuntimeRestart = "restart"

	RuntimeInstall   = "install"
	RuntimeUninstall = "uninstall"
	RuntimeUpdate    = "update"

	RuntimeNpm  = "npm"
	RuntimeYarn = "yarn"
)

var GoDefaultVolumes = map[string]string{
	"${CODE_DIR}": "/app",
	"./run.sh":    "/run.sh",
	"./.env":      "/.env",
	"./mod":       "/go/pkg/mod",
}

var RuntimeDefaultVolumes = map[string]string{
	"${CODE_DIR}": "/app",
	"./run.sh":    "/run.sh",
	"./.env":      "/.env",
}

var PHPDefaultVolumes = map[string]string{
	"${PANEL_WEBSITE_DIR}":                  "/www/",
	"./conf/php.ini":                        "/usr/local/etc/php/php.ini",
	"./conf/conf.d":                         "/usr/local/etc/php/conf.d",
	"./conf/php-fpm.conf":                   "/usr/local/etc/php-fpm.conf",
	"./log":                                 "/var/log/php",
	"./extensions":                          "/usr/local/lib/php/extensions",
	"./supervisor/supervisord.conf":         "/etc/supervisord.conf",
	"./supervisor/supervisor.d/php-fpm.ini": "/etc/supervisor.d/php-fpm.ini",
	"./supervisor/supervisor.d":             "/etc/supervisor.d",
	"./supervisor/log":                      "/var/log/supervisor",
	"./composer":                            "/tmp/composer",
}
