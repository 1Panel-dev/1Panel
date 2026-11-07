package service

import (
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// the mapping layer logs via global.LOG; the test binary never runs the
// agent boot sequence, so install a silent logger to avoid nil derefs
func TestMain(m *testing.M) {
	if global.LOG == nil {
		global.LOG = &logrus.Logger{Out: io.Discard, Formatter: &logrus.TextFormatter{}, Level: logrus.PanicLevel}
	}
	os.Exit(m.Run())
}

func TestMapLoong64Image(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		mapping bool
	}{
		// exact repo rules with major routing / fallbacks
		{"redis:8.10.3", "lcr.loongnix.cn/library/redis:7.4.1", true},
		{"mysql:5.7.44", "lcr.loongnix.cn/library/mysql:5.7.29-an23", true},
		{"mysql:8.4.13", "lcr.loongnix.cn/library/mysql:8.4.3", true},
		{"postgres:18.6-alpine", "lcr.loongnix.cn/library/postgres:17-sid", true},
		{"postgres:13.3", "lcr.loongnix.cn/library/postgres:13.19-alpine3.21", true},
		{"postgres", "lcr.loongnix.cn/library/postgres:15.7-sid", true},
		{"mariadb:11.8.9", "lcr.loongnix.cn/library/mariadb:13.1.1", true},
		{"mariadb:10.5.18", "lcr.loongnix.cn/library/mariadb:10.5.18", true},
		{"wordpress:7.1.4", "lcr.loongnix.cn/library/wordpress:7.0.2-php8.5-fpm-alpine", true},
		{"phpmyadmin:5.2.3", "lcr.loongnix.cn/library/phpmyadmin:5.2.3", true},
		{"elasticsearch:9.5.5", "lcr.loongnix.cn/library/elasticsearch:9.5.5", true},
		{"kibana:9.5.5", "lcr.loongnix.cn/library/kibana:9.4.2", true},
		{"influxdb:3.12.2-core", "lcr.loongnix.cn/library/influxdb:3.9.6-alpine", true},
		{"nats:2.15.2", "lcr.loongnix.cn/library/nats:2.10.24-alpine", true},
		{"mongo:8.3.13", "lcr.loongnix.cn/library/mongo:6.0.26", true},
		{"registry:3.1.2", "lcr.loongnix.cn/library/registry:3.1.1", true},
		{"hashicorp/consul:2.1.0", "lcr.loongnix.cn/library/consul:1.15.2", true},
		{"gitlab/gitlab-ce:19.5.3-ce.0", "lcr.loongnix.cn/gitlab/gitlab-ce:17.6.2", true},
		{"grafana/grafana:13.3.0", "lcr.loongnix.cn/grafana/grafana:13.0.2-alpine", true},
		{"jaegertracing/all-in-one:1.76.0", "lcr.loongnix.cn/jaegertracing/all-in-one:1.6.0", true},
		{"jenkins/jenkins:2.580.2-lts", "lcr.loongnix.cn/jenkins/jenkins:2.583-jdk21", true},
		{"portainer/portainer-ce:2.45.3-alpine", "lcr.loongnix.cn/portainer/portainer-ce:2.11.1", true},
		{"postgis/postgis:18-3.6-alpine", "lcr.loongnix.cn/postgis/postgis:16-3.4", true},
		{"emqx/emqx:6.3.2", "lcr.loongnix.cn/emqx/emqx:6.3.2", true},
		{"2fauth/2fauth:9.0.1", "lcr.loongnix.cn/library/2fauth:latest", true},
		{"pgvector/pgvector:0.8.7-pg18", "lcr.loongnix.cn/library/pgvector:0.8.0", true},
		// base image fallbacks
		{"alpine:3.19", "lcr.loongnix.cn/library/alpine:3.24", true},
		{"busybox:1.36", "lcr.loongnix.cn/library/busybox:latest", true},
		{"python:3.12-slim", "lcr.loongnix.cn/library/python:3.11.7", true},
		// openresty -> local registry, tag kept
		{"1panel/openresty:1.31.8.1-2-1-noble", "127.0.0.1:5000/loong/openresty:1.31.8.1-2-1-noble", true},
		// phase-2 batch 1 self-built apps -> local registry, tag kept
		{"caddy:2.11.8", "127.0.0.1:5000/loong/caddy:2.11.8", true},
		{"filebrowser/filebrowser:v2.63.23", "127.0.0.1:5000/loong/filebrowser:v2.63.23", true},
		{"jeessy/ddns-go:v6.17.7", "127.0.0.1:5000/loong/ddns-go:v6.17.7", true},
		// phase-2 batch 2
		{"joyqi/typecho:1.3.0-php8.2-apache", "127.0.0.1:5000/loong/typecho:1.3.0-php8.2-apache", true},
		{"louislam/uptime-kuma:2.5.9", "127.0.0.1:5000/loong/uptime-kuma:2.5.9", true},
		// phase-2 batch 3
		{"halohub/halo-pro:2.27.2", "127.0.0.1:5000/loong/halo-pro:2.27.2", true},
		{"linuxserver/freshrss:1.30.4", "127.0.0.1:5000/loong/freshrss:1.30.4", true},
		// mega216 batch (2026-11-07)
		{"cloudflare/cloudflared:2026.10.7", "127.0.0.1:5000/loong/cloudflared:2026.10.7", true},
		{"miniflux/miniflux:2.3.5", "127.0.0.1:5000/loong/miniflux:2.3.5", true},
		{"prom/node-exporter:v1.12.1", "127.0.0.1:5000/loong/node-exporter:v1.12.1", true},
		{"containrrr/watchtower:latest", "127.0.0.1:5000/loong/watchtower:latest", true},
		{"snowdreamtech/frpc:0.71.1-alpine", "127.0.0.1:5000/loong/frpc:0.71.1-alpine", true},
		{"snowdreamtech/frps:0.71.1-alpine", "127.0.0.1:5000/loong/frps:0.71.1-alpine", true},
		{"glanceapp/glance:v0.8.7", "127.0.0.1:5000/loong/glance:v0.8.7", true},
		{"foxcpp/maddy:0.9.6", "127.0.0.1:5000/loong/maddy-mail:0.9.6", true},
		{"artalk/artalk-go:2.10", "127.0.0.1:5000/loong/artalk:2.10", true},
		{"drakkan/sftpgo:v2.7.6-alpine", "127.0.0.1:5000/loong/sftpgo:v2.7.6-alpine", true},
		{"yisier1/npc:v0.26.35", "127.0.0.1:5000/loong/npc:v0.26.35", true},
		{"yisier1/nps:v0.26.35", "127.0.0.1:5000/loong/nps:v0.26.35", true},
		{"weishaw/sub2api:0.2.19", "127.0.0.1:5000/loong/sub2api:0.2.19", true},
		{"jokerwho/palworld-server-tool:0.12.2", "127.0.0.1:5000/loong/palworld-server-tool:0.12.2", true},
		{"gitea/act_runner:0.6.1", "127.0.0.1:5000/loong/act-runner:0.6.1", true},
		{"calciumion/new-api:v0.13.2", "127.0.0.1:5000/loong/new-api:v0.13.2", true},
		{"xhofe/alist:v3.66.0", "127.0.0.1:5000/loong/alist:v3.66.0", true},
		{"keycloak/keycloak:26.8.1", "127.0.0.1:5000/loong/keycloak:26.8.1", true},
		{"metabase/metabase:v0.65.3", "127.0.0.1:5000/loong/metabase:v0.65.3", true},
		{"nacos/nacos-server:v3.3.0", "127.0.0.1:5000/loong/nacos:v3.3.0", true},
		{"apache/rocketmq:5.5.0", "127.0.0.1:5000/loong/rocketmq:5.5.0", true},
		{"bladex/sentinel-dashboard:1.8.10", "127.0.0.1:5000/loong/sentinel-dashboard:1.8.10", true},
		{"gotson/komga:1.30.0", "127.0.0.1:5000/loong/komga:1.30.0", true},
		{"bitnamilegacy/kafka:4.0.0", "127.0.0.1:5000/loong/kafka:4.0.0", true},
		{"bitnamilegacy/tomcat:10.1.43", "127.0.0.1:5000/loong/tomcat:10.1.43", true},
		{"corentinth/it-tools:2024.10.22-7ca5933", "127.0.0.1:5000/loong/it-tools:2024.10.22-7ca5933", true},
		{"ghcr.io/gchq/cyberchef:11.5.0", "127.0.0.1:5000/loong/cyberchef:11.5.0", true},
		{"doocs/md:2.1.0", "127.0.0.1:5000/loong/wechatmarkdowneditor:2.1.0", true},
		{"fogforest/yesplaymusic:0.4.9", "127.0.0.1:5000/loong/yesplaymusic:0.4.9", true},
		{"wcjiang/linux-command:1.22.0", "127.0.0.1:5000/loong/linux-command:1.22.0", true},
		{"adminer:6.3.0-standalone", "127.0.0.1:5000/loong/adminer:6.3.0-standalone", true},
		{"ghcr.io/digininja/dvwa:latest", "127.0.0.1:5000/loong/dvwa:latest", true},
		{"roundcube/roundcubemail:1.7.5-apache", "127.0.0.1:5000/loong/roundcube:1.7.5-apache", true},
		{"kodcloud/kodbox:1.70.03", "127.0.0.1:5000/loong/kodbox:1.70.03", true},
		{"bellamy/wallos:5.10.4", "127.0.0.1:5000/loong/wallos:5.10.4", true},
		{"linuxserver/nextcloud:35.0.2", "127.0.0.1:5000/loong/nextcloud:35.0.2", true},
		{"star7th/showdoc:v3.9.6", "127.0.0.1:5000/loong/showdoc:v3.9.6", true},
		{"ddsderek/easyimage:v2.8.7", "127.0.0.1:5000/loong/easyimage2:v2.8.7", true},
		{"tencentci/discuz:v3.4", "127.0.0.1:5000/loong/discuz:v3.4", true},
		{"verdaccio/verdaccio:7.0.2", "127.0.0.1:5000/loong/verdaccio:7.0.2", true},
		{"mongo-express:1.0.2-20", "127.0.0.1:5000/loong/mongo-express:1.0.2-20", true},
		{"rediscommander/redis-commander:latest", "127.0.0.1:5000/loong/redis-commander:latest", true},
		{"nodered/node-red:5.0.9", "127.0.0.1:5000/loong/node-red:5.0.9", true},
		{"linuxserver/qbittorrent:5.3.0", "127.0.0.1:5000/loong/qbittorrent:5.3.0", true},
		{"valkey/valkey:9.0.7", "127.0.0.1:5000/loong/valkey:9.0.7", true},
		{"zabbix/zabbix-agent:alpine-7.4.5", "127.0.0.1:5000/loong/zabbix-agent:alpine-7.4.5", true},
		// mega216 batch (2026-11-07)
		{"cloudflare/cloudflared:2026.10.7", "127.0.0.1:5000/loong/cloudflared:2026.10.7", true},
		{"miniflux/miniflux:2.3.5", "127.0.0.1:5000/loong/miniflux:2.3.5", true},
		{"prom/node-exporter:v1.12.1", "127.0.0.1:5000/loong/node-exporter:v1.12.1", true},
		{"containrrr/watchtower:latest", "127.0.0.1:5000/loong/watchtower:latest", true},
		{"snowdreamtech/frpc:0.71.1-alpine", "127.0.0.1:5000/loong/frpc:0.71.1-alpine", true},
		{"snowdreamtech/frps:0.71.1-alpine", "127.0.0.1:5000/loong/frps:0.71.1-alpine", true},
		{"glanceapp/glance:v0.8.7", "127.0.0.1:5000/loong/glance:v0.8.7", true},
		{"foxcpp/maddy:0.9.6", "127.0.0.1:5000/loong/maddy-mail:0.9.6", true},
		{"artalk/artalk-go:2.10", "127.0.0.1:5000/loong/artalk:2.10", true},
		{"yisier1/npc:v0.26.35", "127.0.0.1:5000/loong/npc:v0.26.35", true},
		{"yisier1/nps:v0.26.35", "127.0.0.1:5000/loong/nps:v0.26.35", true},
		{"gitea/act_runner:0.6.1", "127.0.0.1:5000/loong/act-runner:0.6.1", true},
		{"calciumion/new-api:v0.13.2", "127.0.0.1:5000/loong/new-api:v0.13.2", true},
		{"keycloak/keycloak:26.8.1", "127.0.0.1:5000/loong/keycloak:26.8.1", true},
		{"metabase/metabase:v0.65.3", "127.0.0.1:5000/loong/metabase:v0.65.3", true},
		{"nacos/nacos-server:v3.3.0", "127.0.0.1:5000/loong/nacos:v3.3.0", true},
		{"apache/rocketmq:5.5.0", "127.0.0.1:5000/loong/rocketmq:5.5.0", true},
		{"bladex/sentinel-dashboard:1.8.10", "127.0.0.1:5000/loong/sentinel-dashboard:1.8.10", true},
		{"bitnamilegacy/kafka:4.0.0", "127.0.0.1:5000/loong/kafka:4.0.0", true},
		{"bitnamilegacy/tomcat:10.1.43", "127.0.0.1:5000/loong/tomcat:10.1.43", true},
		{"corentinth/it-tools:2024.10.22-7ca5933", "127.0.0.1:5000/loong/it-tools:2024.10.22-7ca5933", true},
		{"doocs/md:2.1.0", "127.0.0.1:5000/loong/wechatmarkdowneditor:2.1.0", true},
		{"wcjiang/linux-command:1.22.0", "127.0.0.1:5000/loong/linux-command:1.22.0", true},
		{"adminer:6.3.0-standalone", "127.0.0.1:5000/loong/adminer:6.3.0-standalone", true},
		{"roundcube/roundcubemail:1.7.5-apache", "127.0.0.1:5000/loong/roundcube:1.7.5-apache", true},
		{"star7th/showdoc:v3.9.6", "127.0.0.1:5000/loong/showdoc:v3.9.6", true},
		{"rediscommander/redis-commander:latest", "127.0.0.1:5000/loong/redis-commander:latest", true},
		{"linuxserver/qbittorrent:5.3.0", "127.0.0.1:5000/loong/qbittorrent:5.3.0", true},
		{"valkey/valkey:9.0.7", "127.0.0.1:5000/loong/valkey:9.0.7", true},
		{"zabbix/zabbix-agent:alpine-7.4.5", "127.0.0.1:5000/loong/zabbix-agent:alpine-7.4.5", true},
		// no mapping
		{"nginx:stable", "", false},
		{"lcr.loongnix.cn/library/redis:7.4.1", "", false},
		{"redis@sha256:abc123", "", false},
		{"${OPENLIST_IMAGE}:latest", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := mapLoong64Image(c.in)
		if ok != c.mapping {
			t.Errorf("mapLoong64Image(%q) mapping=%v, want %v", c.in, ok, c.mapping)
			continue
		}
		if ok && got != c.want {
			t.Errorf("mapLoong64Image(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMapImagesForLoong64Blocked(t *testing.T) {
	raw := `
services:
  jellyfin:
    image: jellyfin/jellyfin:10.11.11
`
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(raw), &composeMap); err != nil {
		t.Fatal(err)
	}
	// runtime gate: on non-loong64 test hosts nothing happens
	if err := mapImagesForLoong64(composeMap); err != nil {
		t.Logf("non-loong64 host: gate returned early (%v)", err)
	}
}

func TestIsBlockedLoong64Image(t *testing.T) {
	blocked := []string{
		"jellyfin/jellyfin:10.11.11",
		// mega216 batch (2026-11-07)
		"ghcr.io/666ghj/bettafish:v3.0.0",
		"linuxserver/code-server:4.145.0",
		"1panel/comfyui:v0.33.1",
		"harness/harness:3.3.0",
		"ghcr.io/ggml-org/llama.cpp:server",
		"localai/localai:v3.12.1-aio-cpu",
		"mcr.microsoft.com/mssql/server:2025-latest",
		"jc21/nginx-proxy-manager:2.16.0",
		"ollama/ollama:0.40.14",
		"ghcr.io/open-webui/open-webui:v0.13.0",
		"openvpn/openvpn-as:latest",
		"thijsvanloef/palworld-server-docker:v2.8.1",
		"lmsysorg/sglang:v0.5.23",
		"teamspeak:3.13.8",
		"ghcr.io/tiny-craft/tiny-rdm:1.2.7",
		"shuzhiyinhang/vexdb:3.0.0.28146",
		"dockurr/windows:6.06",
		"zhaojun1998/zfile:5.1.0",
	}
	for _, img := range blocked {
		if !isBlockedLoong64Image(img) {
			t.Errorf("%s should be blocked", img)
		}
	}
	allowed := []string{
		"redis:8.10.3",
		"cloudflare/cloudflared:2026.10.7", // mapped to local registry, not blocked
		"library/wordpress:7.1.4",
	}
	for _, img := range allowed {
		if isBlockedLoong64Image(img) {
			t.Errorf("%s should not be blocked", img)
		}
	}
}

// regression: the openresty store template declares BOTH a build context and
// a published image; the mapping must still rewrite the image (it used to
// silently skip any service carrying a "build" section)
func TestMapImagesForLoong64WithBuildSection(t *testing.T) {
	if runtime.GOARCH != "loong64" {
		t.Skip("compose-level gate only fires on linux/loong64")
	}
	raw := `
services:
  openresty:
    build:
      context: ./build
      args:
        - PANEL_OPENRESTY_VERSION=1.31.8.1-2-1-noble
    image: 1panel/openresty:1.31.8.1-2-1-noble
    container_name: ${CONTAINER_NAME}
    restart: always
    network_mode: host
`
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(raw), &composeMap); err != nil {
		t.Fatal(err)
	}
	if err := mapImagesForLoong64(composeMap); err != nil {
		t.Fatalf("mapping failed: %v", err)
	}
	services := composeMap["services"].(map[string]interface{})
	svc := services["openresty"].(map[string]interface{})
	got := svc["image"].(string)
	want := "127.0.0.1:5000/loong/openresty:1.31.8.1-2-1-noble"
	if got != want {
		t.Fatalf("image = %q, want %q", got, want)
	}
}

// compose-level blocked check must abort with the clear Chinese message
func TestMapImagesForLoong64BlockedCompose(t *testing.T) {
	if runtime.GOARCH != "loong64" {
		t.Skip("compose-level gate only fires on linux/loong64")
	}
	raw := `
services:
  jellyfin:
    image: jellyfin/jellyfin:10.11.11
`
	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(raw), &composeMap); err != nil {
		t.Fatal(err)
	}
	err := mapImagesForLoong64(composeMap)
	if err == nil {
		t.Fatal("blocked repo must abort the install")
	}
	if err.Error() != loong64UnsupportedMsg {
		t.Fatalf("error = %q, want %q", err.Error(), loong64UnsupportedMsg)
	}
}
