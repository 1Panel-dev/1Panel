package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrFirewallRuleRevisionConflict = errors.New("firewall rule revision conflict")
	ErrFirewallPersistenceInvalid   = errors.New("invalid firewall persistence record")
)

type IFirewallRuleRepo interface {
	Create(context.Context, *model.FirewallRule) error
	GetByUUID(context.Context, string) (model.FirewallRule, error)
	List(context.Context, ...DBOption) ([]model.FirewallRule, error)
	UpdateWithRevision(context.Context, string, uint, map[string]interface{}) error
	DeleteWithRevision(context.Context, string, uint) error
	DeleteBatchWithRevision(context.Context, []model.FirewallRule) map[string]error
	SaveResetOrder(context.Context, []model.FirewallRule) error
}

type FirewallRuleRepo struct {
	db *gorm.DB
}

func NewIFirewallRuleRepo() IFirewallRuleRepo {
	return &FirewallRuleRepo{}
}

func NewFirewallRuleRepo(db *gorm.DB) *FirewallRuleRepo {
	return &FirewallRuleRepo{db: db}
}

func (r *FirewallRuleRepo) Create(ctx context.Context, rule *model.FirewallRule) error {
	if err := prepareFirewallRule(rule); err != nil {
		return err
	}
	return r.dbFor(ctx).Create(rule).Error
}

func (r *FirewallRuleRepo) GetByUUID(ctx context.Context, ruleUUID string) (model.FirewallRule, error) {
	var rule model.FirewallRule
	err := r.dbFor(ctx).Where("uuid = ?", ruleUUID).First(&rule).Error
	return rule, err
}

func (r *FirewallRuleRepo) List(ctx context.Context, opts ...DBOption) ([]model.FirewallRule, error) {
	var rules []model.FirewallRule
	db := r.dbFor(ctx).Model(&model.FirewallRule{})
	for _, opt := range opts {
		db = opt(db)
	}
	return rules, db.Find(&rules).Error
}

func (r *FirewallRuleRepo) UpdateWithRevision(ctx context.Context, ruleUUID string, expectedRevision uint, updates map[string]interface{}) error {
	updates = sanitizeRuleUpdates(updates)
	updates["revision"] = gorm.Expr("revision + 1")
	result := r.dbFor(ctx).Model(&model.FirewallRule{}).
		Where("uuid = ? AND revision = ?", ruleUUID, expectedRevision).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFirewallRuleRevisionConflict
	}
	return nil
}

func (r *FirewallRuleRepo) DeleteWithRevision(ctx context.Context, ruleUUID string, expectedRevision uint) error {
	result := r.dbFor(ctx).
		Where("uuid = ? AND revision = ?", ruleUUID, expectedRevision).
		Delete(&model.FirewallRule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFirewallRuleRevisionConflict
	}
	return nil
}

func (r *FirewallRuleRepo) DeleteBatchWithRevision(ctx context.Context, rules []model.FirewallRule) map[string]error {
	failures := make(map[string]error)
	for start := 0; start < len(rules); start += 500 {
		batch := rules[start:min(start+500, len(rules))]
		ids := make([][]interface{}, 0, len(batch))
		for _, rule := range batch {
			ids = append(ids, []interface{}{rule.UUID, rule.Revision})
			failures[rule.UUID] = ErrFirewallRuleRevisionConflict
		}
		var deleted []model.FirewallRule
		err := r.dbFor(ctx).Clauses(clause.Returning{Columns: []clause.Column{{Name: "uuid"}}}).
			Where("(uuid, revision) IN ?", ids).Delete(&deleted).Error
		if err != nil {
			for _, rule := range batch {
				failures[rule.UUID] = err
			}
			continue
		}
		for _, rule := range deleted {
			delete(failures, rule.UUID)
		}
	}
	return failures
}

func (r *FirewallRuleRepo) SaveResetOrder(ctx context.Context, rules []model.FirewallRule) error {
	if len(rules) == 0 {
		return nil
	}
	return r.dbFor(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			result := tx.Model(&model.FirewallRule{}).
				Where("uuid = ? AND revision = ?", rule.UUID, rule.Revision).
				Updates(map[string]interface{}{"sequence": rule.Sequence, "priority": rule.Priority, "revision": gorm.Expr("revision + 1")})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrFirewallRuleRevisionConflict
			}
		}
		return nil
	})
}

func (r *FirewallRuleRepo) dbFor(ctx context.Context) *gorm.DB {
	return firewallDB(ctx, r.db)
}

func firewallDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	if tx, ok := ctx.Value(constant.DB).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	if fallback == nil {
		fallback = global.DB
	}
	return fallback.WithContext(ctx)
}

func prepareFirewallRule(rule *model.FirewallRule) error {
	if rule == nil {
		return fmt.Errorf("%w: rule is nil", ErrFirewallPersistenceInvalid)
	}
	if rule.Family == "" || rule.Protocol == "" || rule.Action == "" {
		return fmt.Errorf("%w: atomic rule identity fields are required", ErrFirewallPersistenceInvalid)
	}
	if rule.UUID == "" {
		rule.UUID = uuid.NewString()
	}
	if rule.Revision == 0 {
		rule.Revision = 1
	}
	if rule.Origin == "" {
		rule.Origin = constant.FirewallRuleOriginCreated
	}
	if rule.Owner == "" {
		rule.Owner = constant.FirewallRuleSourceUser
	}
	return nil
}

func sanitizeRuleUpdates(updates map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(updates)+1)
	for key, value := range updates {
		result[key] = value
	}
	delete(result, "id")
	delete(result, "uuid")
	delete(result, "revision")
	delete(result, "created_at")
	return result
}
