package result

var (
	OK = NewError(0, "Ok")

	ErrParam = NewError(400, "InvalidParam")
)
