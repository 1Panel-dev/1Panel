package model

type Firewall struct {
	BaseModel

	Type    string `json:"type"`
	Port    string `json:"port"`    // Deprecated
	Address string `json:"address"` // Deprecated

	Chain       string `json:"chain"`
	Protocol    string `json:"protocol"`
	SrcIP       string `json:"srcIP"`
	SrcPort     string `json:"srcPort"`
	DstIP       string `json:"dstIP"`
	DstPort     string `json:"dstPort"`
	Strategy    string `gorm:"not null" json:"strategy"`
	Description string `json:"description"`
}
