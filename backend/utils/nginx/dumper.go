package nginx

import (
	"bytes"
	"fmt"
	components "github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	//NoIndentStyle default style
	NoIndentStyle = &Style{
		SortDirectives: false,
		StartIndent:    0,
		Indent:         0,
	}

	//IndentedStyle default style
	IndentedStyle = &Style{
		SortDirectives: false,
		StartIndent:    0,
		Indent:         4,
	}

	//NoIndentSortedStyle default style
	NoIndentSortedStyle = &Style{
		SortDirectives: true,
		StartIndent:    0,
		Indent:         0,
	}

	//NoIndentSortedSpaceStyle default style
	NoIndentSortedSpaceStyle = &Style{
		SortDirectives:    true,
		SpaceBeforeBlocks: true,
		StartIndent:       0,
		Indent:            0,
	}
)

type Style struct {
	SortDirectives    bool
	SpaceBeforeBlocks bool
	StartIndent       int
	Indent            int
}

func NewStyle() *Style {
	style := &Style{
		SortDirectives: false,
		StartIndent:    0,
		Indent:         4,
	}
	return style
}

func (s *Style) Iterate() *Style {
	newStyle := &Style{
		SortDirectives:    s.SortDirectives,
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
		buf.WriteString(DumpBlock(d.GetBlock(), style.Iterate()))
		buf.WriteString(fmt.Sprintf("\n%s}", strings.Repeat(" ", style.StartIndent)))
	}
	return buf.String()
}

func DumpBlock(b components.IBlock, style *Style) string {
	var buf bytes.Buffer

	directives := b.GetDirectives()
	if style.SortDirectives {
		sort.SliceStable(directives, func(i, j int) bool {
			return directives[i].GetName() < directives[j].GetName()
		})
	}

	for i, directive := range directives {
		buf.WriteString(DumpDirective(directive, style))
		if i != len(directives)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func DumpConfig(c *components.Config, style *Style) string {
	return DumpBlock(c.Block, style)
}

func WriteConfig(c *components.Config, style *Style) error {
	dir, _ := filepath.Split(c.FilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.FilePath, []byte(DumpConfig(c, style)), 0644)
}
