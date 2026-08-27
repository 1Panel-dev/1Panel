package sync

import (
	"slices"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func SupportsManagedOrder(provider filter.Provider) bool {
	return provider == filter.ProviderIptables || provider == filter.ProviderNftables || provider == filter.ProviderUFW
}

func ManagedOrderDrift(snapshot filter.Snapshot, desiredMarkers []string) (map[string]struct{}, bool) {
	if !SupportsManagedOrder(snapshot.Scope.Provider) || len(desiredMarkers) < 2 {
		return nil, true
	}
	expected := make(map[string]struct{}, len(desiredMarkers))
	for _, marker := range desiredMarkers {
		expected[marker] = struct{}{}
	}
	actual := make([]string, 0, len(desiredMarkers))
	segments := make(map[string]int, len(desiredMarkers))
	segment := 0
	for _, observed := range snapshot.Rules {
		_, wanted := expected[observed.Marker]
		if wanted {
			actual = append(actual, observed.Marker)
			if observed.Protected || observed.ParseStatus == filter.ParseStatusOpaque {
				segment++
				segments[observed.Marker] = segment
				segment++
			} else {
				segments[observed.Marker] = segment
			}
			continue
		}
		if strings.HasPrefix(observed.Marker, "1panel-rule:") &&
			!observed.Protected && observed.ParseStatus != filter.ParseStatusOpaque {
			continue
		}
		segment++
	}

	desired := make([]string, 0, len(actual))
	for _, marker := range desiredMarkers {
		if _, exists := segments[marker]; exists {
			desired = append(desired, marker)
		}
	}
	if slices.Equal(actual, desired) {
		return nil, true
	}
	drifted := make(map[string]struct{}, len(desired))
	for index := range desired {
		if actual[index] != desired[index] {
			drifted[actual[index]] = struct{}{}
			drifted[desired[index]] = struct{}{}
		}
	}
	feasible, previousSegment := true, -1
	for _, marker := range desired {
		if segments[marker] < previousSegment {
			feasible = false
			break
		}
		previousSegment = segments[marker]
	}
	return drifted, feasible
}

func InsertionPosition(snapshot filter.Snapshot, desiredMarkers []string, targetMarker string) (int64, bool) {
	if !SupportsManagedOrder(snapshot.Scope.Provider) {
		return 0, false
	}
	targetIndex := slices.Index(desiredMarkers, targetMarker)
	if targetIndex < 0 {
		return 0, false
	}
	for index := targetIndex - 1; index >= 0; index-- {
		if _, position, exists := ObservedByMarker(snapshot, desiredMarkers[index]); exists {
			return int64(position + 1), true
		}
	}
	for index := targetIndex + 1; index < len(desiredMarkers); index++ {
		if _, position, exists := ObservedByMarker(snapshot, desiredMarkers[index]); exists {
			return int64(position), true
		}
	}
	return 0, false
}

func NextManagedOrderChange(snapshot filter.Snapshot, desiredMarkers []string) (string, int, bool, error) {
	expected := make(map[string]struct{}, len(desiredMarkers))
	for _, marker := range desiredMarkers {
		expected[marker] = struct{}{}
	}
	actual := make([]string, 0, len(desiredMarkers))
	positions := make([]int, 0, len(desiredMarkers))
	for index, observed := range snapshot.Rules {
		if _, exists := expected[observed.Marker]; !exists {
			continue
		}
		actual = append(actual, observed.Marker)
		position := index + 1
		if observed.Locator.Position != nil {
			position = *observed.Locator.Position
		}
		positions = append(positions, position)
	}
	if len(actual) != len(desiredMarkers) {
		return "", 0, false, filter.ErrRuleStale
	}
	for index := range desiredMarkers {
		if actual[index] != desiredMarkers[index] {
			return desiredMarkers[index], positions[index], false, nil
		}
	}
	return "", 0, true, nil
}

func ObservedByMarker(snapshot filter.Snapshot, marker string) (filter.ObservedRule, int, bool) {
	for index, observed := range snapshot.Rules {
		if observed.Marker == marker {
			position := index + 1
			if observed.Locator.Position != nil {
				position = *observed.Locator.Position
			}
			return observed, position, true
		}
	}
	return filter.ObservedRule{}, 0, false
}

func ObservedRule(observed filter.ObservedRule) filter.FirewallRule {
	rule := observed.Rule
	if rule.UUID == "" && strings.HasPrefix(observed.Marker, "1panel-rule:") {
		rule.UUID = strings.TrimSpace(strings.TrimPrefix(observed.Marker, "1panel-rule:"))
	}
	return rule
}
