package components

import (
	"errors"
)

type Server struct {
	Comment    string
	Listens    []*ServerListen
	Directives []IDirective
	Line       int
}

func (s *Server) GetCodeBlock() string {
	return ""
}

func NewServer(directive IDirective) (*Server, error) {
	server := &Server{}
	if block := directive.GetBlock(); block != nil {
		server.Line = directive.GetBlock().GetLine()
		server.Comment = block.GetComment()
		directives := block.GetDirectives()
		for _, dir := range directives {
			switch dir.GetName() {
			case "listen":
				server.Listens = append(server.Listens, NewServerListen(dir.GetParameters(), dir.GetLine()))
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
	if directiveName == "listen" {
		for _, listen := range s.Listens {
			params := []string{listen.Bind}
			params = append(params, listen.Parameters...)
			if listen.DefaultServer != "" {
				params = append(params, DefaultServer)
			}
			directives = append(directives, &Directive{
				Name:       "listen",
				Parameters: params,
			})
		}
	}
	return directives
}

func (s *Server) UpdateDirective(key string, params []string) {
	if key == "" || len(params) == 0 {
		return
	}
	if key == "listen" {
		defaultServer := false
		paramLen := len(params)
		if paramLen > 0 && params[paramLen-1] == "default_server" {
			params = params[:paramLen-1]
			defaultServer = true
		}
		s.UpdateListen(params[0], defaultServer, params[1:]...)
		return
	}

	directives := s.Directives
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
	s.Directives = directives
}

func (s *Server) RemoveDirective(key string, params []string) {
	directives := s.Directives
	var newDirectives []IDirective
	for _, dir := range directives {
		if dir.GetName() == key {
			if len(params) == 0 {
				continue
			}
			oldParams := dir.GetParameters()
			if key == "location" {
				if len(params) == len(oldParams) {
					exist := true
					for i := range params {
						if params[i] != oldParams[i] {
							exist = false
							break
						}
					}
					if exist {
						continue
					}
				}
			} else {
				if oldParams[0] == params[0] {
					continue
				}
			}
		}
		newDirectives = append(newDirectives, dir)
	}
	s.Directives = newDirectives
}

func (s *Server) GetLine() int {
	return s.Line
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
	exist := false
	for _, li := range s.Listens {
		if li.Bind == bind {
			exist = true
			newListens = append(newListens, listen)
		} else {
			newListens = append(newListens, li)
		}
	}
	if !exist {
		newListens = append(newListens, listen)
	}

	s.Listens = newListens
}

func (s *Server) DeleteListen(bind string) {
	var newListens []*ServerListen
	for _, li := range s.Listens {
		if li.Bind != bind {
			newListens = append(newListens, li)
		}
	}
	s.Listens = newListens
}

func (s *Server) DeleteServerName(name string) {
	var names []string
	dirs := s.FindDirectives("server_name")
	params := dirs[0].GetParameters()
	for _, param := range params {
		if param != name {
			names = append(names, param)
		}
	}
	s.UpdateServerName(names)
}

func (s *Server) AddServerName(name string) {
	dirs := s.FindDirectives("server_name")
	params := dirs[0].GetParameters()
	params = append(params, name)
	s.UpdateServerName(params)
}

func (s *Server) UpdateServerName(names []string) {
	s.UpdateDirective("server_name", names)
}

func (s *Server) UpdateRoot(path string) {
	s.UpdateDirective("root", []string{path})
}

func (s *Server) UpdateRootLocation() {
	newDir := Directive{
		Name:       "location",
		Parameters: []string{"/"},
		Block:      &Block{},
	}
	block := &Block{}
	block.Directives = append(block.Directives, &Directive{
		Name:       "root",
		Parameters: []string{"index.html"},
	})
	newDir.Block = block
}

func (s *Server) UpdateRootProxy(proxy []string) {
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
	s.UpdateDirectiveBySecondKey("location", "/", newDir)
}

func (s *Server) UpdateRootProxyForAi(proxy []string) {
	newDir := Directive{
		Name:       "location",
		Parameters: []string{"/"},
		Block:      &Block{},
	}
	block := &Block{}
	block.Directives = []IDirective{
		&Directive{
			Name: "proxy_buffering",
			Parameters: []string{
				"off",
			},
		},
		&Directive{
			Name: "proxy_cache",
			Parameters: []string{
				"off",
			},
		},
		&Directive{
			Name: "proxy_http_version",
			Parameters: []string{
				"1.1",
			},
		},
		&Directive{
			Name: "proxy_set_header",
			Parameters: []string{
				"Connection", "''",
			},
		},
		&Directive{
			Name: "chunked_transfer_encoding",
			Parameters: []string{
				"off",
			},
		},
	}
	block.Directives = append(block.Directives, &Directive{
		Name:       "proxy_pass",
		Parameters: proxy,
	})
	newDir.Block = block
	s.UpdateDirectiveBySecondKey("location", "/", newDir)
}

func (s *Server) UpdatePHPProxy(proxy []string, localPath string) {
	newDir := Directive{
		Name:       "location",
		Parameters: []string{"~ [^/]\\.php(/|$)"},
		Block:      &Block{},
	}
	block := &Block{}
	block.Directives = append(block.Directives, &Directive{
		Name:       "fastcgi_pass",
		Parameters: proxy,
	})
	block.Directives = append(block.Directives, &Directive{
		Name:       "include",
		Parameters: []string{"fastcgi-php.conf"},
	})
	block.Directives = append(block.Directives, &Directive{
		Name:       "include",
		Parameters: []string{"fastcgi_params"},
	})
	if localPath == "" {
		block.Directives = append(block.Directives, &Directive{
			Name:       "set",
			Parameters: []string{"$real_script_name", "$fastcgi_script_name"},
		})
		ifDir := &Directive{
			Name:       "if",
			Parameters: []string{"($fastcgi_script_name ~ \"^(.+?\\.php)(/.+)$\")"},
		}
		ifDir.Block = &Block{
			Directives: []IDirective{
				&Directive{
					Name:       "set",
					Parameters: []string{"$real_script_name", "$1"},
				},
				&Directive{
					Name:       "set",
					Parameters: []string{"$path_info", "$2"},
				},
			},
		}
		block.Directives = append(block.Directives, ifDir)
		block.Directives = append(block.Directives, &Directive{
			Name:       "fastcgi_param",
			Parameters: []string{"SCRIPT_FILENAME", "$document_root$real_script_name"},
		})
		block.Directives = append(block.Directives, &Directive{
			Name:       "fastcgi_param",
			Parameters: []string{"SCRIPT_NAME", "$real_script_name"},
		})
		block.Directives = append(block.Directives, &Directive{
			Name:       "fastcgi_param",
			Parameters: []string{"PATH_INFO", "$path_info"},
		})
	} else {
		block.Directives = append(block.Directives, &Directive{
			Name:       "fastcgi_param",
			Parameters: []string{"SCRIPT_FILENAME", localPath},
		})
	}
	newDir.Block = block
	s.UpdateDirectiveBySecondKey("location", "~ [^/]\\.php(/|$)", newDir)
}

func (s *Server) UpdateDirectiveBySecondKey(name string, key string, directive Directive) {
	directives := s.Directives
	index := -1
	for i, dir := range directives {
		if dir.GetName() == name && dir.GetParameters()[0] == key {
			index = i
			break
		}
	}
	if index > -1 {
		directives[index] = &directive
	} else {
		directives = append(directives, &directive)
	}
	s.Directives = directives
}

func (s *Server) RemoveListenByBind(bind string) {
	var listens []*ServerListen
	for _, listen := range s.Listens {
		if listen.Bind != bind {
			listens = append(listens, listen)
		}
	}
	s.Listens = listens
}

func (s *Server) AddHTTP2HTTPS() {
	newDir := Directive{
		Name:       "if",
		Parameters: []string{"($scheme = http)"},
		Block:      &Block{},
	}
	block := &Block{}
	block.Directives = append(block.Directives, &Directive{
		Name:       "return",
		Parameters: []string{"301", "https://$host$request_uri"},
	})
	newDir.Block = block
	s.UpdateDirectiveBySecondKey("if", "($scheme", newDir)
}

func (s *Server) UpdateAllowIPs(ips []string) {
	index := -1
	for i, directive := range s.Directives {
		if directive.GetName() == "location" && directive.GetParameters()[0] == "/" {
			index = i
			break
		}
	}
	ipDirectives := make([]IDirective, 0)
	for _, ip := range ips {
		ipDirectives = append(ipDirectives, &Directive{
			Name:       "allow",
			Parameters: []string{ip},
		})
	}
	ipDirectives = append(ipDirectives, &Directive{
		Name:       "deny",
		Parameters: []string{"all"},
	})
	if index != -1 {
		newDirectives := append(ipDirectives, s.Directives[index:]...)
		s.Directives = append(s.Directives[:index], newDirectives...)
	} else {
		s.Directives = append(s.Directives, ipDirectives...)
	}
}
