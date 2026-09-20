package apikey_migration

import (
	"errors"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"gorm.io/gorm"
)

const (
	panelAdmin     = "panel_admin"
	enterpriseUser = "enterprise_user"
)

func ToEnterprise(tx *gorm.DB, superAdminID string) error {
	if strings.TrimSpace(superAdminID) == "" {
		return errors.New("missing enterprise super administrator identity")
	}
	return moveOwner(tx, panelAdmin, psession.SuperAdminSessionUserID, enterpriseUser, superAdminID)
}

func ToCommunity(tx *gorm.DB, superAdminID string) error {
	if tx == nil {
		return errors.New("missing Core database")
	}
	if strings.TrimSpace(superAdminID) == "" {
		return errors.New("missing enterprise super administrator identity")
	}
	if tx.Migrator().HasTable(&model.APIKey{}) {
		now := time.Now()
		if err := tx.Model(&model.APIKey{}).
			Where("owner_type = ? AND owner_id <> ? AND status <> ?", enterpriseUser, superAdminID, "Revoked").
			Updates(map[string]interface{}{"status": "Revoked", "secret_ciphertext": "", "active_name": nil, "allow_app_binding": false, "revoked_at": now, "revision": gorm.Expr("revision + 1")}).Error; err != nil {
			return err
		}
	}
	if tx.Migrator().HasTable(&model.LegacyAPIKeyPolicy{}) {
		if err := tx.Where("owner_type = ? AND owner_id <> ?", enterpriseUser, superAdminID).Delete(&model.LegacyAPIKeyPolicy{}).Error; err != nil {
			return err
		}
	}
	return moveOwner(tx, enterpriseUser, superAdminID, panelAdmin, psession.SuperAdminSessionUserID)
}

func moveOwner(tx *gorm.DB, fromType, fromID, toType, toID string) error {
	if tx == nil {
		return errors.New("missing Core database")
	}
	if tx.Migrator().HasTable(&model.APIKey{}) {
		if err := tx.Model(&model.APIKey{}).Where("owner_type = ? AND owner_id = ?", fromType, fromID).
			Updates(map[string]interface{}{"owner_type": toType, "owner_id": toID, "revision": gorm.Expr("revision + 1")}).Error; err != nil {
			return err
		}
	}
	if tx.Migrator().HasTable(&model.LegacyAPIKeyPolicy{}) {
		if err := tx.Model(&model.LegacyAPIKeyPolicy{}).Where("owner_type = ? AND owner_id = ?", fromType, fromID).
			Updates(map[string]interface{}{"owner_type": toType, "owner_id": toID, "revision": gorm.Expr("revision + 1")}).Error; err != nil {
			return err
		}
	}
	return nil
}
