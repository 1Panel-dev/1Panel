package dto

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
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

type NginxKey string

const (
	Index     NginxKey = "index"
	LimitConn NginxKey = "limit-conn"
	SSL       NginxKey = "ssl"
	HttpPer   NginxKey = "http-per"
)

var ScopeKeyMap = map[NginxKey][]string{
	Index:     {"index"},
	LimitConn: {"limit_conn", "limit_rate", "limit_conn_zone"},
	SSL:       {"ssl_certificate", "ssl_certificate_key"},
	HttpPer:   {"server_names_hash_bucket_size", "client_header_buffer_size", "client_max_body_size", "keepalive_timeout", "gzip", "gzip_min_length", "gzip_comp_level"},
}

var StaticFileKeyMap = map[NginxKey]struct {
}{
	SSL:       {},
	LimitConn: {},
}
