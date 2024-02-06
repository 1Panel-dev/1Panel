package parser

import (
	"bufio"
	"bytes"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser/flag"
	"io"
	"strings"
)

type lexer struct {
	reader     *bufio.Reader
	file       string
	line       int
	column     int
	inLuaBlock bool
	Latest     flag.Flag
}

func lex(content string) *lexer {
	return newLexer(bytes.NewBuffer([]byte(content)))
}

func newLexer(r io.Reader) *lexer {
	return &lexer{
		line:   1,
		reader: bufio.NewReader(r),
	}
}

func (s *lexer) scan() flag.Flag {
	s.Latest = s.getNextFlag()
	return s.Latest
}

//func (s *lexer) all() flag.Flags {
//	tokens := make([]flag.Flag, 0)
//	for {
//		v := s.scan()
//		if v.Type == flag.EOF || v.Type == -1 {
//			break
//		}
//		tokens = append(tokens, v)
//	}
//	return tokens
//}

func (s *lexer) getNextFlag() flag.Flag {
	if s.inLuaBlock {
		s.inLuaBlock = false
		flag := s.scanLuaCode()
		return flag
	}
retoFlag:
	ch := s.peek()
	switch {
	case isSpace(ch):
		s.skipWhitespace()
		goto retoFlag
	case isEOF(ch):
		return s.NewToken(flag.EOF).Lit(string(s.read()))
	case ch == ';':
		return s.NewToken(flag.Semicolon).Lit(string(s.read()))
	case ch == '{':
		if isLuaBlock(s.Latest) {
			s.inLuaBlock = true
		}
		return s.NewToken(flag.BlockStart).Lit(string(s.read()))
	case ch == '}':
		return s.NewToken(flag.BlockEnd).Lit(string(s.read()))
	case ch == '#':
		return s.scanComment()
	case ch == '$':
		return s.scanVariable()
	case isQuote(ch):
		return s.scanQuotedString(ch)
	default:
		return s.scanKeyword()
	}
}

func (s *lexer) scanLuaCode() flag.Flag {
	ret := s.NewToken(flag.LuaCode)
	stack := make([]rune, 0, 50)
	code := strings.Builder{}

	for {
		ch := s.read()
		if ch == rune(flag.EOF) {
			panic("unexpected end of file while scanning a string, maybe an unclosed lua code?")
		}
		if ch == '#' {
			code.WriteRune(ch)
			code.WriteString(s.readUntil(isEndOfLine))
			continue
		} else if ch == '}' {
			if len(stack) == 0 {
				_ = s.reader.UnreadRune()
				return ret.Lit(strings.TrimRight(strings.Trim(code.String(), "\n"), "\n  "))
			}
			if stack[len(stack)-1] == '{' {
				stack = stack[0 : len(stack)-1]
			}
		} else if ch == '{' {
			stack = append(stack, ch)
		}
		code.WriteRune(ch)
	}
}

func (s *lexer) peek() rune {
	r, _, _ := s.reader.ReadRune()
	_ = s.reader.UnreadRune()
	return r
}

type runeCheck func(rune) bool

func (s *lexer) readUntil(until runeCheck) string {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.peek(); isEOF(ch) {
			break
		} else if until(ch) {
			break
		} else {
			buf.WriteRune(s.read())
		}
	}

	return buf.String()
}

func (s *lexer) NewToken(tokenType flag.Type) flag.Flag {
	return flag.Flag{
		Type:   tokenType,
		Line:   s.line,
		Column: s.column,
	}
}

func (s *lexer) readWhile(while runeCheck) string {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.peek(); while(ch) {
			buf.WriteRune(s.read())
		} else {
			break
		}
	}
	return buf.String()
}

func (s *lexer) skipWhitespace() {
	s.readWhile(isSpace)
}

func (s *lexer) scanComment() flag.Flag {
	return s.NewToken(flag.Comment).Lit(s.readUntil(isEndOfLine))
}

func (s *lexer) scanQuotedString(delimiter rune) flag.Flag {
	var buf bytes.Buffer
	tok := s.NewToken(flag.QuotedString)
	_, _ = buf.WriteRune(s.read())
	for {
		ch := s.read()

		if ch == rune(flag.EOF) {
			panic("unexpected end of file while scanning a string, maybe an unclosed quote?")
		}

		if ch == '\\' && (s.peek() == delimiter) {
			buf.WriteRune(ch)
			buf.WriteRune(s.read())
			continue
		}

		_, _ = buf.WriteRune(ch)
		if ch == delimiter {
			break
		}
	}

	return tok.Lit(buf.String())
}

func (s *lexer) scanKeyword() flag.Flag {
	var buf bytes.Buffer
	tok := s.NewToken(flag.Keyword)
	prev := s.read()
	buf.WriteRune(prev)
	for {
		ch := s.peek()

		if isSpace(ch) || isEOF(ch) || ch == ';' {
			break
		}

		if ch == '{' {
			if prev == '$' {
				buf.WriteString(s.readUntil(func(r rune) bool {
					return r == '}'
				}))
				buf.WriteRune(s.read()) //consume latest '}'
			} else {
				break
			}
		}
		buf.WriteRune(s.read())
	}

	return tok.Lit(buf.String())
}

func (s *lexer) scanVariable() flag.Flag {
	return s.NewToken(flag.Variable).Lit(s.readUntil(isKeywordTerminator))
}

func (s *lexer) read() rune {
	ch, _, err := s.reader.ReadRune()
	if err != nil {
		return rune(flag.EOF)
	}

	if ch == '\n' {
		s.column = 1
		s.line++
	} else {
		s.column++
	}
	return ch
}

func isQuote(ch rune) bool {
	return ch == '"' || ch == '\'' || ch == '`'
}

func isKeywordTerminator(ch rune) bool {
	return isSpace(ch) || isEndOfLine(ch) || ch == '{' || ch == ';'
}

func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || isEndOfLine(ch)
}

func isEOF(ch rune) bool {
	return ch == rune(flag.EOF)
}

func isEndOfLine(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

func isLuaBlock(t flag.Flag) bool {
	return t.Type == flag.Keyword && strings.HasSuffix(t.Literal, "_by_lua_block")
}
