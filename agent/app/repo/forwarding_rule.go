package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type IForwardingRuleRepo interface {
	List(context.Context) ([]model.ForwardingRule, error)
	ReplaceAll(context.Context, []model.ForwardingRule) error
}

type ForwardingRuleRepo struct{}

func NewIForwardingRuleRepo() IForwardingRuleRepo { return &ForwardingRuleRepo{} }

func (r *ForwardingRuleRepo) List(ctx context.Context) ([]model.ForwardingRule, error) {
	var rules []model.ForwardingRule
	err := global.DB.WithContext(ctx).Order("id ASC").Find(&rules).Error
	return rules, err
}

func (r *ForwardingRuleRepo) ReplaceAll(ctx context.Context, rules []model.ForwardingRule) error {
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.ForwardingRule{}).Error; err != nil {
			return err
		}
		if len(rules) == 0 {
			return nil
		}
		return tx.Create(&rules).Error
	})
}
