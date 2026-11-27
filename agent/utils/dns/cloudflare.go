package dns

import (
	"fmt"
	"net/url"
	"strings"
)

const cloudflareAPI = "https://api.cloudflare.com/client/v4"

type cloudflareClient struct {
	token string
}

func newCloudflare(auth map[string]string) *cloudflareClient {
	return &cloudflareClient{token: auth["apiKey"]}
}

type cfResultInfo struct {
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cfRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Proxied  bool   `json:"proxied"`
	Priority int    `json:"priority"`
	Comment  string `json:"comment"`
}

type cfZonesResp struct {
	Success    bool         `json:"success"`
	Errors     []cfError    `json:"errors"`
	Result     []cfZone     `json:"result"`
	ResultInfo cfResultInfo `json:"result_info"`
}

type cfRecordsResp struct {
	Success    bool         `json:"success"`
	Errors     []cfError    `json:"errors"`
	Result     []cfRecord   `json:"result"`
	ResultInfo cfResultInfo `json:"result_info"`
}

type cfBaseResp struct {
	Success bool      `json:"success"`
	Errors  []cfError `json:"errors"`
}

func (c *cloudflareClient) headers() map[string]string {
	return map[string]string{"Authorization": "Bearer " + c.token}
}

func cfErrors(errs []cfError) error {
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("%d: %s", e.Code, e.Message))
	}
	return fmt.Errorf("cloudflare: %s", strings.Join(msgs, "; "))
}

func (c *cloudflareClient) ListZones() ([]Zone, error) {
	var zones []Zone
	for page := 1; ; page++ {
		var resp cfZonesResp
		u := fmt.Sprintf("%s/zones?per_page=50&page=%d", cloudflareAPI, page)
		if err := doJSON("GET", u, c.headers(), nil, &resp); err != nil {
			return nil, err
		}
		if !resp.Success {
			return nil, cfErrors(resp.Errors)
		}
		for _, z := range resp.Result {
			zones = append(zones, Zone{ID: z.ID, Name: z.Name})
		}
		if page >= resp.ResultInfo.TotalPages {
			break
		}
	}
	return zones, nil
}

func (c *cloudflareClient) ListRecords(zone Zone) ([]Record, error) {
	var records []Record
	for page := 1; ; page++ {
		var resp cfRecordsResp
		u := fmt.Sprintf("%s/zones/%s/dns_records?per_page=100&page=%d", cloudflareAPI, url.PathEscape(zone.ID), page)
		if err := doJSON("GET", u, c.headers(), nil, &resp); err != nil {
			return nil, err
		}
		if !resp.Success {
			return nil, cfErrors(resp.Errors)
		}
		for _, r := range resp.Result {
			records = append(records, Record{
				ID:       r.ID,
				Name:     toRelative(r.Name, zone.Name),
				Type:     r.Type,
				Value:    r.Content,
				TTL:      r.TTL,
				Priority: r.Priority,
				Proxied:  r.Proxied,
				Remark:   r.Comment,
			})
		}
		if page >= resp.ResultInfo.TotalPages {
			break
		}
	}
	return records, nil
}

func (c *cloudflareClient) payload(zone Zone, r Record) map[string]interface{} {
	ttl := r.TTL
	if ttl <= 0 {
		ttl = 1 // automatic
	}
	body := map[string]interface{}{
		"type":    r.Type,
		"name":    toFQDN(r.Name, zone.Name),
		"content": r.Value,
		"ttl":     ttl,
		"comment": r.Remark,
	}
	switch r.Type {
	case "A", "AAAA", "CNAME":
		body["proxied"] = r.Proxied
	case "MX":
		body["priority"] = r.Priority
	}
	return body
}

func (c *cloudflareClient) CreateRecord(zone Zone, r Record) error {
	var resp cfBaseResp
	u := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPI, url.PathEscape(zone.ID))
	if err := doJSON("POST", u, c.headers(), c.payload(zone, r), &resp); err != nil {
		return err
	}
	if !resp.Success {
		return cfErrors(resp.Errors)
	}
	return nil
}

func (c *cloudflareClient) UpdateRecord(zone Zone, r Record) error {
	var resp cfBaseResp
	u := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPI, url.PathEscape(zone.ID), url.PathEscape(r.ID))
	if err := doJSON("PUT", u, c.headers(), c.payload(zone, r), &resp); err != nil {
		return err
	}
	if !resp.Success {
		return cfErrors(resp.Errors)
	}
	return nil
}

func (c *cloudflareClient) DeleteRecord(zone Zone, r Record) error {
	var resp cfBaseResp
	u := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPI, url.PathEscape(zone.ID), url.PathEscape(r.ID))
	if err := doJSON("DELETE", u, c.headers(), nil, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return cfErrors(resp.Errors)
	}
	return nil
}
