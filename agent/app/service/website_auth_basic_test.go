package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
)

func TestApplyPathAuthLocationMatch(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		wantModifier string
		wantMatch    string
		wantDisplay  string
	}{
		{
			name:         "directory path uses prefix priority modifier",
			input:        "/test/",
			wantModifier: "^~",
			wantMatch:    "/test/",
			wantDisplay:  "/test/",
		},
		{
			name:         "path without leading slash is normalized",
			input:        "admin/",
			wantModifier: "^~",
			wantMatch:    "/admin/",
			wantDisplay:  "/admin/",
		},
		{
			name:         "regex path keeps case-insensitive regex modifier",
			input:        "/secret.*",
			wantModifier: "~*",
			wantMatch:    "^/secret.*",
			wantDisplay:  "/secret.*",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := parser.NewStringParser(string(nginx_conf.PathAuth)).Parse()
			if err != nil {
				t.Fatalf("parse PathAuth template: %v", err)
			}
			loc, ok := cfg.Directives[0].(*components.Location)
			if !ok {
				t.Fatalf("expected Location directive, got %T", cfg.Directives[0])
			}
			applyPathAuthLocationMatch(loc, tc.input)
			if loc.Modifier != tc.wantModifier || loc.Match != tc.wantMatch {
				t.Fatalf("got (%q, %q), want (%q, %q)", loc.Modifier, loc.Match, tc.wantModifier, tc.wantMatch)
			}
			if gotDisplay := strings.TrimPrefix(loc.Match, "^"); gotDisplay != tc.wantDisplay {
				t.Fatalf("display path got %q, want %q", gotDisplay, tc.wantDisplay)
			}
		})
	}
}

func TestCollectSyncedPathAuthConfigsAcrossPHPSwitches(t *testing.T) {
	fileOp := files.NewFileOp()
	authDir := t.TempDir()
	confPath := filepath.Join(authDir, "test.conf")

	legacyConf := `location ~* ^/test/ {
    auth_basic "Authorization";
    auth_basic_user_file /www/sites/demo/path_auth/pass/test.pass;
}
`
	if err := os.WriteFile(confPath, []byte(legacyConf), 0o644); err != nil {
		t.Fatal(err)
	}

	serverConf := `server {
    listen 80;
    root /www/sites/demo/index;
    location ~ ^/(\.user.ini|\.htaccess|\.git|\.env|\.svn|\.project|LICENSE|README.md) {
        return 404;
    }
    location ~ [^/]\.php(/|$) {
        fastcgi_pass 127.0.0.1:9001;
        include fastcgi-php.conf;
        include fastcgi_params;
        fastcgi_read_timeout 120s;
    }
    include /www/sites/demo/path_auth/*.conf;
}
`
	parsedServerCfg, err := parser.NewStringParser(serverConf).Parse()
	if err != nil {
		t.Fatal(err)
	}
	server := parsedServerCfg.FindServers()[0]
	if findServerPHPLocation(server) == nil {
		t.Fatal("expected server PHP location to be found")
	}

	// 1. Legacy ~* config + switch to PHP 8.3 (127.0.0.1:9001): upgrades to ^~, nests sensitive deny before PHP location.
	configs, _, err := collectSyncedPathAuthConfigs(fileOp, authDir, true, "127.0.0.1:9001", server)
	if err != nil {
		t.Fatalf("collectSyncedPathAuthConfigs (enable PHP) failed: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 synced config, got %d", len(configs))
	}
	dumped := nginx.DumpConfig(configs[0], nginx.IndentedStyle)
	if !strings.Contains(dumped, "location ^~ /test/ {") {
		t.Fatalf("expected upgraded ^~ location, got:\n%s", dumped)
	}
	denyIdx := strings.Index(dumped, defaultSensitiveDenyMatch)
	phpIdx := strings.Index(dumped, "location ~ [^/]\\.php(/|$) {")
	if denyIdx < 0 || phpIdx < 0 || denyIdx > phpIdx {
		t.Fatalf("expected sensitive-file deny location before nested PHP location, got:\n%s", dumped)
	}
	if !strings.Contains(dumped, "fastcgi_pass 127.0.0.1:9001;") {
		t.Fatalf("expected nested PHP location with fastcgi_pass 127.0.0.1:9001, got:\n%s", dumped)
	}
	if !strings.Contains(dumped, "fastcgi_read_timeout 120s;") {
		t.Fatalf("expected custom fastcgi directive to be preserved, got:\n%s", dumped)
	}
	if err := os.WriteFile(confPath, []byte(dumped), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Round-trip parse and switch PHP 8.3 -> PHP 7.0 (127.0.0.1:9002): updates fastcgi_pass in place without duplicating deny or PHP blocks.
	configs, _, err = collectSyncedPathAuthConfigs(fileOp, authDir, true, "127.0.0.1:9002", server)
	if err != nil {
		t.Fatalf("collectSyncedPathAuthConfigs (update PHP version) failed: %v", err)
	}
	dumped = nginx.DumpConfig(configs[0], nginx.IndentedStyle)
	if strings.Count(dumped, "location ~ [^/]\\.php(/|$)") != 1 {
		t.Fatalf("expected exactly 1 nested PHP location, got:\n%s", dumped)
	}
	if strings.Count(dumped, defaultSensitiveDenyMatch) != 1 {
		t.Fatalf("expected exactly 1 sensitive-file deny location, got:\n%s", dumped)
	}
	if !strings.Contains(dumped, "fastcgi_pass 127.0.0.1:9002;") || strings.Contains(dumped, "127.0.0.1:9001") {
		t.Fatalf("expected fastcgi_pass updated to 127.0.0.1:9002, got:\n%s", dumped)
	}
	if err := os.WriteFile(confPath, []byte(dumped), 0o644); err != nil {
		t.Fatal(err)
	}

	// 3. Switch PHP -> Static: removes nested PHP location while preserving auth_basic and sensitive-file deny.
	configs, _, err = collectSyncedPathAuthConfigs(fileOp, authDir, false, "", server)
	if err != nil {
		t.Fatalf("collectSyncedPathAuthConfigs (switch to static) failed: %v", err)
	}
	dumped = nginx.DumpConfig(configs[0], nginx.IndentedStyle)
	if strings.Contains(dumped, "fastcgi_pass") || strings.Contains(dumped, "location ~ [^/]\\.php(/|$)") {
		t.Fatalf("expected nested PHP location removed for static site, got:\n%s", dumped)
	}
	if !strings.Contains(dumped, "location ^~ /test/ {") || !strings.Contains(dumped, `auth_basic "Authorization";`) {
		t.Fatalf("expected auth_basic preserved, got:\n%s", dumped)
	}
	if !strings.Contains(dumped, defaultSensitiveDenyMatch) || !strings.Contains(dumped, "return 404;") {
		t.Fatalf("expected sensitive-file deny preserved on static site, got:\n%s", dumped)
	}
}

func TestSyncPathAuthPreservesWebsiteDefaultSensitiveDenyOnRoot(t *testing.T) {
	defaultServerCfg, err := parser.NewStringParser(string(nginx_conf.GetWebsiteFile("website_default.conf"))).Parse()
	if err != nil {
		t.Fatalf("parse WebsiteDefault: %v", err)
	}
	server := defaultServerCfg.FindServers()[0]

	authCfg, err := parser.NewStringParser(string(nginx_conf.PathAuth)).Parse()
	if err != nil {
		t.Fatalf("parse PathAuth: %v", err)
	}
	loc := authCfg.Directives[0].(*components.Location)
	applyPathAuthLocationMatch(loc, "/")
	if err := syncPathAuthSensitiveDeny(loc, server); err != nil {
		t.Fatalf("syncPathAuthSensitiveDeny: %v", err)
	}

	dumped := nginx.DumpConfig(authCfg, nginx.IndentedStyle)
	if !strings.Contains(dumped, "location ^~ / {") {
		t.Fatalf("expected priority root auth location, got:\n%s", dumped)
	}
	if !strings.Contains(dumped, defaultSensitiveDenyMatch) || !strings.Contains(dumped, "return 404;") {
		t.Fatalf("expected WebsiteDefault sensitive-file deny nested in ^~ /, got:\n%s", dumped)
	}
}

func TestServerUpdatePHPProxyReplacesInPlace(t *testing.T) {
	serverConf := `server {
    listen 80;
    location ~ [^/]\.php(/|$) {
        fastcgi_pass 127.0.0.1:9000;
        include fastcgi-php.conf;
    }
    include /www/sites/demo/path_auth/*.conf;
}
`
	cfg, err := parser.NewStringParser(serverConf).Parse()
	if err != nil {
		t.Fatal(err)
	}
	server := cfg.FindServers()[0]
	server.UpdatePHPProxy([]string{"127.0.0.1:9001"}, "")
	dumped := nginx.DumpConfig(cfg, nginx.IndentedStyle)
	if strings.Count(dumped, "location ~ [^/]\\.php(/|$)") != 1 {
		t.Fatalf("expected in-place PHP location update without duplicates, got:\n%s", dumped)
	}
	if strings.Index(dumped, "location ~ [^/]\\.php(/|$)") > strings.Index(dumped, "include /www/sites/demo/path_auth/*.conf;") {
		t.Fatalf("expected PHP location to stay in its original position, got:\n%s", dumped)
	}
}
