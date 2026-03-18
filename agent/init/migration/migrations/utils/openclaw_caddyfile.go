package utils

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"

	"gorm.io/gorm"
)

const (
	openclawVersionWithBundledCaddyMigration = "2026.3.13"
	openclawGatewayPort                      = 18789
	openclawCaddyPort                        = 8443
	openclawCaddyDataPerm                    = 0o777
)

func RewriteOpenclawCaddyfileForVersion(tx *gorm.DB, version string) error {
	targetVersion := strings.TrimSpace(version)
	if targetVersion == "" {
		return nil
	}
	var installs []model.AppInstall
	if err := tx.Preload("App").Find(&installs).Error; err != nil {
		return err
	}
	for _, install := range installs {
		if install.App.Key != constant.AppOpenclaw {
			continue
		}
		if strings.TrimSpace(install.Version) != targetVersion {
			continue
		}
		if err := writeOpenclawCatchAllCaddyfile(install.GetPath()); err != nil {
			return err
		}
	}
	return nil
}

func RewriteOpenclawBundledCaddyfile(tx *gorm.DB) error {
	return RewriteOpenclawCaddyfileForVersion(tx, openclawVersionWithBundledCaddyMigration)
}

func writeOpenclawCatchAllCaddyfile(installPath string) error {
	caddyDir := path.Join(installPath, "data", "caddy")
	caddyDataDir := path.Join(caddyDir, "data")
	if err := os.MkdirAll(caddyDataDir, constant.DirPerm); err != nil {
		return err
	}
	if err := os.Chmod(caddyDataDir, openclawCaddyDataPerm); err != nil {
		return err
	}
	content := fmt.Sprintf(`{
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
`, openclawCaddyPort, openclawGatewayPort)
	return os.WriteFile(path.Join(caddyDir, "Caddyfile"), []byte(content), constant.FilePerm)
}
