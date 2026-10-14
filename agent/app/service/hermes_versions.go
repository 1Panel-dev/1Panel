package service

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

// Release names at https://github.com/NousResearch/hermes-agent/releases pair
// these historical calendar tags with their semantic versions. This includes
// every calendar version published in the 1Panel app store. Do not rewrite
// stored versions: backup metadata, restore and rollback retain the original tag.
var hermesHistoricalVersions = map[string]string{
	"2026.3.12":   "0.2.0",
	"2026.3.17":   "0.3.0",
	"2026.3.23":   "0.4.0",
	"2026.3.28":   "0.5.0",
	"2026.3.30":   "0.6.0",
	"2026.4.3":    "0.7.0",
	"2026.4.8":    "0.8.0",
	"2026.4.13":   "0.9.0",
	"2026.4.16":   "0.10.0",
	"2026.4.23":   "0.11.0",
	"2026.4.30":   "0.12.0",
	"2026.5.7":    "0.13.0",
	"2026.5.16":   "0.14.0",
	"2026.5.28":   "0.15.0",
	"2026.5.29":   "0.15.1",
	"2026.5.29.2": "0.15.2",
	"2026.6.5":    "0.16.0",
	"2026.6.19":   "0.17.0",
	"2026.7.1":    "0.18.0",
	"2026.7.7":    "0.18.1",
	"2026.7.7.2":  "0.18.2",
	"2026.7.20":   "0.19.0",
	"2026.7.30":   "0.19.1",
	"2026.8.3":    "0.20.0",
	"2026.8.13":   "0.20.1",
	"2026.8.16":   "0.20.2",
	"2026.8.16.2": "0.20.3",
	"2026.8.18":   "0.20.4",
	"2026.8.19":   "0.20.5",
	"2026.8.27":   "0.20.6",
	"2026.8.31":   "0.21.0",
	"2026.9.7":    "0.21.1",
	"2026.9.11":   "0.21.2",
	"2026.9.14":   "0.21.3",
	"2026.9.21":   "0.21.4",
	"2026.9.24":   "0.21.5",
}

var hermesReleaseVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func isHermesApp(key string) bool {
	return key == constant.AppHermesAgent || key == "local"+constant.AppHermesAgent
}

func hermesComparableVersion(version string) (string, bool) {
	version = strings.TrimPrefix(version, "v")
	if mapped, ok := hermesHistoricalVersions[version]; ok {
		return mapped, true
	}
	parts := hermesReleaseVersion.FindStringSubmatch(version)
	if len(parts) == 0 {
		return "", false
	}
	for i, part := range parts[1:] {
		n, err := strconv.Atoi(part)
		if err != nil || (i == 0 && n >= 2000) {
			// An unknown calendar tag must not become a huge semantic major.
			return "", false
		}
	}
	return version, true
}

func isNewerAppVersion(key, target, current string) bool {
	if !isHermesApp(key) {
		return common.CompareVersion(target, current)
	}
	targetVersion, targetOK := hermesComparableVersion(target)
	currentVersion, currentOK := hermesComparableVersion(current)
	return targetOK && currentOK && common.CompareVersion(targetVersion, currentVersion)
}

func isCrossAppVersion(key, current, target string) bool {
	if !isHermesApp(key) {
		return common.IsCrossVersion(current, target)
	}
	currentVersion, currentOK := hermesComparableVersion(current)
	targetVersion, targetOK := hermesComparableVersion(target)
	return currentOK && targetOK && common.IsCrossVersion(currentVersion, targetVersion)
}

func appVersionSortBefore(key, left, right string) bool {
	if !isHermesApp(key) {
		return common.CompareVersion(left, right)
	}
	lv, lok := hermesComparableVersion(left)
	rv, rok := hermesComparableVersion(right)
	if lok != rok {
		return lok
	}
	if !lok {
		return common.CompareVersion(left, right)
	}
	if lv != rv {
		return common.CompareVersion(lv, rv)
	}
	// Prefer the official semantic spelling when both aliases are available.
	return strings.TrimPrefix(left, "v") == lv && strings.TrimPrefix(right, "v") != rv
}

func sortAppVersions(key string, versions []string) []string {
	if !isHermesApp(key) {
		return common.GetSortedVersions(versions)
	}
	sort.SliceStable(versions, func(i, j int) bool {
		return appVersionSortBefore(key, versions[i], versions[j])
	})
	return versions
}
