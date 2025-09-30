package components

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Location struct {
	Modifier         string
	Match            string
	Cache            bool
	ProxyPass        string
	Host             string
	CacheTime        int
	CacheUint        string
	Comment          string
	Directives       []IDirective
	Line             int
	Parameters       []string
	Replaces         map[string]string
	ServerCacheTime  int
	ServerCacheUint  string
	Cors             bool
	AllowMethods     string
	AllowHeaders     string
	AllowOrigins     string
	AllowCredentials bool
	Preflight        bool
}

func (l *Location) GetCodeBlock() string {
	return ""
}

func NewLocation(directive IDirective) *Location {
	location := &Location{
		Modifier: "",
		Match:    "",
	}
	directives := make([]IDirective, 0)
	if len(directive.GetParameters()) == 0 {
		panic("no enough parameter for location")
	}
	for _, dir := range directive.GetBlock().GetDirectives() {
		directives = append(directives, dir)
		params := dir.GetParameters()
		switch dir.GetName() {
		case "proxy_pass":
			location.ProxyPass = params[0]
		case "proxy_set_header":
			if params[0] == "Host" {
				location.Host = params[1]
			}
		case "proxy_cache":
			location.Cache = true
		case "if":
			if params[0] == "(" && params[1] == "$uri" && params[2] == "~*" {
				dirs := dir.GetBlock().GetDirectives()
				for _, di := range dirs {
					if di.GetName() == "expires" {
						re := regexp.MustCompile(`^(\d+)(\w+)$`)
						matches := re.FindStringSubmatch(di.GetParameters()[0])
						if matches == nil {
							continue
						}
						cacheTime, err := strconv.Atoi(matches[1])
						if err != nil {
							continue
						}
						unit := matches[2]
						location.CacheUint = unit
						location.CacheTime = cacheTime
					}
				}
			}
			if params[0] == "(" && params[1] == "$request_method" && params[2] == `=` && params[3] == `'OPTIONS'` && params[4] == ")" {
				location.Preflight = true
			}
		case "proxy_cache_valid":
			timeParam := params[len(params)-1]
			re := regexp.MustCompile(`^(\d+)(\w+)$`)
			matches := re.FindStringSubmatch(timeParam)
			if matches == nil {
				continue
			}

			cacheTime, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			unit := matches[2]

			location.ServerCacheTime = cacheTime
			location.ServerCacheUint = unit
		case "sub_filter":
			if location.Replaces == nil {
				location.Replaces = make(map[string]string, 0)
			}
			location.Replaces[strings.Trim(params[0], "\"")] = strings.Trim(params[1], "\"")
		case "add_header":
			if params[0] == "Access-Control-Allow-Origin" {
				location.Cors = true
				location.AllowOrigins = params[1]
			}
			if params[0] == "Access-Control-Allow-Methods" {
				location.AllowMethods = params[1]
			}
			if params[0] == "Access-Control-Allow-Headers" {
				location.AllowHeaders = params[1]
			}
			if params[0] == "Access-Control-Allow-Credentials" && params[1] == "true" {
				location.AllowCredentials = true
			}
		}
	}

	params := directive.GetParameters()
	if len(params) == 1 {
		location.Match = params[0]
	} else if len(params) == 2 {
		location.Match = params[1]
		location.Modifier = params[0]
	}
	location.Parameters = directive.GetParameters()
	location.Line = directive.GetLine()
	location.Comment = directive.GetComment()
	location.Directives = directives
	return location
}

func (l *Location) GetName() string {
	return "location"
}

func (l *Location) GetParameters() []string {
	return l.Parameters
}

func (l *Location) GetBlock() IBlock {
	return l
}

func (l *Location) GetComment() string {
	return l.Comment
}

func (l *Location) GetLine() int {
	return l.Line
}

func (l *Location) GetDirectives() []IDirective {
	return l.Directives
}

func (l *Location) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	for _, directive := range l.Directives {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}
	return directives
}

func (l *Location) UpdateDirective(key string, params []string) {
	if key == "" || len(params) == 0 {
		return
	}
	directives := l.Directives
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
	l.Directives = directives
}

func (l *Location) RemoveDirective(key string, params []string) {
	directives := l.Directives
	var newDirectives []IDirective
	for _, dir := range directives {
		if dir.GetName() == key {
			if len(params) > 0 {
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
	l.Directives = newDirectives
}

func (l *Location) ChangePath(Modifier string, Match string) {
	if Match != "" && Modifier != "" {
		l.Parameters = []string{Modifier, Match}
	}
	if Match != "" && Modifier == "" {
		l.Parameters = []string{Match}
	}
	l.Modifier = Modifier
	l.Match = Match
}

func (l *Location) AddCache(cacheTime int, cacheUint, cacheKey string, serverCacheTime int, serverCacheUint string) {
	l.RemoveDirective("add_header", []string{"Cache-Control", "no-cache"})
	l.RemoveDirective("if", []string{"(", "$uri", "~*", `"\.(gif|png|jpg|css|js|woff|woff2)$"`, ")"})
	l.RemoveDirective("if", []string{"(", "$uri", "~*", `"\.(gif|png|jpg|css|js|woff|woff2|jpeg|svg|webp|avif)$"`, ")"})
	directives := l.GetDirectives()
	newDir := &Directive{
		Name:       "if",
		Parameters: []string{"(", "$uri", "~*", `"\.(gif|png|jpg|css|js|woff|woff2|jpeg|svg|webp|avif)$"`, ")"},
		Block:      &Block{},
	}
	block := &Block{}
	block.Directives = append(block.Directives, &Directive{
		Name:       "expires",
		Parameters: []string{strconv.Itoa(cacheTime) + cacheUint},
	})
	newDir.Block = block
	directives = append(directives, newDir)
	l.Directives = directives
	l.UpdateDirective("proxy_ignore_headers", []string{"Set-Cookie", "Cache-Control", "expires"})
	l.UpdateDirective("proxy_cache", []string{cacheKey})
	l.UpdateDirective("proxy_cache_key", []string{"$host$uri$is_args$args"})
	l.UpdateDirective("proxy_cache_valid", []string{"200", "304", "301", "302", strconv.Itoa(serverCacheTime) + serverCacheUint})
	l.Cache = true
	l.CacheTime = cacheTime
	l.CacheUint = cacheUint
}

func (l *Location) RemoveCache(cacheKey string) {
	l.RemoveDirective("if", []string{"(", "$uri", "~*", `"\.(gif|png|jpg|css|js|woff|woff2)$"`, ")"})
	l.RemoveDirective("if", []string{"(", "$uri", "~*", `"\.(gif|png|jpg|css|js|woff|woff2|jpeg|svg|webp|avif)$"`, ")"})
	l.RemoveDirective("proxy_ignore_headers", []string{"Set-Cookie"})
	l.RemoveDirective("proxy_cache", []string{cacheKey})
	l.RemoveDirective("proxy_cache_key", []string{"$host$uri$is_args$args"})
	l.RemoveDirective("proxy_cache_valid", []string{"200"})

	l.UpdateDirective("add_header", []string{"Cache-Control", "no-cache"})

	l.CacheTime = 0
	l.CacheUint = ""
	l.Cache = false
}

func (l *Location) AddSubFilter(subFilters map[string]string) {
	l.RemoveDirective("sub_filter", []string{})
	l.Replaces = subFilters
	for k, v := range subFilters {
		l.UpdateDirective("sub_filter", []string{fmt.Sprintf(`"%s"`, k), fmt.Sprintf(`"%s"`, v)})
	}
	l.UpdateDirective("proxy_set_header", []string{"Accept-Encoding", `""`})
	l.UpdateDirective("sub_filter_once", []string{"off"})
	l.UpdateDirective("sub_filter_types", []string{"*"})
}

func (l *Location) RemoveSubFilter() {
	l.RemoveDirective("sub_filter", []string{})
	l.RemoveDirective("proxy_set_header", []string{"Accept-Encoding", `""`})
	l.RemoveDirective("sub_filter_once", []string{"off"})
	l.RemoveDirective("sub_filter_types", []string{"*"})
	l.Replaces = nil
}

func (l *Location) AddCorsOption() {
	newDir := &Directive{
		Name:       "if",
		Parameters: []string{"(", "$request_method", "=", "'OPTIONS'", ")"},
		Block:      &Block{},
	}
	block := &Block{}
	block.AppendDirectives(&Directive{
		Name:       "add_header",
		Parameters: []string{"Access-Control-Max-Age", "1728000"},
	})
	block.AppendDirectives(&Directive{
		Name:       "add_header",
		Parameters: []string{"Content-Type", "'text/plain;charset=UTF-8'"},
	})
	block.AppendDirectives(&Directive{
		Name:       "add_header",
		Parameters: []string{"Content-Length", "0"},
	})
	block.AppendDirectives(&Directive{
		Name:       "return",
		Parameters: []string{"204"},
	})
	newDir.Block = block
	directives := l.GetDirectives()
	directives = append(directives, newDir)
	l.Directives = directives
}
