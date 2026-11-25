package dns

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubHTTP replaces the package HTTP client for the duration of a test.
func stubHTTP(t *testing.T, fn roundTripFunc) {
	t.Helper()
	old := httpClient
	httpClient = &http.Client{Transport: fn}
	t.Cleanup(func() { httpClient = old })
}

func jsonResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestToRelative(t *testing.T) {
	cases := []struct{ fqdn, zone, want string }{
		{"example.com", "example.com", "@"},
		{"example.com.", "example.com.", "@"},
		{"WWW.Example.com", "example.com", "www"},
		{"a.b.example.com.", "example.com", "a.b"},
	}
	for _, c := range cases {
		if got := toRelative(c.fqdn, c.zone); got != c.want {
			t.Errorf("toRelative(%q, %q) = %q, want %q", c.fqdn, c.zone, got, c.want)
		}
	}
}

func TestToFQDN(t *testing.T) {
	cases := []struct{ host, zone, want string }{
		{"@", "example.com", "example.com"},
		{"", "example.com", "example.com"},
		{"www", "example.com", "www.example.com"},
		{"www.example.com", "example.com", "www.example.com"},
		{"www.", "example.com.", "www.example.com"},
		{"example.com", "example.com", "example.com"},
	}
	for _, c := range cases {
		if got := toFQDN(c.host, c.zone); got != c.want {
			t.Errorf("toFQDN(%q, %q) = %q, want %q", c.host, c.zone, got, c.want)
		}
	}
}

func TestIsSupportedAndNewClient(t *testing.T) {
	for _, typ := range SupportedTypes {
		if !IsSupported(typ) {
			t.Errorf("IsSupported(%q) = false", typ)
		}
		c, err := NewClient(typ, map[string]string{})
		if err != nil || c == nil {
			t.Errorf("NewClient(%q) = %v, %v", typ, c, err)
		}
	}
	if IsSupported("GoDaddy") {
		t.Error("IsSupported(GoDaddy) = true")
	}
	if c, err := NewClient("GoDaddy", nil); err == nil || c != nil {
		t.Errorf("NewClient(GoDaddy) = %v, %v; want unsupported error", c, err)
	}
}

func TestCloudflarePayload(t *testing.T) {
	c := newCloudflare(map[string]string{"apiKey": "tok"})
	zone := Zone{ID: "z1", Name: "example.com"}

	a := c.payload(zone, Record{Name: "www", Type: "A", Value: "1.2.3.4", Proxied: true})
	if a["name"] != "www.example.com" || a["ttl"] != 1 || a["proxied"] != true {
		t.Errorf("unexpected A payload: %v", a)
	}
	if _, ok := a["priority"]; ok {
		t.Errorf("A payload must not carry priority: %v", a)
	}

	mx := c.payload(zone, Record{Name: "@", Type: "MX", Value: "mail.example.com", Priority: 20, TTL: 300})
	if mx["name"] != "example.com" || mx["priority"] != 20 || mx["ttl"] != 300 {
		t.Errorf("unexpected MX payload: %v", mx)
	}
	if _, ok := mx["proxied"]; ok {
		t.Errorf("MX payload must not carry proxied: %v", mx)
	}

	txt := c.payload(zone, Record{Name: "_acme", Type: "TXT", Value: "v"})
	if _, ok := txt["proxied"]; ok {
		t.Errorf("TXT payload must not carry proxied: %v", txt)
	}
}

func TestCloudflareListRecordsPaginates(t *testing.T) {
	calls := 0
	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		if r.URL.Path != "/client/v4/zones/z1/dns_records" {
			t.Errorf("path = %q", r.URL.Path)
		}
		switch r.URL.Query().Get("page") {
		case "1":
			return jsonResp(200, `{"success":true,"result":[{"id":"r1","type":"A","name":"www.example.com","content":"1.1.1.1","ttl":1,"proxied":true}],"result_info":{"page":1,"total_pages":2}}`), nil
		case "2":
			return jsonResp(200, `{"success":true,"result":[{"id":"r2","type":"MX","name":"example.com","content":"mail.example.com","ttl":300,"priority":10,"comment":"mail"}],"result_info":{"page":2,"total_pages":2}}`), nil
		}
		t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		return jsonResp(500, ""), nil
	})

	c := newCloudflare(map[string]string{"apiKey": "tok"})
	records, err := c.ListRecords(Zone{ID: "z1", Name: "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(records) != 2 {
		t.Fatalf("calls = %d, records = %d; want 2, 2", calls, len(records))
	}
	if records[0].Name != "www" || !records[0].Proxied || records[0].Value != "1.1.1.1" {
		t.Errorf("record 0 = %+v", records[0])
	}
	if records[1].Name != "@" || records[1].Priority != 10 || records[1].Remark != "mail" {
		t.Errorf("record 1 = %+v", records[1])
	}
}

func TestCloudflareErrors(t *testing.T) {
	c := newCloudflare(map[string]string{"apiKey": "tok"})
	zone := Zone{ID: "z1", Name: "example.com"}

	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		return jsonResp(200, `{"success":false,"errors":[{"code":1003,"message":"invalid"}]}`), nil
	})
	if err := c.CreateRecord(zone, Record{Name: "a", Type: "A", Value: "1.1.1.1"}); err == nil || !strings.Contains(err.Error(), "1003: invalid") {
		t.Errorf("API error not surfaced: %v", err)
	}

	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		return jsonResp(403, `forbidden`), nil
	})
	if err := c.DeleteRecord(zone, Record{ID: "r1"}); err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("HTTP error not surfaced: %v", err)
	}
}

func TestCloudflareDeleteRecordRequest(t *testing.T) {
	stubHTTP(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodDelete || r.URL.Path != "/client/v4/zones/z1/dns_records/r1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return jsonResp(200, `{"success":true}`), nil
	})
	c := newCloudflare(map[string]string{"apiKey": "tok"})
	if err := c.DeleteRecord(Zone{ID: "z1", Name: "example.com"}, Record{ID: "r1"}); err != nil {
		t.Fatal(err)
	}
}

func TestRoute53Values(t *testing.T) {
	got := r53Values(Record{Type: "TXT", Value: "plain\n\n  \"quoted\"  \nhas \"inner\" quote"})
	want := []string{`"plain"`, `"quoted"`, `"has \"inner\" quote"`}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i].Value != want[i] {
			t.Errorf("TXT value %d = %q, want %q", i, got[i].Value, want[i])
		}
	}

	mx := r53Values(Record{Type: "MX", Priority: 10, Value: "mail.example.com\n20 backup.example.com"})
	if len(mx) != 2 || mx[0].Value != "10 mail.example.com" || mx[1].Value != "20 backup.example.com" {
		t.Errorf("MX values = %v", mx)
	}

	a := r53Values(Record{Type: "A", Value: "1.1.1.1\n2.2.2.2"})
	if len(a) != 2 || a[0].Value != "1.1.1.1" || a[1].Value != "2.2.2.2" {
		t.Errorf("A values = %v", a)
	}
}

func TestRoute53SetFromAndID(t *testing.T) {
	set := r53SetFrom(Zone{Name: "example.com"}, Record{Name: "www", Type: "A", Value: "1.1.1.1"})
	if set.Name != "www.example.com" || set.Type != "A" || set.TTL != 300 || len(set.Records) != 1 {
		t.Errorf("set = %+v", set)
	}
	if got := r53ID("WWW.Example.com.", "A"); got != "A|www.example.com" {
		t.Errorf("r53ID = %q", got)
	}
}

func TestAliDNSRecordParams(t *testing.T) {
	c := newAliDNS(map[string]string{})
	p := c.recordParams(Record{Type: "A", Value: "1.1.1.1"})
	if p["RR"] != "@" || p["TTL"] != "600" || p["Type"] != "A" {
		t.Errorf("params = %v", p)
	}
	if _, ok := p["Priority"]; ok {
		t.Errorf("non-MX params must not carry Priority: %v", p)
	}
	mx := c.recordParams(Record{Name: "mail", Type: "MX", Value: "mx.example.com", TTL: 120})
	if mx["RR"] != "mail" || mx["TTL"] != "120" || mx["Priority"] != "10" {
		t.Errorf("MX params = %v", mx)
	}
}

func TestTencentRecordPayloadAndParseID(t *testing.T) {
	c := newTencent(map[string]string{})
	p := c.recordPayload(Zone{Name: "example.com"}, Record{Type: "A", Value: "1.1.1.1"})
	if p["Domain"] != "example.com" || p["SubDomain"] != "@" || p["TTL"] != 600 {
		t.Errorf("payload = %v", p)
	}
	if _, ok := p["MX"]; ok {
		t.Errorf("non-MX payload must not carry MX: %v", p)
	}
	mx := c.recordPayload(Zone{Name: "example.com"}, Record{Name: "mail", Type: "MX", Value: "mx.example.com", Priority: 5})
	if mx["MX"] != 5 || mx["SubDomain"] != "mail" {
		t.Errorf("MX payload = %v", mx)
	}

	if id, err := parseRecordID("12345"); err != nil || id != 12345 {
		t.Errorf("parseRecordID(12345) = %d, %v", id, err)
	}
	if _, err := parseRecordID("abc"); err == nil {
		t.Error("parseRecordID(abc) must fail")
	}
}
