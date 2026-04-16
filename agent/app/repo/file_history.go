package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type FileHistoryRepo struct{}

type IFileHistoryRepo interface {
	Get(opts ...DBOption) (model.FileHistory, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.FileHistory, error)
	Create(history *model.FileHistory) error
	Delete(opts ...DBOption) error
	DeleteByIDs(ids []uint) error
	WithByID(id uint) DBOption
	WithByFileID(fileID string) DBOption
	WithByPath(path string) DBOption
	WithByPathHash(pathHash string) DBOption
	WithByPathLike(path string) DBOption
	WithByOperation(operation string) DBOption
	WithNotOperation(operation string) DBOption
	WithByRelatedPath(path string) DBOption
}

func NewIFileHistoryRepo() IFileHistoryRepo {
	return &FileHistoryRepo{}
}

func (r *FileHistoryRepo) Get(opts ...DBOption) (model.FileHistory, error) {
	var item model.FileHistory
	db := global.DB.Model(&model.FileHistory{})
	for _, opt := range opts {
		db = opt(db)
	}
	return item, db.First(&item).Error
}

func (r *FileHistoryRepo) Page(limit, offset int, opts ...DBOption) (int64, []model.FileHistory, error) {
	var total int64
	var items []model.FileHistory
	db := global.DB.Model(&model.FileHistory{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if err := db.Order("created_at desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return 0, nil, err
	}
	return total, items, nil
}

func (r *FileHistoryRepo) Create(history *model.FileHistory) error {
	return global.DB.Create(history).Error
}

func (r *FileHistoryRepo) Delete(opts ...DBOption) error {
	db := global.DB.Model(&model.FileHistory{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.FileHistory{}).Error
}

func (r *FileHistoryRepo) DeleteByIDs(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return global.DB.Where("id in ?", ids).Delete(&model.FileHistory{}).Error
}

func (r *FileHistoryRepo) WithByID(id uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

func (r *FileHistoryRepo) WithByFileID(fileID string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("file_id = ?", fileID)
	}
}

func (r *FileHistoryRepo) WithByPath(path string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path = ?", path)
	}
}

func (r *FileHistoryRepo) WithByPathHash(pathHash string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path_hash = ?", pathHash)
	}
}

func (r *FileHistoryRepo) WithByPathLike(path string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path like ?", "%"+path+"%")
	}
}

func (r *FileHistoryRepo) WithByOperation(operation string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("operation = ?", operation)
	}
}

func (r *FileHistoryRepo) WithNotOperation(operation string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("operation <> ?", operation)
	}
}

func (r *FileHistoryRepo) WithByRelatedPath(path string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path = ? OR source_path = ? OR target_path = ?", path, path, path)
	}
}
