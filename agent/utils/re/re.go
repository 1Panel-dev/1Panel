package re

import (
	"fmt"
	"regexp"
)

const (
	NumberAlphaPattern                 = `(\d+)([A-Za-z]+)`
	ComposeDisallowedCharsPattern      = `[^a-z0-9_-]+`
	ComposeEnvVarPattern               = `\$\{([^}]+)\}`
	DiskKeyValuePattern                = `([A-Za-z0-9_]+)=("([^"\\]|\\.)*"|[^ \t]+)`
	FirewalldForwardPattern            = `^port=(\d{1,5}):proto=(.+?):toport=(\d{1,5}):toaddr=(.*)$`
	ValidatorNamePattern               = `^[a-zA-Z\p{Han}]{1}[a-zA-Z0-9_\p{Han}]{0,30}$`
	ValidatorIPPattern                 = `^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}$`
	DomainPattern                      = `^([\w\p{Han}\-\*]{1,100}\.){1,10}([\w\p{Han}\-]{1,24}|[\w\p{Han}\-]{1,24}\.[\w\p{Han}\-]{1,24})(:\d{1,5})?$`
	ProxyCacheZonePattern              = `keys_zone=proxy_cache_zone_of_[\w.]+:(\d+)([kmgt]?)`
	ProxyCacheMaxSizePattern           = `max_size=([0-9.]+)([kmgt]?)`
	ProxyCacheMaxSizeValidationPattern = `max_size=\d+(\.\d+)?[kmgt]?`
	ProxyCacheInactivePattern          = `inactive=(\d+)([smhd])`
	NumberWordPattern                  = `^(\d+)(\w+)$`
	TrailingDigitsPattern              = `_(\d+)$`
	AlertIPPattern                     = `from\s+([0-9.]+)\s+port\s+(\d+)`
	CosDualStackPattern                = `.*cos-dualstack\..*`
	VersionPattern                     = `v(\d+\.\d+\.\d+)`
	PhpAssignmentPattern               = `^\s*([a-z_]+)\s*=\s*(.*)$`
	DurationWithOptionalUnitPattern    = `^(\d+)([smhdw]?)$`
	MysqlGroupPattern                  = `\[*\]`
	AnsiEscapePattern                  = "\x1b\\[[0-9;?]*[A-Za-z]|\x1b=|\x1b>"
	RecycleBinFilePattern              = `_1p_file_1p_(.+)_p_(\d+)_(\d+)`
)

var regexMap = make(map[string]*regexp.Regexp)

// InitRegex compiles all regex patterns and stores them in the map.
// This function should be called once at program startup.
func Init() {
	patterns := []string{
		NumberAlphaPattern,
		ComposeDisallowedCharsPattern,
		ComposeEnvVarPattern,
		DiskKeyValuePattern,
		FirewalldForwardPattern,
		ValidatorNamePattern,
		ValidatorIPPattern,
		DomainPattern,
		ProxyCacheZonePattern,
		ProxyCacheMaxSizePattern,
		ProxyCacheMaxSizeValidationPattern,
		ProxyCacheInactivePattern,
		NumberWordPattern,
		TrailingDigitsPattern,
		AlertIPPattern,
		CosDualStackPattern,
		VersionPattern,
		PhpAssignmentPattern,
		DurationWithOptionalUnitPattern,
		MysqlGroupPattern,
		AnsiEscapePattern,
		RecycleBinFilePattern,
	}

	for _, pattern := range patterns {
		regexMap[pattern] = regexp.MustCompile(pattern)
	}
}

// GetRegex retrieves a compiled regex by its pattern string.
// Panics if the pattern is not found in the map.
func GetRegex(pattern string) *regexp.Regexp {
	regex, exists := regexMap[pattern]
	if !exists {
		panic(fmt.Sprintf("regex pattern not found: %s", pattern))
	}
	return regex
}

// RegisterRegex registers a regex pattern and stores it in the map.
// This function should be called once at program startup.
func RegisterRegex(pattern string) {
	regexMap[pattern] = regexp.MustCompile(pattern)
}
