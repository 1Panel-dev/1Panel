package service

import (
	"encoding/json"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

const (
	vllmAppKeyForUpgrade = "vllm"
	vllmImageEnvKey      = "IMAGE"
	vllmImageTypeNvidia  = "nvidia"
	vllmImageTypeIntel   = "intel"
	vllmImageTypeAscend  = "ascend"
	vllmImageTypeGB10    = "nvidia-gb10-dspark"
)

func resolveVllmVersionFamily(version, image string) string {
	normalizedVersion := strings.ToLower(strings.TrimSpace(version))
	if strings.HasPrefix(normalizedVersion, vllmImageTypeGB10+"-") {
		return vllmImageTypeGB10
	}
	if strings.HasPrefix(normalizedVersion, vllmImageTypeIntel+"-") {
		return vllmImageTypeIntel
	}
	if strings.HasPrefix(normalizedVersion, vllmImageTypeAscend+"-") {
		return vllmImageTypeAscend
	}
	if strings.HasPrefix(normalizedVersion, vllmImageTypeNvidia+"-") {
		return vllmImageTypeNvidia
	}
	normalizedImage := strings.ToLower(strings.TrimSpace(image))
	if strings.Contains(normalizedImage, "vllm-gb10-dspark") {
		return vllmImageTypeGB10
	}
	if strings.Contains(normalizedImage, "intel/") || strings.Contains(normalizedImage, "llm-scaler-vllm") {
		return vllmImageTypeIntel
	}
	if strings.Contains(normalizedImage, "ascend/") || strings.Contains(normalizedImage, "vllm-ascend") {
		return vllmImageTypeAscend
	}
	return vllmImageTypeNvidia
}

func trimVllmVersionFamily(version string) string {
	trimmed := strings.TrimSpace(version)
	normalized := strings.ToLower(trimmed)
	for _, family := range []string{vllmImageTypeGB10, vllmImageTypeNvidia, vllmImageTypeIntel, vllmImageTypeAscend} {
		prefix := family + "-"
		if strings.HasPrefix(normalized, prefix) {
			return strings.TrimSpace(trimmed[len(prefix):])
		}
	}
	return trimmed
}

func buildDefaultVllmImageByVersion(version string) string {
	tag := trimVllmVersionFamily(version)
	family := resolveVllmVersionFamily(version, "")
	if family == vllmImageTypeGB10 {
		return "1panel/vllm-gb10-dspark:" + tag
	}
	if family == vllmImageTypeIntel {
		return "intel/llm-scaler-vllm:" + tag
	}
	if tag != "" && !strings.HasPrefix(strings.ToLower(tag), "v") {
		tag = "v" + tag
	}
	if family == vllmImageTypeAscend {
		return "quay.io/ascend/vllm-ascend:" + tag
	}
	return "vllm/vllm-openai:" + tag
}

func isVllmUpgradeVersionAllowed(currentVersion, targetVersion, currentImage string) bool {
	currentFamily := resolveVllmVersionFamily(currentVersion, currentImage)
	targetFamily := resolveVllmVersionFamily(targetVersion, "")
	return currentFamily == targetFamily
}

func hasVllmVersionFamilyPrefix(version string) bool {
	normalized := strings.ToLower(strings.TrimSpace(version))
	return strings.HasPrefix(normalized, vllmImageTypeGB10+"-") ||
		strings.HasPrefix(normalized, vllmImageTypeNvidia+"-") ||
		strings.HasPrefix(normalized, vllmImageTypeIntel+"-") ||
		strings.HasPrefix(normalized, vllmImageTypeAscend+"-")
}

func isVllmUpgradeCandidate(currentVersion, targetVersion, currentImage string) bool {
	if strings.TrimSpace(currentVersion) == strings.TrimSpace(targetVersion) {
		return false
	}
	if !isVllmUpgradeVersionAllowed(currentVersion, targetVersion, currentImage) {
		return false
	}
	if common.CompareVersion(targetVersion, currentVersion) {
		return true
	}
	return !hasVllmVersionFamilyPrefix(currentVersion) &&
		resolveVllmVersionFamily(targetVersion, "") == vllmImageTypeNvidia &&
		trimVllmVersionFamily(currentVersion) == trimVllmVersionFamily(targetVersion)
}

func buildVllmUpgradeImage(currentImage, currentVersion, targetVersion string) string {
	trimmedImage := strings.TrimSpace(currentImage)
	if trimmedImage == "" || trimmedImage == buildDefaultVllmImageByVersion(currentVersion) {
		return buildDefaultVllmImageByVersion(targetVersion)
	}
	return trimmedImage
}

func loadVllmImageFromEnv(raw string) string {
	envs := make(map[string]interface{})
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	if err := json.Unmarshal([]byte(raw), &envs); err != nil {
		return ""
	}
	if image, ok := envs[vllmImageEnvKey].(string); ok {
		return strings.TrimSpace(image)
	}
	return ""
}

func setVllmImageInEnvContent(content []byte, image string) []byte {
	normalizedImage := strings.TrimSpace(image)
	if normalizedImage == "" {
		return content
	}
	lines := strings.Split(string(content), "\n")
	replaced := false
	for index, line := range lines {
		if strings.HasPrefix(line, vllmImageEnvKey+"=") {
			lines[index] = vllmImageEnvKey + "=" + normalizedImage
			replaced = true
			break
		}
	}
	if !replaced {
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines[len(lines)-1] = vllmImageEnvKey + "=" + normalizedImage
			lines = append(lines, "")
		} else {
			lines = append(lines, vllmImageEnvKey+"="+normalizedImage)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}
