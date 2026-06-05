package repo

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"gorm.io/gorm"
)

type AgentAccountRepo struct{}

type IAgentAccountRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.AgentAccount, error)
	GetFirst(opts ...DBOption) (*model.AgentAccount, error)
	WithByMasterAccountID(masterID uint) DBOption
	Create(account *model.AgentAccount) error
	Save(account *model.AgentAccount) error
	DeleteByID(id uint) error
	List(opts ...DBOption) ([]model.AgentAccount, error)
	CountByProviders(providers []string) (map[string]int64, error)
}

func NewIAgentAccountRepo() IAgentAccountRepo {
	return &AgentAccountRepo{}
}

func (a AgentAccountRepo) Page(page, size int, opts ...DBOption) (int64, []model.AgentAccount, error) {
	var accounts []model.AgentAccount
	db := getDb(opts...).Model(&model.AgentAccount{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&accounts).Error
	return count, accounts, err
}

func (a AgentAccountRepo) GetFirst(opts ...DBOption) (*model.AgentAccount, error) {
	var account model.AgentAccount
	if err := getDb(opts...).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (a AgentAccountRepo) WithByMasterAccountID(masterID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("master_account_id = ?", masterID)
	}
}

func (a AgentAccountRepo) Create(account *model.AgentAccount) error {
	return getDb().Create(account).Error
}

func (a AgentAccountRepo) Save(account *model.AgentAccount) error {
	return getDb().Save(account).Error
}

func (a AgentAccountRepo) DeleteByID(id uint) error {
	return getDb().Delete(&model.AgentAccount{}, id).Error
}

func (a AgentAccountRepo) List(opts ...DBOption) ([]model.AgentAccount, error) {
	var accounts []model.AgentAccount
	if err := getDb(opts...).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (a AgentAccountRepo) CountByProviders(providers []string) (map[string]int64, error) {
	normalizedProviders := normalizeProviders(providers)
	counts := make(map[string]int64, len(normalizedProviders))
	for _, provider := range normalizedProviders {
		counts[provider] = 0
	}
	if len(normalizedProviders) == 0 {
		return counts, nil
	}

	type providerCount struct {
		Provider string
		Count    int64
	}
	var rows []providerCount
	if err := getDb().
		Model(&model.AgentAccount{}).
		Select("provider, COUNT(*) as count").
		Where("provider IN ?", normalizedProviders).
		Group("provider").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.Provider] = row.Count
	}
	return counts, nil
}

func normalizeProviders(providers []string) []string {
	seen := make(map[string]struct{}, len(providers))
	result := make([]string, 0, len(providers))
	for _, provider := range providers {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		if _, ok := seen[provider]; ok {
			continue
		}
		seen[provider] = struct{}{}
		result = append(result, provider)
	}
	return result
}
