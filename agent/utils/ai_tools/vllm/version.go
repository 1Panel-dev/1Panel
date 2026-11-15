package vllm

import "strings"

const (
	gb10VersionPrefix = "nvidia-"
	rtxVersionSuffix  = "-optimized-for-ds-v4.1-flash"
)

func IsRTXVersion(version string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(version)), rtxVersionSuffix)
}

func IsGB10Version(version string) bool {
	return GB10NodeCount(version) > 0
}

func GB10NodeCount(version string) int {
	normalized := strings.ToLower(strings.TrimSpace(version))
	tag := strings.TrimPrefix(normalized, gb10VersionPrefix)
	if tag == normalized {
		return 0
	}
	for suffix, nodes := range map[string]int{
		"-optimized-for-2-gb10-ds-v4-flash":   2,
		"-optimized-for-4-gb10-ds-v4.1-flash": 4,
	} {
		if strings.HasSuffix(tag, suffix) && strings.TrimSpace(strings.TrimSuffix(tag, suffix)) != "" {
			return nodes
		}
	}
	return 0
}

func IsGB10TileLangVersion(version string) bool {
	tag := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(version)), gb10VersionPrefix)
	return strings.TrimPrefix(tag, "v") == "2026.11.15-optimized-for-4-gb10-ds-v4.1-flash"
}

func ResolveImage(image, imageType, tag string) string {
	image = strings.TrimSpace(image)
	tag = strings.TrimSpace(tag)
	defaultImage := ""
	switch imageType {
	case "intel":
		defaultImage = "intel/llm-scaler-vllm:" + tag
	case "ascend":
		if tag != "" && !strings.HasPrefix(strings.ToLower(tag), "v") {
			tag = "v" + tag
		}
		defaultImage = "quay.io/ascend/vllm-ascend:" + tag
	default:
		if tag != "" && !strings.HasPrefix(strings.ToLower(tag), "v") {
			tag = "v" + tag
		}
		defaultImage = "vllm/vllm-openai:" + tag
		_, suffix, _ := strings.Cut(tag, "-")
		switch suffix {
		case "optimized-for-ds-v4-flash":
			defaultImage = "1panel/vllm-gb10-dual-node:" + strings.TrimPrefix(tag, "v")
		case "optimized-for-2-gb10-ds-v4-flash":
			defaultImage = "1panel/vllm:" + strings.TrimPrefix(tag, "v")
		case "optimized-for-4-gb10-ds-v4.1-flash":
			defaultImage = "eugr/spark-vllm-b12x:latest"
			if IsGB10TileLangVersion(tag) {
				defaultImage = "spark-ds41f:1107"
			}
		case "optimized-for-ds-v4.1-flash":
			defaultImage = "1panel/vllm:" + tag
		}
		if defaultImage != "vllm/vllm-openai:"+tag && image == "vllm/vllm-openai:"+tag {
			image = ""
		}
	}
	if image == "" {
		return defaultImage
	}
	return image
}
