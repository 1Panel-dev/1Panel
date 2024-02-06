package nginx

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
)

var (
	IndentedStyle = &Style{
		SpaceBeforeBlocks: false,
		StartIndent:       0,
		Indent:            4,
	}
)

type Style struct {
	SpaceBeforeBlocks bool
	StartIndent       int
	Indent            int
}

func (s *Style) Iterate() *Style {
	newStyle := &Style{
		SpaceBeforeBlocks: s.SpaceBeforeBlocks,
		StartIndent:       s.StartIndent + s.Indent,
		Indent:            s.Indent,
	}
	return newStyle
}

func DumpDirective(d components.IDirective, style *Style) string {
	var buf bytes.Buffer

	if style.SpaceBeforeBlocks && d.GetBlock() != nil {
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintf("%s%s", strings.Repeat(" ", style.StartIndent), d.GetName()))
	if len(d.GetParameters()) > 0 {
		buf.WriteString(fmt.Sprintf(" %s", strings.Join(d.GetParameters(), " ")))
	}
	if d.GetBlock() == nil {
		if d.GetName() != "" {
			buf.WriteRune(';')
			buf.WriteString(" ")
		}
		if d.GetComment() != "" {
			buf.WriteString(d.GetComment())
		}
	} else {
		buf.WriteString(" {")
		if d.GetComment() != "" {
			buf.WriteString(" ")
			buf.WriteString(d.GetComment())
		}
		buf.WriteString("\n")
		buf.WriteString(DumpBlock(d.GetBlock(), style.Iterate(), d.GetBlock().GetLine()))
		buf.WriteString(fmt.Sprintf("\n%s}", strings.Repeat(" ", style.StartIndent)))
	}
	return buf.String()
}

func DumpBlock(b components.IBlock, style *Style, startLine int) string {
	var buf bytes.Buffer

	if b.GetCodeBlock() != "" {
		luaLines := strings.Split(b.GetCodeBlock(), "\n")
		for i, line := range luaLines {
			if strings.Replace(line, " ", "", -1) == "" {
				continue
			}
			buf.WriteString(line)
			if i != len(luaLines)-1 {
				buf.WriteString("\n")
			}
		}
		return buf.String()
	}

	line := startLine
	if b.GetLine() > startLine {
		for i := 0; i < b.GetLine()-startLine; i++ {
			buf.WriteString("\n")
		}
		line = b.GetLine()
	}

	directives := b.GetDirectives()
	for i, directive := range directives {

		if directive.GetLine() > line {
			for i := 0; i < b.GetLine()-line; i++ {
				buf.WriteString("\n")
			}
			line = b.GetLine()
		}

		buf.WriteString(DumpDirective(directive, style))
		if i != len(directives)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func DumpConfig(c *components.Config, style *Style) string {
	return DumpBlock(c.Block, style, 1)
}

func WriteConfig(c *components.Config, style *Style) error {
	return os.WriteFile(c.FilePath, []byte(DumpConfig(c, style)), 0644)
}
