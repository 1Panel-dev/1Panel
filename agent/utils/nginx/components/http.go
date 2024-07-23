package components

import (
	"errors"
)

type Http struct {
	Comment    string
	Servers    []*Server
	Directives []IDirective
	Line       int
}

func (h *Http) GetCodeBlock() string {
	return ""
}

func (h *Http) GetComment() string {
	return h.Comment
}

func NewHttp(directive IDirective) (*Http, error) {
	if block := directive.GetBlock(); block != nil {
		http := &Http{
			Line:       directive.GetBlock().GetLine(),
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

func (h *Http) UpdateDirective(key string, params []string) {
	if key == "" || len(params) == 0 {
		return
	}
	directives := h.GetDirectives()
	index := -1
	for i, dir := range directives {
		if dir.GetName() == key {
			if IsRepeatKey(key) {
				oldParams := dir.GetParameters()
				if !(len(oldParams) > 0 && oldParams[0] == params[0]) {
					continue
				}
			}
			index = i
			break
		}
	}
	newDirective := &Directive{
		Name:       key,
		Parameters: params,
	}
	if index > -1 {
		directives[index] = newDirective
	} else {
		directives = append(directives, newDirective)
	}
	h.Directives = directives
}

func (h *Http) RemoveDirective(key string, params []string) {
	directives := h.GetDirectives()
	var newDirectives []IDirective
	for _, dir := range directives {
		if dir.GetName() == key {
			if IsRepeatKey(key) && len(params) > 0 {
				oldParams := dir.GetParameters()
				if oldParams[0] == params[0] {
					continue
				}
			} else {
				continue
			}
		}
		newDirectives = append(newDirectives, dir)
	}
	h.Directives = newDirectives
}

func (h *Http) GetBlock() IBlock {
	return h
}

func (h *Http) GetLine() int {
	return h.Line
}
