package nginx_conf

import (
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
