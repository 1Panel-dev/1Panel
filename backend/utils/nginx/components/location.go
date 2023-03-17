package components

type Location struct {
	*Directive
	Modifier string
	Match    string
}

func NewLocation(directive *Directive) *Location {
	location := &Location{
		Modifier:  "",
		Match:     "",
		Directive: directive,
	}
	if directive.GetBlock() != nil {
		directive.Comment = directive.GetBlock().GetComment()
	}

	if len(directive.Parameters) == 0 {
		panic("no enough parameter for location")
	}

	if len(directive.Parameters) == 1 {
		location.Match = directive.Parameters[0]
		return location
	} else if len(directive.Parameters) == 2 {
		location.Modifier = directive.Parameters[0]
		location.Match = directive.Parameters[1]
		return location
	}
	return nil
}
