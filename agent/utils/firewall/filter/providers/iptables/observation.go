package iptables

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	native "github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
)

type tableReader interface {
	ListTable(context.Context, filter.Scope) (string, error)
}

type tableRead struct {
	output string
	err    error
}

type tableObservationReader struct {
	reader tableReader
	tables map[string]tableRead
}

func (r *tableObservationReader) ListChain(ctx context.Context, scope filter.Scope) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	key := string(scope.Family) + ":" + scope.Table
	read, exists := r.tables[key]
	if !exists {
		read.output, read.err = r.reader.ListTable(ctx, scope)
		r.tables[key] = read
	}
	return chainOutput(scope, read.output, read.err)
}

func (a *Adapter) NewObservationSession() filter.Adapter {
	reader, ok := a.reader.(tableReader)
	if !ok {
		return a
	}
	return &Adapter{
		reader: &tableObservationReader{reader: reader, tables: make(map[string]tableRead)},
		writer: a.writer, checker: a.checker,
	}
}

func (systemBackend) ListTable(ctx context.Context, scope filter.Scope) (string, error) {
	return native.ReadTable(ctx, scope.Table, scope.Family == filter.FamilyIPv6)
}

func chainOutput(scope filter.Scope, output string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if !containsChainDeclaration(output, scope.Chain) {
		return "", fmt.Errorf("%w: iptables %s chain %s is not initialized", filter.ErrProviderUnavailable, scope.Family, scope.Chain)
	}
	return output, nil
}
