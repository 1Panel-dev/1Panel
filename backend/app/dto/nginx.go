package dto

import "github.com/1Panel-dev/1Panel/backend/utils/nginx/components"

type NginxConfig struct {
	FilePath      string             `json:"filePath"`
	ContainerName string             `json:"containerName"`
	Config        *components.Config `json:"config"`
	OldContent    string             `json:"oldContent"`
}

type NginxConfigReq struct {
	Scope     NginxKey    `json:"scope"`
	Operate   NginxOp     `json:"operate"`
	WebSiteID uint        `json:"webSiteId"`
	Params    interface{} `json:"params"`
}

type NginxScopeReq struct {
	Scope NginxKey `json:"scope"`
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

type NginxScope string

const (
	NginxHttp   NginxScope = "http"
	NginxServer NginxScope = "server"
	NginxEvents NginxScope = "events"
)

var RepeatKeys = map[string]struct {
}{
	"limit_conn":      {},
	"limit_conn_zone": {},
}

type NginxParam struct {
	Name      string   `json:"name"`
	SecondKey string   `json:"secondKey"`
	Params    []string `json:"params"`
}
