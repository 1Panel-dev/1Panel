package components

import (
	"fmt"
	"sort"
	"strings"
)

type UpstreamServer struct {
	Comment    string
	Address    string
	Flags      []string
	Parameters map[string]string
	Line       int
}

func (uss *UpstreamServer) GetName() string {
	return "server"
}

func (uss *UpstreamServer) GetBlock() IBlock {
	return nil
}

func (uss *UpstreamServer) GetParameters() []string {
	return uss.GetDirective().Parameters
}

func (uss *UpstreamServer) GetComment() string {
	return uss.Comment
}

func (uss *UpstreamServer) GetDirective() *Directive {
	directive := &Directive{
		Name:       "server",
		Parameters: make([]string, 0),
		Block:      nil,
	}

	directive.Parameters = append(directive.Parameters, uss.Address)

	paramNames := make([]string, 0)
	for k := range uss.Parameters {
		paramNames = append(paramNames, k)
	}
	sort.Strings(paramNames)

	for _, k := range paramNames {
		directive.Parameters = append(directive.Parameters, fmt.Sprintf("%s=%s", k, uss.Parameters[k]))
	}

	directive.Parameters = append(directive.Parameters, uss.Flags...)

	return directive
}

func NewUpstreamServer(directive IDirective) *UpstreamServer {
	uss := &UpstreamServer{
		Comment:    directive.GetComment(),
		Flags:      make([]string, 0),
		Parameters: make(map[string]string, 0),
		Line:       directive.GetLine(),
	}

	for i, parameter := range directive.GetParameters() {
		if i == 0 {
			uss.Address = parameter
			continue
		}
		if strings.Contains(parameter, "=") {
			s := strings.SplitN(parameter, "=", 2)
			uss.Parameters[s[0]] = s[1]
		} else {
			uss.Flags = append(uss.Flags, parameter)
		}
	}

	return uss
}

func (uss *UpstreamServer) GetLine() int {
	return uss.Line
}
