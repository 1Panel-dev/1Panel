package dns

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	route53Host    = "route53.amazonaws.com"
	route53Region  = "us-east-1" // Route 53 is a global service signed with us-east-1
	route53Service = "route53"
	route53Version = "2013-04-01"
	route53NS      = "https://route53.amazonaws.com/doc/2013-04-01/"
)

type route53Client struct {
	accessKey string
	secretKey string
}

func newRoute53(auth map[string]string) *route53Client {
	return &route53Client{accessKey: auth["accessKey"], secretKey: auth["secretKey"]}
}

type r53Error struct {
	Error struct {
		Code    string `xml:"Code"`
		Message string `xml:"Message"`
	} `xml:"Error"`
}

func (c *route53Client) call(method, path string, query map[string]string, payload []byte, out interface{}) error {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	date := now.Format("20060102")

	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, aliEscape(k)+"="+aliEscape(query[k]))
	}
	canonicalQuery := strings.Join(pairs, "&")

	payloadHash := sha256Hex(string(payload))
	canonicalHeaders := "host:" + route53Host + "\n" + "x-amz-date:" + amzDate + "\n"
	signedHeaders := "host;x-amz-date"
	canonicalRequest := method + "\n" + path + "\n" + canonicalQuery + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHash
	scope := date + "/" + route53Region + "/" + route53Service + "/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + sha256Hex(canonicalRequest)
	kDate := hmacSHA256([]byte("AWS4"+c.secretKey), date)
	kRegion := hmacSHA256(kDate, route53Region)
	kService := hmacSHA256(kRegion, route53Service)
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))
	authorization := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.accessKey, scope, signedHeaders, signature)

	target := "https://" + route53Host + path
	if canonicalQuery != "" {
		target += "?" + canonicalQuery
	}
	req, err := http.NewRequest(method, target, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Host", route53Host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Authorization", authorization)
	if len(payload) > 0 {
		req.Header.Set("Content-Type", "application/xml")
	}
	body, status, err := doRequest(req)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		var e r53Error
		if xml.Unmarshal(body, &e) == nil && e.Error.Code != "" {
			return fmt.Errorf("route53: %s: %s", e.Error.Code, e.Error.Message)
		}
		return fmt.Errorf("route53: status %d: %s", status, truncate(string(body), 300))
	}
	if out != nil {
		return xml.Unmarshal(body, out)
	}
	return nil
}

type r53Zones struct {
	HostedZones []struct {
		ID   string `xml:"Id"`
		Name string `xml:"Name"`
	} `xml:"HostedZones>HostedZone"`
	IsTruncated bool   `xml:"IsTruncated"`
	NextMarker  string `xml:"NextMarker"`
}

func (c *route53Client) ListZones() ([]Zone, error) {
	var zones []Zone
	marker := ""
	for {
		query := map[string]string{"maxitems": "100"}
		if marker != "" {
			query["marker"] = marker
		}
		var resp r53Zones
		if err := c.call("GET", "/"+route53Version+"/hostedzone", query, nil, &resp); err != nil {
			return nil, err
		}
		for _, z := range resp.HostedZones {
			zones = append(zones, Zone{
				ID:   strings.TrimPrefix(z.ID, "/hostedzone/"),
				Name: strings.TrimSuffix(z.Name, "."),
			})
		}
		if !resp.IsTruncated || resp.NextMarker == "" {
			break
		}
		marker = resp.NextMarker
	}
	return zones, nil
}

type r53RecordSet struct {
	Name            string   `xml:"Name"`
	Type            string   `xml:"Type"`
	TTL             int      `xml:"TTL"`
	Values          []string `xml:"ResourceRecords>ResourceRecord>Value"`
	AliasTargetName string   `xml:"AliasTarget>DNSName"`
}

type r53RecordSets struct {
	Sets           []r53RecordSet `xml:"ResourceRecordSets>ResourceRecordSet"`
	IsTruncated    bool           `xml:"IsTruncated"`
	NextRecordName string         `xml:"NextRecordName"`
	NextRecordType string         `xml:"NextRecordType"`
}

func r53Unescape(name string) string {
	return strings.ReplaceAll(name, `\052`, "*")
}

func (c *route53Client) listSets(zone Zone, query map[string]string) ([]r53RecordSet, error) {
	var sets []r53RecordSet
	next := map[string]string{"maxitems": "300"}
	for k, v := range query {
		next[k] = v
	}
	for {
		var resp r53RecordSets
		path := "/" + route53Version + "/hostedzone/" + url.PathEscape(zone.ID) + "/rrset"
		if err := c.call("GET", path, next, nil, &resp); err != nil {
			return nil, err
		}
		sets = append(sets, resp.Sets...)
		// A lookup by name only needs the first page.
		if !resp.IsTruncated || query["name"] != "" {
			break
		}
		next = map[string]string{
			"maxitems": "300",
			"name":     resp.NextRecordName,
			"type":     resp.NextRecordType,
		}
	}
	return sets, nil
}

func r53ID(fqdn, recordType string) string {
	return recordType + "|" + strings.TrimSuffix(strings.ToLower(fqdn), ".")
}

func (c *route53Client) ListRecords(zone Zone) ([]Record, error) {
	sets, err := c.listSets(zone, nil)
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(sets))
	for _, s := range sets {
		if s.AliasTargetName != "" {
			continue // alias records are not editable here
		}
		name := r53Unescape(s.Name)
		records = append(records, Record{
			ID:    r53ID(name, s.Type),
			Name:  toRelative(name, zone.Name),
			Type:  s.Type,
			Value: strings.Join(s.Values, "\n"),
			TTL:   s.TTL,
		})
	}
	return records, nil
}

type r53Change struct {
	Action    string       `xml:"Action"`
	RecordSet r53SetToSend `xml:"ResourceRecordSet"`
}

type r53SetToSend struct {
	Name    string      `xml:"Name"`
	Type    string      `xml:"Type"`
	TTL     int         `xml:"TTL"`
	Records []r53Record `xml:"ResourceRecords>ResourceRecord"`
}

type r53Record struct {
	Value string `xml:"Value"`
}

type r53ChangeRequest struct {
	XMLName xml.Name    `xml:"https://route53.amazonaws.com/doc/2013-04-01/ ChangeResourceRecordSetsRequest"`
	Changes []r53Change `xml:"ChangeBatch>Changes>Change"`
}

func r53Values(r Record) []r53Record {
	var out []r53Record
	for _, line := range strings.Split(r.Value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch r.Type {
		case "TXT", "SPF":
			if !strings.HasPrefix(line, `"`) {
				line = `"` + strings.ReplaceAll(line, `"`, `\"`) + `"`
			}
		case "MX":
			if r.Priority > 0 {
				if _, err := strconv.Atoi(strings.Fields(line)[0]); err != nil {
					line = strconv.Itoa(r.Priority) + " " + line
				}
			}
		}
		out = append(out, r53Record{Value: line})
	}
	return out
}

func r53SetFrom(zone Zone, r Record) r53SetToSend {
	ttl := r.TTL
	if ttl <= 0 {
		ttl = 300
	}
	return r53SetToSend{
		Name:    toFQDN(r.Name, zone.Name),
		Type:    r.Type,
		TTL:     ttl,
		Records: r53Values(r),
	}
}

func (c *route53Client) change(zone Zone, changes []r53Change) error {
	data, err := xml.Marshal(r53ChangeRequest{Changes: changes})
	if err != nil {
		return err
	}
	payload := append([]byte(xml.Header), data...)
	path := "/" + route53Version + "/hostedzone/" + url.PathEscape(zone.ID) + "/rrset"
	return c.call("POST", path, nil, payload, nil)
}

func (c *route53Client) CreateRecord(zone Zone, r Record) error {
	set := r53SetFrom(zone, r)
	if len(set.Records) == 0 {
		return fmt.Errorf("route53: record value is required")
	}
	return c.change(zone, []r53Change{{Action: "CREATE", RecordSet: set}})
}

func (c *route53Client) UpdateRecord(zone Zone, r Record) error {
	set := r53SetFrom(zone, r)
	if len(set.Records) == 0 {
		return fmt.Errorf("route53: record value is required")
	}
	newID := r53ID(set.Name, set.Type)
	if r.ID == "" || r.ID == newID {
		return c.change(zone, []r53Change{{Action: "UPSERT", RecordSet: set}})
	}
	// Name or type changed: delete the old record set and create the new one atomically.
	parts := strings.SplitN(r.ID, "|", 2)
	if len(parts) != 2 {
		return fmt.Errorf("route53: invalid record id %q", r.ID)
	}
	old, err := c.findSet(zone, parts[1], parts[0])
	if err != nil {
		return err
	}
	return c.change(zone, []r53Change{
		{Action: "DELETE", RecordSet: old},
		{Action: "CREATE", RecordSet: set},
	})
}

// findSet fetches an existing record set (route53 requires an exact match to delete it).
func (c *route53Client) findSet(zone Zone, fqdn, recordType string) (r53SetToSend, error) {
	sets, err := c.listSets(zone, map[string]string{"name": fqdn, "type": recordType, "maxitems": "10"})
	if err != nil {
		return r53SetToSend{}, err
	}
	want := strings.TrimSuffix(strings.ToLower(fqdn), ".")
	for _, s := range sets {
		if s.Type == recordType && strings.TrimSuffix(strings.ToLower(r53Unescape(s.Name)), ".") == want {
			out := r53SetToSend{Name: s.Name, Type: s.Type, TTL: s.TTL}
			for _, v := range s.Values {
				out.Records = append(out.Records, r53Record{Value: v})
			}
			return out, nil
		}
	}
	return r53SetToSend{}, fmt.Errorf("route53: record %s %s not found", recordType, fqdn)
}

func (c *route53Client) DeleteRecord(zone Zone, r Record) error {
	parts := strings.SplitN(r.ID, "|", 2)
	if len(parts) != 2 {
		return fmt.Errorf("route53: invalid record id %q", r.ID)
	}
	old, err := c.findSet(zone, parts[1], parts[0])
	if err != nil {
		return err
	}
	return c.change(zone, []r53Change{{Action: "DELETE", RecordSet: old}})
}
