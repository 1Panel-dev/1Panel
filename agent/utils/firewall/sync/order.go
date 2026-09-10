package sync

import (
	"slices"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func RuleOrder(snapshot filter.Snapshot, ordered []filter.InventoryItem) map[string]bool {
	drifted := make(map[string]bool)
	if snapshot.Scope.Provider == filter.ProviderFirewalld {
		return drifted
	}
	markers := make([]string, 0, len(ordered))
	projected := append([]filter.ObservedRule(nil), snapshot.Rules...)
	for _, item := range ordered {
		if item.Desired == nil || item.Desired.Marker == "" {
			continue
		}
		markers = append(markers, item.Desired.Marker)
		if item.Observed != nil {
			for index := range projected {
				if filter.SameLocator(projected[index].Locator, item.Observed.Locator) {
					projected[index].Marker = item.Desired.Marker
					break
				}
			}
		}
	}
	actual := make([]string, 0, len(markers))
	presentMarkers := make(map[string]bool, len(markers))
	for _, rule := range projected {
		if slices.Contains(markers, rule.Marker) {
			actual = append(actual, rule.Marker)
			presentMarkers[rule.Marker] = true
		}
	}
	present := make([]string, 0, len(actual))
	for _, marker := range markers {
		if presentMarkers[marker] {
			present = append(present, marker)
		}
	}
	for index, marker := range present {
		if actual[index] != marker {
			drifted[marker] = true
			drifted[actual[index]] = true
		}
	}
	return drifted
}

func InsertionPosition(snapshot filter.Snapshot, markers []string, target string) *int64 {
	if snapshot.Scope.Provider == filter.ProviderFirewalld {
		return nil
	}
	targetIndex := slices.Index(markers, target)
	for index := targetIndex - 1; index >= 0; index-- {
		for _, rule := range snapshot.Rules {
			if rule.Marker == markers[index] && rule.Locator.Position != nil {
				position := int64(*rule.Locator.Position + 1)
				return &position
			}
		}
	}
	if targetIndex >= 0 {
		for _, marker := range markers[targetIndex+1:] {
			for _, rule := range snapshot.Rules {
				if rule.Marker == marker && rule.Locator.Position != nil {
					position := int64(*rule.Locator.Position)
					return &position
				}
			}
		}
	}
	return nil
}

func DeleteChange(snapshot filter.Snapshot, previous filter.ObservedRule, desired filter.DesiredRule) (filter.DesiredChange, error) {
	beforeKey, err := filter.RuleKey(previous.Rule)
	if err != nil {
		return filter.DesiredChange{}, err
	}
	matches := make([]filter.ObservedRule, 0, 1)
	for _, current := range snapshot.Rules {
		if previous.Marker != "" {
			if current.Marker != previous.Marker {
				continue
			}
		} else if current.Marker != "" || !filter.SameLocator(current.Locator, previous.Locator) {
			continue
		}
		key, err := filter.RuleKey(current.Rule)
		if err == nil && key == beforeKey {
			matches = append(matches, current)
		}
	}
	if len(matches) != 1 {
		return filter.DesiredChange{}, filter.ErrRuleStale
	}
	current := matches[0]
	if err := filter.GuardMutation(current); err != nil {
		return filter.DesiredChange{}, err
	}
	before := ObservedRule(current)
	if before.UUID == "" {
		before.UUID = desired.Rule.UUID
	}
	return filter.DesiredChange{Operation: filter.ChangeDelete, Before: &before, Locator: &current.Locator, UnmarkedAdopted: current.Marker == "" && desired.Origin == filter.RuleOriginAdopted}, nil
}

func ObservedRule(observed filter.ObservedRule) filter.FirewallRule {
	rule := observed.Rule
	if rule.UUID == "" && strings.HasPrefix(observed.Marker, "1panel-rule:") {
		rule.UUID = strings.TrimSpace(strings.TrimPrefix(observed.Marker, "1panel-rule:"))
	}
	return rule
}
