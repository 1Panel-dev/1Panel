package model

import "strings"

type FirewallRule struct {
	UUID     string `gorm:"size:64;primaryKey" json:"uuid"`
	ScopeKey string `gorm:"size:255;not null;index" json:"scopeKey"`
	Provider string `gorm:"size:32;not null" json:"provider"`
	Family   string `gorm:"size:16;not null" json:"family"`
	Location string `gorm:"size:128;not null" json:"location"`

	NativeKind         string `gorm:"size:64;not null" json:"nativeKind"`
	Protocol           string `gorm:"size:32;not null" json:"protocol"`
	SourceAddress      string `gorm:"size:255" json:"sourceAddress"`
	SourcePort         string `gorm:"size:64" json:"sourcePort"`
	DestinationAddress string `gorm:"size:255" json:"destinationAddress"`
	DestinationPort    string `gorm:"size:64" json:"destinationPort"`
	Interface          string `gorm:"size:128" json:"interface"`
	ConnectionStates   string `gorm:"type:text" json:"connectionStates"`
	Action             string `gorm:"size:32;not null" json:"action"`
	Priority           *int   `json:"priority"`
	OrderIndex         *int64 `json:"orderIndex"`
	OrderBucket        string `gorm:"size:64" json:"orderBucket"`
	Description        string `gorm:"type:text" json:"description"`

	RuleKey  string `gorm:"size:80;not null" json:"ruleKey"`
	Origin   string `gorm:"size:32;not null" json:"origin"`
	Owner    string `gorm:"size:320;not null" json:"owner"`
	MatchKey string `gorm:"size:320" json:"matchKey"`
	Revision uint   `gorm:"not null;default:1" json:"revision"`
}

func FirewallRuleOwner(sourceKind, sourceID string) string {
	sourceKind = strings.TrimSpace(sourceKind)
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		return sourceKind
	}
	return sourceKind + ":" + sourceID
}
