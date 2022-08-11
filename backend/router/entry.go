package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	OperationLogRouter
}

var RouterGroupApp = new(RouterGroup)
