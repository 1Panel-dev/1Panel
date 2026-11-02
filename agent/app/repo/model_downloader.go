package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type ModelDownloadTaskRepo struct{}

type IModelDownloadTaskRepo interface {
	Page(page, size int, info, status string, opts ...DBOption) (int64, []model.ModelDownloadTask, error)
	GetFirst(opts ...DBOption) (model.ModelDownloadTask, error)
	List(opts ...DBOption) ([]model.ModelDownloadTask, error)
	Create(task *model.ModelDownloadTask) error
	Update(id uint, values map[string]interface{}) error
	UpdateBy(values map[string]interface{}, opts ...DBOption) error
	DeleteBy(opts ...DBOption) error
	WithByDownloadType(downloadType string) DBOption
	WithBySource(source string) DBOption
	WithByRepoID(repoID string) DBOption
	WithByTargetDir(targetDir string) DBOption
	WithByStatusIn(status []string) DBOption
}

func NewIModelDownloadTaskRepo() IModelDownloadTaskRepo {
	return &ModelDownloadTaskRepo{}
}

func (m *ModelDownloadTaskRepo) WithBySource(source string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("source = ?", source)
	}
}

func (m *ModelDownloadTaskRepo) WithByRepoID(repoID string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("repo_id = ?", repoID)
	}
}

func (m *ModelDownloadTaskRepo) WithByTargetDir(targetDir string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("target_dir = ?", targetDir)
	}
}

func (m *ModelDownloadTaskRepo) WithByStatusIn(status []string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status in ?", status)
	}
}

func (m *ModelDownloadTaskRepo) Page(page, size int, info, status string, opts ...DBOption) (int64, []model.ModelDownloadTask, error) {
	var (
		total int64
		list  []model.ModelDownloadTask
	)
	db := global.DB.Model(&model.ModelDownloadTask{})
	for _, opt := range opts {
		db = opt(db)
	}
	if info != "" {
		db = db.Where("repo_id LIKE ? OR model_name LIKE ?", "%"+info+"%", "%"+info+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if err := db.Order("created_at desc").Limit(size).Offset(size * (page - 1)).Find(&list).Error; err != nil {
		return 0, nil, err
	}
	return total, list, nil
}

func (m *ModelDownloadTaskRepo) GetFirst(opts ...DBOption) (model.ModelDownloadTask, error) {
	var item model.ModelDownloadTask
	db := global.DB.Model(&model.ModelDownloadTask{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&item).Error
	return item, err
}

func (m *ModelDownloadTaskRepo) Create(task *model.ModelDownloadTask) error {
	return global.DB.Create(task).Error
}

func (m *ModelDownloadTaskRepo) List(opts ...DBOption) ([]model.ModelDownloadTask, error) {
	var items []model.ModelDownloadTask
	db := global.DB.Model(&model.ModelDownloadTask{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&items).Error
	return items, err
}

func (m *ModelDownloadTaskRepo) Update(id uint, values map[string]interface{}) error {
	return global.DB.Model(&model.ModelDownloadTask{}).Where("id = ?", id).Updates(values).Error
}

func (m *ModelDownloadTaskRepo) UpdateBy(values map[string]interface{}, opts ...DBOption) error {
	db := global.DB.Model(&model.ModelDownloadTask{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(values).Error
}

func (m *ModelDownloadTaskRepo) DeleteBy(opts ...DBOption) error {
	db := global.DB.Model(&model.ModelDownloadTask{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ModelDownloadTask{}).Error
}

func (m *ModelDownloadTaskRepo) WithByDownloadType(downloadType string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("download_type = ?", downloadType)
	}
}
