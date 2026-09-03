package service

import (
	"strings"
	"testing"
)

const plainNginxConf = `user  root;
worker_processes  auto;

include /usr/local/openresty/nginx/conf/modules-enabled/*.conf;

events {
      use epoll;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    gzip on;
    gzip_comp_level 5;

    limit_conn_zone $binary_remote_addr zone=perip:10m;

    include /usr/local/openresty/nginx/conf/conf.d/*.conf;
    include /usr/local/openresty/nginx/conf/default/*.conf;
}
`

func TestInsertNginxHTTPIncludeBeforeConfD(t *testing.T) {
	got, err := insertNginxHTTPInclude(plainNginxConf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "    "+nginxHTTPIncludeDirective) {
		t.Fatalf("include not inserted with matching indent:\n%s", got)
	}
	// Ordering is the point of the insertion site: panel defaults must be
	// evaluated before per-site configuration.
	httpIdx := strings.Index(got, "conf/http.d/*.conf")
	confDIdx := strings.Index(got, "conf/conf.d/*.conf")
	if httpIdx < 0 || confDIdx < 0 || httpIdx > confDIdx {
		t.Fatalf("http.d must be included before conf.d (http.d=%d conf.d=%d)", httpIdx, confDIdx)
	}
	// The rest of the file must be untouched.
	stripped := strings.Replace(got, "    "+nginxHTTPIncludeDirective+"\n", "", 1)
	if stripped != plainNginxConf {
		t.Fatal("insertion altered content outside the inserted line")
	}
}

func TestInsertNginxHTTPIncludeIsIdempotent(t *testing.T) {
	once, err := insertNginxHTTPInclude(plainNginxConf)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := insertNginxHTTPInclude(once)
	if err != nil {
		t.Fatal(err)
	}
	if twice != once {
		t.Fatal("a second insertion must be a no-op")
	}
}

func TestInsertNginxHTTPIncludeFallsBackToHTTPBlock(t *testing.T) {
	content := strings.Replace(plainNginxConf,
		"    include /usr/local/openresty/nginx/conf/conf.d/*.conf;\n", "", 1)
	got, err := insertNginxHTTPInclude(content)
	if err != nil {
		t.Fatal(err)
	}
	httpIdx := strings.Index(got, "http {")
	incIdx := strings.Index(got, nginxHTTPIncludeDirective)
	if incIdx < 0 || incIdx < httpIdx {
		t.Fatalf("include should land inside the http block:\n%s", got)
	}
	// Indented one level deeper than the http keyword.
	if !strings.Contains(got, "    "+nginxHTTPIncludeDirective) {
		t.Errorf("fallback indentation is wrong:\n%s", got)
	}
}

func TestInsertNginxHTTPIncludeRejectsConfigWithoutHTTPBlock(t *testing.T) {
	if _, err := insertNginxHTTPInclude("events {}\n"); err == nil {
		t.Fatal("a config without an http block must be rejected so callers can degrade")
	}
}

func TestInsertNginxHTTPIncludeIgnoresCommentedIncludes(t *testing.T) {
	commented := strings.Replace(plainNginxConf,
		"    include /usr/local/openresty/nginx/conf/conf.d/*.conf;",
		"    # include /usr/local/openresty/nginx/conf/conf.d/*.conf;", 1)
	got, err := insertNginxHTTPInclude(commented)
	if err != nil {
		t.Fatal(err)
	}
	// The commented conf.d line is not a valid anchor; the fallback must win.
	if strings.Index(got, nginxHTTPIncludeDirective) < strings.Index(got, "http {") {
		t.Fatal("a commented include must not be used as the anchor")
	}
}

func TestInsertNginxHTTPIncludeHandlesCRLF(t *testing.T) {
	crlf := strings.ReplaceAll(plainNginxConf, "\n", "\r\n")
	got, err := insertNginxHTTPInclude(crlf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, nginxHTTPIncludeDirective) {
		t.Fatal("include missing on a CRLF file")
	}
}

func TestNginxHTTPIncludeRe(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"present", plainNginxConf + "    " + nginxHTTPIncludeDirective + "\n", true},
		{"absent", plainNginxConf, false},
		{"commented out", "# " + nginxHTTPIncludeDirective, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := nginxHTTPIncludeRe.MatchString(tc.content); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
