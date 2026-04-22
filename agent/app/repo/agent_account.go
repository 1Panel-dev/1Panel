package repo

import "github.com/1Panel-dev/1Panel/agent/app/model"

type AgentAccountRepo struct{}

type IAgentAccountRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.AgentAccount, error)
	GetFirst(opts ...DBOption) (*model.AgentAccount, error)
	Create(account *model.AgentAccount) error
	Save(account *model.AgentAccount) error
	DeleteByID(id uint) error
	List(opts ...DBOption) ([]model.AgentAccount, error)
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
