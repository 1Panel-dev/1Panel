package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const firewallTransferMigrationID = "firewall-transfer"

type legacyFirewalldForward struct {
	rule forwarding.Rule
	spec string
}

type legacyFirewalldForwardFailure struct {
	spec string
	err  error
}

type firewallTransferSource struct {
	rules      []forwarding.Rule
	firewalld  []legacyFirewalldForward
	provider   string
	cleanupOld func([]legacyFirewalldForward) error
}

type firewallTransfer struct {
	db   *gorm.DB
	load func() (firewallTransferSource, error)
}

// TransferFirewallForwarding migrates the legacy system-backed forwarding
// inventory into the forwarding_rules table once. The migrations table is
// also used as the completion ledger so a failed transfer can be retried on
// the next agent start without introducing another persisted setting.
func TransferFirewallForwarding(ctx context.Context) error {
	transfer := &firewallTransfer{
		db:   global.DB,
		load: loadLegacyFirewallForwarding,
	}
	return transfer.run(ctx)
}

func (t *firewallTransfer) run(ctx context.Context) error {
	if t.db == nil {
		return errors.New("firewall transfer database is required")
	}
	completed, err := firewallTransferCompleted(t.db)
	if err != nil {
		return err
	}
	if completed {
		return nil
	}
	if t.load == nil {
		return errors.New("legacy firewall forwarding loader is required")
	}
	source, err := t.load()
	if err != nil {
		return err
	}
	if err := importLegacyForwardingRules(ctx, t.db, source.rules, source.provider); err != nil {
		return err
	}
	if len(source.firewalld) > 0 {
		if source.cleanupOld == nil {
			return errors.New("legacy firewalld cleanup is required")
		}
		if err := source.cleanupOld(source.firewalld); err != nil {
			return err
		}
	}
	return markFirewallTransferCompleted(t.db)
}

func firewallTransferCompleted(db *gorm.DB) (bool, error) {
	return migrationRecordExists(db, firewallTransferMigrationID)
}

func markFirewallTransferCompleted(db *gorm.DB) error {
	return markMigrationRecord(db, firewallTransferMigrationID)
}

func migrationRecordExists(db *gorm.DB, migrationID string) (bool, error) {
	var count int64
	if err := db.Table("migrations").Where("id = ?", migrationID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check migration record %q: %w", migrationID, err)
	}
	return count > 0, nil
}

func markMigrationRecord(db *gorm.DB, migrationID string) error {
	if err := db.Table("migrations").Clauses(clause.OnConflict{DoNothing: true}).
		Create(map[string]interface{}{"id": migrationID}).Error; err != nil {
		return fmt.Errorf("record migration %q: %w", migrationID, err)
	}
	return nil
}

func importLegacyForwardingRules(ctx context.Context, db *gorm.DB, rules []forwarding.Rule, provider string) error {
	models := make([]model.ForwardingRule, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		normalized, err := forwarding.NormalizeRule(rule)
		if err != nil {
			return fmt.Errorf("normalize legacy forwarding rule: %w", err)
		}
		key := normalized.Identity()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		models = append(models, model.ForwardingRule{
			Family: normalized.Family, Protocol: normalized.Protocol, Port: normalized.Port,
			TargetIP: normalized.TargetIP, TargetPort: normalized.TargetPort, Interface: normalized.Interface,
		})
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(models) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models).Error; err != nil {
				return fmt.Errorf("import legacy forwarding rules: %w", err)
			}
			if err := updateOrCreateSetting(tx, "IptablesForwardStatus", constant.StatusEnable); err != nil {
				return err
			}
			if provider != "" {
				if err := updateOrCreateSetting(tx, "ForwardingBackend", provider); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func updateOrCreateSetting(tx *gorm.DB, key, value string) error {
	result := tx.Model(&model.Setting{}).Where("key = ?", key).Update("value", value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	return tx.Create(&model.Setting{Key: key, Value: value}).Error
}

func loadLegacyFirewallForwarding() (firewallTransferSource, error) {
	source := firewallTransferSource{cleanupOld: cleanupLegacyFirewalldForwarding}
	if cmd.Which("firewall-cmd") {
		firewalldRules, err := listLegacyFirewalldForwarding()
		if err != nil {
			return firewallTransferSource{}, err
		}
		supportedRules, cleanupRules, failures := selectSupportedLegacyFirewalldForwarding(firewalldRules)
		for _, failure := range failures {
			if global.LOG != nil {
				global.LOG.Warnf("skip unsupported legacy firewalld forwarding rule %q: %v", failure.spec, failure.err)
			}
		}
		source.rules = append(source.rules, supportedRules...)
		source.firewalld = append(source.firewalld, cleanupRules...)
		if len(source.rules) > 0 {
			source.provider = "iptables"
		}
		return source, nil
	}
	if _, err := lifecycle.ResolveIptablesCommands(); err != nil {
		return source, nil
	}
	rules, err := listLegacyIptablesForwarding()
	if err != nil {
		return firewallTransferSource{}, err
	}
	source.rules = append(source.rules, rules...)
	return source, nil
}

func selectSupportedLegacyFirewalldForwarding(items []legacyFirewalldForward) (
	[]forwarding.Rule, []legacyFirewalldForward, []legacyFirewalldForwardFailure,
) {
	rules := make([]forwarding.Rule, 0, len(items))
	cleanup := make([]legacyFirewalldForward, 0, len(items))
	failures := make([]legacyFirewalldForwardFailure, 0)
	for _, item := range items {
		if _, err := forwarding.NormalizeRule(item.rule); err != nil {
			failures = append(failures, legacyFirewalldForwardFailure{spec: item.spec, err: err})
			continue
		}
		rules = append(rules, item.rule)
		cleanup = append(cleanup, item)
	}
	return rules, cleanup, failures
}

func listLegacyIptablesForwarding() ([]forwarding.Rule, error) {
	exists, err := iptables_helper.CheckChainExist(iptables_helper.NatTab, forwarding.ChainPreRouting)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	stdout, err := iptables_helper.RunWithStd(
		iptables_helper.NatTab, "-nvL", forwarding.ChainPreRouting, "--line-numbers",
	)
	if err != nil {
		return nil, fmt.Errorf("list legacy iptables forwarding rules: %w", err)
	}
	return parseLegacyIptablesForwarding(stdout), nil
}

func parseLegacyIptablesForwarding(stdout string) []forwarding.Rule {
	rules := make([]forwarding.Rule, 0)
	for _, line := range strings.Split(stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			continue
		}
		rule := forwarding.Rule{
			Family: forwarding.FamilyIPv4, Protocol: loadLegacyIptablesProtocol(fields[4]),
			Interface: fields[6], Port: loadLegacyIptablesSourcePort(fields[11]),
		}
		if len(fields) == 15 && fields[13] == "ports" {
			rule.TargetPort = fields[14]
		}
		if len(fields) == 13 && strings.HasPrefix(fields[12], "to:") {
			target := strings.TrimPrefix(fields[12], "to:")
			separator := strings.LastIndex(target, ":")
			if separator > 0 {
				rule.TargetIP, rule.TargetPort = target[:separator], target[separator+1:]
			}
		}
		if rule.TargetIP == "" {
			rule.TargetIP = "127.0.0.1"
		}
		rule.TargetPort = strings.TrimPrefix(rule.TargetPort, ":")
		rules = append(rules, rule)
	}
	return rules
}

func loadLegacyIptablesProtocol(protocol string) string {
	switch protocol {
	case "6":
		return "tcp"
	case "17":
		return "udp"
	default:
		return protocol
	}
}

func loadLegacyIptablesSourcePort(value string) string {
	port := ""
	if strings.Contains(value, "dpt:") {
		port = strings.ReplaceAll(value, "dpt:", "")
	}
	if strings.Contains(value, "dpts:") {
		port = strings.ReplaceAll(value, "dpts:", "")
	}
	return strings.ReplaceAll(port, ":", "-")
}

func listLegacyFirewalldForwarding() ([]legacyFirewalldForward, error) {
	stdout, err := cmd.NewCommandMgr().RunWithStdout(
		"firewall-cmd", "--permanent", "--zone=public", "--list-forward-ports",
	)
	if err != nil {
		return nil, fmt.Errorf("list legacy firewalld forwarding rules: %w", err)
	}
	rules, failures := parseLegacyFirewalldForwarding(stdout)
	for _, failure := range failures {
		if global.LOG != nil {
			global.LOG.Warnf("skip unsupported legacy firewalld forwarding rule %q: %v", failure.spec, failure.err)
		}
	}
	return rules, nil
}

func parseLegacyFirewalldForwarding(stdout string) ([]legacyFirewalldForward, []legacyFirewalldForwardFailure) {
	result := make([]legacyFirewalldForward, 0)
	failures := make([]legacyFirewalldForwardFailure, 0)
	for _, spec := range strings.Fields(stdout) {
		item, err := parseLegacyFirewalldForward(spec)
		if err != nil {
			failures = append(failures, legacyFirewalldForwardFailure{spec: spec, err: err})
			continue
		}
		result = append(result, item)
	}
	return result, failures
}

func parseLegacyFirewalldForward(spec string) (legacyFirewalldForward, error) {
	if !strings.HasPrefix(spec, "port=") {
		return legacyFirewalldForward{}, errors.New("missing port field")
	}
	port, rest, ok := strings.Cut(strings.TrimPrefix(spec, "port="), ":proto=")
	if !ok {
		return legacyFirewalldForward{}, errors.New("missing protocol field")
	}
	protocol, rest, ok := strings.Cut(rest, ":toport=")
	if !ok {
		return legacyFirewalldForward{}, errors.New("missing target port field")
	}
	targetPort, targetIP, ok := strings.Cut(rest, ":toaddr=")
	if !ok {
		return legacyFirewalldForward{}, errors.New("missing target address field")
	}
	if targetIP == "" {
		targetIP = "127.0.0.1"
	}
	return legacyFirewalldForward{
		rule: forwarding.Rule{
			Family: forwarding.FamilyIPv4, Protocol: protocol, Port: port,
			TargetIP: targetIP, TargetPort: targetPort,
		},
		spec: spec,
	}, nil
}

func cleanupLegacyFirewalldForwarding(rules []legacyFirewalldForward) error {
	manager := cmd.NewCommandMgr()
	for _, item := range rules {
		if err := manager.Run(
			"firewall-cmd", "--permanent", "--zone=public", "--remove-forward-port="+item.spec,
		); err != nil {
			return fmt.Errorf("remove legacy firewalld forwarding rule %q: %w", item.spec, err)
		}
	}
	if len(rules) == 0 {
		return nil
	}
	if err := manager.Run("firewall-cmd", "--reload"); err != nil {
		return fmt.Errorf("reload firewalld after forwarding transfer: %w", err)
	}
	return restartDockerAfterFirewalldReload(cmd.Which, controller.CheckActive, controller.HandleRestart)
}

func restartDockerAfterFirewalldReload(
	which func(string) bool,
	checkActive func(string) (bool, error),
	restart func(string) error,
) error {
	const service = "docker"
	if !which(service) {
		return nil
	}
	active, err := checkActive(service)
	if err != nil {
		return fmt.Errorf("check Docker status after reloading firewalld: %w", err)
	}
	if !active {
		return nil
	}
	if err := restart(service); err != nil {
		return fmt.Errorf("restart Docker after reloading firewalld: %w", err)
	}
	return nil
}
