package repo

import (
	"context"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
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
	CheckDeleteLocal(ctx context.Context, installID uint) (string, error)
	DeleteLocal(ctx context.Context, installID uint) error
	List(opts ...DBOption) ([]model.AgentAccount, error)
	CountTextByProviders(providers []string) (map[string]int64, error)
	SaveLocal(account *model.AgentAccount, models []model.AgentAccountModel) error
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

func (a AgentAccountRepo) CountTextByProviders(providers []string) (map[string]int64, error) {
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
		Scopes(WithTextAPIType()).
		Group("provider").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.Provider] = row.Count
	}
	return counts, nil
}

func (a AgentAccountRepo) SaveLocal(account *model.AgentAccount, models []model.AgentAccountModel) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(account).Error; err != nil {
			return err
		}
		if err := tx.Where("account_id = ?", account.ID).Delete(&model.AgentAccountModel{}).Error; err != nil {
			return err
		}
		if len(models) == 0 {
			return nil
		}
		for i := range models {
			models[i].AccountID = account.ID
			models[i].SortOrder = i + 1
		}
		return tx.Create(&models).Error
	})
}

func (a AgentAccountRepo) CheckDeleteLocal(ctx context.Context, installID uint) (string, error) {
	if installID == 0 {
		return "", nil
	}
	var accountIDs []uint
	db := getTx(ctx)
	if err := db.Model(&model.AgentAccount{}).Where("app_install_id = ?", installID).Pluck("id", &accountIDs).Error; err != nil {
		return "", err
	}
	if len(accountIDs) == 0 {
		return "", nil
	}
	var agentCount int64
	if err := db.Model(&model.Agent{}).Where("account_id IN ?", accountIDs).Count(&agentCount).Error; err != nil {
		return "", err
	}
	if agentCount > 0 {
		return "ErrAgentAccountBound", nil
	}
	var settings []model.Setting
	if err := db.Where("key IN ?", []string{"AIStatus", "AIAccountID", "FileAIStatus", "FileAIAccountID"}).Find(&settings).Error; err != nil {
		return "", err
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = strings.TrimSpace(setting.Value)
	}
	for _, accountID := range accountIDs {
		id := strconv.FormatUint(uint64(accountID), 10)
		if strings.EqualFold(values["AIStatus"], constant.StatusEnable) && values["AIAccountID"] == id {
			return "ErrTerminalAIAccountInUse", nil
		}
		if strings.EqualFold(values["FileAIStatus"], constant.StatusEnable) && values["FileAIAccountID"] == id {
			return "ErrFileAIAccountInUse", nil
		}
	}
	return "", nil
}

func (a AgentAccountRepo) DeleteLocal(ctx context.Context, installID uint) error {
	if installID == 0 {
		return nil
	}
	db := getTx(ctx)
	var accountIDs []uint
	if err := db.Model(&model.AgentAccount{}).Where("app_install_id = ?", installID).Pluck("id", &accountIDs).Error; err != nil {
		return err
	}
	if len(accountIDs) == 0 {
		return nil
	}
	if err := db.Where("account_id IN ?", accountIDs).Delete(&model.AgentAccountModel{}).Error; err != nil {
		return err
	}
	return db.Where("id IN ?", accountIDs).Delete(&model.AgentAccount{}).Error
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
