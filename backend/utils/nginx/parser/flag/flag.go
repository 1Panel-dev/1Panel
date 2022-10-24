package flag

import (
	"fmt"
)

type Type int

const (
	EOF Type = iota
	Eol
	Keyword
	QuotedString
	Variable
	BlockStart
	BlockEnd
	Semicolon
	Comment
	Illegal
	Regex
)

var (
	FlagName = map[Type]string{
		QuotedString: "QuotedString",
		EOF:          "Eof",
		Keyword:      "Keyword",
		Variable:     "Variable",
		BlockStart:   "BlockStart",
		BlockEnd:     "BlockEnd",
		Semicolon:    "Semicolon",
		Comment:      "Comment",
		Illegal:      "Illegal",
		Regex:        "Regex",
	}
)

func (tt Type) String() string {
	return FlagName[tt]
}

type Flag struct {
	Type    Type
	Literal string
	Line    int
	Column  int
}

func (t Flag) String() string {
	return fmt.Sprintf("{Type:%s,Literal:\"%s\",Line:%d,Column:%d}", t.Type, t.Literal, t.Line, t.Column)
}

func (t Flag) Lit(literal string) Flag {
	t.Literal = literal
	return t
}

func (t Flag) EqualTo(t2 Flag) bool {
	return t.Type == t2.Type && t.Literal == t2.Literal
}

type Flags []Flag

func (fs Flags) EqualTo(flags Flags) bool {
	if len(fs) != len(flags) {
		return false
	}
	for i, t := range fs {
		if !t.EqualTo(flags[i]) {
			return false
		}
	}
	return true
}

func (t Flag) Is(typ Type) bool {
	return t.Type == typ
}

func (t Flag) IsParameterEligible() bool {
	return t.Is(Keyword) || t.Is(QuotedString) || t.Is(Variable) || t.Is(Regex)
}
