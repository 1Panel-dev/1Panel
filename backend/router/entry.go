package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	OperationLogRouter
	FileRouter
}

var RouterGroupApp = new(RouterGroup)
