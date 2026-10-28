package mcp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
)

const (
	WebsiteProxyKeyHeader  = "X-Panel-MCP-Website-Key"
	WebsiteProxyPeerHeader = "X-Panel-MCP-Website-Peer"
)

func WebsiteProxyKey(masterKey, bindingID string) (string, error) {
	id, err := uuid.Parse(bindingID)
	if masterKey == "" || err != nil || id == uuid.Nil || id.String() != bindingID {
		return "", errors.New("MCP website proxy identity is unavailable")
	}
	mac := hmac.New(sha256.New, []byte(masterKey))
	mac.Write([]byte("1panel-mcp-website:v1:" + bindingID))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
