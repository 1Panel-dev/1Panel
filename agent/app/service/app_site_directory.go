package service

import (
	"context"
	"errors"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

func ensureCurrentWebsiteDirectory(directory string) error {
	changed := errors.New("website directory changed; inspect the installation before retrying")
	if directory == "" || global.DB == nil {
		return changed
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		ctx := context.WithValue(tx.Statement.Context, constant.DB, tx)
		setting, err := settingRepo.GetWithContext(ctx, settingRepo.WithByKey("WEBSITE_DIR"))
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
			if err := settingRepo.CreateWithContext(ctx, &model.Setting{Key: "WEBSITE_DIR", Value: directory}); err != nil {
				return err
			}
		} else if setting.Value == "" {
			updated, err := settingRepo.UpdateIfMatch(ctx, setting.ID, "", directory)
			if err != nil {
				return err
			}
			if updated != 1 {
				return changed
			}
		}
		return os.MkdirAll(path.Join(directory, "conf.d"), constant.DirPerm)
	})
}
