package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type FavoriteRepo struct{}

type IFavoriteRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.Favorite, error)
	Create(group *model.Favorite) error
	Delete(opts ...DBOption) error
	GetFirst(opts ...DBOption) (model.Favorite, error)
	All() ([]model.Favorite, error)
	ListByPaths(ctx context.Context, paths []string) ([]model.Favorite, error)
	WithByPath(path string) DBOption
}

func NewIFavoriteRepo() IFavoriteRepo {
	return &FavoriteRepo{}
}

func (f *FavoriteRepo) WithByPath(path string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("path = ?", path)
	}
}

func (f *FavoriteRepo) Page(page, size int, opts ...DBOption) (int64, []model.Favorite, error) {
	var (
		favorites []model.Favorite
		count     int64
	)
	count = int64(0)
	db := getDb(opts...).Model(&model.Favorite{})
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&favorites).Error
	return count, favorites, err
}

func (f *FavoriteRepo) Create(favorite *model.Favorite) error {
	return global.DB.Create(favorite).Error
}

func (f *FavoriteRepo) GetFirst(opts ...DBOption) (model.Favorite, error) {
	var favorite model.Favorite
	db := getDb(opts...).Model(&model.Favorite{})
	if err := db.First(&favorite).Error; err != nil {
		return favorite, err
	}
	return favorite, nil
}

func (f *FavoriteRepo) Delete(opts ...DBOption) error {
	db := getDb(opts...).Model(&model.Favorite{})
	return db.Delete(&model.Favorite{}).Error
}

func (f *FavoriteRepo) All() ([]model.Favorite, error) {
	var favorites []model.Favorite
	if err := getDb().Find(&favorites).Error; err != nil {
		return nil, err
	}
	return favorites, nil
}

const filePathQueryBatchSize = 900

func uniqueFilePaths(paths []string) []string {
	unique := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		unique = append(unique, path)
	}
	return unique
}

func (f *FavoriteRepo) ListByPaths(ctx context.Context, paths []string) ([]model.Favorite, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	paths = uniqueFilePaths(paths)
	favorites := make([]model.Favorite, 0, len(paths))
	for start := 0; start < len(paths); start += filePathQueryBatchSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		end := min(start+filePathQueryBatchSize, len(paths))
		db := getTx(ctx).WithContext(ctx)
		firstIDs := db.Model(&model.Favorite{}).Select("MIN(id)").Where("path IN ?", paths[start:end]).Group("path")
		var batch []model.Favorite
		if err := db.Where("id IN (?)", firstIDs).Order("id ASC").Find(&batch).Error; err != nil {
			return nil, err
		}
		favorites = append(favorites, batch...)
	}
	return favorites, nil
}
