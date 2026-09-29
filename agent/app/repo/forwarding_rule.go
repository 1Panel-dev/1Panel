package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type IForwardingRuleRepo interface {
	List(context.Context) ([]model.ForwardingRule, error)
	CreateBatch(context.Context, []model.ForwardingRule) error
	DeleteBatch(context.Context, []uint) error
}

type ForwardingRuleRepo struct{}

func NewIForwardingRuleRepo() IForwardingRuleRepo { return &ForwardingRuleRepo{} }

func (r *ForwardingRuleRepo) List(ctx context.Context) ([]model.ForwardingRule, error) {
	var rules []model.ForwardingRule
	err := global.DB.WithContext(ctx).Order("id ASC").Find(&rules).Error
	return rules, err
}

func (r *ForwardingRuleRepo) CreateBatch(ctx context.Context, rules []model.ForwardingRule) error {
	if len(rules) == 0 {
		return nil
	}
	return global.DB.WithContext(ctx).CreateInBatches(&rules, 500).Error
}

func (r *ForwardingRuleRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return global.DB.WithContext(ctx).Where("id IN ?", ids).Delete(&model.ForwardingRule{}).Error
}
