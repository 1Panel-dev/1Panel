package components

type IBlock interface {
	GetDirectives() []IDirective
	FindDirectives(directiveName string) []IDirective
	RemoveDirective(name string, params []string)
	UpdateDirective(name string, params []string)
	GetComment() string
	GetLine() int
	GetCodeBlock() string
}

type IDirective interface {
	GetName() string
	GetParameters() []string
	GetBlock() IBlock
	GetComment() string
	GetLine() int
}
