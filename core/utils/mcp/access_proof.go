package mcp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/netip"
	"strings"
	"sync"
	"time"
)

const AccessProofHeader = "X-Panel-MCP-Access"

var accessProofKey struct {
	sync.Once
	value []byte
	err   error
}

type accessProof struct {
	IP      string `json:"ip"`
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

func proofKey() ([]byte, error) {
	accessProofKey.Do(func() {
		accessProofKey.value = make([]byte, 32)
		_, accessProofKey.err = rand.Read(accessProofKey.value)
	})
	return accessProofKey.value, accessProofKey.err
}

func proofToken(token string) string {
	value := sha256.Sum256([]byte(token))
	return hex.EncodeToString(value[:])
}

func IssueAccessProof(token, source string, now time.Time) (string, error) {
	address, err := netip.ParseAddr(source)
	if err != nil || address.Zone() != "" || token == "" {
		return "", errors.New("invalid MCP source proof")
	}
	key, err := proofKey()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(accessProof{IP: address.Unmap().String(), Token: proofToken(token), Expires: now.Add(400 * time.Second).Unix()})
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return base64.RawURLEncoding.EncodeToString(data) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func VerifyAccessProof(value, token string, now time.Time) (string, error) {
	if len(value) > 512 {
		return "", Forbidden()
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", Forbidden()
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", Forbidden()
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", Forbidden()
	}
	key, err := proofKey()
	if err != nil {
		return "", Forbidden()
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return "", Forbidden()
	}
	var proof accessProof
	if json.Unmarshal(data, &proof) != nil || !hmac.Equal([]byte(proof.Token), []byte(proofToken(token))) || now.Unix() >= proof.Expires || proof.Expires > now.Add(401*time.Second).Unix() {
		return "", Forbidden()
	}
	address, err := netip.ParseAddr(proof.IP)
	if err != nil || address.Zone() != "" {
		return "", Forbidden()
	}
	return address.Unmap().String(), nil
}
