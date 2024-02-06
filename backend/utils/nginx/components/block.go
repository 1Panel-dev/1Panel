package components

type Block struct {
	Line        int
	Comment     string
	Directives  []IDirective
	IsLuaBlock  bool
	LiteralCode string
}

func (b *Block) GetDirectives() []IDirective {
	return b.Directives
}

func (b *Block) GetComment() string {
	return b.Comment
}

func (b *Block) GetLine() int {
	return b.Line
}

func (b *Block) GetCodeBlock() string {
	return b.LiteralCode
}

func (b *Block) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	for _, directive := range b.GetDirectives() {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}

	return directives
}

func (b *Block) UpdateDirective(key string, params []string) {
	if key == "" || len(params) == 0 {
		return
	}
	directives := b.GetDirectives()
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
	b.Directives = directives
}

func (b *Block) RemoveDirective(key string, params []string) {
	directives := b.GetDirectives()
	var newDirectives []IDirective
	for _, dir := range directives {
		if dir.GetName() == key {
			if IsRepeatKey(key) && len(params) > 0 {
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
	b.Directives = newDirectives
}
