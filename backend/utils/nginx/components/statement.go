package components

type IBlock interface {
	GetDirectives() []IDirective
	FindDirectives(directiveName string) []IDirective
	UpdateDirectives(directiveName string, directive Directive)
	AddDirectives(directive Directive)
	RemoveDirectives(names []string)
	GetComment() string
}

type IDirective interface {
	GetName() string
	GetParameters() []string
	GetBlock() IBlock
	GetComment() string
}

type FileDirective interface {
	isFileDirective()
}

type IncludeDirective interface {
	FileDirective
}
