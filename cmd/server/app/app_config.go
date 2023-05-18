package app

import (
	_ "embed"
)

//go:embed app_config.yml
var Config []byte

//go:embed logo.png
var Logo []byte

//go:embed app_param.yml
var Param []byte
