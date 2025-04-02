package mcp

import (
	_ "embed"
)

//go:embed compose.yml
var DefaultMcpCompose []byte
