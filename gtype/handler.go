package gtype

type Handler interface {
	InitRouting(router Router)
	BeforeRouting(ctx Context)
	AfterRouting(ctx Context)
	ExtendOptApi(router Router, path *Path, preHandle HttpHandle, wsc SocketChannelCollection)
}
