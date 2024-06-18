package nginx_conf

import (
	"embed"
	_ "embed"
)

//go:embed ssl.conf
var SSL []byte

//go:embed  website_default.conf
var WebsiteDefault []byte

//go:embed index.html
var Index []byte

//go:embed index.php
var IndexPHP []byte

//go:embed rewrite/*
var Rewrites embed.FS

//go:embed cache.conf
var Cache []byte

//go:embed proxy.conf
var Proxy []byte

//go:embed proxy_cache.conf
var ProxyCache []byte

//go:embed 404.html
var NotFoundHTML []byte

//go:embed domain404.html
var DomainNotFoundHTML []byte

//go:embed stop.html
var StopHTML []byte
