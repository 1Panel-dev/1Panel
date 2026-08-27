package filter

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
