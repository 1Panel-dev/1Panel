package result

var (
	OK = NewError(0, "Ok")

	ErrParam = NewError(400, "参数不合法")

	ErrUserNot = NewError(20001, "用户不存在")
)
