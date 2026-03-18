package utils

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	openclawutil "github.com/1Panel-dev/1Panel/agent/utils/openclaw"

	"gorm.io/gorm"
)

const (
	openclawVersionWithBundledCaddyMigration = "2026.3.13"
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
		if err := openclawutil.WriteCatchAllCaddyfile(install.GetPath()); err != nil {
			return err
		}
	}
	return nil
}

func RewriteOpenclawBundledCaddyfile(tx *gorm.DB) error {
	return RewriteOpenclawCaddyfileForVersion(tx, openclawVersionWithBundledCaddyMigration)
}
