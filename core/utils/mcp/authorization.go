package mcp

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
)

const toolAuthorizationSnapshotVersion = 1
const toolAuthorizationSnapshotLimit = 2 << 20

type toolAuthorizationSnapshot struct {
	Version       int   `json:"version"`
	PolicyVersion int   `json:"policyVersion"`
	Tool          *Tool `json:"tool"`
}

// ToolAuthorizationSnapshot records the tool the owner explicitly approved.
// It is produced from the trusted catalog, never from credential request data.
func ToolAuthorizationSnapshot(tool Tool) (string, error) {
	catalog, err := ValidateCatalog([]Tool{tool})
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(toolAuthorizationSnapshot{
		Version: toolAuthorizationSnapshotVersion, PolicyVersion: ToolPolicyVersion, Tool: &catalog[0],
	})
	if err != nil {
		return "", err
	}
	if len(data) > toolAuthorizationSnapshotLimit {
		return "", errors.New("MCP authorization snapshot is too large")
	}
	return string(data), nil
}

// ToolAuthorizationAllows accepts only equivalent or provably narrower tools.
// The original snapshot remains the baseline; callers must separately require
// the current schema version and complete policy hash on every invocation.
func ToolAuthorizationAllows(snapshot, grantedPolicyHash string, current Tool) bool {
	if grantedPolicyHash == "" {
		return false
	}
	var saved toolAuthorizationSnapshot
	if DecodeObjectLimit([]byte(snapshot), &saved, toolAuthorizationSnapshotLimit) != nil ||
		saved.Version != toolAuthorizationSnapshotVersion || saved.PolicyVersion != ToolPolicyVersion || saved.Tool == nil ||
		saved.Tool.PolicyHash != grantedPolicyHash || ToolPolicyHash(*saved.Tool) != grantedPolicyHash {
		return false
	}
	// A snapshot must carry its read/write classification explicitly. Unlike
	// the API manifest, that classification is not part of ToolPolicyHash.
	var fields struct {
		Tool map[string]json.RawMessage `json:"tool"`
	}
	var classification *bool
	if json.Unmarshal([]byte(snapshot), &fields) != nil || json.Unmarshal(fields.Tool["readOnly"], &classification) != nil || classification == nil {
		return false
	}
	previous, err := ValidateCatalog([]Tool{*saved.Tool})
	if err != nil {
		return false
	}
	latest, err := ValidateCatalog([]Tool{current})
	if err != nil {
		return false
	}
	old, next := previous[0], latest[0]
	if old.ToolID != next.ToolID || old.ReadOnly && !next.ReadOnly || !authorizationSchemaSubset(old.InputSchema, next.InputSchema, 0) {
		return false
	}
	ops := make(map[string]Operation, len(old.Operations))
	for _, op := range old.Operations {
		ops[op.ID] = op
	}
	for _, op := range next.Operations {
		approved, exists := ops[op.ID]
		if !exists || approved.Method != op.Method || approved.Path != op.Path ||
			!authorizationSchemaSubset(approved.BodySchema, op.BodySchema, 0) ||
			!authorizationResponseSubset(approved.ResponseFields, op.ResponseFields) {
			return false
		}
	}
	return true
}

func authorizationResponseSubset(previous, current []string) bool {
	for _, field := range current {
		// Projections expose scalar leaves. An approved parent path never
		// authorizes new descendants, which could contain sensitive data.
		if !slices.Contains(previous, field) {
			return false
		}
	}
	return true
}

func authorizationSchemaSubset(previous, current map[string]any, depth int) bool {
	if authorizationSchemaEqual(previous, current, depth) {
		return true
	}
	if depth > 16 || validateSchema(previous, depth) != nil || validateSchema(current, depth) != nil {
		return false
	}
	oldType, nextType := previous["type"], current["type"]
	if oldType != nextType && !(oldType == "number" && nextType == "integer") {
		return false
	}
	oldDefault, oldHasDefault := previous["default"]
	nextDefault, nextHasDefault := current["default"]
	if oldHasDefault != nextHasDefault || !reflect.DeepEqual(oldDefault, nextDefault) {
		return false
	}
	if values, ok := previous["enum"].([]any); ok {
		selected, ok := current["enum"].([]any)
		if !ok {
			return false
		}
		for _, value := range selected {
			if !slices.ContainsFunc(values, func(approved any) bool { return reflect.DeepEqual(approved, value) }) {
				return false
			}
		}
	}
	for _, key := range []string{"minimum", "minLength", "minItems"} {
		if bound, ok := previous[key].(float64); ok {
			value, exists := current[key].(float64)
			if !exists || value < bound {
				return false
			}
		}
	}
	for _, key := range []string{"maximum", "maxLength", "maxItems"} {
		if bound, ok := previous[key].(float64); ok {
			value, exists := current[key].(float64)
			if !exists || value > bound {
				return false
			}
		}
	}
	if pattern, ok := previous["pattern"]; ok && !reflect.DeepEqual(pattern, current["pattern"]) {
		return false
	}
	switch nextType {
	case "object":
		oldProperties := previous["properties"].(map[string]any)
		nextProperties := current["properties"].(map[string]any)
		for name, raw := range oldProperties {
			if _, remains := nextProperties[name]; !remains && authorizationSchemaHasDefault(raw.(map[string]any)) {
				// Removing a default can change the meaning of an omitted input.
				return false
			}
		}
		for name, raw := range nextProperties {
			approved, exists := oldProperties[name].(map[string]any)
			if !exists || !authorizationSchemaSubset(approved, raw.(map[string]any), depth+1) {
				return false
			}
		}
		nextRequired, _ := current["required"].([]any)
		if required, ok := previous["required"].([]any); ok {
			for _, field := range required {
				if !slices.Contains(nextRequired, field) {
					return false
				}
			}
		}
	case "array":
		return authorizationSchemaSubset(previous["items"].(map[string]any), current["items"].(map[string]any), depth+1)
	}
	return true
}

func authorizationSchemaHasDefault(schema map[string]any) bool {
	if _, exists := schema["default"]; exists {
		return true
	}
	if properties, ok := schema["properties"].(map[string]any); ok {
		for _, raw := range properties {
			if child, ok := raw.(map[string]any); ok && authorizationSchemaHasDefault(child) {
				return true
			}
		}
	}
	if child, ok := schema["items"].(map[string]any); ok {
		return authorizationSchemaHasDefault(child)
	}
	return false
}

// Input schemas may contain extensions not interpreted by Core. Exact equality
// is safe, while a changed unknown constraint remains subject to confirmation.
// Only schema annotations are ignored, never defaults or enum object contents.
func authorizationSchemaEqual(previous, current map[string]any, depth int) bool {
	if reflect.DeepEqual(previous, current) {
		return true
	}
	if depth > 16 {
		return false
	}
	count := func(schema map[string]any) int {
		n := len(schema)
		if _, exists := schema["description"]; exists {
			n--
		}
		if _, exists := schema["title"]; exists {
			n--
		}
		return n
	}
	if count(previous) != count(current) {
		return false
	}
	for key, approved := range previous {
		if key == "description" || key == "title" {
			continue
		}
		value, exists := current[key]
		if !exists {
			return false
		}
		switch key {
		case "properties":
			oldProperties, oldOK := approved.(map[string]any)
			nextProperties, nextOK := value.(map[string]any)
			if !oldOK || !nextOK || len(oldProperties) != len(nextProperties) {
				return false
			}
			for name, raw := range oldProperties {
				oldSchema, oldOK := raw.(map[string]any)
				nextSchema, nextOK := nextProperties[name].(map[string]any)
				if !oldOK || !nextOK || !authorizationSchemaEqual(oldSchema, nextSchema, depth+1) {
					return false
				}
			}
		case "items":
			oldSchema, oldOK := approved.(map[string]any)
			nextSchema, nextOK := value.(map[string]any)
			if !oldOK || !nextOK || !authorizationSchemaEqual(oldSchema, nextSchema, depth+1) {
				return false
			}
		default:
			if !reflect.DeepEqual(approved, value) {
				return false
			}
		}
	}
	return true
}
