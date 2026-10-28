// Package taskscope separates task visibility and executor ownership from task storage.
package taskscope

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
)

type Rule struct {
	Scope            string
	Restricted       bool
	ExternalExecutor bool
}

type policy struct{ rules []Rule }

var configured atomic.Pointer[policy]

var ErrUnavailable = errors.New("task scope policy is unavailable")

func Configure(rules ...Rule) error {
	copyRules := append([]Rule(nil), rules...)
	if len(copyRules) == 0 {
		return ErrUnavailable
	}
	sort.Slice(copyRules, func(i, j int) bool { return copyRules[i].Scope < copyRules[j].Scope })
	for i, rule := range copyRules {
		if rule.Scope == "" || strings.TrimSpace(rule.Scope) != rule.Scope || strings.ContainsAny(rule.Scope, "\x00\r\n") ||
			i > 0 && copyRules[i-1].Scope == rule.Scope {
			return errors.New("invalid task scope policy")
		}
	}
	next := &policy{rules: copyRules}
	if configured.CompareAndSwap(nil, next) {
		return nil
	}
	if !slices.Equal(configured.Load().rules, next.rules) {
		return errors.New("task scope policy is already configured")
	}
	return nil
}

func ReadExclusions(granted ...string) ([]string, error) {
	p := configured.Load()
	if p == nil {
		return nil, ErrUnavailable
	}
	allowed := make(map[string]bool, len(granted))
	for _, scope := range granted {
		found := false
		for _, rule := range p.rules {
			if rule.Scope == scope && rule.Restricted {
				found = true
				break
			}
		}
		if !found || allowed[scope] {
			return nil, errors.New("invalid task scope grant")
		}
		allowed[scope] = true
	}
	var excluded []string
	for _, rule := range p.rules {
		if rule.Restricted && !allowed[rule.Scope] {
			excluded = append(excluded, rule.Scope)
		}
	}
	return excluded, nil
}

func RecoveryExclusions() ([]string, error) {
	p := configured.Load()
	if p == nil {
		return nil, ErrUnavailable
	}
	var excluded []string
	for _, rule := range p.rules {
		if rule.ExternalExecutor {
			excluded = append(excluded, rule.Scope)
		}
	}
	return excluded, nil
}

type grantKey struct{}

func WithGrants(ctx context.Context, granted ...string) context.Context {
	return context.WithValue(ctx, grantKey{}, append([]string(nil), granted...))
}

func Grants(ctx context.Context) []string {
	granted, _ := ctx.Value(grantKey{}).([]string)
	return append([]string(nil), granted...)
}
