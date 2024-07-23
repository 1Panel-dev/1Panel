package components

import (
	"fmt"
)

type LuaBlock struct {
	Directives []IDirective
	Name       string
	Comment    string
	LuaCode    string
	Line       int
}

func NewLuaBlock(directive IDirective) (*LuaBlock, error) {
	if block := directive.GetBlock(); block != nil {
		lb := &LuaBlock{
			Directives: []IDirective{},
			Name:       directive.GetName(),
			LuaCode:    block.GetCodeBlock(),
		}

		lb.Directives = append(lb.Directives, block.GetDirectives()...)
		return lb, nil
	}
	return nil, fmt.Errorf("%s must have a block", directive.GetName())
}

func (lb *LuaBlock) GetName() string {
	return lb.Name
}

func (lb *LuaBlock) GetParameters() []string {
	return []string{}
}

func (lb *LuaBlock) GetDirectives() []IDirective {
	directives := make([]IDirective, 0)
	directives = append(directives, lb.Directives...)
	return directives
}

func (lb *LuaBlock) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	for _, directive := range lb.GetDirectives() {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}

	return directives
}

func (lb *LuaBlock) GetCodeBlock() string {
	return lb.LuaCode
}

func (lb *LuaBlock) GetBlock() IBlock {
	return lb
}

func (lb *LuaBlock) GetComment() string {
	return lb.Comment
}

func (lb *LuaBlock) RemoveDirective(key string, params []string) {
	directives := lb.Directives
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
	lb.Directives = newDirectives
}

func (lb *LuaBlock) UpdateDirective(key string, params []string) {
	if key == "" || len(params) == 0 {
		return
	}
	directives := lb.Directives
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
	lb.Directives = directives
}

func (lb *LuaBlock) GetLine() int {
	return lb.Line
}
