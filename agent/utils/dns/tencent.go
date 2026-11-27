package dns

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	tencentHost    = "dnspod.tencentcloudapi.com"
	tencentService = "dnspod"
	tencentVersion = "2021-03-23"
	tencentLine    = "默认"
)

type tencentClient struct {
	secretID  string
	secretKey string
}

func newTencent(auth map[string]string) *tencentClient {
	return &tencentClient{secretID: auth["secretID"], secretKey: auth["secretKey"]}
}

func hmacSHA256(key []byte, msg string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return m.Sum(nil)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

type tencentError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

// call invokes a DNSPod API action. out receives the "Response" object.
func (c *tencentClient) call(action string, payload map[string]interface{}, out interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	timestamp := strconv.FormatInt(now.Unix(), 10)
	date := now.Format("2006-01-02")
	contentType := "application/json; charset=utf-8"

	canonicalHeaders := "content-type:" + contentType + "\n" +
		"host:" + tencentHost + "\n" +
		"x-tc-action:" + strings.ToLower(action) + "\n"
	signedHeaders := "content-type;host;x-tc-action"
	canonicalRequest := "POST\n/\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + sha256Hex(string(data))
	scope := date + "/" + tencentService + "/tc3_request"
	stringToSign := "TC3-HMAC-SHA256\n" + timestamp + "\n" + scope + "\n" + sha256Hex(canonicalRequest)
	secretDate := hmacSHA256([]byte("TC3"+c.secretKey), date)
	secretService := hmacSHA256(secretDate, tencentService)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))
	authorization := fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.secretID, scope, signedHeaders, signature)

	req, err := http.NewRequest("POST", "https://"+tencentHost, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Host", tencentHost)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Timestamp", timestamp)
	req.Header.Set("X-TC-Version", tencentVersion)

	body, status, err := doRequest(req)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("tencentcloud: status %d: %s", status, truncate(string(body), 300))
	}
	var envelope struct {
		Response json.RawMessage `json:"Response"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	var errResp struct {
		Error *tencentError `json:"Error"`
	}
	if err := json.Unmarshal(envelope.Response, &errResp); err == nil && errResp.Error != nil {
		return &tencentAPIError{Code: errResp.Error.Code, Message: errResp.Error.Message}
	}
	if out != nil {
		return json.Unmarshal(envelope.Response, out)
	}
	return nil
}

type tencentAPIError struct {
	Code    string
	Message string
}

func (e *tencentAPIError) Error() string {
	return fmt.Sprintf("tencentcloud: %s: %s", e.Code, e.Message)
}

func (c *tencentClient) ListZones() ([]Zone, error) {
	var zones []Zone
	const limit = 500
	for offset := 0; ; offset += limit {
		var resp struct {
			DomainList []struct {
				DomainID uint64 `json:"DomainId"`
				Name     string `json:"Name"`
			} `json:"DomainList"`
		}
		err := c.call("DescribeDomainList", map[string]interface{}{"Offset": offset, "Limit": limit}, &resp)
		if err != nil {
			if apiErr, ok := err.(*tencentAPIError); ok && strings.HasPrefix(apiErr.Code, "ResourceNotFound") {
				break
			}
			return nil, err
		}
		for _, d := range resp.DomainList {
			zones = append(zones, Zone{ID: strconv.FormatUint(d.DomainID, 10), Name: d.Name})
		}
		if len(resp.DomainList) < limit {
			break
		}
	}
	return zones, nil
}

func (c *tencentClient) ListRecords(zone Zone) ([]Record, error) {
	var records []Record
	const limit = 1000
	for offset := 0; ; offset += limit {
		var resp struct {
			RecordList []struct {
				RecordID uint64 `json:"RecordId"`
				Name     string `json:"Name"`
				Type     string `json:"Type"`
				Value    string `json:"Value"`
				TTL      uint64 `json:"TTL"`
				MX       uint64 `json:"MX"`
				Remark   string `json:"Remark"`
			} `json:"RecordList"`
		}
		err := c.call("DescribeRecordList", map[string]interface{}{
			"Domain": zone.Name,
			"Offset": offset,
			"Limit":  limit,
		}, &resp)
		if err != nil {
			// A zone without records is reported as an error by the API.
			if apiErr, ok := err.(*tencentAPIError); ok && strings.HasPrefix(apiErr.Code, "ResourceNotFound") {
				break
			}
			return nil, err
		}
		for _, r := range resp.RecordList {
			records = append(records, Record{
				ID:       strconv.FormatUint(r.RecordID, 10),
				Name:     r.Name,
				Type:     r.Type,
				Value:    r.Value,
				TTL:      int(r.TTL),
				Priority: int(r.MX),
				Remark:   r.Remark,
			})
		}
		if len(resp.RecordList) < limit {
			break
		}
	}
	return records, nil
}

func (c *tencentClient) recordPayload(zone Zone, r Record) map[string]interface{} {
	ttl := r.TTL
	if ttl <= 0 {
		ttl = 600
	}
	host := r.Name
	if host == "" {
		host = "@"
	}
	p := map[string]interface{}{
		"Domain":     zone.Name,
		"SubDomain":  host,
		"RecordType": r.Type,
		"RecordLine": tencentLine,
		"Value":      r.Value,
		"TTL":        ttl,
	}
	if r.Type == "MX" {
		prio := r.Priority
		if prio <= 0 {
			prio = 10
		}
		p["MX"] = prio
	}
	return p
}

func parseRecordID(id string) (uint64, error) {
	v, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid record id %q", id)
	}
	return v, nil
}

func (c *tencentClient) CreateRecord(zone Zone, r Record) error {
	return c.call("CreateRecord", c.recordPayload(zone, r), nil)
}

func (c *tencentClient) UpdateRecord(zone Zone, r Record) error {
	id, err := parseRecordID(r.ID)
	if err != nil {
		return err
	}
	p := c.recordPayload(zone, r)
	p["RecordId"] = id
	return c.call("ModifyRecord", p, nil)
}

func (c *tencentClient) DeleteRecord(zone Zone, r Record) error {
	id, err := parseRecordID(r.ID)
	if err != nil {
		return err
	}
	return c.call("DeleteRecord", map[string]interface{}{"Domain": zone.Name, "RecordId": id}, nil)
}
