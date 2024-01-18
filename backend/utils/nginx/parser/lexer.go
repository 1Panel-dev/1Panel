package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser/flag"
	"io"
)

type lexer struct {
	reader *bufio.Reader
	file   string
	line   int
	column int
	Latest flag.Flag
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
	buf.WriteRune(s.read()) //consume delimiter
	for {
		ch := s.read()

		if ch == rune(flag.EOF) {
			panic("unexpected end of file while scanning a string, maybe an unclosed quote?")
		}

		if ch == '\\' {
			if needsEscape(s.peek(), delimiter) {
				nextch := s.read()
				switch nextch {
				case 'n':
					fmt.Println("n")
					buf.WriteRune('\n')
				case 'r':
					fmt.Println("r")
					buf.WriteRune('\r')
				case 't':
					fmt.Println("t")
					buf.WriteRune('\t')
				case '\\':
					buf.WriteRune('\\')
				default:
					buf.WriteRune('\\')
					buf.WriteRune(nextch)
				}
				continue
			}
		}
		buf.WriteRune(ch)
		if ch == delimiter {
			break
		}
	}

	return tok.Lit(buf.String())
}

func (s *lexer) scanKeyword() flag.Flag {
	return s.NewToken(flag.Keyword).Lit(s.readUntil(isKeywordTerminator))
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

func needsEscape(ch, delimiter rune) bool {
	return ch == delimiter || ch == 'n' || ch == 't' || ch == '\\' || ch == 'r'
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
