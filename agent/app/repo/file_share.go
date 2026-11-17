package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type FileShareRepo struct{}

type IFileShareRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.FileShare, error)
	Create(fileShare *model.FileShare) error
	Save(fileShare *model.FileShare) error
	Delete(opts ...DBOption) error
	GetFirst(opts ...DBOption) (model.FileShare, error)
	All() ([]model.FileShare, error)
	ListByPaths(ctx context.Context, paths []string) ([]model.FileShare, error)
	WithByPath(path string) DBOption
	WithByCode(code string) DBOption
}

func NewIFileShareRepo() IFileShareRepo {
	return &FileShareRepo{}
}

func (r *FileShareRepo) WithByPath(path string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path = ?", path)
	}
}

func (r *FileShareRepo) WithByCode(code string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("token = ?", code)
	}
}

func (r *FileShareRepo) Page(page, size int, opts ...DBOption) (int64, []model.FileShare, error) {
	var (
		items []model.FileShare
		count int64
	)
	db := getDb(opts...).Model(&model.FileShare{})
	db = db.Count(&count)
	err := db.Order("path asc").Limit(size).Offset(size * (page - 1)).Find(&items).Error
	return count, items, err
}

func (r *FileShareRepo) Create(fileShare *model.FileShare) error {
	return global.DB.Create(fileShare).Error
}

func (r *FileShareRepo) Save(fileShare *model.FileShare) error {
	return global.DB.Save(fileShare).Error
}

func (r *FileShareRepo) GetFirst(opts ...DBOption) (model.FileShare, error) {
	var item model.FileShare
	db := getDb(opts...).Model(&model.FileShare{})
	if err := db.First(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

func (r *FileShareRepo) Delete(opts ...DBOption) error {
	db := getDb(opts...).Model(&model.FileShare{})
	return db.Delete(&model.FileShare{}).Error
}

func (r *FileShareRepo) All() ([]model.FileShare, error) {
	var items []model.FileShare
	if err := getDb().Order("path asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *FileShareRepo) ListByPaths(ctx context.Context, paths []string) ([]model.FileShare, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	paths = uniqueFilePaths(paths)
	items := make([]model.FileShare, 0, len(paths))
	for start := 0; start < len(paths); start += filePathQueryBatchSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		end := min(start+filePathQueryBatchSize, len(paths))
		var batch []model.FileShare
		if err := getTx(ctx).WithContext(ctx).Where("path IN ?", paths[start:end]).Order("path ASC").Find(&batch).Error; err != nil {
			return nil, err
		}
		items = append(items, batch...)
	}
	return items, nil
}
