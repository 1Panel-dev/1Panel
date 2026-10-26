package service

import (
	"errors"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

func prepareMCPWebsiteDirectory(req request.AppInstallCreate) error {
	changed := errors.New("MCP website directory changed; inspect the installation before retrying")
	directory, ok := req.Params["WEBSITE_DIR"].(string)
	if !ok || directory == "" || global.DB == nil {
		return changed
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var setting model.Setting
		err := tx.Where("key = ?", "WEBSITE_DIR").First(&setting).Error
		missing := errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !missing {
			return err
		}
		current := setting.Value
		if current == "" {
			current = path.Join(global.Dir.BaseDir, "1panel", "www")
		}
		if directory != current {
			return changed
		}
		if missing {
			if err := tx.Create(&model.Setting{Key: "WEBSITE_DIR", Value: directory}).Error; err != nil {
				return err
			}
		} else if setting.Value == "" {
			result := tx.Model(&model.Setting{}).Where("id = ? AND value = ?", setting.ID, "").Update("value", directory)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return changed
			}
		}
		if err := os.MkdirAll(path.Join(directory, "conf.d"), constant.DirPerm); err != nil {
			return err
		}
		return nil
	})
}
