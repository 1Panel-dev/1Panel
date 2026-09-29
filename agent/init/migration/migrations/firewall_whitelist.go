package migrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle/providers"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const firewalldSSHServiceMigrationID = "20260916-remove-firewalld-ssh-service"

func TransferFirewalldSSHService(ctx context.Context, client lifecycle.Client, syncWhitelist func(context.Context) error) error {
	return transferFirewalldSSHService(ctx, global.DB, client, syncWhitelist)
}

func transferFirewalldSSHService(ctx context.Context, db *gorm.DB, client lifecycle.Client, syncWhitelist func(context.Context) error) error {
	if err := syncWhitelist(ctx); err != nil {
		return err
	}
	if client.Name() != lifecycle.ProviderFirewalld {
		return nil
	}
	var count int64
	if err := db.WithContext(ctx).Table("migrations").Where("id = ?", firewalldSSHServiceMigrationID).Count(&count).Error; err != nil {
		return fmt.Errorf("check firewalld SSH service migration: %w", err)
	}
	if count > 0 {
		return nil
	}
	active, err := client.Status()
	if err != nil || !active {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := providers.RemoveFirewalldSSHService(); err != nil {
		return fmt.Errorf("transfer firewalld SSH access to whitelist: %w", err)
	}
	if err := db.WithContext(ctx).Table("migrations").Clauses(clause.OnConflict{DoNothing: true}).
		Create(map[string]interface{}{"id": firewalldSSHServiceMigrationID}).Error; err != nil {
		return fmt.Errorf("record firewalld SSH service migration: %w", err)
	}
	return nil
}

var MigrateFirewallPortWhitelistSources = &gormigrate.Migration{
	ID: "20260915-migrate-firewall-port-whitelist-sources",
	Migrate: func(tx *gorm.DB) error {
		var setting model.Setting
		err := tx.Where("key = ?", constant.FirewallPortWhiteList).First(&setting).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		rules, err := migrateFirewallPortWhitelist(setting.Value)
		if err != nil {
			return fmt.Errorf("migrate firewall port whitelist: %w", err)
		}
		value, err := json.Marshal(rules)
		if err != nil {
			return err
		}
		if setting.ID == 0 {
			err = tx.Create(&model.Setting{Key: constant.FirewallPortWhiteList, Value: string(value)}).Error
		} else {
			err = tx.Model(&setting).Update("value", string(value)).Error
		}
		if err != nil {
			return err
		}
		return tx.Where("key = ?", "FirewallPortWhiteListPending").Delete(&model.Setting{}).Error
	},
}

type legacyPortWhitelist struct {
	Ports    []string `json:"ports"`
	Family   string   `json:"family"`
	Port     string   `json:"port"`
	Protocol string   `json:"protocol"`
	Type     string   `json:"type"`
	Sources  []string `json:"sources"`
}

func (entry legacyPortWhitelist) singlePortRule() firewall.PortWhitelist {
	rule := firewall.PortWhitelist{Port: entry.Port, Protocol: entry.Protocol, Type: entry.Type, Sources: entry.Sources}
	if strings.TrimSpace(rule.Type) != "" && rule.Port == "" && len(entry.Ports) > 0 {
		rule.Port = entry.Ports[0]
	}
	return rule
}

func migrateFirewallPortWhitelist(value string) ([]firewall.PortWhitelist, error) {
	legacy, err := parseLegacyPortWhitelist(value)
	if err != nil {
		return nil, err
	}
	rules := make([]firewall.PortWhitelist, 0, len(legacy)+5)
	indexes := make(map[string]int)
	key := func(rule firewall.PortWhitelist) string {
		if rule.Type != "" {
			return rule.Type + "/" + rule.Protocol
		}
		return rule.Type + "/" + rule.Protocol + "/" + rule.Port
	}
	for index, entry := range legacy {
		family := strings.ToLower(strings.TrimSpace(entry.Family))
		if family != "" && family != constant.FirewallFamilyIPv4 && family != constant.FirewallFamilyIPv6 {
			return nil, fmt.Errorf("entry #%d: invalid address family %q", index+1, entry.Family)
		}
		rule := entry.singlePortRule()
		if strings.TrimSpace(rule.Protocol) == "" {
			rule.Protocol = "tcp"
		}
		if len(rule.Sources) == 0 {
			rule.Sources = []string{"0.0.0.0/0"}
			if family == constant.FirewallFamilyIPv6 {
				rule.Sources = []string{"::/0"}
			} else if family == "" && strings.TrimSpace(rule.Type) != "" {
				rule.Sources = append(rule.Sources, "::/0")
			}
		}
		rule.Sources, err = firewall.NormalizeWhitelistSources(family, rule.Sources)
		if err != nil {
			return nil, fmt.Errorf("entry #%d: %w", index+1, err)
		}
		normalized, err := service.InitializeFirewallWhitelistPorts([]firewall.PortWhitelist{rule})
		if err != nil {
			return nil, fmt.Errorf("entry #%d: %w", index+1, err)
		}
		rule = normalized[0]
		if existing, found := indexes[key(rule)]; found {
			rules[existing].Sources, err = firewall.NormalizeWhitelistSources("", append(rules[existing].Sources, rule.Sources...))
			if err != nil {
				return nil, err
			}
			continue
		}
		indexes[key(rule)] = len(rules)
		rules = append(rules, rule)
	}

	defaults := []firewall.PortWhitelist{
		{Type: firewall.PortWhitelistTypePanel, Protocol: "tcp"},
		{Type: firewall.PortWhitelistTypeSSH, Protocol: "tcp"},
		{Port: "443", Protocol: "tcp"},
		{Port: "443", Protocol: "udp"},
		{Port: "80", Protocol: "tcp"},
	}
	for _, rule := range defaults {
		index, found := indexes[key(rule)]
		if !found {
			if rule.Type == "" {
				continue
			}
			index = len(rules)
			indexes[key(rule)] = index
			rules = append(rules, rule)
		}
		var ipv4, ipv6 bool
		for _, source := range rules[index].Sources {
			if strings.Contains(source, ":") {
				ipv6 = true
			} else {
				ipv4 = true
			}
		}
		if !ipv4 {
			rules[index].Sources = append(rules[index].Sources, "0.0.0.0/0")
		}
		if !ipv6 {
			rules[index].Sources = append(rules[index].Sources, "::/0")
		}
	}
	return service.InitializeFirewallWhitelistPorts(rules)
}

func parseLegacyPortWhitelist(value string) ([]legacyPortWhitelist, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "null" {
		return nil, nil
	}
	if strings.HasPrefix(value, "[") {
		var rules []legacyPortWhitelist
		err := json.Unmarshal([]byte(value), &rules)
		return rules, err
	}
	items := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' || unicode.IsSpace(r) })
	rules := make([]legacyPortWhitelist, 0, len(items))
	for _, item := range items {
		parts := strings.Split(item, "/")
		rule := legacyPortWhitelist{}
		switch len(parts) {
		case 1:
			rule.Port = parts[0]
		case 2:
			rule.Port, rule.Protocol = parts[0], parts[1]
		case 3:
			rule.Family, rule.Port, rule.Protocol = parts[0], parts[1], parts[2]
		default:
			return nil, fmt.Errorf("invalid legacy whitelist entry %q", item)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
