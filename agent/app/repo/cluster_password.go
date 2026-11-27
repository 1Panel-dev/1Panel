package repo

import (
	"context"
	"encoding/json"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SaveClusterDatabaseCredentials(ctx context.Context, install model.AppInstall, loginKey string, changes map[string]string) error {
	fields := make(map[string]any, 2)
	for name, raw := range map[string]string{"env": install.Env, "param": install.Param} {
		values := map[string]json.RawMessage{}
		if raw != "" {
			if err := json.Unmarshal([]byte(raw), &values); err != nil {
				return err
			}
		}
		if values == nil {
			return gorm.ErrInvalidData
		}
		for key, value := range changes {
			encoded, err := json.Marshal(value)
			if err != nil {
				return err
			}
			values[key] = encoded
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return err
		}
		fields[name] = string(encoded)
	}
	password, updateLogin := changes[loginKey]
	var ciphertext string
	if updateLogin {
		var err error
		ciphertext, err = encrypt.StringEncrypt(password)
		if err != nil {
			return err
		}
	}
	return getTx(ctx).WithContext(ctx).Session(&gorm.Session{Logger: logger.Discard}).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.AppInstall{}).
			Where("id = ? AND env = ? AND COALESCE(param, '') = ?", install.ID, install.Env, install.Param).Updates(fields)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if updateLogin {
			return tx.Model(&model.Database{}).
				Where("app_install_id = ? AND type = ? AND \"from\" = ?", install.ID, install.App.Key, "local").
				Update("password", ciphertext).Error
		}
		return nil
	})
}
