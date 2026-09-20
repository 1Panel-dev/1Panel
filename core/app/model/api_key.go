package model

import "time"

type APIKey struct {
	ID                 string     `gorm:"type:varchar(36);primaryKey" json:"-"`
	OwnerType          string     `gorm:"not null;uniqueIndex:idx_api_key_request;uniqueIndex:idx_api_key_active_name;index:idx_api_key_owner" json:"-"`
	OwnerID            string     `gorm:"not null;uniqueIndex:idx_api_key_request;uniqueIndex:idx_api_key_active_name;index:idx_api_key_owner" json:"-"`
	Name               string     `gorm:"not null" json:"-"`
	ActiveName         *string    `gorm:"uniqueIndex:idx_api_key_active_name" json:"-"`
	Description        string     `json:"-"`
	RequestID          string     `gorm:"not null;uniqueIndex:idx_api_key_request" json:"-"`
	RequestHash        string     `json:"-"`
	SecretCiphertext   string     `gorm:"type:text;not null" json:"-"`
	SecretVersion      int        `gorm:"not null" json:"-"`
	KeyHint            string     `json:"-"`
	Status             string     `gorm:"not null;index" json:"-"`
	IPWhiteList        string     `json:"-"`
	APITrustedProxies  string     `json:"-"`
	APIKeyValidityTime int        `json:"-"`
	ExpiresAt          *time.Time `gorm:"index" json:"-"`
	AllowAppBinding    bool       `json:"-"`
	Revision           uint64     `gorm:"not null" json:"-"`
	CreatedAt          time.Time  `json:"-"`
	UpdatedAt          time.Time  `json:"-"`
	RevokedAt          *time.Time `json:"-"`
}

type LegacyAPIKeyPolicy struct {
	OwnerType       string `gorm:"primaryKey"`
	OwnerID         string `gorm:"primaryKey"`
	AllowAppBinding bool
	Revision        uint64
}
