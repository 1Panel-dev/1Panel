package filter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type ruleIdentity struct {
	Scope              string     `json:"scope"`
	Family             Family     `json:"family"`
	NativeKind         NativeKind `json:"nativeKind"`
	Protocol           string     `json:"protocol"`
	SourceAddress      string     `json:"sourceAddress,omitempty"`
	SourcePort         string     `json:"sourcePort,omitempty"`
	DestinationAddress string     `json:"destinationAddress,omitempty"`
	DestinationPort    string     `json:"destinationPort,omitempty"`
	Interface          string     `json:"interface,omitempty"`
	ConnectionStates   []string   `json:"connectionStates,omitempty"`
	Action             Action     `json:"action"`
	Priority           *int       `json:"priority,omitempty"`
	OrderBucket        string     `json:"orderBucket,omitempty"`
}

type instanceIdentity struct {
	RuleKey     string            `json:"ruleKey"`
	Marker      string            `json:"marker,omitempty"`
	Persistence PersistenceStatus `json:"persistence,omitempty"`
	NativeID    string            `json:"nativeId,omitempty"`
	Canonical   string            `json:"canonical,omitempty"`
	Position    *int              `json:"position,omitempty"`
	ScopeKey    string            `json:"scopeKey"`
	Provider    Provider          `json:"provider"`
}

func RuleKey(rule FirewallRule) (string, error) {
	normalized, err := NormalizeRule(rule)
	if err != nil {
		return "", err
	}
	return normalizedRuleKey(normalized)
}

func RuleMatchKey(rule FirewallRule) (string, error) {
	normalized, err := NormalizeRule(rule)
	if err != nil {
		return "", err
	}
	normalized.Action, normalized.NativeKind, normalized.OrderBucket = "", "", ""
	normalized.Priority = nil
	return normalizedRuleKey(normalized)
}

func OppositeActions(left, right Action) bool {
	return left == ActionAccept && (right == ActionDrop || right == ActionReject) ||
		right == ActionAccept && (left == ActionDrop || left == ActionReject)
}

func SameRuleContent(before, after FirewallRule) (bool, error) {
	before, err := NormalizeRule(before)
	if err != nil {
		return false, err
	}
	after, err = NormalizeRule(after)
	if err != nil {
		return false, err
	}
	previous, err := RuleMatchKey(before)
	if err != nil {
		return false, err
	}
	requested, err := RuleMatchKey(after)
	return err == nil && previous == requested && before.Action == after.Action, err
}

func normalizedRuleKey(normalized FirewallRule) (string, error) {
	identity := ruleIdentity{
		Scope:              normalized.Scope.Key(),
		Family:             normalized.Scope.Family,
		NativeKind:         normalized.NativeKind,
		Protocol:           normalized.Protocol,
		SourceAddress:      normalized.SourceAddress,
		SourcePort:         normalized.SourcePort,
		DestinationAddress: normalized.DestinationAddress,
		DestinationPort:    normalized.DestinationPort,
		Interface:          normalized.Interface,
		ConnectionStates:   normalized.ConnectionStates,
		Action:             normalized.Action,
		Priority:           normalized.Priority,
		OrderBucket:        normalized.OrderBucket,
	}
	return hashJSON(identity)
}

func InstanceKey(rule ObservedRule) (string, error) {
	if rule.ParseStatus != ParseStatusSupported {
		locator, err := validatedLocator(rule.Locator, rule.Rule.Scope)
		if err != nil {
			return "", err
		}
		return opaqueInstanceKey(rule, locator)
	}
	ruleKey, err := RuleKey(rule.Rule)
	if err != nil {
		return "", err
	}
	return instanceKeyWithRuleKey(rule, ruleKey)
}

func instanceKeyWithRuleKey(rule ObservedRule, ruleKey string) (string, error) {
	locator, err := validatedLocator(rule.Locator, rule.Rule.Scope)
	if err != nil {
		return "", err
	}
	return hashJSON(instanceIdentity{
		RuleKey:     ruleKey,
		Marker:      strings.TrimSpace(rule.Marker),
		Persistence: rule.Persistence,
		NativeID:    locator.NativeID,
		Canonical:   locator.Canonical,
		Position:    locator.Position,
		ScopeKey:    locator.ScopeKey,
		Provider:    locator.Provider,
	})
}

func opaqueInstanceKey(rule ObservedRule, locator Locator) (string, error) {
	return hashJSON(struct {
		Raw         string            `json:"raw"`
		Locator     Locator           `json:"locator"`
		Persistence PersistenceStatus `json:"persistence,omitempty"`
	}{Raw: strings.TrimSpace(rule.Raw), Locator: locator, Persistence: rule.Persistence})
}

func NewRuleSet(scope Scope, rules []ObservedRule) (RuleSet, error) {
	scope = scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return RuleSet{}, err
	}
	for _, rule := range rules {
		if rule.Rule.Scope.Normalize().Key() != scope.Key() {
			return RuleSet{}, fmt.Errorf("%w: observed rule scope does not match read scope", ErrInvalidRule)
		}
		if rule.ParseStatus == ParseStatusSupported {
			if _, err := NormalizeRule(rule.Rule); err != nil {
				return RuleSet{}, err
			}
		}
		if _, err := validatedLocator(rule.Locator, scope); err != nil {
			return RuleSet{}, err
		}
	}
	return RuleSet{Scope: scope, Rules: rules}, nil
}

func normalizeLocator(locator Locator, scope Scope) Locator {
	scope = scope.Normalize()
	locator.Provider = Provider(strings.ToLower(strings.TrimSpace(string(locator.Provider))))
	if locator.Provider == "" {
		locator.Provider = scope.Provider
	}
	locator.ScopeKey = strings.TrimSpace(locator.ScopeKey)
	if locator.ScopeKey == "" {
		locator.ScopeKey = scope.Key()
	}
	locator.NativeID = strings.TrimSpace(locator.NativeID)
	locator.Canonical = strings.TrimSpace(locator.Canonical)
	return locator
}

func validatedLocator(locator Locator, scope Scope) (Locator, error) {
	scope = scope.Normalize()
	locator = normalizeLocator(locator, scope)
	if locator.Provider != scope.Provider || locator.ScopeKey != scope.Key() {
		return Locator{}, fmt.Errorf("%w: locator does not belong to scope %q", ErrInvalidRule, scope.Key())
	}
	if (scope.Provider == ProviderIptables || scope.Provider == ProviderNftables || scope.Provider == ProviderUFW) && locator.Position == nil {
		return Locator{}, fmt.Errorf("%w: ordered scope %q requires locator position", ErrInvalidRule, scope.Key())
	}
	return locator, nil
}

func hashJSON(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal firewall identity: %w", err)
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func FindCandidate(candidates []ObservedRule, selected string) (ObservedRule, error) {
	matched := make([]ObservedRule, 0, 1)
	for _, candidate := range candidates {
		identity, err := InstanceKey(candidate)
		if err != nil {
			continue
		}
		if selected != "" && identity == selected {
			matched = append(matched, candidate)
		}
	}
	if len(matched) != 1 {
		return ObservedRule{}, ErrRuleOperation
	}
	return matched[0], nil
}

const ObservedFieldProtocol = "protocol"

// ObservedRuleMatchesExpected compares the semantic fields that the backend
// could actually observe. Partial rules name fields omitted by their native
// listing; those fields are supplied from the expected rule before comparing
// normalized identities.
func ObservedRuleMatchesExpected(observed ObservedRule, expected FirewallRule) bool {
	if observed.ParseStatus == ParseStatusOpaque {
		return false
	}
	hydrated := observed.Rule
	for _, field := range observed.UncertainFields {
		switch field {
		case ObservedFieldProtocol:
			hydrated.Protocol = expected.Protocol
		default:
			return false
		}
	}
	gotKey, gotErr := RuleKey(hydrated)
	wantKey, wantErr := RuleKey(expected)
	return gotErr == nil && wantErr == nil && gotKey == wantKey
}
