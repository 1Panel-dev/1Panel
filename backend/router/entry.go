package router

type RouterGroup struct {
	BaseRouter
	UserRouter
}

var RouterGroupApp = new(RouterGroup)
