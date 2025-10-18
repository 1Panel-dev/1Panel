package ai

import (
	_ "embed"
)

//go:embed compose.yml
var DefaultMcpCompose []byte

//go:embed llm-compose.yml
var DefaultTensorrtLLMCompose []byte
