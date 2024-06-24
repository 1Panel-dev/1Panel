package parser

import (
	"bufio"
	"errors"
	"fmt"
	components "github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser/flag"
	"os"
	"strings"
)

type Parser struct {
	lexer             *lexer
	currentToken      flag.Flag
	followingToken    flag.Flag
	blockWrappers     map[string]func(*components.Directive) (components.IDirective, error)
	directiveWrappers map[string]func(*components.Directive) components.IDirective
}

func NewStringParser(str string) *Parser {
	return NewParserFromLexer(lex(str))
}

func NewParser(filePath string) (*Parser, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	l := newLexer(bufio.NewReader(f))
	l.file = filePath
	p := NewParserFromLexer(l)
	return p, nil
}

func NewParserFromLexer(lexer *lexer) *Parser {
	parser := &Parser{
		lexer: lexer,
	}

	parser.nextToken()
	parser.nextToken()

	parser.blockWrappers = map[string]func(*components.Directive) (components.IDirective, error){
		"http": func(directive *components.Directive) (components.IDirective, error) {
			return parser.wrapHttp(directive), nil
		},
		"server": func(directive *components.Directive) (components.IDirective, error) {
			return parser.wrapServer(directive), nil
		},
		"location": func(directive *components.Directive) (components.IDirective, error) {
			return parser.wrapLocation(directive), nil
		},
		"upstream": func(directive *components.Directive) (components.IDirective, error) {
			return parser.wrapUpstream(directive), nil
		},
		"_by_lua_block": func(directive *components.Directive) (components.IDirective, error) {
			return parser.wrapLuaBlock(directive)
		},
	}

	parser.directiveWrappers = map[string]func(*components.Directive) components.IDirective{
		"server": func(directive *components.Directive) components.IDirective {
			return parser.parseUpstreamServer(directive)
		},
	}

	return parser
}

func (p *Parser) nextToken() {
	p.currentToken = p.followingToken
	p.followingToken = p.lexer.scan()
}

func (p *Parser) curTokenIs(t flag.Type) bool {
	return p.currentToken.Type == t
}

func (p *Parser) followingTokenIs(t flag.Type) bool {
	return p.followingToken.Type == t
}

func (p *Parser) Parse() (*components.Config, error) {
	parsedBlock, err := p.parseBlock(false)
	if err != nil {
		return nil, err
	}
	c := &components.Config{
		FilePath: p.lexer.file,
		Block:    parsedBlock,
	}
	return c, err
}

func (p *Parser) parseBlock(inBlock bool) (*components.Block, error) {
	context := &components.Block{
		Comment:    "",
		Directives: make([]components.IDirective, 0),
		Line:       p.currentToken.Line,
	}

parsingloop:
	for {
		switch {
		case p.curTokenIs(flag.EOF):
			if inBlock {
				return nil, errors.New("unexpected eof in block")
			}
			break parsingloop
		case p.curTokenIs(flag.BlockEnd):
			break parsingloop
		case p.curTokenIs(flag.LuaCode):
			context.IsLuaBlock = true
			context.LiteralCode = p.currentToken.Literal
		case p.curTokenIs(flag.Keyword) || p.curTokenIs(flag.QuotedString):
			s, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			context.Directives = append(context.Directives, s)
		case p.curTokenIs(flag.Comment):
			context.Directives = append(context.Directives, &components.Comment{
				Detail: p.currentToken.Literal,
				Line:   p.currentToken.Line,
			})
		}
		p.nextToken()
	}

	return context, nil
}

func (p *Parser) parseStatement() (components.IDirective, error) {
	d := &components.Directive{
		Name: p.currentToken.Literal,
		Line: p.currentToken.Line,
	}

	for p.nextToken(); p.currentToken.IsParameterEligible(); p.nextToken() {
		d.Parameters = append(d.Parameters, p.currentToken.Literal)
	}

	if p.curTokenIs(flag.Semicolon) {
		if dw, ok := p.directiveWrappers[d.Name]; ok {
			return dw(d), nil
		}
		if p.followingTokenIs(flag.Comment) && p.currentToken.Line == p.followingToken.Line {
			d.Comment = p.followingToken.Literal
			p.nextToken()
		}
		return d, nil
	}

	if p.curTokenIs(flag.BlockStart) {
		inLineComment := ""
		if p.followingTokenIs(flag.Comment) && p.currentToken.Line == p.followingToken.Line {
			inLineComment = p.followingToken.Literal
			p.nextToken()
			p.nextToken()
		}
		block, err := p.parseBlock(false)
		if err != nil {
			return nil, err
		}

		block.Comment = inLineComment
		d.Block = block

		if strings.HasSuffix(d.Name, "_by_lua_block") {
			return p.blockWrappers["_by_lua_block"](d)
		}

		if bw, ok := p.blockWrappers[d.Name]; ok {
			return bw(d)
		}
		return d, nil
	}

	panic(fmt.Errorf("unexpected token %s (%s) on line %d, column %d", p.currentToken.Type.String(), p.currentToken.Literal, p.currentToken.Line, p.currentToken.Column))
}

func (p *Parser) wrapLocation(directive *components.Directive) *components.Location {
	return components.NewLocation(directive)
}

func (p *Parser) wrapServer(directive *components.Directive) *components.Server {
	s, _ := components.NewServer(directive)
	return s
}

func (p *Parser) wrapUpstream(directive *components.Directive) *components.Upstream {
	s, _ := components.NewUpstream(directive)
	return s
}

func (p *Parser) wrapHttp(directive *components.Directive) *components.Http {
	h, _ := components.NewHttp(directive)
	return h
}

func (p *Parser) wrapLuaBlock(directive *components.Directive) (*components.LuaBlock, error) {
	return components.NewLuaBlock(directive)
}

func (p *Parser) parseUpstreamServer(directive *components.Directive) *components.UpstreamServer {
	return components.NewUpstreamServer(directive)
}
