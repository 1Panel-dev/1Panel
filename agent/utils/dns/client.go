// Package dns manages DNS records at third-party DNS providers.
// Clients use only the standard library (REST + request signing) so that no
// extra SDK dependencies are required.
package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
)

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Record struct {
	ID       string `json:"id"`
	Name     string `json:"name"` // host relative to the zone, "@" for the apex
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
	Proxied  bool   `json:"proxied"`
	Remark   string `json:"remark"`
}

type Client interface {
	ListZones() ([]Zone, error)
	ListRecords(zone Zone) ([]Record, error)
	CreateRecord(zone Zone, record Record) error
	UpdateRecord(zone Zone, record Record) error
	DeleteRecord(zone Zone, record Record) error
}

// SupportedTypes lists provider types that support record management.
var SupportedTypes = []string{"CloudFlare", "AliYun", "AWSRoute53", "TencentCloud"}

func IsSupported(dnsType string) bool {
	for _, t := range SupportedTypes {
		if t == dnsType {
			return true
		}
	}
	return false
}

// NewClient builds a record client from the stored account authorization map.
func NewClient(dnsType string, auth map[string]string) (Client, error) {
	switch dnsType {
	case "CloudFlare":
		return newCloudflare(auth), nil
	case "AliYun":
		return newAliDNS(auth), nil
	case "AWSRoute53":
		return newRoute53(auth), nil
	case "TencentCloud":
		return newTencent(auth), nil
	default:
		return nil, buserr.New("ErrDnsRecordUnsupported")
	}
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func doJSON(method, url string, headers map[string]string, payload interface{}, out interface{}) error {
	var reader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	body, status, err := doRequest(req)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("request failed with status %d: %s", status, truncate(string(body), 300))
	}
	if out != nil && len(body) > 0 {
		return json.Unmarshal(body, out)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// toRelative converts an FQDN into a host relative to the zone ("@" for apex).
func toRelative(fqdn, zone string) string {
	fqdn = strings.TrimSuffix(strings.ToLower(fqdn), ".")
	zone = strings.TrimSuffix(strings.ToLower(zone), ".")
	if fqdn == zone {
		return "@"
	}
	return strings.TrimSuffix(fqdn, "."+zone)
}

// toFQDN converts a relative host into an FQDN inside the zone.
func toFQDN(host, zone string) string {
	host = strings.TrimSuffix(host, ".")
	zone = strings.TrimSuffix(zone, ".")
	if host == "" || host == "@" {
		return zone
	}
	if host == zone || strings.HasSuffix(host, "."+zone) {
		return host
	}
	return host + "." + zone
}
