package dto

import (
	"encoding/json"
	"time"
)

type APIKeyItem struct {
	ID                 string     `json:"id"`
	Kind               string     `json:"kind"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	KeyHint            string     `json:"keyHint"`
	Status             string     `json:"status"`
	IPWhiteList        string     `json:"ipWhiteList"`
	APITrustedProxies  string     `json:"apiTrustedProxies"`
	APIKeyValidityTime int        `json:"apiKeyValidityTime"`
	ExpiresAt          *time.Time `json:"expiresAt"`
	AllowAppBinding    bool       `json:"allowAppBinding"`
	Revision           uint64     `json:"revision"`
	CreatedAt          *time.Time `json:"createdAt"`
}

type APIKeySearch struct {
	Page           int  `json:"page"`
	PageSize       int  `json:"pageSize"`
	ExcludeRevoked bool `json:"excludeRevoked"`
}
type APIKeyPage struct {
	Items []APIKeyItem `json:"items"`
	Total int          `json:"total"`
}
type APIKeyFields struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	IPWhiteList        string          `json:"ipWhiteList"`
	APITrustedProxies  string          `json:"apiTrustedProxies"`
	APIKeyValidityTime int             `json:"apiKeyValidityTime"`
	ExpiresAt          json.RawMessage `json:"expiresAt"`
	AllowAppBinding    bool            `json:"allowAppBinding"`
}
type APIKeyCreate struct {
	RequestID string `json:"requestID"`
	APIKeyFields
}
type APIKeyCreated struct {
	Item           APIKeyItem `json:"item"`
	APIKey         string     `json:"apiKey"`
	AlreadyCreated bool       `json:"alreadyCreated,omitempty"`
}
type APIKeyMutation struct {
	ID       string `json:"id"`
	Revision uint64 `json:"revision"`
}
type APIKeyUpdate struct {
	APIKeyMutation
	APIKeyFields
}
type APIKeyStatus struct {
	APIKeyMutation
	Status string `json:"status"`
}
