package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type ScriptRepo struct{}

type IScriptRepo interface {
	Get(opts ...DBOption) (model.ScriptLibrary, error)
	List(opts ...DBOption) ([]model.ScriptLibrary, error)

	SyncAll(data []model.ScriptLibrary) error
}

func NewIScriptRepo() IScriptRepo {
	return &ScriptRepo{}
}

func (u *ScriptRepo) Get(opts ...DBOption) (model.ScriptLibrary, error) {
	var script model.ScriptLibrary
	db := global.DB
	if global.IsMaster {
		db = global.CoreDB
	}
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&script).Error
	return script, err
}

func (u *ScriptRepo) List(opts ...DBOption) ([]model.ScriptLibrary, error) {
	var ops []model.ScriptLibrary
	if global.IsMaster {
		db := global.CoreDB.Model(&model.ScriptLibrary{})
		for _, opt := range opts {
			db = opt(db)
		}
		err := db.Where("is_interactive = ?", false).Find(&ops).Error
		return ops, err
	}
	db := global.DB.Model(&model.ScriptLibrary{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&ops).Error
	return ops, err
}

func (u *ScriptRepo) SyncAll(data []model.ScriptLibrary) error {
	tx := global.DB.Begin()
	var oldScripts []model.ScriptLibrary
	_ = tx.Where("1 = 1").Find(&oldScripts).Error
	oldScriptMap := make(map[string]uint)
	for _, item := range oldScripts {
		oldScriptMap[item.Name] = item.ID
	}
	for _, item := range data {
		if val, ok := oldScriptMap[item.Name]; ok {
			item.ID = val
			delete(oldScriptMap, item.Name)
		} else {
			item.ID = 0
		}
		if err := tx.Model(model.ScriptLibrary{}).Where("id = ?", item.ID).Save(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, val := range oldScriptMap {
		if err := tx.Where("id = ?", val).Delete(&model.ScriptLibrary{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
