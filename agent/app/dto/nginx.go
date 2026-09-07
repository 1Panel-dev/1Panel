package dto

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
)

type NginxFull struct {
	Install    model.AppInstall
	Website    model.Website
	ConfigDir  string
	ConfigFile string
	SiteDir    string
	Dir        string
	RootConfig NginxConfig
	SiteConfig NginxConfig
}

type NginxConfig struct {
	FilePath   string
	Config     *components.Config
	OldContent string
}

type NginxParam struct {
	UpdateScope string
	Name        string
	Params      []string
}

type NginxAuth struct {
	Username string `json:"username"`
	Remark   string `json:"remark"`
}

type NginxPathAuth struct {
	Username string `json:"username"`
	Remark   string `json:"remark"`
	Path     string `json:"path"`
	Name     string `json:"name"`
}

type NginxKey string

const (
	Index      NginxKey = "index"
	LimitConn  NginxKey = "limit-conn"
	SSL        NginxKey = "ssl"
	CACHE      NginxKey = "cache"
	HttpPer    NginxKey = "http-per"
	ProxyCache NginxKey = "proxy-cache"
	Brotli     NginxKey = "brotli"
)

// BrotliKeys are served from the panel-managed http.d file rather than
// nginx.conf, because the module is optional: its directives must disappear
// together with the module, otherwise nginx refuses to start.
var BrotliKeys = []string{"brotli", "brotli_comp_level", "brotli_min_length", "brotli_types"}

var ScopeKeyMap = map[NginxKey][]string{
	Index:     {"index"},
	LimitConn: {"limit_conn", "limit_rate", "limit_conn_zone"},
	SSL:       {"ssl_certificate", "ssl_certificate_key"},
	HttpPer:   {"server_names_hash_bucket_size", "client_header_buffer_size", "client_max_body_size", "keepalive_timeout", "gzip", "gzip_min_length", "gzip_comp_level", "gzip_types", "gzip_vary", "gzip_proxied"},
}

var StaticFileKeyMap = map[NginxKey]struct {
}{
	SSL:        {},
	CACHE:      {},
	ProxyCache: {},
}

type NginxUpstream struct {
	Name      string                `json:"name"`
	Algorithm string                `json:"algorithm"`
	Servers   []NginxUpstreamServer `json:"servers"`
	Content   string                `json:"content"`
}

type NginxUpstreamServer struct {
	Server          string `json:"server"`
	Weight          int    `json:"weight"`
	FailTimeout     int    `json:"failTimeout"`
	FailTimeoutUnit string `json:"failTimeoutUnit"`
	MaxFails        int    `json:"maxFails"`
	MaxConns        int    `json:"maxConns"`
	Flag            string `json:"flag"`
}

var LBAlgorithms = map[string]struct{}{"ip_hash": {}, "least_conn": {}}

var RealIPKeys = map[string]struct{}{"X-Forwarded-For": {}, "X-Real-IP": {}, "CF-Connecting-IP": {}}

type NginxModule struct {
	Name      string             `json:"name"`
	Custom    bool               `json:"custom,omitempty"`
	Script    string             `json:"script"`
	Packages  []string           `json:"packages"`
	Params    string             `json:"params"`
	Enable    bool               `json:"enable"`
	BuildMode string             `json:"buildMode,omitempty"`
	Provider  string             `json:"provider,omitempty"`
	LoadOrder int                `json:"loadOrder,omitempty"`
	Builds    []NginxModuleBuild `json:"builds,omitempty"`
	LastError string             `json:"lastError,omitempty"`
}

type NginxModuleBuild struct {
	Provider  string                `json:"provider"`
	BuildMode string                `json:"buildMode"`
	Status    string                `json:"status"`
	Hash      string                `json:"hash"`
	Target    NginxModuleTarget     `json:"target"`
	Artifacts []NginxModuleArtifact `json:"artifacts,omitempty"`
	Error     string                `json:"error,omitempty"`
	BuiltAt   time.Time             `json:"builtAt,omitempty"`
}

type NginxModuleTarget struct {
	Key              string `json:"key"`
	OpenRestyVersion string `json:"openrestyVersion"`
	Architecture     string `json:"architecture"`
	Image            string `json:"image,omitempty"`
	ImageDigest      string `json:"imageDigest,omitempty"`
	BuilderDigest    string `json:"builderDigest,omitempty"`
}

type NginxModuleArtifact struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}
