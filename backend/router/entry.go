package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	HostRouter
	OperationLogRouter
}

var RouterGroupApp = new(RouterGroup)
