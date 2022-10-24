package components

import (
	"errors"
)

type Http struct {
	Comment    string
	Servers    []*Server
	Directives []IDirective
}

func (h *Http) GetComment() string {
	return h.Comment
}

func NewHttp(directive IDirective) (*Http, error) {
	if block := directive.GetBlock(); block != nil {
		http := &Http{
			Servers:    []*Server{},
			Directives: []IDirective{},
			Comment:    block.GetComment(),
		}

		for _, directive := range block.GetDirectives() {
			if server, ok := directive.(*Server); ok {
				http.Servers = append(http.Servers, server)
				continue
			}
			http.Directives = append(http.Directives, directive)
		}

		return http, nil
	}
	return nil, errors.New("http directive must have a block")
}

func (h *Http) GetName() string {
	return "http"
}

func (h *Http) GetParameters() []string {
	return []string{}
}

func (h *Http) GetDirectives() []IDirective {
	directives := make([]IDirective, 0)
	directives = append(directives, h.Directives...)
	for _, directive := range h.Servers {
		directives = append(directives, directive)
	}
	return directives
}

func (h *Http) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	for _, directive := range h.GetDirectives() {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}

	return directives
}

func (h *Http) UpdateDirectives(directiveName string, directive Directive) {
	directives := make([]IDirective, len(h.GetDirectives()))
	index := -1
	for i, dir := range h.GetDirectives() {
		if dir.GetName() == directiveName {
			index = i
			break
		}
	}
	if index > -1 {
		directives[index] = &directive
	} else {
		directives = append(directives, &directive)
	}
	h.Directives = directives
}

func (h *Http) AddDirectives(directive Directive) {
	directives := append(h.GetDirectives(), &directive)
	h.Directives = directives
}

func (h *Http) GetBlock() IBlock {
	return h
}
