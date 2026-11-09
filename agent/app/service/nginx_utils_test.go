package service

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
)

func TestFindDirectNginxDirectivesSkipsNestedBlocks(t *testing.T) {
	config, err := parser.NewStringParser(`http {
    proxy_read_timeout 5m;
    server {
 proxy_read_timeout 10s;
        location / {
     proxy_read_timeout 2s;
 }
    }
}
`).Parse()
	if err != nil {
		t.Fatal(err)
	}
	dirs := findDirectNginxDirectives(config.FindHttp(), "proxy_read_timeout")
	if len(dirs) != 1 || !reflect.DeepEqual(dirs[0].GetParameters(), []string{"5m"}) {
		t.Fatalf("expected only the http-level 5m, got %d directives", len(dirs))
	}
}

func TestFillNginxIncludedHTTPParams(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		filePath := filepath.Join(dir, name)
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return filePath
	}
	site := write("a.example.com.conf", `server {
    listen 80;
    location / {
        proxy_read_timeout 2s;
        proxy_pass http://127.0.0.1:8080;
    }
}
`)
	global := write("timeout.conf", "proxy_read_timeout 120s;\nproxy_send_timeout 90s;\n")
	broken := write("broken.conf", "proxy_connect_timeout 5s\n")

	params := []response.NginxParam{
		{Name: "proxy_connect_timeout", Params: []string{}},
		{Name: "proxy_send_timeout", Params: []string{"30s"}},
		{Name: "proxy_read_timeout", Params: []string{}},
	}
	got := fillNginxIncludedHTTPParams(params, []string{site, broken, global})

	want := []response.NginxParam{
		// The only candidate fails to parse, so the value stays unknown.
		{Name: "proxy_connect_timeout", Params: []string{}},
		// nginx.conf already sets it; the included value is not consulted.
		{Name: "proxy_send_timeout", Params: []string{"30s"}},
		// The site's location value is skipped, the top-level one is used.
		{Name: "proxy_read_timeout", Params: []string{"120s"}, Source: global},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}
