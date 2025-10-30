package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type IIptablesFilterRuleRepo interface {
	Create(ctx context.Context, rule *model.IptablesFilterRule) error
	Delete(ctx context.Context, id uint) error
	DeleteByChain(ctx context.Context, chain string) error
	GetByID(ctx context.Context, id uint) (model.IptablesFilterRule, error)
	List(ctx context.Context, chains []string) ([]model.IptablesFilterRule, error)
	ListByChain(ctx context.Context, chain string) ([]model.IptablesFilterRule, error)
	GetMaxRuleOrder(ctx context.Context, chain string) (int, error)
}

type IptablesFilterRuleRepo struct{}

func NewIIptablesFilterRuleRepo() IIptablesFilterRuleRepo {
	return &IptablesFilterRuleRepo{}
}

func (r *IptablesFilterRuleRepo) Create(ctx context.Context, rule *model.IptablesFilterRule) error {
	return global.DB.WithContext(ctx).Create(rule).Error
}

func (r *IptablesFilterRuleRepo) Delete(ctx context.Context, id uint) error {
	return global.DB.WithContext(ctx).Where("id = ?", id).Delete(&model.IptablesFilterRule{}).Error
}

func (r *IptablesFilterRuleRepo) DeleteByChain(ctx context.Context, chain string) error {
	return global.DB.WithContext(ctx).Where("chain = ?", chain).Delete(&model.IptablesFilterRule{}).Error
}

func (r *IptablesFilterRuleRepo) GetByID(ctx context.Context, id uint) (model.IptablesFilterRule, error) {
	var rule model.IptablesFilterRule
	err := global.DB.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	return rule, err
}

func (r *IptablesFilterRuleRepo) List(ctx context.Context, chains []string) ([]model.IptablesFilterRule, error) {
	var rules []model.IptablesFilterRule
	query := global.DB.WithContext(ctx)
	if len(chains) > 0 {
		query = query.Where("chain IN ?", chains)
	}
	err := query.Order("chain ASC, rule_order ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (r *IptablesFilterRuleRepo) ListByChain(ctx context.Context, chain string) ([]model.IptablesFilterRule, error) {
	var rules []model.IptablesFilterRule
	err := global.DB.WithContext(ctx).Where("chain = ?", chain).Order("rule_order ASC, id ASC").Find(&rules).Error
	return rules, err
}

func (r *IptablesFilterRuleRepo) GetMaxRuleOrder(ctx context.Context, chain string) (int, error) {
	var maxOrder int
	err := global.DB.WithContext(ctx).Model(&model.IptablesFilterRule{}).
		Where("chain = ?", chain).
		Select("COALESCE(MAX(rule_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}
