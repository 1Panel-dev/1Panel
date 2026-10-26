package mcpruntime

import (
	"regexp"

	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
)

const (
	ContractVersion = contract.ContractVersion
	Protocol        = "streamable-http"
	DefaultPort     = 18081
	maxCatalogBytes = 2 << 20
)

type State struct {
	Installed      bool
	Status         string
	Ready          bool
	Reason         string
	AdapterVersion string
	Image          string
	ImageID        string
}

var catalogSlots = make(chan struct{}, 8)
var versionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?$`)
var bearerPattern = regexp.MustCompile(`(?i)bearer\s+[^\s"']+`)
var sensitivePattern = regexp.MustCompile(`(?i)("?(?:password|token|secret|authorization)"?\s*[:=]\s*)("[^"]*"|[^\s,}]+)`)

func redactLog(value string) string {
	return sensitivePattern.ReplaceAllString(bearerPattern.ReplaceAllString(value, "Bearer [REDACTED]"), "${1}[REDACTED]")
}
