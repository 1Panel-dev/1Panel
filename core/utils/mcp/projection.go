package mcp

import (
	"encoding/json"
	"errors"
	"strings"
)

type projectionNode struct {
	leaf     bool
	children map[string]*projectionNode
}

func projectionTree(fields []string) (*projectionNode, error) {
	root := &projectionNode{children: map[string]*projectionNode{}}
	if len(fields) > 256 {
		return nil, errors.New("too many response fields")
	}
	for _, field := range fields {
		parts := strings.Split(field, ".")
		if len(parts) > 16 {
			return nil, errors.New("response path too deep")
		}
		node := root
		for _, part := range parts {
			if !fieldPart.MatchString(part) || node.leaf {
				return nil, errors.New("invalid response path")
			}
			if node.children[part] == nil {
				node.children[part] = &projectionNode{children: map[string]*projectionNode{}}
			}
			node = node.children[part]
		}
		if node.leaf || len(node.children) > 0 {
			return nil, errors.New("conflicting response fields")
		}
		node.leaf = true
	}
	return root, nil
}

// ProjectResult never copies a complete business object. A manifest lists
// scalar leaves explicitly, and paths transparently traverse arrays.
func ProjectResult(value any, fields []string) (any, error) {
	root, err := projectionTree(fields)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var data any
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return projectValue(data, root, 0)
}
func projectValue(value any, node *projectionNode, depth int) (any, error) {
	if depth > 32 {
		return nil, errors.New("response too deep")
	}
	if value == nil {
		return nil, nil
	}
	if list, ok := value.([]any); ok {
		out := make([]any, 0, len(list))
		for _, item := range list {
			v, err := projectValue(item, node, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	}
	if node.leaf {
		switch value.(type) {
		case string, float64, bool:
			return value, nil
		default:
			return nil, errors.New("response field must be a scalar leaf")
		}
	}
	if len(node.children) == 0 {
		return map[string]any{}, nil
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("response shape changed")
	}
	out := map[string]any{}
	for key, child := range node.children {
		v, exists := obj[key]
		if !exists {
			continue
		}
		selected, err := projectValue(v, child, depth+1)
		if err != nil {
			return nil, err
		}
		out[key] = selected
	}
	return out, nil
}
