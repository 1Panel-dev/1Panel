package vllm

import "strings"

const (
	gb10VersionPrefix = "nvidia-"
	gb10VersionSuffix = "-optimized-for-ds-v4-flash"
	rtxVersionSuffix  = "-optimized-for-ds-v4.1-flash"
)

func IsRTXVersion(version string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(version)), rtxVersionSuffix)
}

func IsGB10Version(version string) bool {
	normalized := strings.ToLower(strings.TrimSpace(version))
	tag := strings.TrimPrefix(normalized, gb10VersionPrefix)
	return tag != normalized && strings.HasSuffix(tag, gb10VersionSuffix) &&
		strings.TrimSpace(strings.TrimSuffix(tag, gb10VersionSuffix)) != ""
}

func GB10Image(version string) string {
	trimmed := strings.TrimSpace(version)
	if !IsGB10Version(trimmed) {
		return ""
	}
	return "1panel/vllm-gb10-dual-node:" + strings.TrimSpace(trimmed[len(gb10VersionPrefix):])
}
