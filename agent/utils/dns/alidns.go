package dns

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const aliDNSEndpoint = "https://alidns.aliyuncs.com/"

type aliDNSClient struct {
	accessKey string
	secretKey string
}

func newAliDNS(auth map[string]string) *aliDNSClient {
	return &aliDNSClient{accessKey: auth["accessKey"], secretKey: auth["secretKey"]}
}

func aliEscape(s string) string {
	s = url.QueryEscape(s)
	s = strings.ReplaceAll(s, "+", "%20")
	s = strings.ReplaceAll(s, "*", "%2A")
	s = strings.ReplaceAll(s, "%7E", "~")
	return s
}

func (c *aliDNSClient) call(action string, params map[string]string, out interface{}) error {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	all := map[string]string{
		"Action":           action,
		"Format":           "JSON",
		"Version":          "2015-01-09",
		"AccessKeyId":      c.accessKey,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   hex.EncodeToString(nonce),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	for k, v := range params {
		all[k] = v
	}
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, aliEscape(k)+"="+aliEscape(all[k]))
	}
	canonical := strings.Join(pairs, "&")
	stringToSign := "POST&" + aliEscape("/") + "&" + aliEscape(canonical)
	mac := hmac.New(sha1.New, []byte(c.secretKey+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	body := canonical + "&Signature=" + aliEscape(signature)

	req, err := http.NewRequest("POST", aliDNSEndpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respBody, status, err := doRequest(req)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		var e struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		}
		if json.Unmarshal(respBody, &e) == nil && e.Code != "" {
			return fmt.Errorf("alidns: %s: %s", e.Code, e.Message)
		}
		return fmt.Errorf("alidns: status %d: %s", status, truncate(string(respBody), 300))
	}
	if out != nil {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func (c *aliDNSClient) ListZones() ([]Zone, error) {
	var zones []Zone
	const pageSize = 100
	for page := 1; ; page++ {
		var resp struct {
			TotalCount int64 `json:"TotalCount"`
			Domains    struct {
				Domain []struct {
					DomainID   string `json:"DomainId"`
					DomainName string `json:"DomainName"`
				} `json:"Domain"`
			} `json:"Domains"`
		}
		err := c.call("DescribeDomains", map[string]string{
			"PageNumber": strconv.Itoa(page),
			"PageSize":   strconv.Itoa(pageSize),
		}, &resp)
		if err != nil {
			return nil, err
		}
		for _, d := range resp.Domains.Domain {
			zones = append(zones, Zone{ID: d.DomainID, Name: d.DomainName})
		}
		if len(resp.Domains.Domain) == 0 || int64(len(zones)) >= resp.TotalCount {
			break
		}
	}
	return zones, nil
}

func (c *aliDNSClient) ListRecords(zone Zone) ([]Record, error) {
	var records []Record
	const pageSize = 500
	for page := 1; ; page++ {
		var resp struct {
			TotalCount    int64 `json:"TotalCount"`
			DomainRecords struct {
				Record []struct {
					RecordID string `json:"RecordId"`
					RR       string `json:"RR"`
					Type     string `json:"Type"`
					Value    string `json:"Value"`
					TTL      int64  `json:"TTL"`
					Priority int64  `json:"Priority"`
					Remark   string `json:"Remark"`
				} `json:"Record"`
			} `json:"DomainRecords"`
		}
		err := c.call("DescribeDomainRecords", map[string]string{
			"DomainName": zone.Name,
			"PageNumber": strconv.Itoa(page),
			"PageSize":   strconv.Itoa(pageSize),
		}, &resp)
		if err != nil {
			return nil, err
		}
		for _, r := range resp.DomainRecords.Record {
			records = append(records, Record{
				ID:       r.RecordID,
				Name:     r.RR,
				Type:     r.Type,
				Value:    r.Value,
				TTL:      int(r.TTL),
				Priority: int(r.Priority),
				Remark:   r.Remark,
			})
		}
		if len(resp.DomainRecords.Record) == 0 || int64(len(records)) >= resp.TotalCount {
			break
		}
	}
	return records, nil
}

func (c *aliDNSClient) recordParams(r Record) map[string]string {
	ttl := r.TTL
	if ttl <= 0 {
		ttl = 600
	}
	host := r.Name
	if host == "" {
		host = "@"
	}
	p := map[string]string{
		"RR":    host,
		"Type":  r.Type,
		"Value": r.Value,
		"TTL":   strconv.Itoa(ttl),
	}
	if r.Type == "MX" {
		prio := r.Priority
		if prio <= 0 {
			prio = 10
		}
		p["Priority"] = strconv.Itoa(prio)
	}
	return p
}

func (c *aliDNSClient) CreateRecord(zone Zone, r Record) error {
	p := c.recordParams(r)
	p["DomainName"] = zone.Name
	return c.call("AddDomainRecord", p, nil)
}

func (c *aliDNSClient) UpdateRecord(zone Zone, r Record) error {
	p := c.recordParams(r)
	p["RecordId"] = r.ID
	return c.call("UpdateDomainRecord", p, nil)
}

func (c *aliDNSClient) DeleteRecord(zone Zone, r Record) error {
	return c.call("DeleteDomainRecord", map[string]string{"RecordId": r.ID}, nil)
}
