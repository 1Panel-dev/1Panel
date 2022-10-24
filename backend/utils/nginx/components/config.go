package components

type Config struct {
	*Block
	FilePath string
}

func (c *Config) FindDirectives(directiveName string) []IDirective {
	return c.Block.FindDirectives(directiveName)
}

func (c *Config) FindUpstreams() []*Upstream {
	var upstreams []*Upstream
	directives := c.Block.FindDirectives("upstream")
	for _, directive := range directives {
		upstreams = append(upstreams, directive.(*Upstream))
	}
	return upstreams
}

func (c *Config) FindServers() []*Server {
	var servers []*Server
	directives := c.Block.FindDirectives("server")
	for _, directive := range directives {
		servers = append(servers, directive.(*Server))
	}
	return servers
}
