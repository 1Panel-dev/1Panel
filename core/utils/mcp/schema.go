package mcp

import (
	"errors"
	"math"
	"reflect"
	"regexp"
	"slices"
	"unicode/utf8"
)

func ValidateInput(schema map[string]any, value any) bool {
	return validateSchema(schema, 0) == nil && validateValue(schema, value, 0) == nil
}

func validateSchema(s map[string]any, depth int) error {
	if depth > 16 {
		return errors.New("MCP schema is too deep")
	}
	typ, ok := s["type"].(string)
	if !ok || !slices.Contains([]string{"object", "string", "integer", "number", "boolean", "array", "null"}, typ) {
		return errors.New("unsupported MCP schema type")
	}
	allowed := []string{"type", "description", "title", "default", "enum"}
	switch typ {
	case "object":
		allowed = append(allowed, "properties", "required", "additionalProperties")
	case "string":
		allowed = append(allowed, "minLength", "maxLength", "pattern")
	case "integer", "number":
		allowed = append(allowed, "minimum", "maximum")
	case "array":
		allowed = append(allowed, "items", "minItems", "maxItems")
	}
	for key := range s {
		if !slices.Contains(allowed, key) {
			return errors.New("unsupported MCP schema keyword")
		}
	}
	if e, ok := s["enum"]; ok {
		values, ok := e.([]any)
		if !ok || len(values) == 0 || len(values) > 256 {
			return errors.New("invalid enum")
		}
	}
	for _, k := range []string{"minimum", "maximum", "minLength", "maxLength", "minItems", "maxItems"} {
		if v, ok := s[k]; ok {
			n, ok := v.(float64)
			if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
				return errors.New("invalid schema limit")
			}
			if k != "minimum" && k != "maximum" && (n < 0 || n != math.Trunc(n)) {
				return errors.New("invalid schema size")
			}
		}
	}
	if p, ok := s["pattern"]; ok {
		text, ok := p.(string)
		if !ok || len(text) > 1024 {
			return errors.New("invalid pattern")
		}
		if _, err := regexp.Compile(text); err != nil {
			return err
		}
	}
	if typ == "object" {
		if extra, ok := s["additionalProperties"].(bool); !ok || extra {
			return errors.New("MCP objects must reject undeclared fields")
		}
		props, ok := s["properties"].(map[string]any)
		if !ok || len(props) > 128 {
			return errors.New("invalid properties")
		}
		for name, raw := range props {
			if !fieldPart.MatchString(name) {
				return errors.New("invalid property name")
			}
			child, ok := raw.(map[string]any)
			if !ok {
				return errors.New("invalid property schema")
			}
			if err := validateSchema(child, depth+1); err != nil {
				return err
			}
		}
		if raw, ok := s["required"]; ok {
			required, ok := raw.([]any)
			if !ok {
				return errors.New("invalid required fields")
			}
			seen := map[string]bool{}
			for _, r := range required {
				name, ok := r.(string)
				if !ok || props[name] == nil || seen[name] {
					return errors.New("invalid required field")
				}
				seen[name] = true
			}
		}
	}
	if typ == "array" {
		child, ok := s["items"].(map[string]any)
		if !ok {
			return errors.New("invalid array schema")
		}
		if err := validateSchema(child, depth+1); err != nil {
			return err
		}
	}
	return nil
}
func validateValue(s map[string]any, v any, depth int) error {
	if depth > 32 {
		return Invalid()
	}
	if raw, ok := s["enum"]; ok {
		matched := false
		for _, e := range raw.([]any) {
			if reflect.DeepEqual(e, v) {
				matched = true
				break
			}
		}
		if !matched {
			return Invalid()
		}
	}
	bound := func(value float64, lower, upper string) bool {
		if n, ok := s[lower].(float64); ok && value < n {
			return false
		}
		if n, ok := s[upper].(float64); ok && value > n {
			return false
		}
		return true
	}
	switch s["type"] {
	case "object":
		obj, ok := v.(map[string]any)
		if !ok {
			return Invalid()
		}
		props := s["properties"].(map[string]any)
		if required, ok := s["required"].([]any); ok {
			for _, r := range required {
				if _, exists := obj[r.(string)]; !exists {
					return Invalid()
				}
			}
		}
		for k, val := range obj {
			raw, ok := props[k]
			if !ok {
				return Invalid()
			}
			if err := validateValue(raw.(map[string]any), val, depth+1); err != nil {
				return err
			}
		}
	case "array":
		a, ok := v.([]any)
		if !ok || !bound(float64(len(a)), "minItems", "maxItems") {
			return Invalid()
		}
		for _, item := range a {
			if err := validateValue(s["items"].(map[string]any), item, depth+1); err != nil {
				return err
			}
		}
	case "string":
		text, ok := v.(string)
		if !ok || !bound(float64(utf8.RuneCountInString(text)), "minLength", "maxLength") {
			return Invalid()
		}
		if p, ok := s["pattern"].(string); ok && !regexp.MustCompile(p).MatchString(text) {
			return Invalid()
		}
	case "number", "integer":
		n, ok := v.(float64)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) || !bound(n, "minimum", "maximum") {
			return Invalid()
		}
		if s["type"] == "integer" && (math.Trunc(n) != n || math.Abs(n) > 9007199254740991) {
			return Invalid()
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return Invalid()
		}
	case "null":
		if v != nil {
			return Invalid()
		}
	default:
		return Invalid()
	}
	return nil
}
