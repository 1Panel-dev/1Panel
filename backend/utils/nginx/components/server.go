package components

import (
	"errors"
)

type Server struct {
	Comment    string
	Listens    []*ServerListen
	Directives []IDirective
}

func NewServer(directive IDirective) (*Server, error) {
	server := &Server{}
	if block := directive.GetBlock(); block != nil {
		server.Comment = block.GetComment()
		directives := block.GetDirectives()
		for _, dir := range directives {

			switch dir.GetName() {
			case "listen":
				server.Listens = append(server.Listens, NewServerListen(dir.GetParameters()))
			default:
				server.Directives = append(server.Directives, dir)
			}
		}
		return server, nil
	}
	return nil, errors.New("server directive must have a block")
}

func (s *Server) GetName() string {
	return "server"
}

func (s *Server) GetParameters() []string {
	return []string{}
}

func (s *Server) GetBlock() IBlock {
	return s
}

func (s *Server) GetComment() string {
	return s.Comment
}

func (s *Server) GetDirectives() []IDirective {
	directives := make([]IDirective, 0)
	for _, ls := range s.Listens {
		directives = append(directives, ls)
	}
	directives = append(directives, s.Directives...)
	return directives
}

func (s *Server) AddListen(bind string, defaultServer bool, params ...string) {
	listen := &ServerListen{
		Bind:       bind,
		Parameters: params,
	}
	if defaultServer {
		listen.DefaultServer = DefaultServer
	}
	s.Listens = append(s.Listens, listen)
}

func (s *Server) UpdateListen(bind string, defaultServer bool, params ...string) {
	listen := &ServerListen{
		Bind:       bind,
		Parameters: params,
	}
	if defaultServer {
		listen.DefaultServer = DefaultServer
	}
	var newListens []*ServerListen
	for _, li := range s.Listens {
		if li.Bind == bind {
			newListens = append(newListens, listen)
		} else {
			newListens = append(newListens, li)
		}
	}

	s.Listens = newListens
}

func (s *Server) UpdateServerName(names []string) {
	serverNameDirective := Directive{
		Name:       "server_name",
		Parameters: names,
	}
	s.UpdateDirectives("server_name", serverNameDirective)
}

func (s *Server) UpdateRootProxy(proxy []string) {
	//TODD 根据ID修改
	dirs := s.FindDirectives("location")
	for _, dir := range dirs {
		location, ok := dir.(*Location)
		if ok && location.Match == "/" {
			newDir := Directive{
				Name:       "location",
				Parameters: []string{"/"},
				Block:      &Block{},
			}
			block := &Block{}
			block.Directives = append(block.Directives, &Directive{
				Name:       "proxy_pass",
				Parameters: proxy,
			})
			newDir.Block = block
			s.UpdateDirectives("location", newDir)
		}
	}
}

func (s *Server) RemoveListenByBind(bind string) {
	index := 0
	listens := s.Listens
	for _, listen := range s.Listens {
		if listen.Bind != bind || len(listen.Parameters) > 0 {
			listens[index] = listen
			index++
		}
	}
	s.Listens = listens
}

func (s *Server) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	for _, directive := range s.Directives {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}

	return directives
}

func (s *Server) UpdateDirectives(directiveName string, directive Directive) {
	directives := make([]IDirective, 0)
	for _, dir := range s.Directives {
		if dir.GetName() == directiveName {
			directives = append(directives, &directive)
		} else {
			directives = append(directives, dir)
		}
	}
	s.Directives = directives
}

func (s *Server) AddDirectives(directive Directive) {
	directives := append(s.Directives, &directive)
	s.Directives = directives
}

func (s *Server) RemoveDirectives(names []string) {
	nameMaps := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameMaps[name] = struct{}{}
	}
	directives := s.GetDirectives()
	var newDirectives []IDirective
	for _, dir := range directives {
		if _, ok := nameMaps[dir.GetName()]; ok {
			continue
		}
		newDirectives = append(newDirectives, dir)
	}
	s.Directives = newDirectives
}
