package service

import (
	"errors"
	"runtime"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
)

// On linux/loong64 hosts most Docker Hub images cannot be pulled because
// they ship no loong64 manifest. This layer rewrites appstore image
// references to loong64-capable equivalents:
//   - lcr.loongnix.cn (Loongson community registry, which rebuilds popular
//     images for loong64)
//   - a local registry (127.0.0.1:5000/loong/...) for images we build on
//     the device (e.g. OpenResty)
//
// All rules are keyed by the source repository path (namespace/name or
// bare docker hub library name). Tag selection: ByMajor[<major>] wins,
// then Fallback; an empty Fallback keeps the requested tag.
// It is a no-op on every other architecture.

type loong64Rule struct {
	Target   string            // target repo path, registry prefix included
	Fallback string            // tag used when the requested tag should be replaced
	ByMajor  map[string]string // source major version -> target tag
}

const (
	loongLCRRegistry      = "lcr.loongnix.cn/"
	loongLocalRegistry    = "127.0.0.1:5000/loong/"
	loong64UnsupportedMsg = "该应用暂无 LoongArch64（龙芯）版本镜像，已列入移植矩阵不支持清单；可在「容器」页用自定义 compose 部署替代镜像"
)

// exact repository-path rules (source repo -> rule)
var loong64RepoRules = map[string]loong64Rule{
	// --- local self-built images (phase-2 batch 1: Go single-binary apps) ---
	"1panel/openresty":        {Target: loongLocalRegistry + "openresty"},   // keep tag
	"caddy":                   {Target: loongLocalRegistry + "caddy"},       // keep tag
	"filebrowser/filebrowser": {Target: loongLocalRegistry + "filebrowser"}, // keep tag
	"jeessy/ddns-go":          {Target: loongLocalRegistry + "ddns-go"},     // keep tag

	// --- local self-built images (phase-2 batch 2) ---
	"joyqi/typecho":        {Target: loongLocalRegistry + "typecho"},     // keep tag
	"louislam/uptime-kuma": {Target: loongLocalRegistry + "uptime-kuma"}, // keep tag

	// --- local self-built images (phase-2 batch 3) ---
	"halohub/halo-pro":     {Target: loongLocalRegistry + "halo-pro"}, // keep tag; community jar under pro tag
	"linuxserver/freshrss": {Target: loongLocalRegistry + "freshrss"}, // keep tag; official FreshRSS recipe, data at /config

	// --- local self-built images (mega216 batch, 2026-11-07) ---
	"cloudflare/cloudflared":         {Target: loongLocalRegistry + "cloudflared"},          // keep tag (mega216)
	"miniflux/miniflux":              {Target: loongLocalRegistry + "miniflux"},             // keep tag (mega216)
	"prom/node-exporter":             {Target: loongLocalRegistry + "node-exporter"},        // keep tag (mega216)
	"containrrr/watchtower":          {Target: loongLocalRegistry + "watchtower"},           // keep tag (mega216)
	"snowdreamtech/frpc":             {Target: loongLocalRegistry + "frpc"},                 // keep tag (mega216)
	"snowdreamtech/frps":             {Target: loongLocalRegistry + "frps"},                 // keep tag (mega216)
	"glanceapp/glance":               {Target: loongLocalRegistry + "glance"},               // keep tag (mega216)
	"foxcpp/maddy":                   {Target: loongLocalRegistry + "maddy-mail"},           // keep tag (mega216)
	"artalk/artalk-go":               {Target: loongLocalRegistry + "artalk"},               // keep tag (mega216)
	"drakkan/sftpgo":                 {Target: loongLocalRegistry + "sftpgo"},               // keep tag (mega216)
	"yisier1/npc":                    {Target: loongLocalRegistry + "npc"},                  // keep tag (mega216)
	"yisier1/nps":                    {Target: loongLocalRegistry + "nps"},                  // keep tag (mega216)
	"weishaw/sub2api":                {Target: loongLocalRegistry + "sub2api"},              // keep tag (mega216)
	"jokerwho/palworld-server-tool":  {Target: loongLocalRegistry + "palworld-server-tool"}, // keep tag (mega216)
	"gitea/act_runner":               {Target: loongLocalRegistry + "act-runner"},           // keep tag (mega216)
	"calciumion/new-api":             {Target: loongLocalRegistry + "new-api"},              // keep tag (mega216)
	"xhofe/alist":                    {Target: loongLocalRegistry + "alist"},                // keep tag (mega216)
	"keycloak/keycloak":              {Target: loongLocalRegistry + "keycloak"},             // keep tag (mega216)
	"metabase/metabase":              {Target: loongLocalRegistry + "metabase"},             // keep tag (mega216)
	"nacos/nacos-server":             {Target: loongLocalRegistry + "nacos"},                // keep tag (mega216)
	"apache/rocketmq":                {Target: loongLocalRegistry + "rocketmq"},             // keep tag (mega216)
	"bladex/sentinel-dashboard":      {Target: loongLocalRegistry + "sentinel-dashboard"},   // keep tag (mega216)
	"gotson/komga":                   {Target: loongLocalRegistry + "komga"},                // keep tag (mega216)
	"bitnamilegacy/kafka":            {Target: loongLocalRegistry + "kafka"},                // keep tag (mega216)
	"bitnamilegacy/tomcat":           {Target: loongLocalRegistry + "tomcat"},               // keep tag (mega216)
	"corentinth/it-tools":            {Target: loongLocalRegistry + "it-tools"},             // keep tag (mega216)
	"ghcr.io/gchq/cyberchef":         {Target: loongLocalRegistry + "cyberchef"},            // keep tag (mega216)
	"doocs/md":                       {Target: loongLocalRegistry + "wechatmarkdowneditor"}, // keep tag (mega216)
	"fogforest/yesplaymusic":         {Target: loongLocalRegistry + "yesplaymusic"},         // keep tag (mega216)
	"wcjiang/linux-command":          {Target: loongLocalRegistry + "linux-command"},        // keep tag (mega216)
	"adminer":                        {Target: loongLocalRegistry + "adminer"},              // keep tag (mega216)
	"ghcr.io/digininja/dvwa":         {Target: loongLocalRegistry + "dvwa"},                 // keep tag (mega216)
	"roundcube/roundcubemail":        {Target: loongLocalRegistry + "roundcube"},            // keep tag (mega216)
	"kodcloud/kodbox":                {Target: loongLocalRegistry + "kodbox"},               // keep tag (mega216)
	"bellamy/wallos":                 {Target: loongLocalRegistry + "wallos"},               // keep tag (mega216)
	"linuxserver/nextcloud":          {Target: loongLocalRegistry + "nextcloud"},            // keep tag (mega216)
	"star7th/showdoc":                {Target: loongLocalRegistry + "showdoc"},              // keep tag (mega216)
	"ddsderek/easyimage":             {Target: loongLocalRegistry + "easyimage2"},           // keep tag (mega216)
	"tencentci/discuz":               {Target: loongLocalRegistry + "discuz"},               // keep tag (mega216)
	"verdaccio/verdaccio":            {Target: loongLocalRegistry + "verdaccio"},            // keep tag (mega216)
	"mongo-express":                  {Target: loongLocalRegistry + "mongo-express"},        // keep tag (mega216)
	"rediscommander/redis-commander": {Target: loongLocalRegistry + "redis-commander"},      // keep tag (mega216)
	"nodered/node-red":               {Target: loongLocalRegistry + "node-red"},             // keep tag (mega216)
	"linuxserver/qbittorrent":        {Target: loongLocalRegistry + "qbittorrent"},          // keep tag (mega216)
	"valkey/valkey":                  {Target: loongLocalRegistry + "valkey"},               // keep tag (mega216)
	"zabbix/zabbix-agent":            {Target: loongLocalRegistry + "zabbix-agent"},         // keep tag (mega216)
	// --- databases / queues ---
	"redis":     {Target: loongLCRRegistry + "library/redis", Fallback: "7.4.1", ByMajor: map[string]string{"5": "5.0.14", "6": "6.2.13", "7": "7.4.1", "8": "7.4.1"}},
	"mysql":     {Target: loongLCRRegistry + "library/mysql", Fallback: "8.4.3", ByMajor: map[string]string{"5": "5.7.29-an23", "8": "8.4.3"}},
	"mariadb":   {Target: loongLCRRegistry + "library/mariadb", Fallback: "13.1.1", ByMajor: map[string]string{"10": "10.5.18", "12": "13.1.1", "13": "13.1.1"}},
	"postgres":  {Target: loongLCRRegistry + "library/postgres", Fallback: "15.7-sid", ByMajor: map[string]string{"13": "13.19-alpine3.21", "14": "13.19-alpine3.21", "15": "15.7-sid", "16": "15.7-sid", "17": "17-sid", "18": "17-sid"}},
	"mongo":     {Target: loongLCRRegistry + "library/mongo", Fallback: "6.0.26"},
	"memcached": {Target: loongLCRRegistry + "library/memcached", Fallback: "1.6.21-alpine"},
	"rabbitmq":  {Target: loongLCRRegistry + "library/rabbitmq", Fallback: "4.1.2-management"},
	"emqx/emqx": {Target: loongLCRRegistry + "emqx/emqx", Fallback: "6.3.2"},
	"nats":      {Target: loongLCRRegistry + "library/nats", Fallback: "2.10.24-alpine"},

	// --- search / observability ---
	"elasticsearch": {Target: loongLCRRegistry + "library/elasticsearch", Fallback: "9.5.5", ByMajor: map[string]string{"9": "9.5.5"}},
	"kibana":        {Target: loongLCRRegistry + "library/kibana", Fallback: "9.4.2", ByMajor: map[string]string{"9": "9.4.2"}},
	"logstash":      {Target: loongLCRRegistry + "library/logstash", Fallback: "9.5.5", ByMajor: map[string]string{"9": "9.5.5"}},
	"influxdb":      {Target: loongLCRRegistry + "library/influxdb", Fallback: "3.9.6-alpine"},

	// --- web / cms ---
	"phpmyadmin":        {Target: loongLCRRegistry + "library/phpmyadmin", Fallback: "5.2.3"},
	"wordpress":         {Target: loongLCRRegistry + "library/wordpress", Fallback: "7.0.2-php8.5-fpm-alpine", ByMajor: map[string]string{"7": "7.0.2-php8.5-fpm-alpine"}},
	"yourls":            {Target: loongLCRRegistry + "library/yourls", Fallback: "1.10.6"},
	"crazymax/flarum":   {Target: loongLCRRegistry + "library/flarum", Fallback: "latest"},
	"postgis/postgis":   {Target: loongLCRRegistry + "postgis/postgis", Fallback: "16-3.4"},
	"pgvector/pgvector": {Target: loongLCRRegistry + "library/pgvector", Fallback: "0.8.0"},

	// --- ops / ci ---
	"registry":                 {Target: loongLCRRegistry + "library/registry", Fallback: "3.1.1", ByMajor: map[string]string{"3": "3.1.1"}},
	"jenkins/jenkins":          {Target: loongLCRRegistry + "jenkins/jenkins", Fallback: "2.583-jdk21"},
	"gitlab/gitlab-ce":         {Target: loongLCRRegistry + "gitlab/gitlab-ce", Fallback: "17.6.2"},
	"grafana/grafana":          {Target: loongLCRRegistry + "grafana/grafana", Fallback: "12.0.8", ByMajor: map[string]string{"13": "13.0.2-alpine", "9": "9.5.2"}},
	"jaegertracing/all-in-one": {Target: loongLCRRegistry + "jaegertracing/all-in-one", Fallback: "1.6.0"},
	"hashicorp/consul":         {Target: loongLCRRegistry + "library/consul", Fallback: "1.15.2"},
	"portainer/portainer-ce":   {Target: loongLCRRegistry + "portainer/portainer-ce", Fallback: "2.11.1"},

	// --- misc ---
	"neo4j":         {Target: loongLCRRegistry + "library/neo4j", Fallback: "5.26.32"},
	"2fauth/2fauth": {Target: loongLCRRegistry + "library/2fauth", Fallback: "latest"},
}

// basename fallback rules for base images referenced by any compose file
var loong64BaseRules = map[string]loong64Rule{
	"alpine":  {Target: loongLCRRegistry + "library/alpine", Fallback: "3.24"},
	"busybox": {Target: loongLCRRegistry + "library/busybox", Fallback: "latest"},
	"python":  {Target: loongLCRRegistry + "library/python", Fallback: "3.11.7"},
	"node":    {Target: loongLCRRegistry + "library/node", Fallback: "v20.19.3-debian"},
}

// repositories confirmed unportable; installs are rejected with a clear message
var loong64BlockedRepos = map[string]bool{
	"jellyfin/jellyfin": true, // .NET runtime has no official loong64 port
	// mega216 evaluation batch (2026-11-07), reasons per app
	"ghcr.io/666ghj/bettafish":            true, // Weibo-sentiment AI stack, heavy deps, no loong64 GPU support
	"linuxserver/code-server":             true, // VS Code build chain too heavy for loong64
	"1panel/comfyui":                      true, // requires torch, no official loong64 wheels
	"harness/harness":                     true, // closed-source CI platform, heavy stack
	"ghcr.io/ggml-org/llama.cpp":          true, // LLM inference impractical on 8G RAM CPU-only device
	"localai/localai":                     true, // x86-optimized llama.cpp backends, no practical loong64 use
	"mcr.microsoft.com/mssql/server":      true, // SQL Server has no loong64 distribution
	"jc21/nginx-proxy-manager":            true, // build chain too heavy; use built-in website module
	"ollama/ollama":                       true, // GPU/CUDA stack, no GPU here; CPU mode useless on 8G RAM
	"ghcr.io/open-webui/open-webui":       true, // depends on torch, no official loong64 wheels
	"openvpn/openvpn-as":                  true, // Access Server ships closed amd64/arm64 binaries only
	"thijsvanloef/palworld-server-docker": true, // steamcmd+wine x86 game server, meaningless on loong64
	"lmsysorg/sglang":                     true, // GPU inference framework, no loong64 build
	"teamspeak":                           true, // closed-source x86 binary
	"ghcr.io/tiny-craft/tiny-rdm":         true, // desktop GUI app wrapper, not fit for NAS use
	"shuzhiyinhang/vexdb":                 true, // bank closed-source vector DB, no public loong64 channel
	"dockurr/windows":                     true, // QEMU x86 guest needs KVM, no x86 virt on loong64
	"zhaojun1998/zfile":                   true, // ZFile 5.x closed source, x86/arm binaries only
}

// mapImagesForLoong64 rewrites service images in a parsed docker-compose map.
// It returns an error when a service references an image known to be
// unportable, so the caller can abort the install with a clear message.
func mapImagesForLoong64(composeMap map[string]interface{}) error {
	if runtime.GOARCH != "loong64" || composeMap == nil {
		return nil
	}
	servicesValue, ok := composeMap["services"]
	if !ok || servicesValue == nil {
		return nil
	}
	services, ok := servicesValue.(map[string]interface{})
	if !ok {
		return nil
	}
	// note: services MAY also carry a "build" section (e.g. openresty ships a
	// build context for dynamic modules alongside its published image); the
	// rule table below is an explicit allowlist, so mapping stays safe — the
	// rewritten image is present in the local registry and compose never
	// needs to invoke the build.
	for name, svcValue := range services {
		svc, ok := svcValue.(map[string]interface{})
		if !ok {
			continue
		}
		raw, ok := svc["image"].(string)
		if !ok {
			continue
		}
		image := strings.Trim(strings.TrimSpace(raw), `"'`)
		if isBlockedLoong64Image(image) {
			return errors.New(loong64UnsupportedMsg)
		}
		mapped, ok := mapLoong64Image(image)
		if !ok {
			continue
		}
		svc["image"] = mapped
		global.LOG.Infof("[loong64] app service %q image %s -> %s", name, image, mapped)
	}
	return nil
}

func isBlockedLoong64Image(image string) bool {
	repo, _ := splitImageRef(image)
	return loong64BlockedRepos[repo]
}

func mapLoong64Image(image string) (string, bool) {
	if image == "" || strings.Contains(image, "${") || strings.Contains(image, "loongnix") ||
		strings.HasPrefix(image, "127.0.0.1:") || strings.Contains(image, "@") {
		return "", false
	}
	repo, tag := splitImageRef(image)
	if rule, ok := loong64RepoRules[repo]; ok {
		return renderLoong64Rule(rule, tag), true
	}
	base := repo[strings.LastIndex(repo, "/")+1:]
	if rule, ok := loong64BaseRules[base]; ok {
		return renderLoong64Rule(rule, tag), true
	}
	return "", false
}

func renderLoong64Rule(rule loong64Rule, wantTag string) string {
	if rule.Fallback == "" {
		// keep requested tag
		return joinImageTag(rule.Target, wantTag)
	}
	if rule.ByMajor != nil && wantTag != "" {
		major := wantTag
		if dot := strings.IndexAny(wantTag, ".-"); dot > 0 {
			major = wantTag[:dot]
		}
		if byMajor, ok := rule.ByMajor[major]; ok {
			return joinImageTag(rule.Target, byMajor)
		}
	}
	return joinImageTag(rule.Target, rule.Fallback)
}

func splitImageRef(image string) (repo, tag string) {
	image = strings.Trim(strings.TrimSpace(image), `"'`)
	repo, tag = image, ""
	if idx := strings.LastIndex(image, ":"); idx > strings.LastIndex(image, "/") {
		repo, tag = image[:idx], image[idx+1:]
	}
	return repo, tag
}

func joinImageTag(repo, tag string) string {
	if tag == "" {
		return repo
	}
	return repo + ":" + tag
}
