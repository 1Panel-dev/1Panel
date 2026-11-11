package vllm

import "strings"

const (
	gb10VersionPrefix = "nvidia-"
	gb10VersionSuffix = "-optimized-for-2-gb10-ds-v4-flash"
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
