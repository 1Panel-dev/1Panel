package openclaw

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

const (
	gatewayPort   = 18789
	caddyPort     = 8443
	caddyDataPerm = 0o777
)

func BuildCatchAllCaddyfile() string {
	return fmt.Sprintf(`{
    admin off
    auto_https disable_redirects
    skip_install_trust
    storage file_system {
        root /data/caddy
    }
}

https://:%d {
    bind 0.0.0.0
    tls internal {
        on_demand
    }
    reverse_proxy 127.0.0.1:%d
}
`, caddyPort, gatewayPort)
}

func WriteCatchAllCaddyfile(installPath string) error {
	caddyDir := path.Join(installPath, "data", "caddy")
	caddyDataDir := path.Join(caddyDir, "data")
	if err := os.MkdirAll(caddyDataDir, constant.DirPerm); err != nil {
		return err
	}
	if err := os.Chmod(caddyDataDir, caddyDataPerm); err != nil {
		return err
	}
	return os.WriteFile(path.Join(caddyDir, "Caddyfile"), []byte(BuildCatchAllCaddyfile()), constant.FilePerm)
}
