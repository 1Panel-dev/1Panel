package model

type Firewall struct {
	BaseModel

	Type         string `gorm:"not null" json:"type"`
	FirewallType string `gorm:"not null" json:"firewallType"`
	Port         string `gorm:"not null" json:"port"`    // Deprecated
	Address      string `gorm:"not null" json:"address"` // Deprecated

	Chain       string `json:"chain"`
	Protocol    string `json:"protocol"`
	SrcIP       string `json:"srcIP"`
	SrcPort     string `json:"srcPort"`
	DstIP       string `json:"dstIP"`
	DstPort     string `json:"dstPort"`
	Strategy    string `gorm:"not null" json:"strategy"`
	Description string `json:"description"`
}
