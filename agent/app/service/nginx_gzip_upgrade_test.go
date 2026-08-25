package service

import (
	"strings"
	"testing"
)

const stockNginxConf = `user  root;
worker_processes  auto;

include /usr/local/openresty/nginx/conf/modules-enabled/*.conf;

events {
      use epoll;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    server_names_hash_bucket_size 512;
    keepalive_requests 5000;

    gzip on;
    gzip_min_length  1k;
    gzip_buffers     4 16k;
    gzip_http_version 1.1;
    gzip_comp_level 2;
    gzip_types     text/plain application/javascript application/x-javascript text/javascript text/css application/xml;
    gzip_vary on;
    gzip_proxied   expired no-cache no-store private auth;
    gzip_disable   "MSIE [1-6]\.";

    limit_conn_zone $binary_remote_addr zone=perip:10m;

    include /usr/local/openresty/nginx/conf/http.d/*.conf;
    include /usr/local/openresty/nginx/conf/conf.d/*.conf;
}
`

func TestIsStockNginxGzipConfig(t *testing.T) {
	if !isStockNginxGzipConfig(stockNginxConf) {
		t.Fatal("factory configuration should be detected as stock")
	}
}

func TestIsStockNginxGzipConfigRejectsTunedValues(t *testing.T) {
	cases := map[string]string{
		"comp level changed": strings.Replace(stockNginxConf, "gzip_comp_level 2;", "gzip_comp_level 6;", 1),
		"gzip disabled":      strings.Replace(stockNginxConf, "gzip on;", "gzip off;", 1),
		"types extended": strings.Replace(stockNginxConf,
			"application/xml;", "application/xml application/json;", 1),
		"directive removed": strings.Replace(stockNginxConf, "    gzip_vary on;\n", "", 1),
		"directive added": strings.Replace(stockNginxConf, "    gzip_vary on;\n",
			"    gzip_vary on;\n    gzip_static on;\n", 1),
	}
	for name, content := range cases {
		if isStockNginxGzipConfig(content) {
			t.Errorf("%s: tuned configuration must not be rewritten", name)
		}
	}
}

func TestIsStockNginxGzipConfigRejectsDuplicateDirective(t *testing.T) {
	content := strings.Replace(stockNginxConf, "    gzip on;\n", "    gzip on;\n    gzip on;\n", 1)
	if isStockNginxGzipConfig(content) {
		t.Fatal("a duplicated directive indicates a hand-edited config")
	}
}

func TestRewriteNginxGzipDirectives(t *testing.T) {
	result := rewriteNginxGzipDirectives(stockNginxConf)

	for _, expected := range []string{
		"    gzip_comp_level 5;",
		"    gzip_proxied any;",
		"    gzip on;",
		"    gzip_vary on;",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected directive missing: %s\n%s", expected, result)
		}
	}
	if !strings.Contains(result, "application/json") {
		t.Error("gzip_types should now cover application/json")
	}
	if strings.Contains(result, "gzip_disable") {
		t.Error("obsolete gzip_disable should have been dropped")
	}
	if strings.Contains(result, "gzip_comp_level 2;") {
		t.Error("stale comp level should have been replaced")
	}
	// Everything outside the gzip block must survive untouched.
	for _, keep := range []string{
		"server_names_hash_bucket_size 512;",
		"keepalive_requests 5000;",
		"limit_conn_zone $binary_remote_addr zone=perip:10m;",
		"include /usr/local/openresty/nginx/conf/http.d/*.conf;",
		"include /usr/local/openresty/nginx/conf/conf.d/*.conf;",
		"include /usr/local/openresty/nginx/conf/modules-enabled/*.conf;",
		"user  root;",
	} {
		if !strings.Contains(result, keep) {
			t.Errorf("unrelated line was altered or dropped: %s", keep)
		}
	}
	if !strings.HasSuffix(result, "}\n") {
		t.Error("trailing newline was not preserved")
	}
}

func TestRewriteNginxGzipDirectivesIsIdempotent(t *testing.T) {
	once := rewriteNginxGzipDirectives(stockNginxConf)
	twice := rewriteNginxGzipDirectives(once)
	if once != twice {
		t.Errorf("rewrite is not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}

func TestRewriteNginxGzipDirectivesAppendsMissing(t *testing.T) {
	// gzip_proxied absent from the source must be appended, not silently lost.
	content := strings.Replace(stockNginxConf,
		"    gzip_proxied   expired no-cache no-store private auth;\n", "", 1)
	result := rewriteNginxGzipDirectives(content)
	if !strings.Contains(result, "gzip_proxied any;") {
		t.Errorf("missing directive was not appended:\n%s", result)
	}
	if !strings.Contains(result, "limit_conn_zone $binary_remote_addr zone=perip:10m;") {
		t.Error("appending must not clobber following lines")
	}
}

func TestRewriteNginxGzipDirectivesKeepsGzipLikeNames(t *testing.T) {
	// gunzip and proxy_set_header must survive: only directives whose name
	// starts with "gzip" are managed here.
	content := "http {\n    gunzip on;\n    gzip on;\n    proxy_set_header Accept-Encoding gzip;\n}\n"
	result := rewriteNginxGzipDirectives(content)
	if !strings.Contains(result, "gunzip on;") {
		t.Error("gunzip directive must be preserved")
	}
	if !strings.Contains(result, "proxy_set_header Accept-Encoding gzip;") {
		t.Error("proxy_set_header must be preserved")
	}
	if !strings.Contains(result, "    gzip on;") {
		t.Error("gzip directive should be kept in place")
	}
}
