package service

import (
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

const aiGatewayAppKey = "ai-gateway"

func isAIGatewayApp(appKey string) bool {
	return appKey == aiGatewayAppKey || appKey == "local"+aiGatewayAppKey
}

func isAIGatewayPostgresVersion(version string) bool {
	return strings.HasSuffix(version, "-pg") || strings.HasSuffix(version, "-postgres")
}

// Database variants are deployment choices, not interchangeable upgrades.
func isAIGatewayUpgradeVersionAllowed(appKey, currentVersion, targetVersion string) bool {
	return !isAIGatewayApp(appKey) ||
		isAIGatewayPostgresVersion(currentVersion) == isAIGatewayPostgresVersion(targetVersion)
}

func sortAIGatewayVersions(versions []string) {
	sort.SliceStable(versions, func(i, j int) bool {
		if common.CompareVersion(versions[i], versions[j]) {
			return true
		}
		if common.CompareVersion(versions[j], versions[i]) {
			return false
		}
		// Prefer SQLite only when the numeric versions are equal.
		return !isAIGatewayPostgresVersion(versions[i]) && isAIGatewayPostgresVersion(versions[j])
	})
}
