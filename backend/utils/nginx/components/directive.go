package components

type Directive struct {
	Block      IBlock
	Name       string
	Comment    string
	Parameters []string
}

func (d *Directive) GetComment() string {
	return d.Comment
}

func (d *Directive) GetName() string {
	return d.Name
}

func (d *Directive) GetParameters() []string {
	return d.Parameters
}

func (d *Directive) GetBlock() IBlock {
	return d.Block
}
