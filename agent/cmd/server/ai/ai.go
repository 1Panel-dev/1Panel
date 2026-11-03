package ai

import (
	_ "embed"
)

//go:embed compose.yml
var DefaultMcpCompose []byte
