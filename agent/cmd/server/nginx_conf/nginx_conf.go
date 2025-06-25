package nginx_conf

import (
	"embed"
	_ "embed"
	"io"
)

//go:embed ssl.conf
var SSL []byte

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

//go:embed path_auth.conf
var PathAuth []byte

//go:embed upstream.conf
var Upstream []byte

//go:embed sse.conf
var SSE []byte

//go:embed *.json *.conf
var websitesFiles embed.FS

func GetWebsiteFile(filename string) []byte {
	file, err := websitesFiles.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()
	res, _ := io.ReadAll(file)
	return res
}
