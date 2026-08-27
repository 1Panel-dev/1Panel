package filter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type CheckFlagCodec struct {
	secret  []byte
	version int
}

type checkFlagClaims struct {
	Version            int                          `json:"version"`
	Provider           Provider                     `json:"provider"`
	ScopeKey           string                       `json:"scopeKey"`
	RuleDigest         string                       `json:"ruleDigest"`
	SnapshotRevision   string                       `json:"snapshotRevision"`
	ManagedRevision    string                       `json:"managedRevision"`
	Decision           CheckDecision                `json:"decision"`
	Classification     CheckClassification          `json:"classification"`
	AllowedActions     []CheckAction                `json:"allowedActions"`
	AdoptionCandidates []checkFlagAdoptionCandidate `json:"adoptionCandidates,omitempty"`
}

type checkFlagAdoptionCandidate struct {
	InstanceKey string  `json:"instanceKey"`
	Locator     Locator `json:"locator"`
}

type CreateAuthorization struct {
	Operation ChangeOperation
	Locator   *Locator
}

func NewCheckFlagCodec(secret []byte, version int) *CheckFlagCodec {
	return &CheckFlagCodec{secret: append([]byte(nil), secret...), version: version}
}

func (c *CheckFlagCodec) Sign(result RuleCheckResult, snapshot Snapshot, managedRevision string) (string, error) {
	ruleDigest, err := ruleDigest(result.RequestedRule)
	if err != nil {
		return "", err
	}
	claims := checkFlagClaims{
		Version:          c.version,
		Provider:         result.RequestedRule.Scope.Provider,
		ScopeKey:         result.RequestedRule.Scope.Key(),
		RuleDigest:       ruleDigest,
		SnapshotRevision: snapshot.Revision,
		ManagedRevision:  managedRevision,
		Decision:         result.Decision,
		Classification:   result.Classification,
		AllowedActions:   result.AllowedActions,
	}
	if result.Classification == CheckClassificationExactExternal {
		claims.AdoptionCandidates = make([]checkFlagAdoptionCandidate, 0, len(result.Candidates))
		for _, candidate := range result.Candidates {
			claims.AdoptionCandidates = append(claims.AdoptionCandidates, checkFlagAdoptionCandidate{
				InstanceKey: candidate.InstanceKey,
				Locator:     candidate.Locator,
			})
		}
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signature := c.signature(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (c *CheckFlagCodec) Authorize(
	checkFlag string,
	action CheckAction,
	adoptInstanceKey string,
	rule FirewallRule,
	snapshot Snapshot,
	managedRevision string,
) (CreateAuthorization, error) {
	claims, err := c.parse(checkFlag)
	if err != nil {
		return CreateAuthorization{}, err
	}
	ruleDigest, err := ruleDigest(rule)
	if err != nil {
		return CreateAuthorization{}, err
	}
	if claims.Version != c.version ||
		claims.Provider != rule.Scope.Provider ||
		claims.ScopeKey != rule.Scope.Key() ||
		claims.RuleDigest != ruleDigest ||
		claims.SnapshotRevision != snapshot.Revision ||
		claims.ManagedRevision != managedRevision {
		return CreateAuthorization{}, fmt.Errorf("%w: firewall or managed rules changed", ErrRuleCheckRequired)
	}
	if claims.Decision != CheckDecisionReady && claims.Decision != CheckDecisionConfirmationRequired {
		return CreateAuthorization{}, ErrRuleOperation
	}
	if !containsCheckAction(claims.AllowedActions, action) {
		return CreateAuthorization{}, ErrRuleOperation
	}

	switch action {
	case CheckActionCreate, CheckActionCreateAnyway:
		if strings.TrimSpace(adoptInstanceKey) != "" {
			return CreateAuthorization{}, ErrRuleOperation
		}
		return CreateAuthorization{Operation: ChangeCreate}, nil
	case CheckActionAdopt, CheckActionSelectAdopt:
		for _, candidate := range claims.AdoptionCandidates {
			if candidate.InstanceKey == adoptInstanceKey && adoptInstanceKey != "" {
				locator := candidate.Locator
				return CreateAuthorization{Operation: ChangeAdopt, Locator: &locator}, nil
			}
		}
		return CreateAuthorization{}, ErrRuleOperation
	default:
		return CreateAuthorization{}, ErrRuleOperation
	}
}

func (c *CheckFlagCodec) parse(checkFlag string) (checkFlagClaims, error) {
	parts := strings.Split(strings.TrimSpace(checkFlag), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return checkFlagClaims{}, ErrRuleCheckRequired
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return checkFlagClaims{}, ErrRuleCheckRequired
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, c.signature(payload)) {
		return checkFlagClaims{}, ErrRuleCheckRequired
	}
	var claims checkFlagClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return checkFlagClaims{}, ErrRuleCheckRequired
	}
	return claims, nil
}

func (c *CheckFlagCodec) signature(payload []byte) []byte {
	mac := hmac.New(sha256.New, c.secret)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func ruleDigest(rule FirewallRule) (string, error) {
	payload, err := json.Marshal(rule)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func containsCheckAction(actions []CheckAction, expected CheckAction) bool {
	for _, action := range actions {
		if action == expected {
			return true
		}
	}
	return false
}
