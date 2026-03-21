package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PdnsClient struct {
	apiURL string
	apiKey string
	client *http.Client
}

type PdnsZone struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Kind           string   `json:"kind"`
	DNSSec         bool     `json:"dnssec"`
	Serial         int64    `json:"serial"`
	NotifiedSerial int64    `json:"notified_serial"`
	Masters        []string `json:"masters"`
	RRSets         []RRSet  `json:"rrsets"`
}

type RRSet struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	TTL        int       `json:"ttl"`
	ChangeType string    `json:"changetype,omitempty"`
	Records    []Record  `json:"records"`
	Comments   []Comment `json:"comments,omitempty"`
}

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type Comment struct {
	Content    string `json:"content"`
	Account    string `json:"account"`
	ModifiedAt int64  `json:"modified_at"`
}

type PdnsZoneCreate struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Masters     []string `json:"masters,omitempty"`
	Nameservers []string `json:"nameservers"`
	SOAEdit     string   `json:"soa_edit"`
	SOAEditAPI  string   `json:"soa_edit_api"`
	RRSets      []RRSet  `json:"rrsets,omitempty"`
}

type PdnsError struct {
	Error string `json:"error"`
}

func NewPdnsClient(apiURL, apiKey string) *PdnsClient {
	return &PdnsClient{
		apiURL: strings.TrimRight(apiURL, "/"),
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *PdnsClient) baseURL() string {
	return fmt.Sprintf("%s/api/v1/servers/localhost", c.apiURL)
}

func (c *PdnsClient) doRequest(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL()+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

func (c *PdnsClient) ListZones() ([]PdnsZone, error) {
	body, status, err := c.doRequest("GET", "/zones", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("PowerDNS API error (status %d): %s", status, string(body))
	}
	var zones []PdnsZone
	if err := json.Unmarshal(body, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (c *PdnsClient) GetZone(zoneID string) (*PdnsZone, error) {
	body, status, err := c.doRequest("GET", fmt.Sprintf("/zones/%s", zoneID), nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("PowerDNS API error (status %d): %s", status, string(body))
	}
	var zone PdnsZone
	if err := json.Unmarshal(body, &zone); err != nil {
		return nil, err
	}
	return &zone, nil
}

func (c *PdnsClient) CreateZone(req PdnsZoneCreate) error {
	body, status, err := c.doRequest("POST", "/zones", req)
	if err != nil {
		return err
	}
	if status != http.StatusCreated {
		var pdnsErr PdnsError
		_ = json.Unmarshal(body, &pdnsErr)
		return fmt.Errorf("PowerDNS create zone failed (status %d): %s", status, pdnsErr.Error)
	}
	return nil
}

func (c *PdnsClient) DeleteZone(zoneID string) error {
	_, status, err := c.doRequest("DELETE", fmt.Sprintf("/zones/%s", zoneID), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("PowerDNS delete zone failed (status %d)", status)
	}
	return nil
}

func (c *PdnsClient) PatchRecords(zoneID string, rrsets []RRSet) error {
	payload := map[string]interface{}{
		"rrsets": rrsets,
	}
	body, status, err := c.doRequest("PATCH", fmt.Sprintf("/zones/%s", zoneID), payload)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		var pdnsErr PdnsError
		_ = json.Unmarshal(body, &pdnsErr)
		return fmt.Errorf("PowerDNS patch records failed (status %d): %s", status, pdnsErr.Error)
	}
	return nil
}

func (c *PdnsClient) GetVersion() (string, error) {
	body, status, err := c.doRequest("GET", "", nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("PowerDNS API error (status %d)", status)
	}
	var server struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &server); err != nil {
		return "", err
	}
	return server.Version, nil
}

// canonicalName ensures a domain name ends with a dot (FQDN)
func CanonicalName(name string) string {
	if !strings.HasSuffix(name, ".") {
		return name + "."
	}
	return name
}
