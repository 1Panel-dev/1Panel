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
	FilePath   string             `json:"filePath"`
	Config     *components.Config `json:"config"`
	OldContent string             `json:"oldContent"`
}

type NginxConfigReq struct {
	Scope     NginxKey    `json:"scope"`
	Operate   NginxOp     `json:"operate"`
	WebsiteID uint        `json:"webSiteId"`
	Params    interface{} `json:"params"`
}

type NginxScopeReq struct {
	Scope NginxKey `json:"scope"`
}

type NginxStatus struct {
	Active   string `json:"active"`
	Accepts  string `json:"accepts"`
	Handled  string `json:"handled"`
	Requests string `json:"requests"`
	Reading  string `json:"reading"`
	Writing  string `json:"writing"`
	Waiting  string `json:"waiting"`
}

type NginxKey string

const (
	Index     NginxKey = "index"
	LimitConn NginxKey = "limit-conn"
	SSL       NginxKey = "ssl"
	HttpPer   NginxKey = "http-per"
)

type NginxOp string

const (
	ConfigNew    NginxOp = "add"
	ConfigUpdate NginxOp = "update"
	ConfigDel    NginxOp = "delete"
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

type NginxParam struct {
	UpdateScope string   `json:"scope"`
	Name        string   `json:"name"`
	Params      []string `json:"params"`
}
