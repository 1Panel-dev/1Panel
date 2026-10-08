package repo

import (
	"context"
	"errors"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingRepo struct{}

type ISettingRepo interface {
	GetList(opts ...DBOption) ([]model.Setting, error)
	Get(opts ...DBOption) (model.Setting, error)
	GetValueByKey(key string) (string, error)
	Create(key, value string) error
	Update(key, value string) error
	WithByKey(key string) DBOption

	UpdateOrCreate(key, value string) error
	UpdateValues(map[string]string) error

	GetDescription(opts ...DBOption) (model.CommonDescription, error)
	GetDescriptionList(opts ...DBOption) ([]model.CommonDescription, error)
	CreateDescription(data *model.CommonDescription) error
	SaveDescriptions(context.Context, []model.CommonDescription) error
	UpdateDescription(id string, val map[string]interface{}) error
	DelDescription(id string) error
	DeleteDescriptions(context.Context, string, []string, bool) (int64, error)
	WithDescriptionIDs(ids []string) DBOption
	WithByDescriptionID(id string) DBOption
}

func NewISettingRepo() ISettingRepo {
	return &SettingRepo{}
}

func (s *SettingRepo) GetList(opts ...DBOption) ([]model.Setting, error) {
	var settings []model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&settings).Error
	return settings, err
}

func (s *SettingRepo) Create(key, value string) error {
	setting := &model.Setting{
		Key:   key,
		Value: value,
	}
	return global.DB.Create(setting).Error
}

func (s *SettingRepo) Get(opts ...DBOption) (model.Setting, error) {
	var settings model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&settings).Error
	return settings, err
}

func (s *SettingRepo) GetValueByKey(key string) (string, error) {
	var setting model.Setting
	if err := global.DB.Model(&model.Setting{}).Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingRepo) WithByKey(key string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}

func (s *SettingRepo) Update(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}

func (s *SettingRepo) UpdateOrCreate(key, value string) error {
	var setting model.Setting
	result := global.DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return global.DB.Create(&model.Setting{Key: key, Value: value}).Error
		}
		return result.Error
	}
	return global.DB.Model(&setting).UpdateColumn("value", value).Error
}

func (s *SettingRepo) UpdateValues(values map[string]string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		for key, value := range values {
			var setting model.Setting
			err := tx.Where("key = ?", key).First(&setting).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&model.Setting{Key: key, Value: value}).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if err := tx.Model(&setting).UpdateColumn("value", value).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SettingRepo) GetDescriptionList(opts ...DBOption) ([]model.CommonDescription, error) {
	var lists []model.CommonDescription
	db := global.DB.Model(&model.CommonDescription{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&lists).Error
	return lists, err
}
func (s *SettingRepo) GetDescription(opts ...DBOption) (model.CommonDescription, error) {
	var data model.CommonDescription
	db := global.DB.Model(&model.CommonDescription{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&data).Error
	return data, err
}
func (s *SettingRepo) CreateDescription(data *model.CommonDescription) error {
	return global.DB.Create(data).Error
}

func (s *SettingRepo) SaveDescriptions(ctx context.Context, descriptions []model.CommonDescription) error {
	return global.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"description"}),
	}).CreateInBatches(&descriptions, 100).Error
}

func (s *SettingRepo) UpdateDescription(id string, val map[string]interface{}) error {
	return global.DB.Model(&model.CommonDescription{}).Where("id = ?", id).Updates(val).Error
}
func (s *SettingRepo) DelDescription(id string) error {
	return global.DB.Where("id = ?", id).Delete(&model.CommonDescription{}).Error
}

func (s *SettingRepo) DeleteDescriptions(ctx context.Context, kind string, ids []string, emptyOnly bool) (int64, error) {
	var deleted int64
	for start := 0; start < len(ids); start += 500 {
		query := global.DB.WithContext(ctx).Where("type = ? AND id IN ?", kind, ids[start:min(start+500, len(ids))])
		if emptyOnly {
			query = query.Where("description = ? AND is_pinned = ?", "", false)
		}
		result := query.Delete(&model.CommonDescription{})
		deleted += result.RowsAffected
		if result.Error != nil {
			return deleted, result.Error
		}
	}
	return deleted, nil
}
func (s *SettingRepo) WithByDescriptionID(id string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func (s *SettingRepo) WithDescriptionIDs(ids []string) DBOption {
	return func(db *gorm.DB) *gorm.DB { return db.Where("id IN ?", ids) }
}
