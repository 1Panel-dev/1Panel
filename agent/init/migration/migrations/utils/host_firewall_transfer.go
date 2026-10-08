package utils

import (
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
	"gorm.io/gorm"
)

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

type legacyFirewallRuleDescription struct {
	UUID               string
	Provider           string
	ScopeKey           string
	Location           string
	Family             string
	NativeKind         string
	Protocol           string
	SourceAddress      string
	SourcePort         string
	DestinationAddress string
	DestinationPort    string
	Interface          string
	ConnectionStates   string
	CompatibilityError string
	Action             string
	Priority           *int
	Description        string
	Owner              string
}

func MigrateHostFirewallDescriptions(db *gorm.DB) error {
	providers := []filter.Provider{filter.ProviderIptables, filter.ProviderNftables, filter.ProviderFirewalld, filter.ProviderUFW}
	descriptions := make(map[string][]string)
	add := func(id, description string) {
		if description != "" && !slices.Contains(descriptions[id], description) {
			descriptions[id] = append(descriptions[id], description)
		}
	}
	addRules := func(rules []filter.FirewallRule, description string) bool {
		matched := false
		for _, rule := range rules {
			var err error
			switch rule.Scope.Provider {
			case filter.ProviderFirewalld:
				rule, err = (&filterfirewalld.Adapter{}).PrepareRule(rule)
			case filter.ProviderUFW:
				rule, err = (&filterufw.Adapter{}).PrepareRule(rule)
			}
			if err != nil {
				continue
			}
			id, err := filter.DescriptionID(filter.ObservedRule{Rule: rule, ParseStatus: filter.ParseStatusSupported})
			if err == nil {
				add(id, description)
				matched = true
			}
		}
		return matched
	}
	if db.Migrator().HasTable("firewalls") {
		var records []legacyHostFirewallRecord
		if err := db.Table("firewalls").Order("id ASC").Find(&records).Error; err != nil {
			return err
		}
		for _, record := range records {
			if record.Description == "" {
				continue
			}
			matched := false
			for _, provider := range providers {
				rules, err := legacyHostFirewallRules(record, provider)
				if err == nil && addRules(rules, record.Description) {
					matched = true
				}
			}
			if !matched {
				add(fmt.Sprintf("firewall:legacy:firewalls:%d", record.ID), record.Description)
			}
		}
	}
	if db.Migrator().HasTable("firewall_rules") {
		var records []legacyFirewallRuleDescription
		if err := db.Table("firewall_rules").Order("uuid ASC").Find(&records).Error; err != nil {
			return err
		}
		for _, record := range records {
			if record.Description == "" {
				continue
			}
			matched := false
			for _, provider := range providers {
				rules, err := legacyFirewallDescriptionRules(record, provider)
				if err == nil && addRules(rules, record.Description) {
					matched = true
				}
			}
			if !matched {
				add("firewall:legacy:firewall_rules:"+record.UUID, record.Description)
			}
		}
	}
	if len(descriptions) == 0 {
		return nil
	}
	var existing []model.CommonDescription
	if err := db.Where("type = ?", "firewall").Find(&existing).Error; err != nil {
		return err
	}
	byID := make(map[string]model.CommonDescription, len(existing))
	for _, description := range existing {
		byID[description.ID] = description
	}
	ids := make([]string, 0, len(descriptions))
	for id := range descriptions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		value := strings.Join(descriptions[id], "\n")
		if current, ok := byID[id]; ok {
			if current.Description == value {
				continue
			}
			if current.Description == "" {
				if err := db.Model(&model.CommonDescription{}).Where("id = ?", id).Update("description", value).Error; err != nil {
					return err
				}
				continue
			}
			id = "firewall:legacy:description:" + strings.TrimPrefix(id, "firewall:")
			if _, ok := byID[id]; ok {
				continue
			}
		}
		record := model.CommonDescription{ID: id, Type: "firewall", Description: value}
		if strings.HasPrefix(id, "firewall:legacy:") {
			record.DetailType = "legacy"
		}
		if err := db.Create(&record).Error; err != nil {
			return err
		}
	}
	return nil
}

func legacyFirewallDescriptionRules(record legacyFirewallRuleDescription, provider filter.Provider) ([]filter.FirewallRule, error) {
	if record.CompatibilityError != "" {
		return nil, errUnsupportedLegacyHostFirewallRule
	}
	source := filter.Provider(strings.ToLower(strings.TrimSpace(record.Provider)))
	scopeParts := strings.Split(record.ScopeKey, ":")
	if source == "" && len(scopeParts) > 1 {
		source = filter.Provider(scopeParts[0])
	}
	if source != "" && source != provider {
		return nil, nil
	}
	base := filter.FirewallRule{
		Protocol: record.Protocol, SourceAddress: record.SourceAddress, SourcePort: record.SourcePort,
		DestinationAddress: record.DestinationAddress, DestinationPort: record.DestinationPort,
		Interface: record.Interface, Action: filter.Action(record.Action),
	}
	if record.ConnectionStates != "" {
		base.ConnectionStates = strings.Split(record.ConnectionStates, ",")
	}
	if source != "" {
		base.NativeKind = filter.NativeKind(record.NativeKind)
	}
	switch base.NativeKind {
	case filter.NativeKindOpaque, filter.NativeKindZoneService, filter.NativeKindUFWApplication:
		return nil, errUnsupportedLegacyHostFirewallRule
	}
	if provider != filter.ProviderUFW && strings.EqualFold(base.Protocol, "all") && base.SourcePort == "" && base.DestinationPort != "" {
		base.Protocol = "tcp/udp"
	}
	if provider == filter.ProviderFirewalld {
		base.Priority = record.Priority
	}
	families := []filter.Family{filter.Family(record.Family)}
	if families[0] == "" {
		families[0] = legacyRuleFamily(base.SourceAddress, base.DestinationAddress)
	}
	if provider != filter.ProviderFirewalld && families[0] == filter.FamilyInet {
		if base.SourceAddress != "" || base.DestinationAddress != "" || base.Protocol == "icmpv6" {
			families = []filter.Family{legacyRuleFamily(base.SourceAddress, base.DestinationAddress)}
			if base.Protocol == "icmpv6" {
				families[0] = filter.FamilyIPv6
			}
		} else {
			families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
		}
	}
	var result []filter.FirewallRule
	for _, family := range families {
		rule := base
		rule.Scope = filter.Scope{Provider: provider, Family: family, Direction: filter.DirectionInput}
		switch provider {
		case filter.ProviderIptables, filter.ProviderNftables:
			rule.Scope.Table, rule.Scope.Chain = "filter", filter.IptablesInputChain
			if source != "" {
				if record.Location != "" {
					rule.Scope.Chain = record.Location
				}
				if len(scopeParts) == 5 {
					rule.Scope.Table, rule.Scope.Chain = scopeParts[2], scopeParts[3]
				}
			}
		case filter.ProviderFirewalld:
			rule.Scope.Zone = filter.FirewalldInputZone
			if source != "" && record.Location != "" {
				rule.Scope.Zone = record.Location
			}
		case filter.ProviderUFW:
			rule.Scope.Chain = filter.UFWInputChain
		}
		expanded, err := filter.ExpandAtomicRules(rule)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
		if source == "" && (provider == filter.ProviderIptables || provider == filter.ProviderNftables) && strings.HasPrefix(record.Owner, constant.FirewallRuleSourceSecurity+":"+constant.FirewallSystemAcceptedPortSourcePrefix) {
			for _, item := range expanded {
				item.Scope.Chain = filter.BasicBeforeChain
				result = append(result, item)
			}
		}
	}
	return result, nil
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
		if provider != filter.ProviderIptables || legacyIptablesAdvancedChain(record.Chain) {
			return nil, fmt.Errorf("%w: advanced rule for provider %q", errUnsupportedLegacyHostFirewallRule, provider)
		}
	}

	switch provider {
	case filter.ProviderIptables:
		rule.Scope = filter.Scope{
			Provider: provider, Family: legacyRuleFamily(rule.SourceAddress, rule.DestinationAddress), Table: "filter",
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

func legacyIptablesAdvancedChain(chain string) bool {
	switch strings.ToUpper(strings.TrimSpace(chain)) {
	case "1PANEL_INPUT", "1PANEL_OUTPUT":
		return true
	default:
		return false
	}
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
	if legacyUFWSinglePortAllProtocols(record, rule.DestinationPort) {
		rule.Protocol = "all"
	}
	if legacyAddressIsEmpty(rule.SourceAddress) && legacyAddressIsEmpty(rule.DestinationAddress) {
		rule.Scope.Family = filter.FamilyInet
	} else {
		rule.Scope.Family = legacyRuleFamily(rule.SourceAddress, rule.DestinationAddress)
	}
	return filter.ExpandAtomicRules(rule)
}

func legacyUFWSinglePortAllProtocols(record legacyHostFirewallRecord, port string) bool {
	if !strings.EqualFold(strings.TrimSpace(record.Type), "port") {
		return false
	}
	protocol := strings.ToLower(strings.TrimSpace(record.Protocol))
	if protocol != "tcp/udp" && protocol != "udp/tcp" {
		return false
	}
	port = strings.TrimSpace(port)
	return port != "" && !strings.Contains(port, ",") && !strings.Contains(port, "-")
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
