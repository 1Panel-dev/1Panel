package components

type Comment struct {
	Detail string
}

func (c *Comment) GetName() string {
	return ""
}

func (c *Comment) GetParameters() []string {
	return []string{}
}

func (c *Comment) GetBlock() IBlock {
	return nil
}

func (c *Comment) GetComment() string {
	return c.Detail
}
