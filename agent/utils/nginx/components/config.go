package components

type Config struct {
	*Block
	FilePath string
}

func (c *Config) FindServers() []*Server {
	var servers []*Server
	directives := c.Block.FindDirectives("server")
	for _, directive := range directives {
		servers = append(servers, directive.(*Server))
	}
	return servers
}

func (c *Config) FindHttp() *Http {
	var http *Http
	directives := c.Block.FindDirectives("http")
	if len(directives) > 0 {
		http = directives[0].(*Http)
	}

	return http
}

func (c *Config) FindUpstreams() []*Upstream {
	var upstreams []*Upstream
	directives := c.Block.FindDirectives("upstream")
	for _, directive := range directives {
		upstreams = append(upstreams, directive.(*Upstream))
	}
	return upstreams
}

var repeatKeys = map[string]struct {
}{
	"limit_conn":       {},
	"limit_conn_zone":  {},
	"set":              {},
	"if":               {},
	"proxy_set_header": {},
	"location":         {},
	"include":          {},
	"sub_filter":       {},
	"add_header":       {},
}

func IsRepeatKey(key string) bool {
	if _, ok := repeatKeys[key]; ok {
		return true
	}
	return false
}
