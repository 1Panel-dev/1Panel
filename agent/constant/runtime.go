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
