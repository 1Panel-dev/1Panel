package components

import (
	"errors"
)

type Upstream struct {
	UpstreamName    string
	UpstreamServers []*UpstreamServer
	Directives      []IDirective
	Comment         string
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

func (us *Upstream) UpdateDirectives(directiveName string, directive Directive) {
	directives := make([]IDirective, 0)
	for _, dir := range us.GetDirectives() {
		if dir.GetName() == directiveName {
			directives = append(directives, &directive)
		} else {
			directives = append(directives, dir)
		}
	}
	us.Directives = directives
}

func (us *Upstream) AddDirectives(directive Directive) {
	directives := append(us.GetDirectives(), &directive)
	us.Directives = directives
}

func (us *Upstream) RemoveDirectives(names []string) {
	nameMaps := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameMaps[name] = struct{}{}
	}
	directives := us.GetDirectives()
	var newDirectives []IDirective
	for _, dir := range directives {
		if _, ok := nameMaps[dir.GetName()]; ok {
			continue
		}
		newDirectives = append(newDirectives, dir)
	}
	us.Directives = newDirectives
}
