package migrations

import "github.com/1Panel-dev/1Panel/agent/app/model"

// legacyFirewallMigration exists only so historical migrations can still be
// replayed. Runtime firewall code must not depend on the retired firewalls table.
type legacyFirewallMigration struct {
	model.BaseModel

	Type    string
	Port    string
	Address string

	Chain       string
	Protocol    string
	SrcIP       string
	SrcPort     string
	DstIP       string
	DstPort     string
	Strategy    string `gorm:"not null"`
	Description string
}

func (legacyFirewallMigration) TableName() string {
	return "firewalls"
}
