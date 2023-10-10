package constant

const (
	ResourceLocal    = "local"
	ResourceAppstore = "appstore"

	RuntimeNormal     = "normal"
	RuntimeError      = "error"
	RuntimeBuildIng   = "building"
	RuntimeStarting   = "starting"
	RuntimeRunning    = "running"
	RuntimeReCreating = "recreating"
	RuntimeStopped    = "stopped"
	RuntimeUnhealthy  = "unhealthy"
	RuntimeCreating   = "creating"

	RuntimePHP  = "php"
	RuntimeNode = "node"

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
