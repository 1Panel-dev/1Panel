package utils

import (
	"context"
	"sort"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestTransferHostFirewallImportsIptablesRecords(t *testing.T) {
	db := newHostFirewallTransferTestDB(t, true)
	records := []legacyHostFirewallRecord{
		{Type: "port", Protocol: "tcp/udp", SrcIP: "Anywhere", DstPort: "80", Strategy: "accept", Description: "web"},
		{Type: "address", SrcIP: "10.0.0.8", Strategy: "drop", Description: "blocked host"},
		{Chain: "1PANEL_BASIC_BEFORE", Protocol: "tcp", DstPort: "22", Strategy: "accept", Description: "ssh first"},
		{Chain: "1PANEL_INPUT", Protocol: "tcp", DstPort: "9000", Strategy: "accept", Description: "unsupported chain"},
	}
	if err := db.Table("firewalls").Create(&records).Error; err != nil {
		t.Fatal(err)
	}

	if err := transferHostFirewall(context.Background(), db, filter.ProviderIptables); err != nil {
		t.Fatal(err)
	}
	rules := loadTransferredHostFirewallRules(t, db)
	if len(rules) != 4 {
		t.Fatalf("transferred %d rules, want 4", len(rules))
	}
	assertTransferredRuleDefaults(t, rules)

	protocols := make([]string, 0, 2)
	foundAddress, foundAdvanced := false, false
	for _, rule := range rules {
		switch rule.Description {
		case "web":
			protocols = append(protocols, rule.Protocol)
			if rule.Location != filter.IptablesInputChain || rule.Family != string(filter.FamilyIPv4) || rule.DestinationPort != "80" {
				t.Fatalf("unexpected port rule: %#v", rule)
			}
		case "blocked host":
			foundAddress = rule.SourceAddress == "10.0.0.8/32" && rule.Action == string(filter.ActionDrop)
		case "ssh first":
			foundAdvanced = rule.Location == "1PANEL_BASIC_BEFORE" && rule.DestinationPort == "22"
		case "unsupported chain":
			t.Fatal("unsupported legacy chain was imported")
		}
	}
	sort.Strings(protocols)
	if len(protocols) != 2 || protocols[0] != "tcp" || protocols[1] != "udp" {
		t.Fatalf("port protocols = %#v", protocols)
	}
	if !foundAddress || !foundAdvanced {
		t.Fatalf("address=%v advanced=%v", foundAddress, foundAdvanced)
	}
	assertHostFirewallTransferCompleted(t, db)

	if err := transferHostFirewall(context.Background(), db, filter.ProviderIptables); err != nil {
		t.Fatal(err)
	}
	if got := len(loadTransferredHostFirewallRules(t, db)); got != 4 {
		t.Fatalf("retry transferred %d rules, want 4", got)
	}
}

func TestTransferHostFirewallMapsFirewalldRepresentations(t *testing.T) {
	db := newHostFirewallTransferTestDB(t, true)
	records := []legacyHostFirewallRecord{
		{Type: "port", Protocol: "tcp/udp", DstPort: "80", Strategy: "accept", Description: "zone ports"},
		{Type: "port", Protocol: "tcp", DstPort: "81", Strategy: "drop", Description: "dual family deny"},
		{Type: "port", Protocol: "tcp", SrcIP: "2001:db8::8", DstPort: "443", Strategy: "accept", Description: "v6 source"},
		{Type: "address", SrcIP: "10.0.0.9", Strategy: "drop", Description: "v4 address"},
	}
	if err := db.Table("firewalls").Create(&records).Error; err != nil {
		t.Fatal(err)
	}

	if err := transferHostFirewall(context.Background(), db, filter.ProviderFirewalld); err != nil {
		t.Fatal(err)
	}
	rules := loadTransferredHostFirewallRules(t, db)
	if len(rules) != 6 {
		t.Fatalf("transferred %d rules, want 6", len(rules))
	}
	zonePorts, denyFamilies := 0, make(map[string]bool)
	for _, rule := range rules {
		if rule.Location != filter.FirewalldInputZone {
			t.Fatalf("unexpected firewalld location %q", rule.Location)
		}
		switch rule.Description {
		case "zone ports":
			zonePorts++
			if rule.NativeKind != string(filter.NativeKindZonePort) || rule.Family != string(filter.FamilyInet) {
				t.Fatalf("unexpected zone port: %#v", rule)
			}
		case "dual family deny":
			if rule.NativeKind != string(filter.NativeKindRichRule) {
				t.Fatalf("unexpected deny native kind: %#v", rule)
			}
			denyFamilies[rule.Family] = true
		case "v6 source":
			if rule.Family != string(filter.FamilyIPv6) || rule.SourceAddress != "2001:db8::8/128" {
				t.Fatalf("unexpected v6 rule: %#v", rule)
			}
		}
	}
	if zonePorts != 2 || !denyFamilies[string(filter.FamilyIPv4)] || !denyFamilies[string(filter.FamilyIPv6)] {
		t.Fatalf("zonePorts=%d denyFamilies=%#v", zonePorts, denyFamilies)
	}
}

func TestTransferHostFirewallExpandsUFWFamilies(t *testing.T) {
	db := newHostFirewallTransferTestDB(t, true)
	records := []legacyHostFirewallRecord{
		{Type: "port", Protocol: "tcp", DstPort: "8080", Strategy: "accept", Description: "dual family port"},
		{Type: "address", SrcIP: "10.0.0.1-10.0.0.2", Strategy: "drop", Description: "from to"},
	}
	if err := db.Table("firewalls").Create(&records).Error; err != nil {
		t.Fatal(err)
	}

	if err := transferHostFirewall(context.Background(), db, filter.ProviderUFW); err != nil {
		t.Fatal(err)
	}
	rules := loadTransferredHostFirewallRules(t, db)
	if len(rules) != 3 {
		t.Fatalf("transferred %d rules, want 3", len(rules))
	}
	portFamilies := make(map[string]bool)
	foundFromTo := false
	for _, rule := range rules {
		if rule.NativeKind != string(filter.NativeKindUFWRule) || rule.Location != filter.UFWInputChain {
			t.Fatalf("unexpected ufw representation: %#v", rule)
		}
		if rule.Description == "dual family port" {
			portFamilies[rule.Family] = true
		}
		if rule.Description == "from to" {
			foundFromTo = rule.SourceAddress == "10.0.0.1/32" && rule.DestinationAddress == "10.0.0.2/32"
		}
	}
	if !portFamilies[string(filter.FamilyIPv4)] || !portFamilies[string(filter.FamilyIPv6)] || !foundFromTo {
		t.Fatalf("portFamilies=%#v foundFromTo=%v", portFamilies, foundFromTo)
	}
}

func TestTransferHostFirewallWithoutLegacyTableOnlyMarksCompletion(t *testing.T) {
	db := newHostFirewallTransferTestDB(t, false)
	if err := transferHostFirewall(context.Background(), db, filter.ProviderNftables); err != nil {
		t.Fatal(err)
	}
	if got := len(loadTransferredHostFirewallRules(t, db)); got != 0 {
		t.Fatalf("transferred %d rules without a legacy table", got)
	}
	assertHostFirewallTransferCompleted(t, db)
}

func TestTransferHostFirewallRestoresMissingDescription(t *testing.T) {
	db := newHostFirewallTransferTestDB(t, true)
	record := legacyHostFirewallRecord{
		Type: "port", Protocol: "tcp", DstPort: "443", Strategy: "accept", Description: "legacy tls",
	}
	if err := db.Table("firewalls").Create(&record).Error; err != nil {
		t.Fatal(err)
	}
	domainRules, err := legacyHostFirewallRules(record, filter.ProviderIptables)
	if err != nil {
		t.Fatal(err)
	}
	existing, err := hostFirewallRuleModel(domainRules[0])
	if err != nil {
		t.Fatal(err)
	}
	existing.Description = ""
	existing.Origin = constant.FirewallRuleOriginCreated
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}

	if err := transferHostFirewall(context.Background(), db, filter.ProviderIptables); err != nil {
		t.Fatal(err)
	}
	rules := loadTransferredHostFirewallRules(t, db)
	if len(rules) != 1 || rules[0].Description != "legacy tls" || rules[0].Origin != constant.FirewallRuleOriginCreated {
		t.Fatalf("unexpected existing rule after transfer: %#v", rules)
	}
}

func newHostFirewallTransferTestDB(t *testing.T, withLegacyTable bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.FirewallRule{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE migrations (id VARCHAR(255) PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	if withLegacyTable {
		if err := db.Table("firewalls").AutoMigrate(&legacyHostFirewallRecord{}); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func loadTransferredHostFirewallRules(t *testing.T, db *gorm.DB) []model.FirewallRule {
	t.Helper()
	var rules []model.FirewallRule
	if err := db.Order("scope_key ASC, rule_key ASC").Find(&rules).Error; err != nil {
		t.Fatal(err)
	}
	return rules
}

func assertTransferredRuleDefaults(t *testing.T, rules []model.FirewallRule) {
	t.Helper()
	for _, rule := range rules {
		if rule.UUID == "" || rule.RuleKey == "" || rule.ScopeKey == "" || rule.Revision != 1 ||
			rule.Origin != constant.FirewallRuleOriginAdopted || rule.Owner != constant.FirewallRuleSourceUser || rule.MatchKey != "" {
			t.Fatalf("unexpected transferred defaults: %#v", rule)
		}
	}
}

func assertHostFirewallTransferCompleted(t *testing.T, db *gorm.DB) {
	t.Helper()
	completed, err := migrationRecordExists(db, hostFirewallTransferMigrationID)
	if err != nil {
		t.Fatal(err)
	}
	if !completed {
		t.Fatal("host firewall transfer was not marked complete")
	}
}
