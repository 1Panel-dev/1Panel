package utils

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const hostFirewallTransferMigrationID = "host-firewall-transfer"

var errUnsupportedLegacyHostFirewallRule = errors.New("unsupported legacy host firewall rule")

type legacyHostFirewallRecord struct {
	ID          uint
	Type        string
	Port        string
	Address     string
	Chain       string
	Protocol    string
	SrcIP       string
	SrcPort     string
	DstIP       string
	DstPort     string
	Strategy    string
	Description string
}

// TransferHostFirewall imports the legacy firewalls table into firewall_rules.
// It intentionally does not inspect or mutate the system firewall: the normal
// inventory merge associates imported rows with observed rules by RuleKey.
func TransferHostFirewall(ctx context.Context, provider string) error {
	if global.DB == nil {
		return errors.New("host firewall transfer database is required")
	}
	return transferHostFirewall(ctx, global.DB, filter.Provider(strings.ToLower(strings.TrimSpace(provider))))
}

func transferHostFirewall(ctx context.Context, db *gorm.DB, provider filter.Provider) error {
	if db == nil {
		return errors.New("host firewall transfer database is required")
	}
	completed, err := migrationRecordExists(db, hostFirewallTransferMigrationID)
	if err != nil || completed {
		return err
	}
	if !isLegacyHostFirewallProvider(provider) {
		return fmt.Errorf("unsupported legacy host firewall provider %q", provider)
	}

	models := make([]model.FirewallRule, 0)
	if db.Migrator().HasTable("firewalls") {
		var records []legacyHostFirewallRecord
		if err := db.WithContext(ctx).Table("firewalls").Order("id ASC").Find(&records).Error; err != nil {
			return fmt.Errorf("load legacy host firewall records: %w", err)
		}
		models = convertLegacyHostFirewallRecords(records, provider)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := importLegacyHostFirewallRules(tx, models); err != nil {
			return err
		}
		return markMigrationRecord(tx, hostFirewallTransferMigrationID)
	})
}

func isLegacyHostFirewallProvider(provider filter.Provider) bool {
	switch provider {
	case filter.ProviderIptables, filter.ProviderNftables, filter.ProviderFirewalld, filter.ProviderUFW:
		return true
	default:
		return false
	}
}

func convertLegacyHostFirewallRecords(records []legacyHostFirewallRecord, provider filter.Provider) []model.FirewallRule {
	converted := make([]model.FirewallRule, 0, len(records))
	byIdentity := make(map[string]int)
	for _, record := range records {
		rules, err := legacyHostFirewallRules(record, provider)
		if err != nil {
			if global.LOG != nil {
				global.LOG.Warnf("skip legacy host firewall record %d during transfer: %v", record.ID, err)
			}
			continue
		}
		for _, rule := range rules {
			item, err := hostFirewallRuleModel(rule)
			if err != nil {
				if global.LOG != nil {
					global.LOG.Warnf("skip legacy host firewall record %d during transfer: %v", record.ID, err)
				}
				continue
			}
			identity := item.ScopeKey + "\x00" + item.RuleKey
			if index, exists := byIdentity[identity]; exists {
				if item.Description != "" {
					converted[index].Description = item.Description
				}
				continue
			}
			byIdentity[identity] = len(converted)
			converted = append(converted, item)
		}
	}
	return converted
}

func legacyHostFirewallRules(record legacyHostFirewallRecord, provider filter.Provider) ([]filter.FirewallRule, error) {
	sourceAddress := record.SrcIP
	if sourceAddress == "" {
		sourceAddress = record.Address
	}
	destinationPort := record.DstPort
	if destinationPort == "" {
		destinationPort = record.Port
	}
	rule := filter.FirewallRule{
		Protocol:           record.Protocol,
		SourceAddress:      sourceAddress,
		SourcePort:         record.SrcPort,
		DestinationAddress: record.DstIP,
		DestinationPort:    destinationPort,
		Action:             filter.Action(record.Strategy),
		Description:        record.Description,
	}

	switch strings.ToLower(strings.TrimSpace(record.Type)) {
	case "port":
		rule.SourcePort = ""
		rule.DestinationAddress = ""
		rule.DestinationPort = destinationPort
	case "address", "ip":
		rule.Protocol = "all"
		rule.SourcePort = ""
		rule.DestinationPort = ""
	default:
		if provider != filter.ProviderIptables {
			return nil, fmt.Errorf("%w: advanced rule for provider %q", errUnsupportedLegacyHostFirewallRule, provider)
		}
	}

	switch provider {
	case filter.ProviderIptables:
		rule.Scope = filter.Scope{
			Provider: provider, Family: filter.FamilyIPv4, Table: "filter",
			Chain: legacyIptablesChain(record), Direction: filter.DirectionInput,
		}
		rule.NativeKind = filter.NativeKindRule
	case filter.ProviderFirewalld:
		return legacyFirewalldHostRules(record, rule)
	case filter.ProviderUFW:
		return legacyUFWHostRules(record, rule)
	default:
		return nil, fmt.Errorf("%w: provider %q", errUnsupportedLegacyHostFirewallRule, provider)
	}
	return filter.ExpandAtomicRules(rule)
}

func legacyIptablesChain(record legacyHostFirewallRecord) string {
	typeName := strings.ToLower(strings.TrimSpace(record.Type))
	if typeName == "port" || typeName == "address" || typeName == "ip" {
		return filter.IptablesInputChain
	}
	return strings.TrimSpace(record.Chain)
}

func legacyFirewalldHostRules(record legacyHostFirewallRecord, rule filter.FirewallRule) ([]filter.FirewallRule, error) {
	rule.Scope = filter.Scope{
		Provider: filter.ProviderFirewalld, Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
	}
	typeName := strings.ToLower(strings.TrimSpace(record.Type))
	if typeName == "port" && legacyActionIsAccept(record.Strategy) && legacyAddressIsEmpty(record.SrcIP) {
		rule.Scope.Family = filter.FamilyInet
		rule.NativeKind = filter.NativeKindZonePort
		return filter.ExpandAtomicRules(rule)
	}

	rule.NativeKind = filter.NativeKindRichRule
	if legacyAddressIsEmpty(rule.SourceAddress) && legacyAddressIsEmpty(rule.DestinationAddress) {
		return expandLegacyFamilies(rule, filter.FamilyIPv4, filter.FamilyIPv6)
	}
	rule.Scope.Family = legacyRuleFamily(rule.SourceAddress, rule.DestinationAddress)
	return filter.ExpandAtomicRules(rule)
}

func legacyUFWHostRules(record legacyHostFirewallRecord, rule filter.FirewallRule) ([]filter.FirewallRule, error) {
	rule.Scope = filter.Scope{
		Provider: filter.ProviderUFW, Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	rule.NativeKind = filter.NativeKindUFWRule
	if strings.EqualFold(strings.TrimSpace(record.Type), "address") || strings.EqualFold(strings.TrimSpace(record.Type), "ip") {
		rule.SourceAddress, rule.DestinationAddress = splitLegacyUFWAddress(rule.SourceAddress)
	}
	if legacyAddressIsEmpty(rule.SourceAddress) && legacyAddressIsEmpty(rule.DestinationAddress) {
		rule.Scope.Family = filter.FamilyInet
	} else {
		rule.Scope.Family = legacyRuleFamily(rule.SourceAddress, rule.DestinationAddress)
	}
	return filter.ExpandAtomicRules(rule)
}

func expandLegacyFamilies(rule filter.FirewallRule, families ...filter.Family) ([]filter.FirewallRule, error) {
	result := make([]filter.FirewallRule, 0, len(families))
	for _, family := range families {
		item := rule
		item.Scope.Family = family
		expanded, err := filter.ExpandAtomicRules(item)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

func legacyActionIsAccept(action string) bool {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "accept", "allow":
		return true
	default:
		return false
	}
}

func legacyAddressIsEmpty(address string) bool {
	address = strings.ToLower(strings.TrimSpace(address))
	return address == "" || address == "any" || strings.HasPrefix(address, "anywhere")
}

func legacyRuleFamily(addresses ...string) filter.Family {
	for _, value := range addresses {
		value = strings.TrimSpace(value)
		if prefix, err := netip.ParsePrefix(value); err == nil {
			if prefix.Addr().Is6() && !prefix.Addr().Is4In6() {
				return filter.FamilyIPv6
			}
			continue
		}
		if address, err := netip.ParseAddr(value); err == nil && address.Is6() && !address.Is4In6() {
			return filter.FamilyIPv6
		}
	}
	return filter.FamilyIPv4
}

func splitLegacyUFWAddress(value string) (string, string) {
	value = strings.TrimSpace(value)
	if source, destination, ok := strings.Cut(value, "-"); ok && legacyIPOrPrefix(source) && legacyIPOrPrefix(destination) {
		return strings.TrimSpace(source), strings.TrimSpace(destination)
	}
	return value, ""
}

func legacyIPOrPrefix(value string) bool {
	value = strings.TrimSpace(value)
	if _, err := netip.ParseAddr(value); err == nil {
		return true
	}
	_, err := netip.ParsePrefix(value)
	return err == nil
}

func hostFirewallRuleModel(rule filter.FirewallRule) (model.FirewallRule, error) {
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return model.FirewallRule{}, err
	}
	ruleKey, err := filter.RuleKey(normalized)
	if err != nil {
		return model.FirewallRule{}, err
	}
	return model.FirewallRule{
		UUID:               uuid.NewString(),
		ScopeKey:           normalized.Scope.Key(),
		Provider:           string(normalized.Scope.Provider),
		Family:             string(normalized.Scope.Family),
		Location:           legacyHostFirewallLocation(normalized.Scope),
		NativeKind:         string(normalized.NativeKind),
		Protocol:           normalized.Protocol,
		SourceAddress:      normalized.SourceAddress,
		SourcePort:         normalized.SourcePort,
		DestinationAddress: normalized.DestinationAddress,
		DestinationPort:    normalized.DestinationPort,
		Interface:          normalized.Interface,
		ConnectionStates:   strings.Join(normalized.ConnectionStates, ","),
		Action:             string(normalized.Action),
		Priority:           normalized.Priority,
		OrderIndex:         normalized.OrderIndex,
		OrderBucket:        normalized.OrderBucket,
		Description:        normalized.Description,
		RuleKey:            ruleKey,
		Origin:             constant.FirewallRuleOriginAdopted,
		Owner:              constant.FirewallRuleSourceUser,
		Revision:           1,
	}, nil
}

func legacyHostFirewallLocation(scope filter.Scope) string {
	switch scope.Provider {
	case filter.ProviderFirewalld:
		return scope.Zone
	default:
		return scope.Chain
	}
}

func importLegacyHostFirewallRules(tx *gorm.DB, rules []model.FirewallRule) error {
	var existing []model.FirewallRule
	if err := tx.Find(&existing).Error; err != nil {
		return fmt.Errorf("load current host firewall rules: %w", err)
	}
	byIdentity := make(map[string]model.FirewallRule, len(existing))
	for _, item := range existing {
		byIdentity[item.ScopeKey+"\x00"+item.RuleKey] = item
	}
	for _, item := range rules {
		identity := item.ScopeKey + "\x00" + item.RuleKey
		if current, exists := byIdentity[identity]; exists {
			if current.Description == "" && item.Description != "" {
				if err := tx.Model(&model.FirewallRule{}).Where("uuid = ?", current.UUID).
					Update("description", item.Description).Error; err != nil {
					return fmt.Errorf("restore legacy host firewall description: %w", err)
				}
			}
			continue
		}
		if err := tx.Create(&item).Error; err != nil {
			return fmt.Errorf("import legacy host firewall rule: %w", err)
		}
		byIdentity[identity] = item
	}
	return nil
}
