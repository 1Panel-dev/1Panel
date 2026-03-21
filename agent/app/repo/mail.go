package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

// MailDomain repo

type MailDomainRepo struct{}

type IMailDomainRepo interface {
	List(opts ...DBOption) ([]model.MailDomain, error)
	GetFirst(opts ...DBOption) (model.MailDomain, error)
	Create(ctx context.Context, domain *model.MailDomain) error
	Save(ctx context.Context, domain *model.MailDomain) error
	Update(id uint, vars map[string]interface{}) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	WithName(name string) DBOption
}

func NewIMailDomainRepo() IMailDomainRepo {
	return &MailDomainRepo{}
}

func (d *MailDomainRepo) List(opts ...DBOption) ([]model.MailDomain, error) {
	var domains []model.MailDomain
	db := global.DB.Model(&model.MailDomain{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&domains).Error; err != nil {
		return domains, err
	}
	return domains, nil
}

func (d *MailDomainRepo) GetFirst(opts ...DBOption) (model.MailDomain, error) {
	var domain model.MailDomain
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&domain).Error; err != nil {
		return domain, err
	}
	return domain, nil
}

func (d *MailDomainRepo) Create(ctx context.Context, domain *model.MailDomain) error {
	return getTx(ctx).Create(domain).Error
}

func (d *MailDomainRepo) Save(ctx context.Context, domain *model.MailDomain) error {
	return getTx(ctx).Save(domain).Error
}

func (d *MailDomainRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.MailDomain{}).Where("id = ?", id).Updates(vars).Error
}

func (d *MailDomainRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.MailDomain{}).Error
}

func (d *MailDomainRepo) WithName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}

// MailAccount repo

type MailAccountRepo struct{}

type IMailAccountRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.MailAccount, error)
	List(opts ...DBOption) ([]model.MailAccount, error)
	GetFirst(opts ...DBOption) (model.MailAccount, error)
	Create(ctx context.Context, account *model.MailAccount) error
	Save(ctx context.Context, account *model.MailAccount) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	WithDomainID(domainID uint) DBOption
	WithEmail(email string) DBOption
	WithLikeInfo(info string) DBOption
}

func NewIMailAccountRepo() IMailAccountRepo {
	return &MailAccountRepo{}
}

func (d *MailAccountRepo) Page(page, size int, opts ...DBOption) (int64, []model.MailAccount, error) {
	var accounts []model.MailAccount
	db := global.DB.Model(&model.MailAccount{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&accounts).Error; err != nil {
		return count, accounts, err
	}
	return count, accounts, nil
}

func (d *MailAccountRepo) List(opts ...DBOption) ([]model.MailAccount, error) {
	var accounts []model.MailAccount
	db := global.DB.Model(&model.MailAccount{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&accounts).Error; err != nil {
		return accounts, err
	}
	return accounts, nil
}

func (d *MailAccountRepo) GetFirst(opts ...DBOption) (model.MailAccount, error) {
	var account model.MailAccount
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&account).Error; err != nil {
		return account, err
	}
	return account, nil
}

func (d *MailAccountRepo) Create(ctx context.Context, account *model.MailAccount) error {
	return getTx(ctx).Create(account).Error
}

func (d *MailAccountRepo) Save(ctx context.Context, account *model.MailAccount) error {
	return getTx(ctx).Save(account).Error
}

func (d *MailAccountRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.MailAccount{}).Error
}

func (d *MailAccountRepo) WithDomainID(domainID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if domainID == 0 {
			return g
		}
		return g.Where("domain_id = ?", domainID)
	}
}

func (d *MailAccountRepo) WithEmail(email string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("email = ?", email)
	}
}

func (d *MailAccountRepo) WithLikeInfo(info string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		return g.Where("email like ? OR username like ?", "%"+info+"%", "%"+info+"%")
	}
}

// MailAlias repo

type MailAliasRepo struct{}

type IMailAliasRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.MailAlias, error)
	List(opts ...DBOption) ([]model.MailAlias, error)
	GetFirst(opts ...DBOption) (model.MailAlias, error)
	Create(ctx context.Context, alias *model.MailAlias) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	WithDomainID(domainID uint) DBOption
	WithSource(source string) DBOption
	WithLikeInfo(info string) DBOption
}

func NewIMailAliasRepo() IMailAliasRepo {
	return &MailAliasRepo{}
}

func (d *MailAliasRepo) Page(page, size int, opts ...DBOption) (int64, []model.MailAlias, error) {
	var aliases []model.MailAlias
	db := global.DB.Model(&model.MailAlias{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&aliases).Error; err != nil {
		return count, aliases, err
	}
	return count, aliases, nil
}

func (d *MailAliasRepo) List(opts ...DBOption) ([]model.MailAlias, error) {
	var aliases []model.MailAlias
	db := global.DB.Model(&model.MailAlias{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&aliases).Error; err != nil {
		return aliases, err
	}
	return aliases, nil
}

func (d *MailAliasRepo) GetFirst(opts ...DBOption) (model.MailAlias, error) {
	var alias model.MailAlias
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&alias).Error; err != nil {
		return alias, err
	}
	return alias, nil
}

func (d *MailAliasRepo) Create(ctx context.Context, alias *model.MailAlias) error {
	return getTx(ctx).Create(alias).Error
}

func (d *MailAliasRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.MailAlias{}).Error
}

func (d *MailAliasRepo) WithDomainID(domainID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if domainID == 0 {
			return g
		}
		return g.Where("domain_id = ?", domainID)
	}
}

func (d *MailAliasRepo) WithSource(source string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("source = ?", source)
	}
}

func (d *MailAliasRepo) WithLikeInfo(info string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		return g.Where("source like ? OR target like ?", "%"+info+"%", "%"+info+"%")
	}
}
