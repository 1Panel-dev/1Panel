package migrations

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddFirewallRuleTable = &gormigrate.Migration{
	ID: "20260803-add-firewall-v2-tables",
	Migrate: func(tx *gorm.DB) error {
		return migrateFirewallRuleTable(tx)
	},
}

var RemoveFirewallLegacyInfluence = &gormigrate.Migration{
	ID: "20260807-remove-firewall-legacy-influence",
	Migrate: func(tx *gorm.DB) error {
		return removeFirewallLegacyInfluence(tx)
	},
}

func migrateFirewallRuleTable(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&model.FirewallRule{}); err != nil {
		return err
	}

	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_firewall_rules_scope_rule
			ON firewall_rules(scope_key, rule_key)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_firewall_rules_scope_match
			ON firewall_rules(scope_key, match_key) WHERE match_key <> ''`,
		`CREATE INDEX IF NOT EXISTS idx_firewall_rules_owner
			ON firewall_rules(owner)`,
	}
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func removeFirewallLegacyInfluence(tx *gorm.DB) error {
	if err := tx.Where("origin = ?", "legacy").Delete(&model.FirewallRule{}).Error; err != nil {
		return err
	}

	return tx.Where("key = ?", "FirewallV2").Delete(&model.Setting{}).Error
}
