package dto

import "github.com/1Panel-dev/1Panel/backend/utils/nginx/components"

type NginxConfig struct {
	FilePath      string             `json:"filePath"`
	ContainerName string             `json:"containerName"`
	Config        *components.Config `json:"config"`
	OldContent    string             `json:"oldContent"`
}

type NginxConfigReq struct {
	Scope     NginxScope  `json:"scope"`
	Operate   NginxOp     `json:"operate"`
	WebSiteID uint        `json:"webSiteId" validate:"required"`
	Params    interface{} `json:"params"`
}

type NginxScope string

const (
	Index     NginxScope = "index"
	LimitConn NginxScope = "limit-conn"
)

type NginxOp string

const (
	ConfigNew    NginxOp = "add"
	ConfigUpdate NginxOp = "update"
	ConfigDel    NginxOp = "delete"
)

var ScopeKeyMap = map[NginxScope][]string{
	Index:     {"index"},
	LimitConn: {"limit_conn", "limit_rate", "limit_conn_zone"},
}

var RepeatKeys = map[string]struct {
}{
	"limit_conn":      {},
	"limit_conn_zone": {},
}

type NginxParam struct {
	Name        string   `json:"name"`
	SecondKey   string   `json:"secondKey"`
	IsRepeatKey bool     `json:"isRepeatKey"`
	Params      []string `json:"params"`
}
