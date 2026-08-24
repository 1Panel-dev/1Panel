package repo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFirewallRuleRepoRevision(t *testing.T) {
	db := newFirewallRepoTestDB(t)
	repository := NewFirewallRuleRepo(db)
	ctx := context.Background()
	rule := newFirewallRuleModel("sha256:rule-one")

	if err := repository.Create(ctx, &rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}
	if rule.UUID == "" || rule.Revision != 1 {
		t.Fatalf("rule defaults were not applied: %#v", rule)
	}

	if err := repository.UpdateWithRevision(ctx, rule.UUID, 2, map[string]interface{}{"description": "updated"}); !errors.Is(err, ErrFirewallRuleRevisionConflict) {
		t.Fatalf("expected revision conflict, got %v", err)
	}
	if err := repository.UpdateWithRevision(ctx, rule.UUID, 1, map[string]interface{}{"description": "updated"}); err != nil {
		t.Fatalf("update rule: %v", err)
	}

	updated, err := repository.GetByUUID(ctx, rule.UUID)
	if err != nil {
		t.Fatalf("get updated rule: %v", err)
	}
	if updated.Revision != 2 || updated.Description != "updated" {
		t.Fatalf("unexpected updated rule: %#v", updated)
	}

	if err := repository.DeleteWithRevision(ctx, rule.UUID, updated.Revision-1); !errors.Is(err, ErrFirewallRuleRevisionConflict) {
		t.Fatalf("expected revision conflict, got %v", err)
	}
	if err := repository.DeleteWithRevision(ctx, rule.UUID, updated.Revision); err != nil {
		t.Fatalf("delete rule: %v", err)
	}
	if _, err := repository.GetByUUID(ctx, rule.UUID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected deleted rule to be absent, got %v", err)
	}
	var deletedCount int64
	if err := db.Model(&model.FirewallRule{}).Where("uuid = ?", rule.UUID).Count(&deletedCount).Error; err != nil || deletedCount != 0 {
		t.Fatalf("rule was not hard deleted: count=%d err=%v", deletedCount, err)
	}
}

func TestFirewallRepositoriesRejectIncompleteRecords(t *testing.T) {
	db := newFirewallRepoTestDB(t)
	ctx := context.Background()
	if err := NewFirewallRuleRepo(db).Create(ctx, &model.FirewallRule{}); !errors.Is(err, ErrFirewallPersistenceInvalid) {
		t.Fatalf("expected invalid rule error, got %v", err)
	}
}

func TestFirewallRepositoriesUseContextTransaction(t *testing.T) {
	db := newFirewallRepoTestDB(t)
	ruleRepo := NewFirewallRuleRepo(db)
	wantRollback := errors.New("rollback")

	err := db.Transaction(func(tx *gorm.DB) error {
		ctx := context.WithValue(context.Background(), constant.DB, tx)
		rule := newFirewallRuleModel("sha256:transaction")
		if err := ruleRepo.Create(ctx, &rule); err != nil {
			return err
		}
		return wantRollback
	})
	if !errors.Is(err, wantRollback) {
		t.Fatalf("expected rollback error, got %v", err)
	}

	var ruleCount int64
	if err := db.Model(&model.FirewallRule{}).Count(&ruleCount).Error; err != nil {
		t.Fatalf("count rules: %v", err)
	}
	if ruleCount != 0 {
		t.Fatalf("transaction did not roll back: rules=%d", ruleCount)
	}
}

func newFirewallRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("load sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&model.FirewallRule{}); err != nil {
		t.Fatalf("migrate models: %v", err)
	}
	return db
}

func newFirewallRuleModel(ruleKey string) model.FirewallRule {
	return model.FirewallRule{
		ScopeKey:        "iptables:ipv4:filter:1PANEL_BASIC:input",
		Provider:        "iptables",
		Family:          "ipv4",
		Location:        "1PANEL_BASIC",
		NativeKind:      "rule",
		Protocol:        "tcp",
		DestinationPort: "22",
		Action:          "accept",
		RuleKey:         ruleKey,
	}
}
