package components

import (
	"errors"
)

type Upstream struct {
	UpstreamName    string
	UpstreamServers []*UpstreamServer
	Directives      []IDirective
	Comment         string
	Line            int
}

func (us *Upstream) GetCodeBlock() string {
	return ""
}

func (us *Upstream) GetName() string {
	return "upstream"
}

func (us *Upstream) GetParameters() []string {
	return []string{us.UpstreamName}
}

func (us *Upstream) GetBlock() IBlock {
	return us
}

func (us *Upstream) GetComment() string {
	return us.Comment
}

func (us *Upstream) GetDirectives() []IDirective {
	directives := make([]IDirective, 0)
	directives = append(directives, us.Directives...)
	for _, uss := range us.UpstreamServers {
		directives = append(directives, uss)
	}
	return directives
}

func NewUpstream(directive IDirective) (*Upstream, error) {
	parameters := directive.GetParameters()
	us := &Upstream{
		UpstreamName: parameters[0],
		Line:         directive.GetLine(),
	}

	if block := directive.GetBlock(); block != nil {
		us.Comment = block.GetComment()
		for _, d := range block.GetDirectives() {
			if d.GetName() == "server" {
				us.UpstreamServers = append(us.UpstreamServers, NewUpstreamServer(d))
			} else {
				us.Directives = append(us.Directives, d)
			}
		}
		return us, nil
	}

	return nil, errors.New("missing upstream block")
}

func (us *Upstream) AddServer(server *UpstreamServer) {
	us.UpstreamServers = append(us.UpstreamServers, server)
}

func (us *Upstream) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	for _, directive := range us.Directives {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}

	return directives
}

func (us *Upstream) UpdateDirective(key string, params []string) {
	if key == "" || len(params) == 0 {
		return
	}
	directives := us.GetDirectives()
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
	us.Directives = directives
}

func (us *Upstream) RemoveDirective(key string, params []string) {
	directives := us.GetDirectives()
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
	us.Directives = newDirectives
}

func (us *Upstream) GetLine() int {
	return us.Line
}
