package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"gorm.io/gorm"
)

type IWebsiteTemplateRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteTemplate, error)
	GetFirst(opts ...DBOption) (*model.WebsiteTemplate, error)
	List(opts ...DBOption) ([]model.WebsiteTemplate, error)
	Create(template *model.WebsiteTemplate) error
	Save(template *model.WebsiteTemplate) error
	DeleteBy(opts ...DBOption) error
	WithName(name string) DBOption
	WithType(templateType string) DBOption
}

func NewIWebsiteTemplateRepo() IWebsiteTemplateRepo {
	return &WebsiteTemplateRepo{}
}

type WebsiteTemplateRepo struct {
}

func (w *WebsiteTemplateRepo) WithName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name like ?", "%"+name+"%")
	}
}

func (w *WebsiteTemplateRepo) WithType(templateType string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", templateType)
	}
}

func (w *WebsiteTemplateRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteTemplate, error) {
	var templates []model.WebsiteTemplate
	db := getDb(opts...).Model(&model.WebsiteTemplate{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&templates).Error
	return count, templates, err
}

func (w *WebsiteTemplateRepo) GetFirst(opts ...DBOption) (*model.WebsiteTemplate, error) {
	var template model.WebsiteTemplate
	db := getDb(opts...).Model(&model.WebsiteTemplate{})
	if err := db.First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (w *WebsiteTemplateRepo) List(opts ...DBOption) ([]model.WebsiteTemplate, error) {
	var templates []model.WebsiteTemplate
	err := getDb(opts...).Model(&model.WebsiteTemplate{}).Find(&templates).Error
	return templates, err
}

func (w *WebsiteTemplateRepo) Create(template *model.WebsiteTemplate) error {
	return getDb().Create(template).Error
}

func (w *WebsiteTemplateRepo) Save(template *model.WebsiteTemplate) error {
	return getDb().Save(template).Error
}

func (w *WebsiteTemplateRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebsiteTemplate{}).Error
}

type IWebsiteTemplateOutputRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteTemplateOutput, error)
	GetFirst(opts ...DBOption) (*model.WebsiteTemplateOutput, error)
	List(opts ...DBOption) ([]model.WebsiteTemplateOutput, error)
	Create(output *model.WebsiteTemplateOutput) error
	Save(output *model.WebsiteTemplateOutput) error
	DeleteBy(opts ...DBOption) error
	WithByTemplateID(templateID uint) DBOption
}

func NewIWebsiteTemplateOutputRepo() IWebsiteTemplateOutputRepo {
	return &WebsiteTemplateOutputRepo{}
}

type WebsiteTemplateOutputRepo struct {
}

func (w *WebsiteTemplateOutputRepo) WithByTemplateID(templateID uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("template_id = ?", templateID)
	}
}

func (w *WebsiteTemplateOutputRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteTemplateOutput, error) {
	var outputs []model.WebsiteTemplateOutput
	db := getDb(opts...).Model(&model.WebsiteTemplateOutput{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&outputs).Error
	return count, outputs, err
}

func (w *WebsiteTemplateOutputRepo) GetFirst(opts ...DBOption) (*model.WebsiteTemplateOutput, error) {
	var output model.WebsiteTemplateOutput
	db := getDb(opts...).Model(&model.WebsiteTemplateOutput{})
	if err := db.First(&output).Error; err != nil {
		return nil, err
	}
	return &output, nil
}

func (w *WebsiteTemplateOutputRepo) List(opts ...DBOption) ([]model.WebsiteTemplateOutput, error) {
	var outputs []model.WebsiteTemplateOutput
	err := getDb(opts...).Model(&model.WebsiteTemplateOutput{}).Find(&outputs).Error
	return outputs, err
}

func (w *WebsiteTemplateOutputRepo) Create(output *model.WebsiteTemplateOutput) error {
	return getDb().Create(output).Error
}

func (w *WebsiteTemplateOutputRepo) Save(output *model.WebsiteTemplateOutput) error {
	return getDb().Save(output).Error
}

func (w *WebsiteTemplateOutputRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebsiteTemplateOutput{}).Error
}
